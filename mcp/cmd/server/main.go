package main

import (
	"context"

	"github.com/dolthub/dolt-mcp/mcp/pkg"
	"github.com/dolthub/dolt-mcp/mcp/pkg/db"
)

func main() {
	port := 8080
	config := db.Config{
		Host: "0.0.0.0",
		Port: 3306,
		User: "root",
		Password: "",
		DatabaseName: "test",
	}
	srv, err := pkg.NewMCPHTTPServer(
		config,
		port,
		pkg.WithToolSet(&pkg.PrimitiveToolSetV1{}))
	if err != nil {
		panic(err)
	}
	srv.ListenAndServe(context.Background())
}

