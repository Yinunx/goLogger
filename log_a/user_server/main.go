package main

import (
	"log_a/logger"
	"time"
)

func initLogger(name, logPath, logName string, level string) (err error) {
	m := make(map[string]string, 8)
	m["log_path"] = logPath
	m["log_name"] = logName
	m["log_level"] = level
	//err = logger.InitLogger("file", m)
	err = logger.InitLogger(name, m)
	if err != nil {
		return
	}

	logger.Debug("init logger success")
	return
}

func Run() {
	for {
		logger.Debug("user server is running")
		time.Sleep(time.Second)
	}
}

func main() {
	//initLogger("c:/loggos/", "user_server", "debug")
	initLogger("console", "c:/loggos/", "user_server", "debug")
	Run()
	return
}
