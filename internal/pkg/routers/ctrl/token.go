package ctrl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/auth-engine/config"
	"github.com/auth-engine/internal/pkg/dao"
	"github.com/auth-engine/internal/pkg/models"
	"github.com/auth-engine/internal/pkg/routers/ctrl/common"
	"github.com/auth-engine/internal/pkg/services"
	"github.com/auth-engine/internal/pkg/utils"
	"github.com/auth-engine/pkg/constants"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"golang.org/x/xerrors"
)

type TokenHandler struct {
	TokenService services.ITokenService
	AppConfig    *config.AppConfig
}

func NewTokenHandler(ts services.ITokenService, appConfig *config.AppConfig, r *server.Hertz) *TokenHandler {
	handler := &TokenHandler{TokenService: ts, AppConfig: appConfig}
	authRouter := r.Group("/apis/auth.engine.io")
	authRouter.GET("/ping", handler.Ping)
	authRouter.POST("/token/auth", common.Handle(handler.TokenAuth))

	router := r.Group("/apis/auth.engine.io/v1")
	router.Use(common.VerifyAuthorization())
	router.GET("/envs", common.Handle(handler.ListEnvs))
	router.POST("/workspaces/:workspaceId/tokens/allEnv/list", common.Handle(handler.AllEnvList))
	router.POST("/workspaces/:workspaceId/tokens/list", common.Handle(handler.List))
	router.POST("/workspaces/:workspaceId/tokens/add", common.Handle(handler.CreateToken))
	router.GET("/workspaces/:workspaceId/tokens/:tokenId", common.Handle(handler.GetToken))
	router.PUT("/workspaces/:workspaceId/tokens/:tokenId", common.Handle(handler.UpdateToken))
	router.DELETE("/workspaces/:workspaceId/tokens/:tokenId", common.Handle(handler.DeleteToken))
	router.POST("/workspaces/:workspaceId/tokens/:tokenId/export", common.Handle(handler.ExportTokenFile))
	router.POST("/workspaces/:workspaceId/tokens/import", common.Handle(handler.ImportTokenFile))
	return handler
}

// Ping 测试
func (h *TokenHandler) Ping(_ context.Context, c *app.RequestContext) {
	token := c.Request.Header.Get("Authorization")
	hlog.Infof("Authorization token: %s", token)
	res := map[string]string{"message": "pong"}
	if token != "" {
		res["authorization"] = token
	}
	c.JSON(http.StatusOK, res)
}

