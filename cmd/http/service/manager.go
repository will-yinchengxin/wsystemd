package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
	"wsystemd/cmd/cluster"
	"wsystemd/cmd/http/consts"
	"wsystemd/cmd/http/core"
	"wsystemd/cmd/http/dto/dao"
	"wsystemd/cmd/http/dto/entity"
	"wsystemd/cmd/http/params"
	"wsystemd/cmd/log"
	"wsystemd/cmd/process"
	"wsystemd/cmd/utils"

	"github.com/go-kit/kit/log/level"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	SplitTag = ">>>>>"
)

func CreateClusterModeJob(req params.JobCfg) (interface{}, *utils.CodeType) {
	if val, ok := core.CoreConfig["singlemode"]; ok && val.(bool) {
		return createJobLocal(req)
	}

	localNode, err := process.GetHostName()
	if err != nil {
		level.Error(log.Logger).Log("GetHostName Err", err.Error())
		return nil, utils.ServerErr
	}

	// 使用 etcd 获取节点负载
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{cluster.GetEtcdAddr()},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, utils.ServerErr
	}
	defer cli.Close()

	nodeStats, err := cluster.GetWorkerTaskCounts(cli)
	if err != nil {
		// 如果获取失败，回退到数据库查询
		var taskDao = &dao.Task{}
		nodeStats, err = taskDao.WithContext(context.Background()).GetNodeTaskCount()
		if err != nil {
			level.Error(log.Logger).Log("GetNodeTaskCount Err", err.Error())
			return nil, utils.DBErr
		}
	}

	targetNode := cluster.FindLeastLoadedNode(nodeStats)

	if targetNode == localNode {
		return createJobLocal(req)
	}

	worker, err := cluster.GetWorkerInfo(targetNode)
	if err != nil {
		level.Error(log.Logger).Log("GetWorkerInfo Err", err.Error())
		return nil, utils.ServerErr
	}

	response, err := cluster.ForwardToWorker(worker, "/v1/jobs/submit", req)
	if err != nil {
		level.Error(log.Logger).Log("ForwardRequest Err", err.Error())
		return nil, utils.ServerErr
	}

	return response, &utils.CodeType{}
}

func createJobLocal(req params.JobCfg) (interface{}, *utils.CodeType) {
	var (
		taskModel entity.Task
		taskDao   = &dao.Task{}
		res       = make(map[string]interface{})
		pid       int
		uuid      = utils.GetID(32)
		err       error
	)

	if req.DoOnce {
		if req.BigOne != "" {
			return bigOne(req)
		}
		return doOnceJob(req)
	}

	pid, err = process.PManager.StartProc(req.Run.Cmd, req.Run.Args, req.Run.Outfile, req.Run.Errfile, uuid)
	if err != nil {
		level.Error(log.Logger).Log("CreateSingleModeJob Err", err.Error())
		return nil, utils.StartJobFail
	}

	taskModel = buildTaskModel(req, pid, uuid)

	if err := taskDao.WithContext(context.Background()).Create(&taskModel); err != nil {
		level.Error(log.Logger).Log("CreateSingleModeJob Err", err.Error())
		return res, utils.DBErr
	}

	// 更新内存中的任务计数
	if !req.DoOnce && cluster.WkMg != nil {
		cluster.WkMg.IncrTaskCount()
	}

	res["pid"] = pid
	res["id"] = uuid
	res["ctime"] = utils.GetCTime()
	res["run"] = req.Run

	return res, &utils.CodeType{}
}

func doOnceJob(req params.JobCfg) (interface{}, *utils.CodeType) {
	res := make(map[string]interface{})
	uuid := utils.GetID(32)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, req.Run.Cmd, req.Run.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		level.Error(log.Logger).Log("Do OnceJob Start Error", err.Error(), "arg", req.Run)
		return nil, utils.StartJobFail
	}

	err = cmd.Wait()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			level.Error(log.Logger).Log("Do OnceJob Out Time", context.DeadlineExceeded, "arg", req.Run)
			return nil, utils.JobExecOutTime
		}
		level.Error(log.Logger).Log("Do OnceJob Err", err.Error(), "arg", req.Run)
		return nil, utils.StartJobFail
	}

	res["id"] = uuid
	res["ctime"] = utils.GetCTime()
	res["run"] = req.Run
	return res, &utils.CodeType{}
}

func buildTaskModel(req params.JobCfg, pid int, jobId string) entity.Task {
	now := time.Now()
	taskModel := entity.Task{
		Args:          strings.Join(req.Run.Args, SplitTag),
		Cmd:           req.Run.Cmd,
		Errfile:       req.Run.Errfile,
		Outfile:       req.Run.Outfile,
		Pid:           pid,
		JobId:         jobId,
		Dc:            req.Dc,
		Ip:            req.Ip,
		LoadMethod:    req.LoadMethod,
		CreateTime:    now,
		UpdateTime:    now,
		HeartBeatTime: now,
	}

	if req.DoOnce {
		taskModel.DoOnce = consts.DoOnce
	} else {
		taskModel.DoOnce = consts.NotDoOnce
	}

	if req.Node == "" {
		hName, _ := process.GetHostName()
		taskModel.Node = hName
	} else {
		taskModel.Node = req.Node
	}

	return taskModel
}

