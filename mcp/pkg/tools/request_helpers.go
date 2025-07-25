package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GetRequiredStringArgumentFromCallToolRequest(request mcp.CallToolRequest, argument string) (string, error) {
	value, ok := request.GetArguments()[argument].(string)
	if !ok {
		err := status.Errorf(codes.InvalidArgument, "%s not defined", argument)
		return "", err
	}
	if value == "" {
		err := status.Errorf(codes.InvalidArgument, "%s not defined", argument)
		return "", err
	}
	return value, nil
}

func GetStringArgumentFromCallToolRequest(request mcp.CallToolRequest, argument string) string {
	value, ok := request.GetArguments()[argument].(string)
	if !ok {
		return ""
	}
	return value
}

func GetBooleanArgumentFromCallToolRequest(request mcp.CallToolRequest, argument string) bool {
	value, ok := request.GetArguments()[argument].(bool)
	if !ok {
		return false
	}
	return value
}