// CheckValidityPolicy 检查 validity_policy 是否合法
func CheckValidityPolicy(policyType string, policys []*models.ValidityPolicy) *common.Error {
	// policy_type: 策略类型，可以是 DAILY（每天）、WEEKLY（每周）、DATE_RANGE（日期范围）
	// 对于 ‘DAILY’ 和 ‘WEEKLY’ 类型的策略，定义策略生效的开始和结束时间，例如：startTime: 09:00, endTime: 18:00
	// 对于 ‘WEEKLY’ 类型的策略，定义策略生效的开始和结束日期，例如：startDay: MONDAY, endDay: FRIDAY
	// 对于 ‘DATE_RANGE’ 类型的策略，定义策略生效的开始和结束日期，例如：startDate: 2022-01-01, endDate: 2022-12-31
	for _, policy := range policys {
		switch policyType {
		case constants.DaliyPolicyType:
			if policy.StartTime == nil || policy.EndTime == nil {
				return common.NewCtrlError(400, xerrors.New("startTime and endTime is required."))
			}
			check, startTime := utils.CheckTimeFormat(*policy.StartTime)
			if !check {
				return common.NewCtrlError(400, xerrors.New("startTime is invalid."))
			}
			check, endTime := utils.CheckTimeFormat(*policy.EndTime)
			if !check {
				return common.NewCtrlError(400, xerrors.New("endTime is invalid."))
			}
			if startTime.After(endTime) {
				return common.NewCtrlError(400, xerrors.New("startTime must be before endTime."))
			}
		case constants.WeeklyPolicyType:
			if policy.StartDay == "" || policy.EndDay == "" {
				return common.NewCtrlError(400, xerrors.New("startDay and endDay is required."))
			}
			if _, ok := constants.WeeklyDayMap[policy.StartDay]; !ok {
				return common.NewCtrlError(400, xerrors.New("startDay is invalid."))
			}
			if policy.StartTime != nil && policy.EndTime != nil {
				check, startTime := utils.CheckTimeFormat(*policy.StartTime)
				if !check {
					return common.NewCtrlError(400, xerrors.New("startTime is invalid."))
				}
				check, endTime := utils.CheckTimeFormat(*policy.StartTime)
				if !check {
					return common.NewCtrlError(400, xerrors.New("endTime is invalid."))
				}
				if startTime.After(endTime) {
					return common.NewCtrlError(400, xerrors.New("startTime must be before endTime."))
				}
			}
		case constants.DateRangePolicyType:
			if policy.StartDate == nil || policy.EndDate == nil {
				return common.NewCtrlError(400, xerrors.New("startDate and endDate is required."))
			}
			check, startDate := utils.CheckTimeFormat(*policy.StartDate)
			if !check {
				return common.NewCtrlError(400, xerrors.New("startDate is invalid."))
			}
			check, endDate := utils.CheckTimeFormat(*policy.EndDate)
			if !check {
				return common.NewCtrlError(400, xerrors.New("endDate is invalid."))
			}
			if startDate.After(endDate) {
				return common.NewCtrlError(400, xerrors.New("startDate must be before endDate."))
			}
		}
	}
	return nil
}

// AuthValidityPolicy 认证，当前时间是否在 validity_policy 中
// tokenAuthInfos: validity_policy 列表
// currentTime:当前时间
// return: true/false, true表示 currentTime 在 validity_policy 中，false表示不在
func AuthValidityPolicy(currentTime time.Time, tokenAuthInfos []*dao.TokenAuthInfo) bool {
	// 检查 currentTime 是否在有效时间范围内，只要有一个策略通过即可
	for _, info := range tokenAuthInfos {
		switch info.PolicyType {
		case constants.DaliyPolicyType:
			// DAILY 策略，检查当前时间是否在 startTime 和 endTime 之间
			_, currentTimeFormat := utils.CheckTimeFormat(currentTime.Format(constants.DaliyTimeFormat))
			currentTimeFormat = currentTimeFormat.AddDate(1, 0, 0)
			if currentTimeFormat.After(*info.StartTime) && currentTimeFormat.Before(*info.EndTime) {
				return true
			}
		case constants.WeeklyPolicyType:
			// WEEKLY 策略，检查当前时间是否在 startDay 和 endDay 之间
			currentWeek := int(currentTime.Weekday())
			// 周日是0，变成7，表示（周一 -> 周日）
			if currentWeek == 0 {
				currentWeek = 7
			}
			// StartDay 和 EndDay 是字符串，如果是空，直接跳过
			if info.StartDay == "" || info.EndDay == "" {
				continue
			}
			startDayWeek := constants.WeeklyDayMap[info.StartDay]
			endDayWeek := constants.WeeklyDayMap[info.EndDay]
			if startDayWeek <= currentWeek && currentWeek <= endDayWeek {
				if info.StartTime == nil || info.EndTime == nil {
					return true
				} else {
					// 如果 startTime 和 endTime 不为空，则检查当前时间是否在 startTime 和 endTime 之间
					_, currentTimeFormat := utils.CheckTimeFormat(currentTime.Format(constants.DaliyTimeFormat))
					currentTimeFormat = currentTimeFormat.AddDate(1, 0, 0)
					if currentTimeFormat.After(*info.StartTime) && currentTimeFormat.Before(*info.EndTime) {
						return true
					}
				}
			}
		case constants.DateRangePolicyType:
			// DATE_RANGE 策略，检查当前时间是否在 startDate 和 endDate 之间
			if currentTime.After(*info.StartDate) && currentTime.Before(*info.EndDate) {
				return true
			}
		}
	}
	return false
}

