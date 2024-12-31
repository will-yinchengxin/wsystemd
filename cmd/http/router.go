package http

import (
	"github.com/gin-gonic/gin"
	"wsystemd/cmd/http/handler"
)

func initRouter(engine *gin.Engine) {
	engine.POST("/v1/jobs/submit", handler.StartJob)
	engine.PUT("/v1/jobs/:id/stop", handler.StopJob)
	engine.POST("/v1/jobs/stopBigOne", handler.StopBigOne)
	engine.POST("/v1/agent/tasks/report", handler.ReportJob)
	engine.POST("/v1/job/list", handler.JobList)
	engine.POST("/v1/job/info", handler.JobInfo)
}
