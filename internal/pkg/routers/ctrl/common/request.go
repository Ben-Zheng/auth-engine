package common

import (
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/auth-engine/internal/pkg/dao"
	"github.com/auth-engine/internal/pkg/models"
)

type CustomReqContext struct {
	*app.RequestContext
}

func (c *CustomReqContext) Params2(name1, name2 string) (string, string) {
	return c.Param(name1), c.Param(name2)
}

func (c *CustomReqContext) Params3(name1, name2, name3 string) (string, string, string) {
	return c.Param(name1), c.Param(name2), c.Param(name3)
}

func (c *CustomReqContext) Params4(name1, name2, name3, name4 string) (string, string, string, string) {
	return c.Param(name1), c.Param(name2), c.Param(name3), c.Param(name4)
}

// ParamInt 获取Param并转为int
func (c *CustomReqContext) ParamInt(name string) (int, error) {
	paramStr := c.Param(name)
	return strconv.Atoi(paramStr)
}

func (c *CustomReqContext) GetOrDefaultPageParam(page *dao.PageParam) *dao.PageParam {
	if page == nil {
		return &dao.PageParam{
			PageSize: 10,
			PageNum:  1,
		}
	}
	if page.PageSize == 0 {
		page.PageSize = 10
	}
	if page.PageNum == 0 {
		page.PageNum = 1
	}
	return page
}

func (c *CustomReqContext) GetUser() (*models.UserInfo, bool) {
	// get AuthResult from ctx
	authResult, exists := c.Get(AuthResultKey)
	if !exists {
		return nil, false
	}
	// assert authResult as AuthResult struct
	result, ok := authResult.(models.AuthResult)
	if !ok {
		return nil, false
	}
	userInfo := &models.UserInfo{
		ID:       result.Sub,
		Username: result.PreferredUsername,
	}
	return userInfo, true
}
