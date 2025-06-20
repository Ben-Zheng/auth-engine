package models

type DataResult[T any] struct {
	Success   bool   `json:"success"`   // 是否成功
	Message   string `json:"message"`   // 提示信息
	Code      int    `json:"code"`      // 状态码
	Result    T      `json:"result"`    // 返回数据
	Timestamp int64  `json:"timestamp"` // 时间戳
}

type IPage[Y any] struct {
	Records     []Y   `json:"records"`     // 当前页数据
	Total       int64 `json:"total"`       // 总记录数
	Size        int   `json:"size"`        // 每页记录数
	Current     int   `json:"current"`     // 当前页
	SearchCount bool  `json:"searchCount"` // 是否执行分页
	Pages       int   `json:"pages"`       // 总页数
	Offset      int   `json:"offset"`      // 偏移量
}