func StopBigOne(req params.BigOne) *utils.CodeType {
	if val, ok := core.CoreConfig["singlemode"]; ok && val.(bool) {
		return stopBigOneLocal(req)
	}

	var taskDao = &dao.Task{}
	info, err := taskDao.WithContext(context.Background()).FindByJobId(req.BigOneJobId)
	if err != nil {
		return utils.DBErr
	}
	if info.ID <= 0 {
		return &utils.CodeType{}
	}

	localNode, err := process.GetHostName()
	if err != nil {
		level.Error(log.Logger).Log("GetHostName Err", err.Error())
		return utils.ServerErr
	}

	if info.Node != localNode {
		worker, err := cluster.GetWorkerInfo(info.Node)
		if err != nil {
			level.Error(log.Logger).Log("GetWorkerInfo Err", err.Error())
			return utils.ServerErr
		}

		_, err = cluster.ForwardToWorker(worker, "/v1/jobs/stopBigOne", req)
		if err != nil {
			level.Error(log.Logger).Log("ForwardRequest Err", err.Error())
			return utils.ServerErr
		}
		return &utils.CodeType{}
	}

	return stopBigOneLocal(req)
}

func stopBigOneLocal(req params.BigOne) *utils.CodeType {
	var taskDao = &dao.Task{}
	info, err := taskDao.WithContext(context.Background()).FindByJobId(req.BigOneJobId)
	if err != nil {
		return utils.DBErr
	}
	if info.ID <= 0 {
		return &utils.CodeType{}
	}

	status, err := process.PManager.StopProc(info.JobId, info.Pid, false)
	if err != nil || status != 0 {
		if err != nil {
			level.Error(log.Logger).Log("StopSingleModeJob Err", err.Error(), "pid", info.Pid)
		}
		return utils.StopJobFail
	}
	return &utils.CodeType{}
}

func SingleJobReporter(req params.JobReporter) *utils.CodeType {
	if val, ok := core.CoreConfig["singlemode"]; ok && val.(bool) {
		return reportJobLocal(req)
	}

	segs := strings.Split(req.Token, ":")
	if len(segs) < 2 {
		level.Error(log.Logger).Log("Invalid task token", req.Token)
		return utils.ReqParamErr
	}
	nodeName := segs[0]

	localNode, err := process.GetHostName()
	if err != nil {
		level.Error(log.Logger).Log("GetHostName Err", err.Error())
		return utils.ServerErr
	}

	if nodeName != localNode {
		worker, err := cluster.GetWorkerInfo(nodeName)
		if err != nil {
			level.Error(log.Logger).Log("GetWorkerInfo Err", err.Error())
			return utils.ServerErr
		}

		_, err = cluster.ForwardToWorker(worker, "/v1/agent/tasks/report", req)
		if err != nil {
			level.Error(log.Logger).Log("ForwardRequest Err", err.Error())
			return utils.ServerErr
		}
		return &utils.CodeType{}
	}

	return reportJobLocal(req)
}

func reportJobLocal(req params.JobReporter) *utils.CodeType {
	segs := strings.Split(req.Token, ":")
	if len(segs) < 2 {
		level.Error(log.Logger).Log("Invalid task token", req.Token)
		return utils.ReqParamErr
	}
	nodeName := segs[0]

	contail, err := process.HNContail(nodeName)
	if err != nil {
		level.Error(log.Logger).Log("Get HostName Err", err.Error())
		return utils.ServerErr
	}
	if !contail {
		return utils.ServerNotExist
	}

	var taskDao = &dao.Task{}
	tInfo, err := taskDao.WithContext(context.Background()).FindByNodeAndPid(nodeName, req.Pid)
	if err != nil {
		level.Error(log.Logger).Log("DB FindByNodeAndPid Err", err.Error())
		return utils.DBErr
	}
	if tInfo.ID <= 0 {
		level.Error(log.Logger).Log("FindByNodeAndPid DBRecorderNotExist")
		return utils.DBRecorderNotExist
	}

	level.Info(log.Logger).Log("msg", "client heart beat report")
	err = taskDao.WithContext(context.Background()).UpdateHeartBeatTime(tInfo.ID)
	if err != nil {
		level.Error(log.Logger).Log("UpdateInfoIfExist Err", err.Error())
		return utils.DBErr
	}
	return &utils.CodeType{}
}

func bigOne(req params.JobCfg) (interface{}, *utils.CodeType) {
	var (
		taskModel entity.Task
		taskDao   = &dao.Task{}
		res       = make(map[string]interface{})
		pid       int
		uuid      = utils.GetID(32)
		err       error
	)

	pid, err = process.PManager.StartProc(req.Run.Cmd, req.Run.Args, req.Run.Outfile, req.Run.Errfile, uuid)
	if err != nil {
		level.Error(log.Logger).Log("CreateSingleModeJob Err", err.Error())
		return nil, utils.StartJobFail
	}

	taskModel = buildTaskModel(req, pid, uuid)
	taskModel.BigOne = "bigOne"

	err = taskDao.WithContext(context.Background()).Create(&taskModel)
	if err != nil {
		level.Error(log.Logger).Log("CreateSingleModeJob Err", err.Error())
		return res, utils.DBErr
	}

	res["pid"] = pid
	res["id"] = uuid
	res["ctime"] = utils.GetCTime()
	res["run"] = req.Run

	return res, &utils.CodeType{}
}

