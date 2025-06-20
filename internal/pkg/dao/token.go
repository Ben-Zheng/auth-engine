package dao

import (
	"time"

	"github.com/auth-engine/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/gorm"
)

type TokenAuthInfo struct {
	UpdateTime           time.Time  `json:"updateTime"`           // Token 更新时间
	TokenID              string     `json:"tokenId"`              // token ID
	Token                string     `json:"token"`                // token
	ExpiredTime          *time.Time `json:"expiredTime"`          // token 过期时间
	AppScenarioName      string     `json:"appScenarioName"`      // 应用场景名称
	ModelName            string     `json:"modelName"`            // 模型名称
	EnvName              string     `json:"envName"`              // 环境名称
	MaxConcurrency       int        `json:"maxConcurrency"`       // 最高并发量
	EnableValidityPolicy bool       `json:"enableValidityPolicy"` // 是否启用有效期策略
	PolicyType           string     `json:"policyType"`           // 策略类型
	StartTime            *time.Time `json:"startTime"`            // 开始时间 08:00:00
	EndTime              *time.Time `json:"endTime"`              // 结束时间 12:00:00
	StartDay             string     `json:"startDay"`             // 开始日
	EndDay               string     `json:"endDay"`               // 结束日
	StartDate            *time.Time `json:"startDate"`            // 开始日期 2023-01-01
	EndDate              *time.Time `json:"endDate"`              // 结束日期 2023-01-31
}

type TokenQueryParam struct {
	AppScenarioName string `json:"appScenarioName"` // 应用场景名称
	ModelName       string `json:"modelName"`       // 模型名称
	EnvName         string `json:"envName"`         // 环境名称
}

type TokenEntity struct {
	CommonModel
	ID                   string     `json:"id" gorm:"column:id;type:varchar(64);not null;primaryKey;comment:ID"`                                                           // ID
	WorkspaceID          string     `json:"workspaceId" gorm:"column:workspace_id;type:varchar(64);not null;index:idx_workspace_id;comment:工作空间ID"`                        // 工作空间ID
	Token                string     `json:"token" gorm:"column:token;type:varchar(255);not null;index:idx_token,unique;comment:token"`                                     // token
	ExpiredTime          *time.Time `json:"expiredTime" gorm:"column:expired_time;type:datetime;null;comment:过期时间"`                                                        // 过期时间
	AppScenarioName      string     `json:"appScenarioName" gorm:"column:app_scenario_name;type:varchar(255);not null;index:idx_app_model_env_name,unique;comment:应用场景名称"` // 应用场景名称
	ModelName            string     `json:"modelName" gorm:"column:model_name;type:varchar(255);not null;index:idx_app_model_env_name,unique;comment:模型名称"`                // 模型名称
	EnvName              string     `json:"envName" gorm:"column:env_name;type:varchar(255);not null;index:idx_app_model_env_name,unique;comment:环境名称"`                    // 环境名称
	EnableValidityPolicy bool       `json:"enableValidityPolicy" gorm:"column:enable_validity_policy;type:tinyint(1);not null;comment:是否启用有效期策略"`                          // 是否启用有效期策略
	PolicyType           string     `json:"policyType" gorm:"column:policy_type;type:enum('DAILY', 'WEEKLY', 'DATERANGE', '');null;comment:策略类型"`                          // 策略类型
	MaxConcurrency       int        `json:"maxConcurrency" gorm:"column:max_concurrency;type:int;not null;comment:最高并发量"`                                                  // 最高并发量
}

func (*TokenEntity) TableName() string {
	return "tokens"
}

func init() {
	registerInjector(func(d *daoInit) {
		setupTableModel(d, &TokenEntity{})
	})
}

type TokenDao struct {
	DB *gorm.DB
}

type ITokenDao interface {
	GetTokenAuthInfo(token string) ([]*TokenAuthInfo, error)
	QueryPageList(
		workspaceID string, pageParam PageParam, orderParam OrderParam, queryParam TokenQueryParam,
	) (int64, []*TokenEntity, error)
	CheckTokenExists(appScenarioName, modelName, envName, id string) (bool, error)
	Create(req *TokenEntity) error
	Get(id string) (*TokenEntity, error)
	Update(id string, req *TokenEntity) error
	Delete(id string) error
}

