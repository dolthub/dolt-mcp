package integration_tests

import (
	"context"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

type TestClient struct {
	c *client.Client
}

type TestClientOption func(tc *TestClient)

func WithNotificationHandler(handler func(notification mcp.JSONRPCNotification)) TestClientOption {
	return func(tc *TestClient) {
		tc.c.OnNotification(handler)
	}
}

func NewMCPHTTPTestClient(httpURL string, opts ...TestClientOption) (*TestClient, error) {
	// Create HTTP transport
	httpTransport, err := transport.NewStreamableHTTP(httpURL)
	// NOTE: the default streamableHTTP transport is not 100% identical to the stdio client.
	// By default, it could not receive global notifications (e.g. toolListChanged).
	// You need to enable the `WithContinuousListening()` option to establish a long-live connection,
	// and receive the notifications any time the server sends them.
	//
	//   httpTransport, err := transport.NewStreamableHTTP(*httpURL, transport.WithContinuousListening())
	if err != nil {
		return nil, err
	}

	tc := &TestClient{c: client.NewClient(httpTransport)}

	for _, opt := range opts {
		opt(tc)
	}

	return tc, nil
}

func (c *TestClient) Initialize(ctx context.Context) (*mcp.InitializeResult, error) {
	initializeParams := mcp.InitializeParams{
		ClientInfo: mcp.Implementation{
			Name:    "Dolt MCP test client",
			Version: "1.0.0",
		},
		Capabilities:    mcp.ClientCapabilities{},
		ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
	}
	request := mcp.InitializeRequest{
		Params: initializeParams,
	}
	return c.c.Initialize(ctx, request)
}

func (c *TestClient) ListTools(ctx context.Context) (*mcp.ListToolsResult, error) {
	return c.c.ListTools(ctx, mcp.ListToolsRequest{})
}

func (c *TestClient) CallTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return c.c.CallTool(ctx, request)
}
