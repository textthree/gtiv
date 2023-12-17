package pionturnclient

import (
	"testing"
)

// 一个测试用例文件中可以有多个测试函数
// 测试用例函数名称必须以 Test 开头，一般来说就是 Test+被测试函数名
// 测试用例函数的参数必须是 *testing.t
func TestPionUdpClient(t *testing.T) {
	// 当出现错误时，可以使用 t.Fatal() 来格式化输出错误信息，并退出程序。
	// 可以使用 t.Logf() 方法来输出相应日志
	PionUdpClient()
}