// TokenAuth Token 认证
// @Summary  Token 认证
// @Tags Token 管理
// @Accept  json
// @Produce  json
// @Param data body models.TokenAuthBody true "issue params"
// @Router /apis/auth-engine.io/token/auth [post]
// @Success 200 object models.DataResult[string] "成功后返回"
// @Security Bearer
func (h *TokenHandler) TokenAuth(_ context.Context, c *common.CustomReqContext) (any, error) {
	request := &models.TokenAuthBody{}
	if err := c.Bind(request); err != nil {
		return nil, xerrors.Errorf("Failed to bind body: %w", err)
	}
	token := request.Token
	if token == "" {
		hlog.Error("Token is empty")
		return nil, common.NewCtrlError(401, xerrors.New("Token is empty."))
	}
	tokenAuthInfos, err := h.TokenService.GetTokenAuthInfo(token)
	if err != nil {
		return nil, common.NewCtrlError(500, xerrors.Errorf("Internal Server Error: Failed to get token auth info, Err: %w", err))
	}
	if len(tokenAuthInfos) == 0 {
		hlog.Error("Token is invalid, no token info found")
		return nil, common.NewCtrlError(403, xerrors.New("Forbidden: Token is invalid."))
	}
	// 获取当前时间,
	currentTime := time.Now()
	expiredTime := tokenAuthInfos[0].ExpiredTime
	if expiredTime != nil && currentTime.After(*expiredTime) {
		hlog.Error("Token is expired")
		return nil, common.NewCtrlError(403, xerrors.New("Forbidden: Token is expired."))
	}
	// 生效策略开启，检查 currentTime 是否在有效时间范围内
	if tokenAuthInfos[0].EnableValidityPolicy {
		policyPass := AuthValidityPolicy(currentTime, tokenAuthInfos)
		if !policyPass {
			hlog.Error("Token is invalid, no policy pass")
			return nil, common.NewCtrlError(403, xerrors.New("Forbidden: Token is invalid for no policy pass"))
		}
	}
	hlog.Infof("Token Auth success, Token: %s", token)
	return "success", nil
}

// ListEnvs 获取环境列表
// @Summary  获取环境列表
// @Tags Token 管理
// @Accept  json
// @Produce  json
// @Param data body models.TokenListReq true "issue params"
// @Router /apis/auth.engine.io/v1/envs [get]
// @Success 200 object models.DataResult[[]config.EnvConf] "成功后返回"
// @Security Bearer
func (h *TokenHandler) ListEnvs(_ context.Context, c *common.CustomReqContext) (any, error) {
	envConfs := config.EnvConfs
	return envConfs, nil
}

// List 当前部署环境下的Token列表
// @Summary  当前部署环境下的Token列表
// @Tags Token 管理
// @Accept  json
// @Produce  json
// @Param workspaceId path string true "工作空间ID"
// @Param data body models.TokenListReq true "issue params"
// @Router /apis/auth.engine.io/v1/workspaces/:workspaceId/tokens/list [post]
// @Success 200 object models.DataResult[models.IPage[models.TokenBaseEntity]] "成功后返回"
// @Security Bearer
func (h *TokenHandler) List(_ context.Context, c *common.CustomReqContext) (any, error) {
	request := &models.TokenListReq{}
	if err := c.Bind(request); err != nil {
		return nil, xerrors.Errorf("Failed to bind body: %w", err)
	}
	// 设置当前部署环境名称
	request.QueryParam.EnvName = config.CurrentEnvName
	workspaceID := c.Param("workspaceId")
	pageParam := c.GetOrDefaultPageParam(&request.PageParam)
	count, res, err := h.TokenService.QueryPageList(workspaceID, *pageParam, request.OrderParam, request.QueryParam)
	if err != nil {
		return nil, xerrors.Errorf("Failed to query page list, Err: %w", err)
	}
	return common.BuildPageResp(res, count, request.PageParam), nil
}

