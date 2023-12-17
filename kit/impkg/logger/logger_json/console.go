package logger_json

import (
	"fmt"
	"strconv"
	"strings"
)

/*************************** 带颜色输出系列 - 开发调试控制台输出用 *************************/

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
	str = fmt.Sprintf("%c[%d;%d;%dm%s%c[0m", 0x1B, 1, 97, color, str, 0x1B)
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
