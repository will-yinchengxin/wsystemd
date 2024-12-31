package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"wsystemd/cmd/http/consts"
	"wsystemd/cmd/http/params"
	"wsystemd/cmd/http/service"
	"wsystemd/cmd/utils"
)

// StartJob 开启任务
func StartJob(ctx *gin.Context) {
	var (
		vd  = utils.NewValidator()
		req = params.JobCfg{}
	)
	if errMsg := vd.ParseJson(ctx, &req); errMsg != "" {
		utils.MessageError(ctx, errMsg)
		return
	}
	var (
		res      interface{}
		codeType *utils.CodeType
	)
	if req.LoadMethod == "" {
		req.LoadMethod = consts.Load_Method_HASH
	}
	res, codeType = service.CreateClusterModeJob(req)
	if codeType.Code != 0 {
		utils.Error(ctx, codeType)
		return
	}
	utils.Out(ctx, res)
}

// StopJob  停止任务
func StopJob(ctx *gin.Context) {
	jobId := ctx.Param("id")
	if jobId == "" {
		utils.MessageError(ctx, "id 不能为空")
		return
	}
	var (
		codeType *utils.CodeType
	)
	codeType = service.StopSingleModeJob(jobId, true)
	if codeType.Code != 0 {
		utils.Error(ctx, codeType)
		return
	}
	utils.Success(ctx)
	return
}

func StopBigOne(ctx *gin.Context) {
	var (
		vd  = utils.NewValidator()
		req = params.BigOne{}
	)
	if errMsg := vd.ParseJson(ctx, &req); errMsg != "" {
		utils.MessageError(ctx, errMsg)
		return
	}
	codeType := service.StopBigOne(req)
	if codeType.Code != 0 {
		utils.Error(ctx, codeType)
		return
	}
	utils.Success(ctx)
}

// ReportJob 报告任务
func ReportJob(ctx *gin.Context) {
	var (
		vd  = utils.NewValidator()
		req = params.JobReporter{}
	)
	fmt.Println(ctx.Request.URL)
	if errMsg := vd.ParseQuery(ctx, &req); errMsg != "" {
		utils.MessageError(ctx, errMsg)
		return
	}
	var (
		codeType *utils.CodeType
	)
	codeType = service.SingleJobReporter(req)
	if codeType.Code != 0 {
		utils.Error(ctx, codeType)
		return
	}
	utils.Success(ctx)
}

// JobList 任务列表
func JobList(ctx *gin.Context) {

}

// JobInfo 任务详情
func JobInfo(ctx *gin.Context) {

}

func checkParam(req params.JobCfg) string {
	if req.Node == "" && req.Ip == "" {
		return "Node Ip 至少填写一个"
	}
	return ""
}