// AllEnvList 所有环境Token列表
// @Summary  所有环境Token列表
// @Tags Token 管理
// @Accept  json
// @Produce  json
// @Param workspaceId path string true "工作空间ID"
// @Param data body models.TokenListReq true "issue params"
// @Router /apis/auth.engine.io/v1/workspaces/:workspaceId/tokens/allEnv/list [post]
// @Success 200 object models.DataResult[models.IPage[models.TokenBaseEntity]] "成功后返回"
// @Security Bearer
func (h *TokenHandler) AllEnvList(_ context.Context, c *common.CustomReqContext) (any, error) {
	request := &models.TokenListReq{}
	if err := c.Bind(request); err != nil {
		return nil, xerrors.Errorf("Failed to bind body: %w", err)
	}
	workspaceID := c.Param("workspaceId")
	pageParam := c.GetOrDefaultPageParam(&request.PageParam)
	count, res, err := h.TokenService.QueryPageList(workspaceID, *pageParam, request.OrderParam, request.QueryParam)
	if err != nil {
		return nil, xerrors.Errorf("Failed to query page list, Err: %w", err)
	}
	return common.BuildPageResp(res, count, request.PageParam), nil
}

// GetToken Token详情
// @Summary  Token详情
// @Tags Token 管理
// @Accept  json
// @Produce  json
// @Param workspaceId path string true "工作空间ID"
// @Param tokenId path string true "Token ID"
// @Router /apis/auth.engine.io/v1/workspaces/:workspaceId/tokens/:tokenId [get]
// @Success 200 object models.DataResult[models.TokenResp] "成功后返回"
// @Security Bearer
func (h *TokenHandler) GetToken(_ context.Context, c *common.CustomReqContext) (any, error) {
	tokenID := c.Param("tokenId")
	if tokenID == "" {
		return nil, common.NewCtrlError(400, xerrors.New("Token ID is required."))
	}
	tokenEntity, err := h.TokenService.Get(tokenID)
	if err != nil {
		return nil, xerrors.Errorf("Failed to get token, Err: %w", err)
	}
	return tokenEntity, nil
}

// CreateToken  Token创建
// @Summary  Token创建
// @Tags Token 管理
// @Accept  json
// @Produce  json
// @Param workspaceId path string true "工作空间ID"
// @Param data body models.CreateTokenReq true "issue params"
// @Router /apis/auth.engine.io/v1/workspaces/:workspaceId/tokens/add [post]
// @Success 200 object models.DataResult[dao.TokenEntity] "成功后返回"
// @Security Bearer
func (h *TokenHandler) CreateToken(_ context.Context, c *common.CustomReqContext) (any, error) {
	userInfo, exists := c.GetUser()
	if !exists {
		return nil, xerrors.Errorf("get userInfo failed")
	}
	request := &models.CreateTokenReq{}
	if err := c.Bind(request); err != nil {
		return nil, xerrors.Errorf("Failed to bind body: %w", err)
	}
	if request.AppScenarioName == "" {
		return nil, common.NewCtrlError(400, xerrors.New("App Scenario Name is required."))
	}
	if request.ModelName == "" {
		return nil, common.NewCtrlError(400, xerrors.New("Model Name is required."))
	}
	workspaceID := c.Param("workspaceId")
	if workspaceID == "" {
		return nil, common.NewCtrlError(400, xerrors.New("Workspace ID is required."))
	}
	request.WorkspaceID = workspaceID
	// 校验过期时间
	if request.ExpiredTime != nil {
		ok, expiredTime := utils.CheckDateTimeFormat(*request.ExpiredTime)
		if !ok {
			return nil, common.NewCtrlError(400, xerrors.New("ExpiredTime is invalid."))
		}
		if expiredTime.Before(time.Now()) {
			return nil, common.NewCtrlError(400, xerrors.New("ExpiredTime is before now."))
		}
		expiredTimeStr := expiredTime.Format(constants.TimeFormat)
		request.ExpiredTime = &expiredTimeStr
	}
	// policy_type: 策略类型，可以是 DAILY（每天）、WEEKLY（每周）、DATE_RANGE（日期范围）
	if request.EnableValidityPolicy {
		policyType := request.PolicyType
		if policyType == "" || len(request.ValidityPolicy) == 0 {
			return nil, common.NewCtrlError(400, xerrors.New("policyType and validityPolicy is required."))
		}
		// 当前只做 DAILY 类型
		if policyType != constants.DaliyPolicyType {
			return nil, common.NewCtrlError(400, xerrors.New("policyType is invalid."))
		}
		// 校验策略
		if err := CheckValidityPolicy(policyType, request.ValidityPolicy); err != nil {
			return nil, err
		}
	}
	if request.EnvName == "" {
		request.EnvName = config.CurrentEnvName
	}
	// check token exists
	exists, err := h.TokenService.CheckTokenExists(request.AppScenarioName, request.ModelName, request.EnvName, "")
	if err != nil {
		return nil, xerrors.Errorf("Failed to check token exists, Err: %w", err)
	}
	if exists {
		return nil, common.NewCtrlError(409, xerrors.New("Token with appScenarioName and modelName already exists."))
	}
	tokenEntity, err := h.TokenService.Create(request, userInfo, "")
	if err != nil {
		return nil, xerrors.Errorf("Failed to create token, Err: %w", err)
	}
	return tokenEntity, nil
}

