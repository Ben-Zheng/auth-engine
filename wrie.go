//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/auth-engine/internal/pkg/dao"
	"github.com/auth-engine/internal/pkg/global"
	"github.com/auth-engine/internal/pkg/global/servers"
	"github.com/auth-engine/internal/pkg/routers/ctrl"
	"github.com/auth-engine/internal/pkg/routers/routeinit"
	"github.com/auth-engine/internal/pkg/services"
)

var initSet = wire.NewSet(
	global.ConfigInit,
	global.LogInit,
	global.DBInit,
	global.GhippoAuthInit,
	global.ServerOptInit,
	global.WebHertzInit,
	global.WebServerInit,
	routeinit.NewInit,
)

var controllerSet = wire.NewSet(
	ctrl.NewSwaggerCtrl,
	ctrl.NewTokenHandler,
	ctrl.NewWorkspaceHandler,
)

var serviceSet = wire.NewSet(
	services.NewClientService,
	services.NewTokenService,
)

var daoSet = wire.NewSet(
	dao.NewTokenDao,
	dao.NewTokenValidityPolicyDao,
)

func InitProject() (*servers.WebServer, error) {
	panic(wire.Build(initSet, controllerSet, serviceSet, daoSet))
}
