package ctrl

import (
	"context"
	"strconv"

	"github.com/auth-engine/config"
	"github.com/auth-engine/internal/pkg/models"
	"github.com/auth-engine/internal/pkg/routers/ctrl/common"
	"github.com/auth-engine/internal/pkg/services"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"golang.org/x/xerrors"
)

type WorkspaceHandler struct {
	IClientService services.IClientService
	AppConfig      *config.AppConfig
}

func NewWorkspaceHandler(iClientService services.IClientService, r *server.Hertz) *WorkspaceHandler {
	handler := &WorkspaceHandler{IClientService: iClientService}
	router := r.Group("/apis/auth.engine.io/v1/auth")
	router.Use(common.VerifyAuthorization())
	router.GET("/workspaces/list", common.Handle(handler.ListVisibleWorkspaces))
	return handler
}

// ListVisibleWorkspaces 用于用户列出可见的工作空间
// @Summary  用于用户列出可见的工作空间
// @Tags WorkSpace 管理
// @Accept  json
// @Produce  json
// @Router /apis/auth.engine.io/v1/auth/workspaces/list [get]
// @Success 200 object models.DataResult[[]models.ListVisibleWorkspacesResponse] "成功后返回"
// @Security Bearer
func (h *WorkspaceHandler) ListVisibleWorkspaces(ctx context.Context, c *common.CustomReqContext) (any, error) {
	result := make([]models.ListVisibleWorkspacesResponse, 0)
	wss, err := h.IClientService.ListVisibleWorkspaces(ctx)
	if err != nil {
		hlog.Errorf("Failed to list visible workspaces, Err: %v", err)
		return nil, xerrors.Errorf("Failed to list visible workspaces, Err: %w", err)
	}
	// 遍历获取到的可见工作空间
	for i := range wss {
		ws := wss[i]
		item := models.ListVisibleWorkspacesResponse{
			WorkspaceID: strconv.Itoa(ws.WorkspaceId),
			Alias:       ws.Alias,
		}
		result = append(result, item)
	}
	return result, nil
}
