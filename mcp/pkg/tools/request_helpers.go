package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	DatabaseCallToolArgumentName = "database"
)

func GetDatabaseArgumentFromCallToolRequest(request mcp.CallToolRequest) (string, error) {
	database, ok := request.GetArguments()[DatabaseCallToolArgumentName].(string)
	if !ok {
		err := status.Errorf(codes.InvalidArgument, "%s not defined", DatabaseCallToolArgumentName)
		return "", err
	}
	if database == "" {
		err := status.Errorf(codes.InvalidArgument, "%s not defined", DatabaseCallToolArgumentName)
		return "", err
	}
	return database, nil
}