// NewAiDatasetDao return the dao interface
func NewTokenDao(db *gorm.DB) ITokenDao {
	if db == nil {
		db = GetDB()
	}
	return &TokenDao{DB: db}
}

func (d *TokenDao) GetTokenAuthInfo(token string) ([]*TokenAuthInfo, error) {
	var tokenAuthInfos []*TokenAuthInfo
	selectFields := `
	tokens.update_time,
	tokens.id as token_id,
	tokens.token,
	tokens.expired_time,
	tokens.app_scenario_name,
	tokens.model_name,
	tokens.env_name,
	tokens.max_concurrency,
	tokens.enable_validity_policy,
	tokens.policy_type,
	tvp.start_time,
	tvp.end_time,
	tvp.start_day,
	tvp.end_day,
	tvp.start_date,
	tvp.end_date`
	err := d.DB.Select(selectFields).
		Joins("left join `token_validity_policies` as tvp on tvp.token_id = tokens.id").
		Where("tokens.token = ? and tokens.del_flag = 0 and tokens.env_name = ?", token, config.CurrentEnvName).
		Debug().Table("tokens").Find(&tokenAuthInfos).Error
	if err != nil {
		hlog.Errorf("get token auth info failed, err: %v", err)
		return nil, err
	}
	return tokenAuthInfos, nil
}

func (d *TokenDao) QueryPageList(
	workspaceID string, pageParam PageParam, orderParam OrderParam, queryParam TokenQueryParam,
) (int64, []*TokenEntity, error) {
	var (
		count  int64
		tokens []*TokenEntity
	)

	query := d.DB.Debug().Where("del_flag=0 and workspace_id = ?", workspaceID)
	if queryParam.AppScenarioName != "" {
		query = query.Where("app_scenario_name like concat('%',?,'%')  ", queryParam.AppScenarioName)
	}
	if queryParam.ModelName != "" {
		query = query.Where("model_name like concat('%',?,'%')  ", queryParam.ModelName)
	}
	if queryParam.EnvName != "" {
		query = query.Where("env_name = ?", queryParam.EnvName)
	}
	err := query.Table("tokens").Count(&count).Error
	if err != nil {
		return 0, nil, err
	}
	switch orderParam.Column {
	case "createTime":
		query = query.Order("create_time" + " " + orderParam.Order)
	}

	if pageParam.PageSize != -1 {
		query = query.Offset((pageParam.PageNum - 1) * pageParam.PageSize).Limit(pageParam.PageSize)
	}
	err = query.Find(&tokens).Error
	if err != nil {
		return 0, nil, err
	}
	return count, tokens, nil
}

func (d *TokenDao) CheckTokenExists(appScenarioName, modelName, envName, id string) (bool, error) {
	query := d.DB.Debug().Model(&TokenEntity{}).Where(
		"app_scenario_name = ? AND model_name = ? AND env_name = ? AND del_flag = 0",
		appScenarioName, modelName, envName,
	)
	if id != "" {
		query = query.Where("id != ?", id)
	}
	var count int64
	err := query.Limit(1).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (d *TokenDao) Create(req *TokenEntity) error {
	return d.DB.Debug().Create(req).Error
}

func (d *TokenDao) Get(id string) (*TokenEntity, error) {
	var tokenEntity TokenEntity
	err := d.DB.Debug().Where("id = ? and del_flag = 0", id).First(&tokenEntity).Error
	if err != nil {
		return nil, err
	}
	return &tokenEntity, nil
}

func (d *TokenDao) Update(id string, req *TokenEntity) error {
	updates := map[string]interface{}{
		"expired_time":           req.ExpiredTime,
		"app_scenario_name":      req.AppScenarioName,
		"model_name":             req.ModelName,
		"enable_validity_policy": req.EnableValidityPolicy,
		"policy_type":            req.PolicyType,
		"max_concurrency":        req.MaxConcurrency,
	}
	if req.PolicyType == "" {
		updates["policy_type"] = gorm.Expr("NULL")
	}
	return d.DB.Debug().Model(&TokenEntity{}).Where("id = ?", id).Updates(updates).Error
}

func (d TokenDao) Delete(id string) error {
	// return d.DB.Debug().Model(&TokenEntity{}).Where("id = ?", id).Update("del_flag", 1).Error
	return d.DB.Debug().Model(&TokenEntity{}).Where("id = ?", id).Delete(&TokenEntity{}).Error
}
