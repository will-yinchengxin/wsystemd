package log

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	LogWriter = logrus.New()
)

func init() {
	LogWriter.SetFormatter(&logrus.JSONFormatter{})
	LogWriter.SetOutput(&lumberjack.Logger{
		Filename:   "./logs/task_err.log",
		MaxSize:    100,
		MaxBackups: 7,
		MaxAge:     1,
		Compress:   true,
	})

	// 	defer func() {
	//		logTmp := logger.WithFields(logrus.Fields{
	//			"url":      req.Url,
	//			"method":   req.Method,
	//			"request":  req.Body,
	//			"response": result,
	//		})
	//		if err != nil {
	//			logTmp.Errorf("%s", err.Error())
	//		} else {
	//			logTmp.Infof("remote request success")
	//		}
	//		utils.ErrorLog(err)
	//	}()
}
