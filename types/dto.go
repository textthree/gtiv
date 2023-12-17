package types

// 嵌套结构体赋值，需要先 new 出实例再赋值，不能再声明结构体时直接在花括号内赋值
type BaseRes struct {
	ApiCode    int    `swaggertype:"integer"`
	ApiMessage string `swaggertype:"string"`
}

// 通用无数据结构，在控制器中返回数据可以不按 dto 中定义的输入输出类型来返回，但 swager 就没类型了
type AnyDataRes struct {
	ApiCode    int    `swaggertype:"integer"`
	ApiMessage string `swaggertype:"string"`
	Data       interface{}
}
