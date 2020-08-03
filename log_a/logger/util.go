package logger

import (
	"fmt"
	"path"
	"runtime"
	"time"
)

type LogData struct {
	Message      string //消息
	TimeStr      string //时间
	LevelStr     string //级别
	Filename     string //文件名
	FuncName     string //函数名
	LineNo       int    //行号
	WarnAndFatal bool   //是warn就写到warn字段里面去
}

//util.go 10
//代码行号
//pc 计数器 funcName 指定的函数
func GetLineInfo() (fileName string, funcName string, lineNo int) {
	pc, file, line, ok := runtime.Caller(4) //0表示util, 1表示掉Getline, 2再往上一层
	if ok {
		fileName = file
		funcName = runtime.FuncForPC(pc).Name()
		lineNo = line
	}
	return
}

/*
1.当然业务用打日志的方法，我们把日志相关的数组写入到chanenl(队列)
然后我们有一个后台线程不断地从chan里面获取这些日志，最终写入到文件中

*/
func WriteLog(level int, format string, args ...interface{}) *LogData {

	now := time.Now()
	nowStr := now.Format("2006-01-02 15:04:05.999 ") //时间

	levelStr := getLevelText(level)

	fileName, funcName, lineNo := GetLineInfo() //文件名，函数名，行号

	fileName = path.Base(fileName) //去除绝对路径得到文件名
	funcName = path.Base(funcName) //去除绝对路径得到函数名

	msg := fmt.Sprintf(format, args...)

	logData := &LogData{
		Message:      msg,
		TimeStr:      nowStr,
		LevelStr:     levelStr,
		Filename:     fileName,
		FuncName:     funcName,
		LineNo:       lineNo,
		WarnAndFatal: false,
	}

	if level == LogLevelError || level == LogLevelWarn || level == LogLevelFatal {
		logData.WarnAndFatal = true
	}

	return logData
	//fmt.Fprintf(file, "%s %s (%s:%s:%d) %s\n", nowStr, levelStr, fileName, funcName, lineNo, msg)
}
