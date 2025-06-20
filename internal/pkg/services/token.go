package services

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/samber/lo"
	"golang.org/x/xerrors"

	"github.com/auth-engine/config"
	"github.com/auth-engine/internal/pkg/dao"
	"github.com/auth-engine/internal/pkg/models"
	"github.com/auth-engine/pkg/constants"
)

type ITokenService interface {
	GetTokenAuthInfo(token string) ([]*dao.TokenAuthInfo, error)
	CheckTokenExists(appScenarioName, modelName, envName, id string) (bool, error)
	GenerateToken() (string, error)
	QueryPageList(
		workspaceID string, pageParam dao.PageParam, orderParam dao.OrderParam, queryParam dao.TokenQueryParam,
	) (int64, []*models.TokenBaseEntity, error)
	Create(req *models.CreateTokenReq, userInfo *models.UserInfo, tokenStr string) (*dao.TokenEntity, error)
	Get(id string) (*models.TokenResp, error)
	Update(id string, req *models.UpdateTokenReq, tokenEntity *dao.TokenEntity, userInfo *models.UserInfo) (*dao.TokenEntity, error)
	Delete(id string) error
	FindTokenEntity(id string) (*dao.TokenEntity, error)
}

type TokenService struct {
	TokenDao  dao.ITokenDao
	PolicyDao dao.ITokenValidityPolicyDao
}

func NewTokenService(tokenDao dao.ITokenDao, policyDao dao.ITokenValidityPolicyDao) ITokenService {
	return &TokenService{
		TokenDao:  tokenDao,
		PolicyDao: policyDao,
	}
}

func (s *TokenService) GenerateToken() (string, error) {
	// Define key prefix
	prefix := "sk-"

	// Generate 32 random bytes
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Encode the random bytes in base64 and trim padding
	randomString := base64.URLEncoding.EncodeToString(randomBytes)
	randomString = strings.TrimRight(randomString, "=")

	// Combine prefix with random string
	apiKey := fmt.Sprintf("%s%s", prefix, randomString)

	return apiKey, nil
}

func (s *TokenService) CheckTokenExists(appScenarioName, modelName, envName, id string) (bool, error) {
	return s.TokenDao.CheckTokenExists(appScenarioName, modelName, envName, id)
}

func (s *TokenService) QueryPageList(
	workspaceID string, pageParam dao.PageParam, orderParam dao.OrderParam, queryParam dao.TokenQueryParam,
) (int64, []*models.TokenBaseEntity, error) {
	count, tokens, err := s.TokenDao.QueryPageList(workspaceID, pageParam, orderParam, queryParam)
	if err != nil {
		return 0, nil, xerrors.Errorf("failed to query token list: %v", err)
	}
	envMap := config.GetEnvMap()
	items := lo.Map(tokens, func(token *dao.TokenEntity, _ int) *models.TokenBaseEntity {
		var expiredTime string
		if token.ExpiredTime != nil {
			expiredTime = token.ExpiredTime.Format(constants.TimeFormat)
		}
		item := &models.TokenBaseEntity{
			CreateTime:           token.CreateTime.Format(constants.TimeFormat),
			UpdateTime:           token.UpdateTime.Format(constants.TimeFormat),
			UpdateBy:             token.UpdateBy,
			CreateBy:             token.CreateBy,
			ID:                   token.ID,
			WorkspaceID:          token.WorkspaceID,
			ExpiredTime:          &expiredTime,
			AppScenarioName:      token.AppScenarioName,
			ModelName:            token.ModelName,
			EnvName:              token.EnvName,
			EnvAlias:             envMap[token.EnvName], // 环境别名
			MaxConcurrency:       token.MaxConcurrency,
			EnableValidityPolicy: token.EnableValidityPolicy,
			PolicyType:           token.PolicyType,
			Token:                token.Token,
		}
		return item
	})
	return count, items, nil
}

