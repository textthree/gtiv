package logger

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

// debug函数
func P(data interface{}, exit ...int) {
	fmt.Println("--------------------------------------- T3Go --------------------------------------------")
	Cgreen("funcs.P()")
	// 获取 reflect.Value
	rVal := reflect.ValueOf(data)
	rKind := rVal.Kind().String()
	rType := rVal.Type().String()
	fmt.Printf("%v -> %v:\n", rKind, rType)
	switch rKind {
	case "array", "slice":
		switch rType {
		case "[]int":
			array := data.([]int)
			for k, v := range array {
				fmt.Println("[", k, "]", "=>", v)
			}
		case "[]int64":
			array := data.([]int64)
			for k, v := range array {
				fmt.Println("[", k, "]", "=>", v)
			}
		case "[]string":
			array := data.([]string)
			for k, v := range array {
				fmt.Println("[", k, "]", "=>", v)
			}
		case "[]uint8":
			fmt.Printf("%v", string(rVal.Bytes()))
		case "[]map[string]interface {}":
			array := data.([]map[string]interface{})
			for k, v := range array {
				fmt.Println("[", k, "]", "=>", "map (")
				for key, val := range v {
					fmt.Println("      ", "[", key, "]", "=>", val)
				}
				fmt.Println("),")
			}
		case "[]interface {}":
			array := data.([]interface{})
			for k, v := range array {
				fmt.Println("\t", "[", k, "]", "=>", v)
			}
		default:
			fmt.Println(data)
		}
	case "map":
		switch rType {
		case "map[string]interface {}":
			for k, v := range data.(map[string]interface{}) {
				fmt.Println("\t", "[", k, "]", "=>", v)
			}
		default:
			fmt.Println(data)
		}
	case "struct":
		t := reflect.TypeOf(data)
		v := reflect.ValueOf(data)
		for k := 0; k < t.NumField(); k++ {
			fmt.Println("\t", "[", t.Field(k).Name, "]", "=>", v.Field(k).Interface())
		}
	default:
		fmt.Println(data)
	}
	fmt.Printf("\n")
	if len(exit) > 0 {
		os.Exit(0)
	}
}

// 控制台消息提示
func Warning(message string) {
	fmt.Printf("--------------------------------------- T3Go --------------------------------------------\n")
	fmt.Printf("%c[%d;%d;%dm%s%c[0m\n", 0x1B, 1, 97, 31, message, 0x1B)
	fmt.Printf("-----------------------------------------------------------------------------------------\n")
}

// sql 调试，按顺序将 ? 替换为传入的值
func SqlDebug(sql string, args ...interface{}) {
	if len(args) > 0 {
		for _, v := range args {
			value := trim(tostring(v))
			sql = strings.Replace("?", "'"+value+"'", sql, 1)
		}
	}
	Cpink(sql, 33)
}