// UpdateToken  Token更新
// @Summary  Token更新
// @Tags Token 管理
// @Accept  json
// @Produce  json
// @Param workspaceId path string true "工作空间ID"
// @Param data body models.UpdateTokenReq true "issue params"
// @Router /apis/auth.engine.io/v1/workspaces/:workspaceId/tokens/:tokenId [put]
// @Success 200 object models.DataResult[dao.TokenEntity] "成功后返回"
// @Security Bearer
func (h *TokenHandler) UpdateToken(_ context.Context, c *common.CustomReqContext) (any, error) {
	userInfo, exists := c.GetUser()
	if !exists {
		return nil, xerrors.Errorf("get userInfo failed")
	}
	request := &models.UpdateTokenReq{}
	if err := c.Bind(request); err != nil {
		return nil, xerrors.Errorf("Failed to bind body: %w", err)
	}
	tokenID := c.Param("tokenId")
	if tokenID == "" {
		return nil, common.NewCtrlError(400, xerrors.New("Token ID is required."))
	}
	if request.AppScenarioName == "" {
		return nil, common.NewCtrlError(400, xerrors.New("App Scenario Name is required."))
	}
	if request.ModelName == "" {
		return nil, common.NewCtrlError(400, xerrors.New("Model Name is required."))
	}
	// 校验过期时间
	if request.ExpiredTime != nil {
		ok, expiredTime := utils.CheckDateTimeFormat(*request.ExpiredTime)
		if !ok {
			return nil, common.NewCtrlError(400, xerrors.New("ExpiredTime is invalid."))
		}
		if expiredTime.Before(time.Now()) {
			return nil, common.NewCtrlError(400, xerrors.New("ExpiredTime is before now."))
		}
		expiredTimeStr := expiredTime.Format(constants.TimeFormat)
		request.ExpiredTime = &expiredTimeStr
	}
	// policy_type: 策略类型，可以是 DAILY（每天）、WEEKLY（每周）、DATE_RANGE（日期范围）
	if request.EnableValidityPolicy {
		policyType := request.PolicyType
		if policyType == "" || len(request.ValidityPolicy) == 0 {
			return nil, common.NewCtrlError(400, xerrors.New("policyType and validityPolicy is required."))
		}
		// 当前只做 DAILY 类型
		if policyType != constants.DaliyPolicyType {
			return nil, common.NewCtrlError(400, xerrors.New("policyType is invalid."))
		}
		// 校验策略
		if err := CheckValidityPolicy(policyType, request.ValidityPolicy); err != nil {
			return nil, err
		}
	}
	tokenQuery, err := h.TokenService.FindTokenEntity(tokenID)
	if err != nil {
		return nil, xerrors.Errorf("failed to get token: %v", err)
	}
	// check token exists
	exists, err = h.TokenService.CheckTokenExists(request.AppScenarioName, request.ModelName, tokenQuery.EnvName, tokenID)
	if err != nil {
		return nil, xerrors.Errorf("Failed to check token exists, Err: %w", err)
	}
	if exists {
		return nil, common.NewCtrlError(409, xerrors.New("Token with appScenarioName and modelName already exists."))
	}
	tokenEntity, err := h.TokenService.Update(tokenID, request, tokenQuery, userInfo)
	if err != nil {
		return nil, xerrors.Errorf("Failed to update token, Err: %w", err)
	}
	return tokenEntity, nil
}