func (s *TokenService) GeneratePolicyEntitys(tokenID, policyType string, validityPolicy []*models.ValidityPolicy) []*dao.TokenValidityPolicy {
	policyEntitys := lo.Map(validityPolicy, func(policy *models.ValidityPolicy, _ int) *dao.TokenValidityPolicy {
		policyEntity := &dao.TokenValidityPolicy{
			TokenID:    tokenID,
			PolicyType: policyType,
			StartDay:   policy.StartDay,
			EndDay:     policy.EndDay,
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		}
		if policy.StartTime != nil && *policy.StartTime != "" {
			startTime, _ := time.ParseInLocation(constants.DaliyTimeFormat, *policy.StartTime, time.Local)
			startTime = startTime.AddDate(1, 0, 0)
			policyEntity.StartTime = &startTime
		}
		if policy.EndTime != nil && *policy.EndTime != "" {
			endTime, _ := time.ParseInLocation(constants.DaliyTimeFormat, *policy.EndTime, time.Local)
			endTime = endTime.AddDate(1, 0, 0)
			policyEntity.EndTime = &endTime
		}
		if policy.StartDate != nil && *policy.StartDate != "" {
			startDate, _ := time.ParseInLocation(constants.DateRangeTimeFormat, *policy.StartDate, time.Local)
			policyEntity.StartDate = &startDate
		}
		if policy.EndDate != nil && *policy.EndDate != "" {
			endDate, _ := time.ParseInLocation(constants.DateRangeTimeFormat, *policy.EndDate, time.Local)
			policyEntity.EndDate = &endDate
		}
		return policyEntity
	})
	return policyEntitys
}

func (s *TokenService) Create(req *models.CreateTokenReq, userInfo *models.UserInfo, tokenStr string) (*dao.TokenEntity, error) {
	tokenID := dao.NewUUID()
	tokenEntity := &dao.TokenEntity{
		ID:                   tokenID,
		WorkspaceID:          req.WorkspaceID,
		AppScenarioName:      req.AppScenarioName,
		ModelName:            req.ModelName,
		EnvName:              req.EnvName,
		EnableValidityPolicy: req.EnableValidityPolicy,
		MaxConcurrency:       req.MaxConcurrency,
		PolicyType:           req.PolicyType,
		CommonModel: dao.CommonModel{
			CreateBy:   userInfo.Username,
			CreateTime: time.Now(),
			UpdateBy:   userInfo.Username,
			UpdateTime: time.Now(),
		},
	}
	if req.ExpiredTime != nil && *req.ExpiredTime != "" {
		expiredTime, _ := time.ParseInLocation(constants.TimeFormat, *req.ExpiredTime, time.Local)
		tokenEntity.ExpiredTime = &expiredTime
	}
	if tokenStr == "" {
		token, err := s.GenerateToken()
		if err != nil {
			return nil, xerrors.Errorf("failed to generate token: %v", err)
		}
		tokenEntity.Token = token
	} else {
		tokenEntity.Token = tokenStr
	}
	// 创建有效期策略
	if req.EnableValidityPolicy {
		policyEntitys := s.GeneratePolicyEntitys(tokenID, req.PolicyType, req.ValidityPolicy)
		if err := s.PolicyDao.BatchCreate(policyEntitys, 100); err != nil {
			return nil, xerrors.Errorf("failed to create token validity policy: %v", err)
		}
	}
	if err := s.TokenDao.Create(tokenEntity); err != nil {
		return nil, xerrors.Errorf("failed to create token: %v", err)
	}
	return tokenEntity, nil
}

