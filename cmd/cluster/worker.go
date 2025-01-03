package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"wsystemd/cmd/http/consts"
	"wsystemd/cmd/http/dto/dao"
	"wsystemd/cmd/log"
	"wsystemd/cmd/utils"

	"github.com/go-kit/kit/log/level"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	WkMg *WorkerManager
)

type Worker struct {
	ID        string
	Hostname  string
	IP        string
	Port      string
	Status    string
	LastBeat  time.Time
	Resources ResourceInfo
}

type ResourceInfo struct {
	CPUUsage    float64
	MemoryUsage float64
	LoadUsage   float64
	TaskCount   int
}

type WorkerManager struct {
	etcd      *clientv3.Client
	worker    Worker
	taskCount int64
	mu        sync.RWMutex
}

type Task struct {
	ID         string
	Command    string
	Args       []string
	Status     string
	WorkerID   string
	CreateTime time.Time
}

func NewWorkerManager(endpoints []string, workerID string) (*WorkerManager, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	hostname, err := utils.GetHostName()
	if err != nil {
		return nil, err
	}

	ip, err := utils.GetLocalIP()
	if err != nil {
		return nil, err
	}

	cpuUsage, err := utils.GetCPUUsage()
	if err != nil {
		return nil, err
	}

	memUsage, err := utils.GetMemoryUsage()
	if err != nil {
		return nil, err
	}

	loadUsage, err := utils.GetLoadAverage()
	if err != nil {
		return nil, err
	}

	WkMg = &WorkerManager{
		etcd: cli,
		worker: Worker{
			ID:       workerID,
			Hostname: hostname,
			IP:       ip,
			Port:     consts.ServerPort,
			Status:   "active",
			LastBeat: time.Now(),
			Resources: ResourceInfo{
				CPUUsage:    cpuUsage,
				MemoryUsage: memUsage,
				LoadUsage:   loadUsage,
				TaskCount:   0,
			},
		},
	}
	return WkMg, nil
}

func (wm *WorkerManager) Register(ctx context.Context) error {
	workerKey := fmt.Sprintf("/workers/%s", wm.worker.ID)
	resp, err := wm.etcd.Get(context.Background(), workerKey)
	if err != nil {
		return err
	}

	if len(resp.Kvs) > 0 {
		wm.etcd.Delete(context.Background(), workerKey)
	}

	if err := wm.updateResourceInfo(); err != nil {
		level.Warn(log.Logger).Log("msg", "Failed to update initial resource info", "error", err)
	}

	workerData := map[string]interface{}{
		"id":       wm.worker.ID,
		"hostname": wm.worker.Hostname,
		"ip":       wm.worker.IP,
		"port":     wm.worker.Port,
		"status":   wm.worker.Status,
		"lastBeat": time.Now(),
		"resources": map[string]interface{}{
			"cpuUsage":    wm.worker.Resources.CPUUsage,
			"memoryUsage": wm.worker.Resources.MemoryUsage,
			"loadUsage":   wm.worker.Resources.LoadUsage,
			"taskCount":   wm.GetTaskCount(),
		},
	}
	data, err := json.Marshal(workerData)
	if err != nil {
		return fmt.Errorf("failed to marshal worker data: %v", err)
	}

	_, err = wm.etcd.Put(context.Background(), workerKey, string(data))
	if err != nil {
		return fmt.Errorf("failed to put worker data to etcd: %v", err)
	}

	level.Info(log.Logger).Log("msg", "Worker registered successfully", "id", wm.worker.ID)

	go wm.updateWorkerStatus(ctx)

	return nil
}

