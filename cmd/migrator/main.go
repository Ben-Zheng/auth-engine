package main

import (
	"fmt"
	"io"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"

	"github.com/auth-engine/internal/pkg/dao"
)

func main() {
	stmts, err := gormschema.New("mysql").Load(
		&dao.TokenEntity{},
		&dao.TokenValidityPolicy{},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load gorm schema: %v\n", err)
		os.Exit(1)
	}
	io.WriteString(os.Stdout, stmts)
}