func (s *TokenService) Get(id string) (*models.TokenResp, error) {
	token, err := s.TokenDao.Get(id)
	if err != nil {
		return nil, xerrors.Errorf("failed to get token: %v", err)
	}
	envMap := config.GetEnvMap()
	var expiredTime string
	if token.ExpiredTime != nil {
		expiredTime = token.ExpiredTime.Format(constants.TimeFormat)
	}
	resp := &models.TokenResp{
		CreateTime:           token.CreateTime,
		UpdateTime:           token.UpdateTime,
		UpdateBy:             token.UpdateBy,
		CreateBy:             token.CreateBy,
		ID:                   token.ID,
		WorkspaceID:          token.WorkspaceID,
		ExpiredTime:          &expiredTime,
		AppScenarioName:      token.AppScenarioName,
		EnvName:              token.EnvName,
		EnvAlias:             envMap[token.EnvName], // 环境别名
		ModelName:            token.ModelName,
		MaxConcurrency:       token.MaxConcurrency,
		EnableValidityPolicy: token.EnableValidityPolicy,
		PolicyType:           token.PolicyType,
		Token:                token.Token,
	}
	if !token.EnableValidityPolicy {
		return resp, nil
	}
	policys, err := s.PolicyDao.ListByTokenID(id, token.PolicyType)
	if err != nil {
		return nil, xerrors.Errorf("failed to get token policy: %v", err)
	}
	policyData := lo.Map(policys, func(policy *dao.TokenValidityPolicy, _ int) *models.ValidityPolicy {
		var startTime, endTime, sartDate, endDate string
		if policy.StartTime != nil {
			startTime = policy.StartTime.Format(constants.DaliyTimeFormat)
		}
		if policy.EndTime != nil {
			endTime = policy.EndTime.Format(constants.DaliyTimeFormat)
		}
		if policy.StartDate != nil {
			sartDate = policy.StartDate.Format(constants.DateRangeTimeFormat)
		}
		if policy.EndDate != nil {
			endDate = policy.EndDate.Format(constants.DateRangeTimeFormat)
		}
		return &models.ValidityPolicy{
			ID:         int32(policy.ID),
			TokenID:    policy.TokenID,
			PolicyType: policy.PolicyType,
			StartTime:  &startTime,
			EndTime:    &endTime,
			StartDay:   policy.StartDay,
			EndDay:     policy.EndDay,
			StartDate:  &sartDate,
			EndDate:    &endDate,
		}
	})
	resp.ValidityPolicy = policyData
	return resp, nil
}

func (s *TokenService) Update(id string, req *models.UpdateTokenReq, tokenEntity *dao.TokenEntity, userInfo *models.UserInfo) (*dao.TokenEntity, error) {
	tokenEntity.AppScenarioName = req.AppScenarioName
	tokenEntity.ModelName = req.ModelName
	tokenEntity.EnableValidityPolicy = req.EnableValidityPolicy
	tokenEntity.MaxConcurrency = req.MaxConcurrency
	tokenEntity.PolicyType = req.PolicyType
	tokenEntity.UpdateBy = userInfo.Username
	tokenEntity.UpdateTime = time.Now()
	tokenEntity.ExpiredTime = nil
	if req.ExpiredTime != nil && *req.ExpiredTime != "" {
		expiredTime, _ := time.ParseInLocation(constants.TimeFormat, *req.ExpiredTime, time.Local)
		tokenEntity.ExpiredTime = &expiredTime
	}
	// 先删除原有的策略，在创建
	if err := s.PolicyDao.DeleteByTokenID(id); err != nil {
		return nil, xerrors.Errorf("failed to delete old token validity policy: %v", err)
	}
	if req.EnableValidityPolicy {
		policyEntitys := s.GeneratePolicyEntitys(id, req.PolicyType, req.ValidityPolicy)
		if err := s.PolicyDao.BatchCreate(policyEntitys, 100); err != nil {
			return nil, xerrors.Errorf("failed to create token validity policy: %v", err)
		}
	}
	if err := s.TokenDao.Update(id, tokenEntity); err != nil {
		return nil, xerrors.Errorf("failed to update token: %v", err)
	}
	return tokenEntity, nil
}

func (s *TokenService) FindTokenEntity(id string) (*dao.TokenEntity, error) {
	return s.TokenDao.Get(id)
}

func (s *TokenService) Delete(id string) error {
	if err := s.PolicyDao.DeleteByTokenID(id); err != nil {
		return xerrors.Errorf("failed to delete old token validity policy: %v", err)
	}
	if err := s.TokenDao.Delete(id); err != nil {
		return xerrors.Errorf("failed to delete token: %v", err)
	}
	return nil
}

func (s *TokenService) GetTokenAuthInfo(token string) ([]*dao.TokenAuthInfo, error) {
	return s.TokenDao.GetTokenAuthInfo(token)
}
