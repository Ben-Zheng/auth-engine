package routeinit

import (
	"github.com/auth-engine/internal/pkg/routers/ctrl"
)

type RouteInit struct{}

// NewInit
func NewInit(
	_ *ctrl.SwaggerCtrl,
	_ *ctrl.TokenHandler,
	_ *ctrl.WorkspaceHandler,
) *RouteInit {
	return &RouteInit{}
}