func StopSingleModeJob(jobId string, delete bool) *utils.CodeType {
	if val, ok := core.CoreConfig["singlemode"]; ok && val.(bool) {
		return stopJobLocal(jobId, delete)
	}

	var taskDao = &dao.Task{}
	taskInfo, err := taskDao.WithContext(context.Background()).FindByJobId(jobId)
	if err != nil {
		return utils.DBErr
	}

	localNode, err := process.GetHostName()
	if err != nil {
		level.Error(log.Logger).Log("GetHostName Err", err.Error())
		return utils.ServerErr
	}

	if taskInfo.Node != localNode {
		worker, err := cluster.GetWorkerInfo(taskInfo.Node)
		if err != nil {
			level.Error(log.Logger).Log("GetWorkerInfo Err", err.Error())
			return utils.ServerErr
		}

		_, err = cluster.ForwardToWorker(worker, fmt.Sprintf("/v1/jobs/%s/stop", jobId), nil)
		if err != nil {
			level.Error(log.Logger).Log("ForwardRequest Err", err.Error())
			return utils.ServerErr
		}
		return &utils.CodeType{}
	}

	return stopJobLocal(jobId, delete)
}

func stopJobLocal(jobId string, delete bool) *utils.CodeType {
	var taskDao = &dao.Task{}
	pid, exist := process.PManager.JobExist(jobId)
	if !exist {
		return utils.StopNotExist
	}

	status, err := process.PManager.StopProc(jobId, pid, delete)
	if err != nil || status != 0 {
		if err != nil {
			level.Error(log.Logger).Log("StopSingleModeJob Err", err.Error(), "pid", pid)
		}
		return utils.StopJobFail
	}

	if delete {
		err = taskDao.WithContext(context.Background()).DeleteByJobId(jobId)
		if err != nil {
			level.Error(log.Logger).Log("DeleteByJobId Err", err.Error())
			return utils.DBErr
		}
	}

	if delete && cluster.WkMg != nil {
		cluster.WkMg.DecrTaskCount()
	}

	return &utils.CodeType{}
}

func CheckClientAlive() error {
	var (
		batchSize = 1000
		taskDao   = &dao.Task{}
		now       = time.Now()
	)

	// 获取本机节点名
	hostName, err := process.GetHostName()
	if err != nil {
		return err
	}

	// 如果是集群模式，只处理本节点的任务
	if val, ok := core.CoreConfig["singlemode"]; ok && !val.(bool) {
		level.Info(log.Logger).Log("msg", "Running in cluster mode, checking local tasks only", "node", hostName)
	}

	maxId, _ := taskDao.WithContext(context.Background()).GetMaxCount(hostName)
	minId, _ := taskDao.WithContext(context.Background()).GetMinCount(hostName)
	if maxId == 0 {
		return nil
	}

	proc := process.PManager
	for {
		list, err := taskDao.GetList(minId, int64(batchSize), hostName)
		if err != nil {
			return err
		}
		if len(list) == 0 {
			return nil
		}
		minId = list[len(list)-1].ID

		level.Debug(log.Logger).Log("msg", "Checking client alive status",
			"minId", minId, "maxId", maxId, "node", hostName)

		for _, task := range list {
			if now.Sub(task.HeartBeatTime).Minutes() > 2 {
				if proc.IsAlive(task.Pid) {
					// 任务存在，更新心跳时间
					if err = taskDao.WithContext(context.Background()).UpdateHeartBeatTime(task.ID); err != nil {
						level.Error(log.Logger).Log("msg", "Failed to update heartbeat time",
							"id", task.ID, "pid", task.Pid, "error", err)
						continue
					}
				} else {
					level.Info(log.Logger).Log("msg", "Restarting dead task",
						"jobId", task.JobId, "node", hostName)

					// 停止旧任务
					StopSingleModeJob(task.JobId, false)

					// 重启任务
					procPid, err := proc.StartProc(task.Cmd,
						strings.Split(task.Args, SplitTag),
						task.Outfile, task.Errfile, task.JobId)
					if err != nil {
						level.Error(log.Logger).Log("msg", "Failed to restart task",
							"cmd", task.Cmd, "args", task.Args, "error", err)
						continue
					}

					// 更新PID
					if err = taskDao.WithContext(context.Background()).UpdatePid(task.ID, procPid); err != nil {
						level.Error(log.Logger).Log("msg", "Failed to update PID",
							"id", task.ID, "pid", procPid, "error", err)
						continue
					}
				}
			}
		}

		if minId >= maxId {
			return nil
		}
	}
}
