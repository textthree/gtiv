package dto

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"time"
)

// 嵌套结构体赋值，需要先 new 出实例再赋值，不能再声明结构体时直接在花括号内赋值
type BaseRes struct {
	ApiCode    int    `dc:"业务代码" v:"required"`
	ApiMessage string `dc:"接口描述信息" v:"required"`
}

// 通用无数据结构，在控制器中返回数据可以不按 dto 中定义的输入输出类型来返回，但 swager 就没类型了
type AnyDataRes struct {
	ApiCode    int         `dc:"业务代码" v:"required"`
	ApiMessage string      `dc:"接口描述信息" v:"required"`
	Data       interface{} `v:"required"`
}

// gorm 格式化时间 start //
type TimeToUnix struct {
	time.Time
}

func (t TimeToUnix) MarshalJSON() ([]byte, error) {
	//格式化秒
	seconds := t.Unix()
	return []byte(strconv.FormatInt(seconds, 10)), nil
}
func (t TimeToUnix) Value() (driver.Value, error) {
	var zeroTime time.Time
	if t.Time.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return t.Time, nil
}
func (t *TimeToUnix) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = TimeToUnix{Time: value}
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

// gorm 格式化时间 end //