// DeleteToken  Token删除
// @Summary  Token删除
// @Tags Token 管理
// @Accept  json
// @Produce  json
// @Param workspaceId path string true "工作空间ID"
// @Router /apis/auth.engine.io/v1/workspaces/:workspaceId/tokens/:tokenId [delete]
// @Success 200 object models.DataResult[string] "成功后返回"
// @Security Bearer
func (h *TokenHandler) DeleteToken(_ context.Context, c *common.CustomReqContext) (any, error) {
	tokenID := c.Param("tokenId")
	if tokenID == "" {
		return nil, common.NewCtrlError(400, xerrors.New("Token ID is required."))
	}
	err := h.TokenService.Delete(tokenID)
	if err != nil {
		return nil, xerrors.Errorf("Failed to delete token exists, Err: %w", err)
	}
	return "deelte token success", nil
}

// ExportTokenFile  Token导出
// @Summary  Token导出
// @Tags Token 管理
// @Accept  json
// @Produce  json
// @Param workspaceId path string true "工作空间ID"
// @Param tokenId path string true "Token ID"
// @Router /apis/auth.engine.io/v1/workspaces/:workspaceId/tokens/:tokenId/export [post]
// @Success 200 object models.DataResult[string] "成功后返回"
// @Security Bearer
func (h *TokenHandler) ExportTokenFile(ctx context.Context, c *common.CustomReqContext) (any, error) {
	request := &models.ExportTokenFileReq{}
	if err := c.Bind(request); err != nil {
		return nil, xerrors.Errorf("Failed to bind body: %w", err)
	}
	if request.EnvName == "" {
		return nil, common.NewCtrlError(400, xerrors.New("EnvName is required."))
	}
	tokenID := c.Param("tokenId")
	if tokenID == "" {
		return nil, common.NewCtrlError(400, xerrors.New("Token ID is required."))
	}
	tokenEntity, err := h.TokenService.Get(tokenID)
	if err != nil {
		return nil, xerrors.Errorf("Failed to get token, Err: %w", err)
	}
	// 将token转换为JSON
	jsonData, err := json.Marshal(tokenEntity)
	if err != nil {
		return nil, xerrors.Errorf("Failed to marshal token, Err: %w", err)
	}
	// 加密JSON数据
	encryptedData, err := utils.EncryptData(jsonData)
	if err != nil {
		return nil, xerrors.Errorf("Failed to encrypt data, Err: %w", err)
	}
	// 设置文件名,格式为：日期-appScenarioName-modelName-envName.txt
	filename := fmt.Sprintf(
		"%s-%s-%s-%s.txt",
		time.Now().Format("2006-01-02"), tokenEntity.AppScenarioName, tokenEntity.ModelName, tokenEntity.EnvName,
	)
	// 设置响应头
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	// Write the encrypted data to the response
	// 返回加密后的二进制数据
	c.Data(http.StatusOK, "application/octet-stream", encryptedData)
	return nil, nil
}

