package logger

import (
	"fmt"
	"os"
	"strconv"
)

//2018/3/26 0:01.383 DEBUG logDebug.go:29 this is a debug log
//2006-01-02 15:04:05.999
type FileLogger struct {
	level       int
	logPath     string
	logName     string
	file        *os.File      //正常文件句柄
	warnFile    *os.File      //错误的文件句柄
	LogDataChan chan *LogData //队列
}

func NewFileLogger(config map[string]string) (log LogInterface, err error) {
	logPath, ok := config["log_path"]
	if !ok {
		err = fmt.Errorf("not found log_path")
		return
	}

	logName, ok := config["log_name"]
	if !ok {
		err = fmt.Errorf("not found log_name")
		return
	}

	logLevel, ok := config["log_level"]
	if !ok {
		err = fmt.Errorf("not found log_level")
		return
	}

	logChanSize, ok := config["log_chan_size"]
	if !ok {
		logChanSize = "50000"
	}

	chanSize, err := strconv.Atoi(logChanSize) //字符串转为整型
	if err != nil {
		chanSize = 50000
	}

	level := getLogLevel(logLevel)

	log = &FileLogger{
		level:       level,
		logPath:     logPath,
		logName:     logName,
		LogDataChan: make(chan *LogData, chanSize),
	}

	log.Init()
	return
}

//初始化
func (f *FileLogger) Init() {
	filename := fmt.Sprintf("%s/%s.log", f.logPath, f.logName)
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		panic(fmt.Sprintf("open faile %s failed, err: %v", filename, err))
	}

	f.file = file

	//写错误日志和fatal日志的文件
	filename = fmt.Sprintf("%s/%s.log.wf", f.logPath, f.logName)
	file, err = os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		panic(fmt.Sprintf("open faile %s failed, err: %v", filename, err))
	}

	f.warnFile = file
	go f.writeLogBackground()
}

func (f *FileLogger) writeLogBackground() {
	for logData := range f.LogDataChan { //如果是空的就阻塞
		var file *os.File = f.file
		if logData.WarnAndFatal {
			file = f.warnFile
		}
		fmt.Fprintf(file, "%s %s (%s:%s:%d) %s\n", logData.TimeStr, logData.LevelStr, logData.Filename, logData.FuncName, logData.LineNo, logData.Message)
	}
}
func (f *FileLogger) SetLevel(level int) {
	if level < LogLevelDebug || level > LogLevelFatal {
		level = LogLevelDebug
	}
	f.level = level
}

func (f *FileLogger) Debug(format string, args ...interface{}) {
	if f.level > LogLevelDebug {
		return
	}
	logData := WriteLog(LogLevelDebug, format, args...)
	select {
	case f.LogDataChan <- logData: //队列没有满
	default: //队列满了

	}
	f.LogDataChan <- logData
}

func (f *FileLogger) Trace(format string, args ...interface{}) {
	if f.level > LogLevelTrace {
		return
	}
	logData := WriteLog(LogLevelTrace, format, args...)
	select {
	case f.LogDataChan <- logData: //队列没有满
	default: //队列满了

	}
	f.LogDataChan <- logData
}

func (f *FileLogger) Info(format string, args ...interface{}) {
	if f.level > LogLevelInfo {
		return
	}
	logData := WriteLog(LogLevelInfo, format, args...)
	select {
	case f.LogDataChan <- logData: //队列没有满
	default: //队列满了

	}
	f.LogDataChan <- logData
}

func (f *FileLogger) Warn(format string, args ...interface{}) {
	if f.level > LogLevelWarn {
		return
	}
	logData := WriteLog(LogLevelWarn, format, args...)
	select {
	case f.LogDataChan <- logData: //队列没有满
	default: //队列满了

	}
	f.LogDataChan <- logData
}

func (f *FileLogger) Fatal(format string, args ...interface{}) {
	if f.level > LogLevelFatal {
		return
	}
	logData := WriteLog(LogLevelFatal, format, args...)
	select {
	case f.LogDataChan <- logData: //队列没有满
	default: //队列满了

	}
	f.LogDataChan <- logData
}

func (f *FileLogger) Error(format string, args ...interface{}) {
	if f.level > LogLevelError {
		return
	}
	logData := WriteLog(LogLevelError, format, args...)
	select {
	case f.LogDataChan <- logData: //队列没有满
	default: //队列满了
	}

	f.LogDataChan <- logData
}

func (f *FileLogger) Close() {
	f.file.Close()
	f.warnFile.Close()
}
