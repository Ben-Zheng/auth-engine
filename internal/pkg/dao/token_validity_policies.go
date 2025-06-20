package dao

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type TokenValidityPolicy struct {
	ID          uint       `json:"id" gorm:"column:id;primaryKey;autoIncrement;comment:ID"`                                                                                       // ID
	WorkspaceID string     `json:"workspaceId" gorm:"column:workspace_id;type:varchar(64);not null;index:idx_workspace_id;comment:工作空间ID"`                                        // 工作空间ID
	TokenID     string     `json:"tokenId" gorm:"column:token_id;type:varchar(64);not null;index:idx_token_id;comment:token ID"`                                                  // token ID
	PolicyType  string     `json:"policyType" gorm:"column:policy_type;type:enum('DAILY', 'WEEKLY', 'DATERANGE');not null;comment:策略类型"`                                          // 策略类型
	StartTime   *time.Time `json:"startTime" gorm:"column:start_time;type:time;null;comment:开始时间"`                                                                                // 开始时间 08:00:00
	EndTime     *time.Time `json:"endTime" gorm:"column:end_time;type:time;null;comment:结束时间"`                                                                                    // 结束时间 12:00:00
	StartDay    string     `json:"startDay" gorm:"column:start_day;type:enum('MONDAY', 'TUESDAY', 'WEDNESDAY', 'THURSDAY', 'FRIDAY', 'SATURDAY', 'SUNDAY', '');null;comment:开始日"` // 开始日
	EndDay      string     `json:"endDay" gorm:"column:end_day;type:enum('MONDAY', 'TUESDAY', 'WEDNESDAY', 'THURSDAY', 'FRIDAY', 'SATURDAY', 'SUNDAY', '');null;comment:结束日"`     // 结束日
	StartDate   *time.Time `json:"startDate" gorm:"column:start_date;type:date;null;comment:开始日期"`                                                                                // 开始日期 2023-01-01
	EndDate     *time.Time `json:"endDate" gorm:"column:end_date;type:date;null;comment:结束日期"`
	CreateTime  time.Time  `json:"createTime,omitempty" gorm:"column:create_time;type:datetime;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdateTime  time.Time  `json:"updateTime,omitempty" gorm:"column:update_time;type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间"` // 结束日期 2023-01-31
}

func (*TokenValidityPolicy) TableName() string {
	return "token_validity_policies"
}

func init() {
	registerInjector(func(d *daoInit) {
		setupTableModel(d, &TokenValidityPolicy{})
	})
}

type TokenValidityPolicyDao struct {
	DB *gorm.DB
}

type ITokenValidityPolicyDao interface {
	ListByTokenID(tokenID, policyType string) ([]*TokenValidityPolicy, error)
	BatchCreate(reqList []*TokenValidityPolicy, batchSize int) error
	DeleteByTokenID(tokenID string) error
}

// NewAiDatasetDao return the dao interface
func NewTokenValidityPolicyDao(db *gorm.DB) ITokenValidityPolicyDao {
	if db == nil {
		db = GetDB()
	}
	return &TokenValidityPolicyDao{DB: db}
}

func (d *TokenValidityPolicyDao) ListByTokenID(tokenID, policyType string) ([]*TokenValidityPolicy, error) {
	var records []*TokenValidityPolicy
	query := d.DB.Debug().Where("token_id = ?", tokenID)
	if policyType != "" {
		query = query.Where("policy_type = ?", policyType)
	}
	if err := query.Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (d *TokenValidityPolicyDao) BatchCreate(reqList []*TokenValidityPolicy, batchSize int) error {
	if err := d.DB.Debug().CreateInBatches(reqList, batchSize).Error; err != nil {
		return fmt.Errorf("batch insert failed: %w", err)
	}
	return nil
}

func (d *TokenValidityPolicyDao) DeleteByTokenID(tokenID string) error {
	if err := d.DB.Debug().Where("token_id = ?", tokenID).Delete(&TokenValidityPolicy{}).Error; err != nil {
		return err
	}
	return nil
}