// ImportTokenFile 导入Token文件
// @Summary  导入Token文件
// @Tags Token 管理
// @Accept  multipart/form-data
// @Produce json
// @Param files formData file true "文件"
// @Router  /apis/auth.engine.io/v1/workspaces/:workspaceId/tokens/import [POST]
// @Success 200 object models.DataResult[string] "成功后返回"
func (h *TokenHandler) ImportTokenFile(ctx context.Context, c *common.CustomReqContext) (any, error) {
	userInfo, exists := c.GetUser()
	if !exists {
		return nil, xerrors.Errorf("get userInfo failed")
	}
	workspaceID := c.Param("workspaceId")
	file, _ := c.FormFile("file")
	if file.Filename == "" {
		return nil, common.NewCtrlError(http.StatusBadRequest, xerrors.New("文件名为空,不允许上传！"))
	}
	if file.Size == 0 {
		return nil, common.NewCtrlError(http.StatusBadRequest, xerrors.New(file.Filename+"文件为空,不允许上传！"))
	}
	if file.Size > constants.ImportTokenFileMaxSize {
		return nil, common.NewCtrlError(http.StatusBadRequest, xerrors.New(file.Filename+"文件过大,只允许上传512MB以内的文件！"))
	}
	suffix := filepath.Ext(file.Filename)
	if suffix != ".txt" {
		return nil, common.NewCtrlError(http.StatusBadRequest, xerrors.Errorf("不支持的文件类型: %s", suffix))
	}
	openFile, err := file.Open()
	if err != nil {
		return nil, xerrors.Errorf("Failed to open file, Err: %w", err)
	}
	defer openFile.Close()
	// 读取文件内容
	fileBytes, err := io.ReadAll(openFile)
	if err != nil {
		return nil, xerrors.Errorf("Failed to read file, Err: %w", err)
	}
	// 解密数据
	decryptedData, err := utils.DecryptData(fileBytes)
	if err != nil {
		return nil, xerrors.Errorf("Failed to decrypt data, Err: %w", err)
	}
	// 解析JSON数据
	var tokenEntity *models.TokenResp
	err = json.Unmarshal(decryptedData, &tokenEntity)
	if err != nil {
		return nil, xerrors.Errorf("Failed to unmarshal token, Err: %w", err)
	}
	// 校验参数
	request := &models.CreateTokenReq{
		WorkspaceID:          workspaceID,
		AppScenarioName:      tokenEntity.AppScenarioName,
		ModelName:            tokenEntity.ModelName,
		ExpiredTime:          tokenEntity.ExpiredTime,
		EnvName:              tokenEntity.EnvName,
		MaxConcurrency:       tokenEntity.MaxConcurrency,
		EnableValidityPolicy: tokenEntity.EnableValidityPolicy,
		PolicyType:           tokenEntity.PolicyType,
		ValidityPolicy:       tokenEntity.ValidityPolicy,
	}
	if request.AppScenarioName == "" {
		return nil, common.NewCtrlError(400, xerrors.New("App Scenario Name is required."))
	}
	if request.ModelName == "" {
		return nil, common.NewCtrlError(400, xerrors.New("Model Name is required."))
	}
	if request.EnvName != config.CurrentEnvName {
		return nil, common.NewCtrlError(400, xerrors.New("Env Name is not match."))
	}
	if request.EnableValidityPolicy {
		policyType := request.PolicyType
		if policyType == "" || len(request.ValidityPolicy) == 0 {
			return nil, common.NewCtrlError(400, xerrors.New("policyType and validityPolicy is required."))
		}
		// 当前只做 DAILY 类型
		if policyType != constants.DaliyPolicyType {
			return nil, common.NewCtrlError(400, xerrors.New("policyType is invalid."))
		}
		if err := CheckValidityPolicy(policyType, request.ValidityPolicy); err != nil {
			return nil, err
		}
	}
	// check token exists
	exists, err = h.TokenService.CheckTokenExists(request.AppScenarioName, request.ModelName, request.EnvName, "")
	if err != nil {
		return nil, xerrors.Errorf("Failed to check token exists, Err: %w", err)
	}
	if exists {
		return nil, common.NewCtrlError(409, xerrors.New("Token with appScenarioName and modelName already exists."))
	}
	tokenRes, err := h.TokenService.Create(request, userInfo, tokenEntity.Token)
	if err != nil {
		return nil, xerrors.Errorf("Failed to create token, Err: %w", err)
	}
	return tokenRes, nil
}
