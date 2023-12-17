package logger

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type (
	Level int
)

// 定义日志级别
const (
	LevelFatal = iota // 0
	LevelError        // 1
	LevelWarning
	LevelInfo
	LevelDebug
)

var logLevel Level = LevelDebug

type logger struct {
	_log *log.Logger
	// 小于等于此级别的 level 才会被输出，例如设置为 LevelInfo 则除 Debug 以外的其他log都会输出
	logLevel Level
}

var _log = New()

// New 实例化，外部直接调用 log.XXXX
func New() *logger {
	return &logger{
		_log: log.New(os.Stderr, "", log.Lshortfile|log.LstdFlags),
		// 日志级别，默认最低级，即开启所有级别的 log 输出
		logLevel: LevelDebug,
	}
}

// 设置日志级别
func SetLogLevel(level Level) {
	_log.setLogLevel(level)
}

func (l *logger) setLogLevel(level Level) {
	l.logLevel = level
}

func Fatal(v ...interface{}) {
	_log.output(LevelFatal, v...)
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	_log.output(LevelFatal, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func Error(v ...interface{}) {
	_log.output(LevelError, v...)
}

func Errorf(format string, v ...interface{}) {
	_log.output(LevelError, fmt.Sprintf(format, v...))
}

func Warn(v ...interface{}) {
	_log.output(LevelWarning, v...)
}

func Warnf(format string, v ...interface{}) {
	_log.output(LevelWarning, fmt.Sprintf(format, v...))
}

func Info(v ...interface{}) {
	_log.output(LevelInfo, v...)
}

func Infof(format string, v ...interface{}) {
	_log.output(LevelInfo, fmt.Sprintf(format, v...))
}

func Debug(v ...interface{}) {
	_log.output(LevelDebug, v...)
}

func Debugf(format string, v ...interface{}) {
	_log.output(LevelDebug, fmt.Sprintf(format, v...))
}

func (l *logger) output(level Level, v ...interface{}) error {
	if l.logLevel < level {
		return nil
	}
	// 最后一个参数可以传 deep:数字来设置 calldepth
	length := len(v)
	calldepth := 3
	last := tostring(v[length-1])
	if len(last) >= 4 && last[:4] == "deep" {
		arr := strings.Split(last, ":")
		calldepth, _ = strconv.Atoi(arr[1])
		length--
	}
	// 格式化输出样式
	str := ""
	formatStr := "%v"
	for i := 0; i < length; i++ {
		str += fmt.Sprintf(formatStr, v[i])
	}
	switch level {
	case LevelFatal:
		formatStr = "\033[35m[FATAL]\033[0m %s"
	case LevelError:
		formatStr = "\033[31m[ERROR]\033[0m %s"
	case LevelWarning:
		formatStr = "\033[33m[WARN]\033[0m %s"
	case LevelInfo:
		formatStr = "\033[32m[INFO]\033[0m %s"
	case LevelDebug:
		formatStr = "\033[36m[DEBUG]\033[0m %s"
	}
	str = fmt.Sprintf(formatStr, str)
	//	s := fmt.Sprintf(formatStr, v...)
	// fmt.Printf("%c[%d;%d;%dm%s%c[0m", 0x1B, 1, 97, 31, message, 0x1B)
	//s = fmt.Sprintf("%c[%d;%d;%dm%s%c[0m", 0x1B, 1, 97, 31, v, 0x1B)
	return l._log.Output(calldepth, str)
}

/*************************** 带颜色输出系列 *************************/

func (l *logger) outputWithColor(color int, v ...interface{}) error {
	// 最后一个参数可以传 deep:数字来设置 calldepth
	length := len(v)
	calldepth := 3
	last := tostring(v[length-1])
	if len(last) >= 4 && last[:4] == "deep" {
		arr := strings.Split(last, ":")
		calldepth, _ = strconv.Atoi(arr[1])
		length--
	}
	// 格式化输出样式
	str := ""
	for i := 0; i < length; i++ {
		str += tostring(v[i])
	}
	str = fmt.Sprintf("\n%c[%d;%d;%dm%s%c[0m", 0x1B, 1, 97, color, str, 0x1B)
	return l._log.Output(calldepth, str)
}

func Cblack(v ...interface{}) {
	_log.outputWithColor(30, v...)
}

func Cred(v ...interface{}) {
	_log.outputWithColor(31, v...)
}

func Cgreen(v ...interface{}) {
	_log.outputWithColor(32, v...)
}

func Cyellow(v ...interface{}) {
	_log.outputWithColor(33, v...)
}

func Cblue(v ...interface{}) {
	_log.outputWithColor(34, v...)
}

func Cpink(v ...interface{}) {
	_log.outputWithColor(35, v...)
}

func Ccyan(v ...interface{}) {
	_log.outputWithColor(36, v...)
}

func Csilver(v ...interface{}) {
	_log.outputWithColor(37, v...)
}
