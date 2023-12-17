package dbkit

import (
	"github.com/text3cn/goodle/kit/strkit"
)

// 传入 page、rows 返回 sql 语句的 limit
// args[0] 页码（page）
// args[1] 每页显示多少条（rows），默认 15
// args[2] 是否前端上拉加载更多方式分页，每次多取一条判断是否有更多，控制器返回时截取掉多余的那条。1.是 0.否
// 	if len(list) > req.Rows {
//		list = list[:len(list)-1]
//	}
// args[3] 限制每页最多取多少条，默认 50
// @return skip length = LIMIT skip, length
func SqlLimitStatement(args ...int) string {
	var page, rows, maxRows = 1, 15, 50
	if len(args) > 0 && args[0] > 0 {
		page = args[0]
	}
	if len(args) > 1 && args[1] > 0 {
		rows = args[1]
	}
	var skip, length string
	if len(args) > 3 {
		maxRows = args[3]
	}
	if rows > maxRows {
		rows = maxRows
	}
	skip = strkit.Tostring((page - 1) * rows)
	if len(args) > 2 && args[2] == 1 {
		length = strkit.Tostring(rows + 1)
	} else {
		length = strkit.Tostring(rows)
	}
	return " LIMIT " + skip + "," + length
}
