package dao

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/oklog/ulid"
	"gorm.io/gorm"
)

type OrderParam struct {
	Column string `json:"column" command:"排序字段"`                 // 排序字段
	Order  string `json:"order" command:"排序方式" enums:"asc,desc"` // 排序方式
}

type PageParam struct {
	PageNum  int `json:"pageNum" command:"pageNum" query:"pageNum"`    // 页码
	PageSize int `json:"pageSize" command:"pageSize" query:"pageSize"` // 每页数量
}

type CommonModel struct {
	CreateTime time.Time `json:"createTime,omitempty" gorm:"column:create_time;type:datetime;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdateTime time.Time `json:"updateTime,omitempty" gorm:"column:update_time;type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间"`
	DelTime    time.Time `json:"delTime,omitempty" gorm:"column:del_time;type:datetime;default:'9999-12-31 00:00:00';comment:逻辑删除时间"`
	UpdateBy   string    `json:"updateBy,omitempty" gorm:"column:update_by;type:varchar(36);comment:更新人"`
	CreateBy   string    `json:"createBy,omitempty" gorm:"column:create_by;type:varchar(36);comment:创建人"`
	DelFlag    int       `json:"delFlag,omitempty" gorm:"column:del_flag;type:int;size:32;default:0;comment:逻辑删除标志【0 ： 未删除 1： 已删除】"`
}

func (o *OrderParam) GetOrderCondition() string {
	return o.Column + " " + o.Order
}

// CalculatePagination 计算并返回分页的 offset 和 limit.
func CalculatePagination(pageParam PageParam) (int, int) {
	var offset, limit int
	if pageParam.PageSize > 0 {
		limit = pageParam.PageSize
		if pageParam.PageNum > 0 {
			offset = (pageParam.PageNum - 1) * pageParam.PageSize
		}
	}
	return offset, limit
}

// https://github.com/ulid/spec
// uuid sortable by time.
// nolint:gosec
// 创建一个新的ULID
func NewUUID() string {
	now := time.Now()
	return ulid.MustNew(ulid.Timestamp(now), ulid.Monotonic(rand.New(rand.NewSource(now.UnixNano())), 0)).String()
}

// RecordExists 通用存在性检查函数.
func RecordExists(query *gorm.DB) (bool, error) {
	// 使用查询创建适当的 DB 对象并检查记录是否存在
	var count int64
	err := query.Limit(1).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("counting records: %w", err)
	}
	return count > 0, nil
}
