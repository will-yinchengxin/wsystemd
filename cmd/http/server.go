package http

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/log/level"
	"github.com/sirupsen/logrus"
	"io"
	"strconv"
	"strings"
	"time"
	"wsystemd/cmd/http/core"
	"wsystemd/cmd/http/middlewares"
	"wsystemd/cmd/log"
)

func SetupRouter() (*gin.Engine, []func(), error) {
	cleanFun := make([]func(), 0)
	if err := core.FetchCoreConfig(); err != nil {
		return nil, nil, err
	}
	cleanFun = append(cleanFun, core.InitMysql())
	gin.DefaultWriter = io.MultiWriter()
	engine := gin.Default()
	engine.Use(func(c *gin.Context) {
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		var bodyMap map[string]interface{}
		_ = json.Unmarshal(bodyBytes, &bodyMap)
		c.Set("bodyMap", bodyMap)
	})
	engine.Use(middlewares.Cors())
	engine.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		if strings.Contains(param.Path, "/v1/agent/tasks/report") {
			return ""
		}
		header := make(map[string]interface{})
		for k, v := range param.Request.Header {
			header[k] = v[0]
		}

		body := make(map[string]interface{})
		if len(param.Keys["bodyMap"].(map[string]interface{})) > 0 {
			body = param.Keys["bodyMap"].(map[string]interface{})
		} else {
			for k, v := range param.Request.Form {
				body[k] = v[0]
			}
		}
		logMap := logrus.Fields{
			"clientIP":    param.ClientIP,
			"timestamp":   param.TimeStamp.Unix(),
			"time":        time.Now().Format("2006-01-02 15:04:05"),
			"method":      param.Method,
			"path":        param.Path,
			"proto":       param.Request.Proto,
			"status":      param.StatusCode,
			"duration":    strconv.Itoa(int(param.Latency.Milliseconds())) + "ms",
			"userAgent":   param.Request.UserAgent(),
			"errorMsg":    param.ErrorMessage,
			"resBodySize": param.BodySize,
			"resData":     param.Keys["resData"],
		}

		level.Info(log.Logger).Log("Header", header, "Body", body, "Trace", logMap)
		return ""
	}))

	initRouter(engine)
	level.Info(log.Logger).Log("msg", "SetupRouter Success")
	return engine, cleanFun, nil
}
