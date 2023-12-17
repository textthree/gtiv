package logger_json

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strconv"
	"time"
)

type (
	Level int
)

// 定义日志级别
const (
	LevelPrint = iota // 0
	LevelFatal        // 1
	LevelError
	LevelWarning
	LevelInfo
	LevelDebug
	LevelTrace
)

var logLevel Level = LevelTrace

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
		logLevel: LevelTrace,
	}
}

// 设置日志级别
func SetLogLevel(level Level) {
	_log.setLogLevel(level)
}

func (l *logger) setLogLevel(level Level) {
	l.logLevel = level
}

func Print(v ...interface{}) {
	_log.output(LevelPrint, v...)
}

func Printf(format string, v ...interface{}) {
	_log.output(LevelPrint, fmt.Sprintf(format, v...))
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

func Trace(v ...interface{}) {
	_log.output(LevelTrace, v...)
}

func Tracef(format string, v ...interface{}) {
	_log.output(LevelTrace, fmt.Sprintf(format, v...))
}

func (l *logger) output(level Level, v ...interface{}) error {
	if l.logLevel < level {
		return nil
	}
	// 最后一个参数可以传 deep:数字来设置 calldepth
	length := len(v)
	out := map[string]interface{}{}
	data := []interface{}{}
	// level
	switch level {
	case LevelPrint:
		out["level"] = 70
	case LevelFatal:
		out["level"] = 60
	case LevelError:
		out["level"] = 50
	case LevelWarning:
		out["level"] = 40
	case LevelInfo:
		out["level"] = 30
	case LevelDebug:
		out["level"] = 20
	case LevelTrace:
		out["level"] = 10
	}
	// message
	for i := 0; i < length; i++ {
		data = append(data, v[i])
	}
	if len(data) == 1 {
		out["message"] = data[0]
	} else {
		for k, _ := range data {
			if err, ok := data[1].(error); ok {
				data[k] = err.Error()
			}
		}
		out["message"] = data
	}
	// time
	out["time"] = time.Now().UnixNano() / 1e6
	out["pid"] = os.Getpid()
	out["ppid"] = os.Getppid()
	hostname, err := os.Hostname()
	if err != nil {
		hostname = ""
	}
	out["hostname"] = hostname
	if out["level"].(int) >= 40 {
		out["stack"] = string(debug.Stack())
	}
	jsonStr, err := json.Marshal(out)
	if err != nil {
		fmt.Println("json 序列化失败")
	}
	fmt.Println(string(jsonStr))
	return nil
}

// 类型转字符串值
// 浮点型 3.0将会转换成字符串3, "3"
// 非数值或字符类型的变量将会被转换成JSON格式字符串
func tostring(value interface{}) string {
	var key string
	if value == nil {
		return key
	}
	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}
	return key
}
