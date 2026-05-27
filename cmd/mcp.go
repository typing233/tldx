package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
	"tldx/internal/checker"
	"tldx/internal/domain"
	"tldx/internal/rdap"
)

func runMCP(_ []string) int {
	fmt.Fprintf(os.Stderr, "tldx MCP server starting (stdio transport)...\n")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		var req jsonRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			writeJSONRPCError(req.ID, -32700, "Parse error")
			continue
		}

		switch req.Method {
		case "initialize":
			writeJSONRPCResult(req.ID, map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{},
				},
				"serverInfo": map[string]interface{}{
					"name":    "tldx",
					"version": version,
				},
			})
		case "notifications/initialized":
			// No response needed
		case "tools/list":
			writeJSONRPCResult(req.ID, map[string]interface{}{
				"tools": []map[string]interface{}{
					{
						"name":        "check_domains",
						"description": "Check domain name availability via RDAP",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"keywords": map[string]interface{}{
									"type":        "array",
									"items":       map[string]string{"type": "string"},
									"description": "Keywords to check",
								},
								"tlds": map[string]interface{}{
									"type":        "array",
									"items":       map[string]string{"type": "string"},
									"description": "TLDs to combine with keywords",
								},
								"prefixes": map[string]interface{}{
									"type":        "array",
									"items":       map[string]string{"type": "string"},
									"description": "Prefixes to prepend to keywords",
								},
								"suffixes": map[string]interface{}{
									"type":        "array",
									"items":       map[string]string{"type": "string"},
									"description": "Suffixes to append to keywords",
								},
							},
							"required": []string{"keywords"},
						},
					},
				},
			})
		case "tools/call":
			handleToolCall(req)
		default:
			writeJSONRPCError(req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method))
		}
	}

	return 0
}

type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func writeJSONRPCResult(id interface{}, result interface{}) {
	resp := jsonRPCResponse{JSONRPC: "2.0", ID: id, Result: result}
	data, _ := json.Marshal(resp)
	fmt.Fprintf(os.Stdout, "%s\n", data)
}

func writeJSONRPCError(id interface{}, code int, message string) {
	resp := jsonRPCResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: message}}
	data, _ := json.Marshal(resp)
	fmt.Fprintf(os.Stdout, "%s\n", data)
}

func handleToolCall(req jsonRPCRequest) {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		writeJSONRPCError(req.ID, -32602, "Invalid params")
		return
	}

	if params.Name != "check_domains" {
		writeJSONRPCError(req.ID, -32602, fmt.Sprintf("Unknown tool: %s", params.Name))
		return
	}

	var toolArgs struct {
		Keywords []string `json:"keywords"`
		TLDs     []string `json:"tlds"`
		Prefixes []string `json:"prefixes"`
		Suffixes []string `json:"suffixes"`
	}
	if err := json.Unmarshal(params.Arguments, &toolArgs); err != nil {
		writeJSONRPCError(req.ID, -32602, "Invalid tool arguments")
		return
	}

	if len(toolArgs.Keywords) == 0 {
		writeJSONRPCResult(req.ID, map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": "Error: no keywords provided"},
			},
			"isError": true,
		})
		return
	}

	tlds := toolArgs.TLDs
	if len(tlds) == 0 {
		tlds = []string{"com"}
	}

	candidates := domain.Generate(domain.GenerateConfig{
		Keywords: toolArgs.Keywords,
		Prefixes: toolArgs.Prefixes,
		Suffixes: toolArgs.Suffixes,
		TLDs:     tlds,
	})

	if len(candidates) == 0 {
		writeJSONRPCResult(req.ID, map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": "No candidate domains generated"},
			},
		})
		return
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	bootstrap := rdap.NewBootstrap(ctx, httpClient)
	rdapClient := rdap.NewClient(httpClient, bootstrap, 1)
	pool := checker.NewPool(rdapClient, 3)

	results := pool.Run(ctx, candidates)

	type domainResult struct {
		Domain    string `json:"domain"`
		Available bool   `json:"available"`
		Error     string `json:"error,omitempty"`
	}

	var checked []domainResult
	var availableList []string
	for r := range results {
		dr := domainResult{Domain: r.Domain, Available: r.Available}
		if r.Error != nil {
			dr.Error = r.Error.Error()
		} else if r.Available {
			availableList = append(availableList, r.Domain)
		}
		checked = append(checked, dr)
	}

	output, _ := json.MarshalIndent(checked, "", "  ")
	summary := fmt.Sprintf("Checked %d domains, %d available.\n\n%s", len(checked), len(availableList), string(output))

	writeJSONRPCResult(req.ID, map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "text", "text": summary},
		},
	})
}
