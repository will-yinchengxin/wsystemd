package utils

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

var (
	SUCCESS            = &CodeType{200, "请求成功"}
	DBErr              = &CodeType{2000, "db err"}
	DBRecorderNotExist = &CodeType{2000, "db recorder not exist"}

	StartJobFail   = &CodeType{1001, "启动任务失败, 请检查启动命令"}
	StopNotExist   = &CodeType{1002, "停止任务失败, JobId 不存在"}
	StopJobFail    = &CodeType{1003, "停止任务失败, 请重试"}
	ReqParamErr    = &CodeType{1004, "请求参数错误, 请检查"}
	ServerErr      = &CodeType{1005, "服务端异常"}
	ServerNotExist = &CodeType{1006, "HostName 不存在"}
	JobExecOutTime = &CodeType{1006, "任务执行超时"}

	NoAvailableWorker = &CodeType{2001, "没有可用的 Worker"}
)

type CodeType struct {
	Code int    `json:"code"`
	Msg  string `json:"message"`
}

type RepType struct {
	CodeType
	Data interface{} `json:"data"`
}

func SendResponse(rep *CodeType, data interface{}) *RepType {
	r := new(RepType)
	r.Code = rep.Code
	r.Msg = rep.Msg
	r.Data = data
	return r
}

func Out(ctx *gin.Context, data interface{}) {
	retData := SendResponse(SUCCESS, data)
	ctx.JSON(http.StatusOK, retData)
	ctx.Abort()
	return
}

func Success(ctx *gin.Context) {
	retData := SendResponse(SUCCESS, map[string]interface{}{})
	ctx.JSON(http.StatusOK, retData)
	ctx.Abort()
	return
}

func Error(ctx *gin.Context, rep *CodeType) {
	retData := SendResponse(rep, map[string]interface{}{})
	ctx.JSON(http.StatusOK, retData)
	ctx.Abort()
	return
}

func ErrorWithData(ctx *gin.Context, rep *CodeType, data interface{}) {
	retData := SendResponse(rep, data)
	ctx.JSON(http.StatusOK, retData)
	ctx.Abort()
	return
}

func MessageError(ctx *gin.Context, msg string) {
	retData := SendResponse(&CodeType{Code: 5003, Msg: msg}, map[string]interface{}{})
	ctx.JSON(http.StatusOK, retData)
	ctx.Abort()
	return
}
