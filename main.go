package main

import (
	"fmt"

	_ "github.com/auth-engine/docs"
)

//go:generate go run -mod=vendor github.com/swaggo/swag/cmd/swag init --parseDependency --parseInternal

// @title           Engine
// @version         1.0
// @description     Engine
// @BasePath  		/api/auth.engine.io

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description 以'Bearer'开头，后面跟Token
func main() {
	project, err := InitProject()
	if err != nil {
		panic(fmt.Sprintf("init error: %+v\n", err))
	}
	project.Run()
}
