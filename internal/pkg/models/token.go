package models

import (
	"time"

	"github.com/auth-engine/internal/pkg/dao"
)

type TokenAuthBody struct {
	Token string `json:"token"`
}

type ValidityPolicy struct {
	ID         int32   `json:"id"`          // ID
	TokenID    string  `json:"tokenId"`     // token ID
	PolicyType string  `json:"policyType" ` // 策略类型，可以是 ‘DAILY’（每天）、‘WEEKLY’（每周）、‘DATE_RANGE’（日期范围）。
	StartTime  *string `json:"startTime"`   // 开始时间 08:00:00, 对于 ‘DAILY’ 和 ‘WEEKLY’ 类型的策略，定义策略生效的开始和结束时间。
	EndTime    *string `json:"endTime"`     // 结束时间 12:00:00
	StartDay   string  `json:"startDay"`    // 开始日, 对于 ‘WEEKLY’ 类型的策略，定义策略生效的开始和结束星期几。
	EndDay     string  `json:"endDay"`      // 结束日,
	StartDate  *string `json:"startDate"`   // 开始日期 2023-01-01, 对于 ‘DATE_RANGE’ 类型的策略，定义策略生效的开始和结束日期。
	EndDate    *string `json:"endDate"`     // 结束日期 2023-01-31
}

type TokenBaseEntity struct {
	CreateTime           string  `json:"createTime,omitempty"`
	UpdateTime           string  `json:"updateTime,omitempty"`
	UpdateBy             string  `json:"updateBy,omitempty"`
	CreateBy             string  `json:"createBy,omitempty"`
	ID                   string  `json:"id"`                    // ID
	WorkspaceID          string  `json:"workspaceId"`           // 工作空间ID
	ExpiredTime          *string `json:"expiredTime,omitempty"` // 过期时间
	AppScenarioName      string  `json:"appScenarioName"`       // 应用场景名称
	ModelName            string  `json:"modelName"`             // 模型名称
	EnvName              string  `json:"envName"`               // 环境名称
	EnvAlias             string  `json:"envAlias"`              // 环境别名，用于展示
	MaxConcurrency       int     `json:"maxConcurrency"`        // 最大并发数
	EnableValidityPolicy bool    `json:"enableValidityPolicy"`  // 是否启用有效期策略
	PolicyType           string  `json:"policyType"`            // 策略类型，可以是 ‘DAILY’（每天）、‘WEEKLY’（每周）、‘DATE_RANGE’（日期范围）。
	Token                string  `json:"token"`                 // token
}

type TokenResp struct {
	CreateTime           time.Time         `json:"createTime,omitempty"`
	UpdateTime           time.Time         `json:"updateTime,omitempty"`
	UpdateBy             string            `json:"updateBy,omitempty"`
	CreateBy             string            `json:"createBy,omitempty"`
	ID                   string            `json:"id"`                    // ID
	WorkspaceID          string            `json:"workspaceId"`           // 工作空间ID
	ExpiredTime          *string           `json:"expiredTime,omitempty"` // 过期时间
	AppScenarioName      string            `json:"appScenarioName"`       // 应用场景名称
	ModelName            string            `json:"modelName"`             // 模型名称
	EnvName              string            `json:"envName"`               // 环境名称
	EnvAlias             string            `json:"envAlias"`              // 环境别名，用于展示
	MaxConcurrency       int               `json:"maxConcurrency"`        // 最大并发数
	EnableValidityPolicy bool              `json:"enableValidityPolicy"`  // 是否启用有效期策略
	PolicyType           string            `json:"policyType"`            // 策略类型，可以是 ‘DAILY’（每天）、‘WEEKLY’（每周）、‘DATE_RANGE’（日期范围）。
	Token                string            `json:"token"`                 // token
	ValidityPolicy       []*ValidityPolicy `json:"validityPolicy"`        // 有效期策略
}

type TokenListReq struct {
	WorkspaceID string              `json:"workspaceId"`
	PageParam   dao.PageParam       `json:"pageParam"`  // 分页参数
	OrderParam  dao.OrderParam      `json:"orderParam"` // 排序参数
	QueryParam  dao.TokenQueryParam `json:"queryParam"` // 查询参数
}

type CreateTokenReq struct {
	WorkspaceID          string            `json:"workspaceId"`          // 工作空间ID
	ExpiredTime          *string           `json:"expiredTime"`          // 过期时间
	AppScenarioName      string            `json:"appScenarioName"`      // 应用场景名称
	ModelName            string            `json:"modelName"`            // 模型名称
	EnvName              string            `json:"envName"`              // 环境名称
	MaxConcurrency       int               `json:"maxConcurrency"`       // 最大并发数
	EnableValidityPolicy bool              `json:"enableValidityPolicy"` // 是否启用有效期策略
	PolicyType           string            `json:"policyType"`           // 策略类型，可以是 ‘DAILY’（每天）、‘WEEKLY’（每周）、‘DATE_RANGE’（日期范围）。
	ValidityPolicy       []*ValidityPolicy `json:"validityPolicy"`       // 有效期策略
}

type UpdateTokenReq struct {
	WorkspaceID          string            `json:"workspaceId"`          // 工作空间ID
	ExpiredTime          *string           `json:"expiredTime"`          // 过期时间           `json:"expiredTime"`          // 过期时间
	AppScenarioName      string            `json:"appScenarioName"`      // 应用场景名称
	ModelName            string            `json:"modelName"`            // 模型名称
	MaxConcurrency       int               `json:"maxConcurrency"`       // 最大并发数
	EnableValidityPolicy bool              `json:"enableValidityPolicy"` // 是否启用有效期策略
	PolicyType           string            `json:"policyType"`           // 策略类型，可以是 ‘DAILY’（每天）、‘WEEKLY’（每周）、‘DATE_RANGE’（日期范围）。
	ValidityPolicy       []*ValidityPolicy `json:"validityPolicy"`       // 有效期策略
}

type ExportTokenFileReq struct {
	WorkspaceID string `json:"workspaceId"`
	EnvName     string `json:"envName"`
	EnvAlias    string `json:"envAlias"`
}
