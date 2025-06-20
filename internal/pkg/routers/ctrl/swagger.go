package ctrl

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/swagger"
	swaggerFiles "github.com/swaggo/files"
)

type SwaggerCtrl struct{}

func NewSwaggerCtrl(r *server.Hertz) *SwaggerCtrl {
	r.GET("/swagger/*any", swagger.WrapHandler(swaggerFiles.Handler, swagger.DefaultModelsExpandDepth(10)))
	return &SwaggerCtrl{}
}
