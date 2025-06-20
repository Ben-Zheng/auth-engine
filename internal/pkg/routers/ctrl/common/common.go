package common

import (
	"context"
	"math"
	"net/http"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"

	"github.com/auth-engine/internal/pkg/dao"
	"github.com/auth-engine/internal/pkg/models"
)

// CustomHandler 自定义的handler处理器
// result: 返回的结果,如果接口正常返回，会把result build为统一包装好的 model.DataResult 返回
// err: 如果接口返回错误，会把err build为统一包装好的 model.DataResult 返回
// 如果result和error都为nil的话，不写任何数据
type CustomHandler = func(ctx context.Context, reqContext *CustomReqContext) (result any, err error)

func Handle(handler CustomHandler) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		result, err := handler(ctx, &CustomReqContext{RequestContext: c})
		if err != nil {
			hlog.Errorf("An error occurred: %+v\n", err)
			resp := BuildErrResp(err)
			c.JSON(resp.Code, resp)
		} else {
			// result为空的话，不写任何数据
			if result != nil {
				resp := BuildSuccessResp(result)
				c.JSON(http.StatusOK, resp)
			}
		}
	}
}

func BuildErrResp(err error) *models.DataResult[any] {
	code := http.StatusInternalServerError
	if ctrlErr, ok := err.(*Error); ok {
		code = ctrlErr.Code
	}
	return &models.DataResult[any]{
		Success:   false,
		Message:   err.Error(),
		Code:      code,
		Result:    "",
		Timestamp: time.Now().UnixMilli(),
	}
}

func BuildSuccessResp(result any) *models.DataResult[any] {
	return &models.DataResult[any]{
		Success:   true,
		Message:   "",
		Code:      http.StatusOK,
		Result:    result,
		Timestamp: time.Now().UnixMilli(),
	}
}

func BuildPageResp[T any](records []T, count int64, pageParam dao.PageParam) models.IPage[T] {
	var pages int
	if pageParam.PageSize == 0 {
		pages = 0
	} else {
		pages = int(math.Ceil(float64(count) / float64(pageParam.PageSize)))
	}

	return models.IPage[T]{
		Records:     records,
		Total:       count,
		SearchCount: true,
		Current:     pageParam.PageNum,
		Pages:       pages,
		Size:        pageParam.PageSize,
	}
}