func (wm *WorkerManager) updateWorkerStatus(ctx context.Context) {
	level.Info(log.Logger).Log("msg", "Worker status update routine started")
	workerKey := fmt.Sprintf("/workers/%s", wm.worker.ID)
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			level.Info(log.Logger).Log("msg", "Worker stopping, cleaning up registration", "id", wm.worker.ID)
			_, err := wm.etcd.Delete(context.Background(), workerKey)
			if err != nil {
				level.Error(log.Logger).Log("msg", "Failed to delete worker info from etcd", "error", err, "id", wm.worker.ID)
			} else {
				level.Info(log.Logger).Log("msg", "Worker registration cleaned up successfully", "id", wm.worker.ID)
			}
			return
		case <-ticker.C:
			if err := wm.updateResourceInfo(); err != nil {
				level.Error(log.Logger).Log("msg", "Failed to update resource info", "error", err)
				continue
			}

			workerData := map[string]interface{}{
				"id":       wm.worker.ID,
				"hostname": wm.worker.Hostname,
				"ip":       wm.worker.IP,
				"port":     wm.worker.Port,
				"status":   wm.worker.Status,
				"lastBeat": time.Now(),
				"resources": map[string]interface{}{
					"cpuUsage":    wm.worker.Resources.CPUUsage,
					"memoryUsage": wm.worker.Resources.MemoryUsage,
					"loadUsage":   wm.worker.Resources.LoadUsage,
					"taskCount":   wm.GetTaskCount(),
				},
			}

			data, err := json.Marshal(workerData)
			if err != nil {
				level.Error(log.Logger).Log("msg", "Failed to marshal worker data", "error", err)
				continue
			}

			_, err = wm.etcd.Put(context.Background(), workerKey, string(data))
			if err != nil {
				level.Error(log.Logger).Log("msg", "Failed to update worker info", "error", err, "id", wm.worker.ID)
			}
		}
	}
}

func (wm *WorkerManager) IncrTaskCount() {
	wm.mu.Lock()
	wm.taskCount++
	wm.mu.Unlock()
}

func (wm *WorkerManager) DecrTaskCount() {
	wm.mu.Lock()
	wm.taskCount--
	wm.mu.Unlock()
}

func (wm *WorkerManager) GetTaskCount() int64 {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.taskCount
}

func (wm *WorkerManager) updateResourceInfo() error {
	cpuUsage, err := utils.GetCPUUsage()
	if err != nil {
		return fmt.Errorf("get CPU usage error: %v", err)
	}

	memUsage, err := utils.GetMemoryUsage()
	if err != nil {
		return fmt.Errorf("get memory usage error: %v", err)
	}

	loadUsage, err := utils.GetLoadAverage()
	if err != nil {
		return fmt.Errorf("get load average error: %v", err)
	}

	var taskDao = &dao.Task{}
	taskCount, err := taskDao.WithContext(context.Background()).GetTargetNodeTaskCount(wm.worker.Hostname)
	if err != nil {
		return fmt.Errorf("get task count error: %v", err)
	}

	wm.mu.Lock()
	wm.worker.Resources.CPUUsage = cpuUsage
	wm.worker.Resources.MemoryUsage = memUsage
	wm.worker.Resources.LoadUsage = loadUsage
	wm.worker.Resources.TaskCount = int(taskCount)
	wm.mu.Unlock()

	return nil
}

func (wm *WorkerManager) getWorkBase() (map[string]ResourceInfo, error) {
	resp, err := wm.etcd.Get(context.Background(), "/workers/", clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	resources := make(map[string]ResourceInfo)
	for _, kv := range resp.Kvs {
		var workerData map[string]interface{}
		if err := json.Unmarshal(kv.Value, &workerData); err != nil {
			continue
		}

		hostname := workerData["hostname"].(string)
		if resourceData, ok := workerData["resources"].(map[string]interface{}); ok {
			resources[hostname] = ResourceInfo{
				CPUUsage:    resourceData["cpuUsage"].(float64),
				MemoryUsage: resourceData["memoryUsage"].(float64),
				LoadUsage:   resourceData["loadUsage"].(float64),
				TaskCount:   int(resourceData["taskCount"].(int)),
			}
		}
	}
	return resources, nil
}
