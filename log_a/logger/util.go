package logger

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"time"
)

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

func WriteLog(file *os.File, level int, format string, args ...interface{}) {

	now := time.Now()
	nowStr := now.Format("2006-01-02 15:04:05.999 ") //时间

	levelStr := getLevelText(level)

	fileName, funcName, lineNo := GetLineInfo() //文件名，函数名，行号

	fileName = path.Base(fileName) //去除绝对路径得到文件名
	funcName = path.Base(funcName) //去除绝对路径得到函数名

	msg := fmt.Sprintf(format, args...)

	fmt.Fprintf(file, "%s %s (%s:%s:%d) %s\n", nowStr, levelStr, fileName, funcName, lineNo, msg)
}
