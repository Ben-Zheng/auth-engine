// tools.go
//go:build tools
// +build tools

package tools

import (
	_ "github.com/google/wire"
	_ "github.com/swaggo/swag/cmd/swag"
)
