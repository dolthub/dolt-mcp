package tools

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestShowProcesslistSQL(t *testing.T) {
	if got := showProcesslistSQL(false); got != "SHOW PROCESSLIST;" {
		t.Fatalf("full=false: got %q", got)
	}
	if got := showProcesslistSQL(true); got != "SHOW FULL PROCESSLIST;" {
		t.Fatalf("full=true: got %q", got)
	}
}

func TestKillProcessSQL(t *testing.T) {
	if got := killProcessSQL(123, false); got != "KILL 123;" {
		t.Fatalf("kill_query=false: got %q", got)
	}
	if got := killProcessSQL(123, true); got != "KILL QUERY 123;" {
		t.Fatalf("kill_query=true: got %q", got)
	}
}

func TestGetRequiredPositiveProcessIDFromCallToolRequest(t *testing.T) {
	t.Run("accepts int-like values", func(t *testing.T) {
		cases := []struct {
			name string
			val  any
			want int64
		}{
			{"int", int(7), 7},
			// JSON numbers often come through as float64.
			{"float64 integral", float64(7), 7},
			// With RequireInt, fractional float64 values are truncated.
			{"float64 fractional truncates", float64(7.9), 7},
			// Strings are parsed with strconv.Atoi by RequireInt.
			{"string int", "7", 7},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				req := mcp.CallToolRequest{
					Params: mcp.CallToolParams{
						Arguments: map[string]any{KillProcessIDArgName: tc.val},
					},
				}
				got, err := getRequiredPositiveProcessIDFromCallToolRequest(req)
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				if got != tc.want {
					t.Fatalf("got %d, want %d", got, tc.want)
				}
			})
		}
	})

	t.Run("rejects invalid values", func(t *testing.T) {
		cases := []any{
			nil,
			int(0),
			int(-1),
			float64(0),
			float64(-2),
			"abc",
		}

		for _, val := range cases {
			t.Run("", func(t *testing.T) {
				req := mcp.CallToolRequest{
					Params: mcp.CallToolParams{
						Arguments: map[string]any{KillProcessIDArgName: val},
					},
				}
				_, err := getRequiredPositiveProcessIDFromCallToolRequest(req)
				if err == nil {
					t.Fatalf("expected err, got nil")
				}
			})
		}
	})

	t.Run("missing argument", func(t *testing.T) {
		req := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]any{},
			},
		}
		_, err := getRequiredPositiveProcessIDFromCallToolRequest(req)
		if err == nil {
			t.Fatalf("expected err, got nil")
		}
	})
}
