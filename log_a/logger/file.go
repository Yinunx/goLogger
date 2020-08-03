package logger

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

//2018/3/26 0:01.383 DEBUG logDebug.go:29 this is a debug log
//2006-01-02 15:04:05.999
type FileLogger struct {
	level         int
	logPath       string
	logName       string
	file          *os.File      //正常文件句柄
	warnFile      *os.File      //错误的文件句柄
	LogDataChan   chan *LogData //队列
	logSplitType  int           //哪一种切分方式
	logSplitSize  int64         //切分的大小
	lastSplitHour int           //上次切分的小时数
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

	var logSplitType int = LogSplitTypeHour
	var logSplitSize int64
	logSplitStr, ok := config["log_split_type"]
	if !ok {
		logSplitStr = "hour"
	} else {
		if logSplitStr == "size" {
			logSplitSizeStr, ok := config["log_split_size"]
			if !ok {
				logSplitSizeStr = "104857600" //100M
			}

			logSplitSize, err = strconv.ParseInt(logSplitSizeStr, 10, 64) //10进制 64位整数
			if err != nil {
				logSplitSize = 104857600
			}

			logSplitType = LogSplitTypeSize
		} else {
			logSplitType = LogSplitTypeHour
		}
	}

	chanSize, err := strconv.Atoi(logChanSize) //字符串转为整型
	if err != nil {
		chanSize = 50000
	}

	level := getLogLevel(logLevel)

	log = &FileLogger{
		level:         level,
		logPath:       logPath,
		logName:       logName,
		LogDataChan:   make(chan *LogData, chanSize),
		logSplitSize:  logSplitSize,
		logSplitType:  logSplitType,
		lastSplitHour: time.Now().Hour(), //当前的小时数
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

//按照时间切割日志
func (f *FileLogger) splitFileHour(warnFile bool) {
	now := time.Now()
	hour := now.Hour()
	if hour == f.lastSplitHour { //如果是日志里的时间就不要切分
		return
	}

	//先备份再切割
	var backupFilename string
	f.lastSplitHour = hour
	var filename string //老的日志路径重新构造

	if warnFile {
		backupFilename = fmt.Sprintf("%s%s.log.wf%04d%02d%02d%02d", f.logPath, f.logName, now.Year(), now.Month(), now.Day(), f.lastSplitHour)
		filename = fmt.Sprintf("%s%s.log.wf", f.logPath, f.logName)
	} else {
		backupFilename = fmt.Sprintf("%s%s.log%04d%02d%02d%02d", f.logPath, f.logName, now.Year(), now.Month(), now.Day(), f.lastSplitHour)
		filename = fmt.Sprintf("%s%s.log", f.logPath, f.logName)
	}
	//当前日志关掉
	file := f.file
	if warnFile {
		file = f.warnFile
	}

	file.Close()
	os.Rename(filename, backupFilename) //备份文件

	//重新打开日志文件
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		return
	}

	//时间切割的逻辑
	if warnFile {
		f.warnFile = file
	} else {
		f.file = file
	}
}

//按照大小切割
func (f *FileLogger) splitFileSize(warnFile bool) {

	file := f.file
	if warnFile {
		file = f.warnFile
	}

	//获取当前文件的大小
	statInfo, err := file.Stat()
	if err != nil {
		return
	}

	fileSize := statInfo.Size()
	if fileSize <= f.logSplitSize { //没达到上限
		return
	}

	//先备份再切割
	var backupFilename string
	var filename string //老的日志路径重新构造

	now := time.Now()

	if warnFile {
		backupFilename = fmt.Sprintf("%s%s.log.wf%04d%02d%02d%02d%02d%02d", f.logPath, f.logName, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
		filename = fmt.Sprintf("%s%s.log.wf", f.logPath, f.logName)
	} else {
		backupFilename = fmt.Sprintf("%s%s.log%04d%02d%02d%02d", f.logPath, f.logName, now.Year(), now.Month(), now.Day(), now.Day(), now.Hour(), now.Minute(), now.Second())
		filename = fmt.Sprintf("%s%s.log", f.logPath, f.logName)
	}

	//当前日志关掉
	file.Close()
	os.Rename(filename, backupFilename) //备份文件

	//重新打开日志文件
	file, err = os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		return
	}

	//如果关闭是哪个文件，如果wf重新打开
	if warnFile {
		f.warnFile = file
	} else {
		f.file = file
	}
}

//检测日志切分
func (f *FileLogger) checkSplitFile(warnFile bool) {
	if f.logSplitType == LogSplitTypeHour {
		f.splitFileHour(warnFile) //按时间切割日志
	} else {
		f.splitFileSize(warnFile) //按大小切割
	}
}

func (f *FileLogger) writeLogBackground() {
	for logData := range f.LogDataChan { //如果是空的就阻塞
		var file *os.File = f.file
		if logData.WarnAndFatal {
			file = f.warnFile
		}

		f.checkSplitFile(logData.WarnAndFatal)
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
