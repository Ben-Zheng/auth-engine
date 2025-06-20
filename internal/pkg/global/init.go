package global

import (
	"time"

	"ghippo.io/api/wssdk/v1alpha1"
	"ghippo.io/api/wssdk/v1alpha1/types"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/logger/accesslog"
	"github.com/hertz-contrib/logger/slog"
	"golang.org/x/xerrors"
	"gorm.io/gorm"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/auth-engine/config"
	"github.com/auth-engine/internal/pkg/dao"
	"github.com/auth-engine/internal/pkg/global/servers"
	"github.com/auth-engine/internal/pkg/routers/routeinit"
	"github.com/auth-engine/pkg/constants"
	"github.com/auth-engine/pkg/log"
)

// 配置文件初始化
func ConfigInit() (*config.AppConfig, error) {
	return config.InitConfig()
}

// 日志初始化
func LogInit(appConfig *config.AppConfig) (*slog.Logger, error) {
	return log.HertzLogInit(appConfig.Level)
}

// 数据库初始化
func DBInit(appConfig *config.AppConfig) (*gorm.DB, error) {
	return dao.Connect(appConfig)
}

// GhippoAuthInit 初始化Ghippo认证
func GhippoAuthInit(appConfig *config.AppConfig) (types.Interface, error) {
	// 从配置文件中构建restConfig
	restConfig, err := clientcmd.BuildConfigFromFlags("", appConfig.Kube.ConfigPath)
	if err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}
	// 创建WSSDKClient
	sdk, err := v1alpha1.NewWSSDKClient(restConfig, constants.ProductName, time.Second, []types.ResourceType{types.ResourceTypeCluster, types.ResourceTypeSharedCluster, types.ResourceTypeClusterNamespace})
	if err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}
	// 启动sdk
	err = sdk.Start()
	if err != nil {
		return nil, xerrors.Errorf("GhippoAuthInit failed: %w", err)
	}
	return sdk, nil
}

// 服务初始化
func ServerOptInit(appConfig *config.AppConfig) (*servers.ServerOpt, error) {
	return servers.NewServerOpt(appConfig.Server, appConfig.Tracer)
}

// hertz 框架初始化
func WebHertzInit(opt *servers.ServerOpt) *server.Hertz {
	web := server.New(opt.Opts...)
	web.Use(accesslog.New(accesslog.WithFormat("[${time}] ${status} - ${latency} ${method} ${path} ${queryParams}")))
	web.Use(opt.AppHandler...)
	return web
}

// WebServer 初始化
func WebServerInit(appConfig *config.AppConfig, ghippoSdk types.Interface, routeInit *routeinit.RouteInit, hertzServer *server.Hertz) (*servers.WebServer, error) {
	return servers.NewWebServer(hertzServer, ghippoSdk, appConfig.Tracer, &routeinit.RouteInit{})
}
