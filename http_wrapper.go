package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// HTTPWrapper wraps the MCP server with an HTTP interface
type HTTPWrapper struct {
	binaryPath string
	mu         sync.Mutex
	client     *mcp.Client
	session    *mcp.ClientSession
}

// NewHTTPWrapper creates a new HTTP wrapper
func NewHTTPWrapper(binaryPath string) *HTTPWrapper {
	return &HTTPWrapper{
		binaryPath: binaryPath,
	}
}

// ensureSession ensures we have an active MCP session
func (h *HTTPWrapper) ensureSession(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.session != nil {
		return nil
	}

	if h.client == nil {
		h.client = mcp.NewClient(&mcp.Implementation{Name: "http-wrapper", Version: "1.0.0"}, nil)
	}

	transport := &mcp.CommandTransport{Command: exec.Command(h.binaryPath)}
	session, err := h.client.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	h.session = session
	return nil
}

// JSONRPCRequest represents a JSON-RPC request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// handleRequest handles an HTTP request and forwards it to MCP
func (h *HTTPWrapper) handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Ensure we have a session
	if err := h.ensureSession(ctx); err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to MCP server: %v", err), http.StatusInternalServerError)
		return
	}

	// Read request body
	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Handle different methods
	var result interface{}
	var err error

	switch req.Method {
	case "initialize":
		// Already initialized, return success
		result = map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"serverInfo": map[string]interface{}{
				"name":    "diningbot",
				"version": "1.0.0",
			},
		}

	case "tools/list":
		h.mu.Lock()
		session := h.session
		h.mu.Unlock()

		listResult, listErr := session.ListTools(ctx, &mcp.ListToolsParams{})
		if listErr != nil {
			err = listErr
		} else {
			result = listResult
		}

	case "tools/call":
		h.mu.Lock()
		session := h.session
		h.mu.Unlock()

		// Extract params
		paramsMap, ok := req.Params.(map[string]interface{})
		if !ok {
			err = fmt.Errorf("invalid params")
			break
		}

		toolName, _ := paramsMap["name"].(string)
		arguments, _ := paramsMap["arguments"].(map[string]interface{})

		callParams := &mcp.CallToolParams{
			Name:      toolName,
			Arguments: arguments,
		}

		callResult, callErr := session.CallTool(ctx, callParams)
		if callErr != nil {
			err = callErr
		} else {
			result = callResult
		}

	default:
		err = fmt.Errorf("unknown method: %s", req.Method)
	}

	// Build response
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	if err != nil {
		resp.Error = map[string]interface{}{
			"code":    -32000,
			"message": err.Error(),
		}
	} else {
		resp.Result = result
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	// Determine binary path
	binaryPath := "./diningbot"
	if _, err := exec.LookPath(binaryPath); err != nil {
		// Try to build it
		if buildErr := exec.Command("go", "build", "-o", binaryPath).Run(); buildErr != nil {
			log.Fatalf("Failed to build diningbot: %v", buildErr)
		}
	}

	wrapper := NewHTTPWrapper(binaryPath)

	http.HandleFunc("/mcp", wrapper.handleRequest)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><body>
			<h1>MCP HTTP Wrapper</h1>
			<p>Send JSON-RPC requests to <code>/mcp</code></p>
			<p>Example:</p>
			<pre>curl -X POST http://localhost:8080/mcp -H "Content-Type: application/json" -d '{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/list"
}'</pre>
		</body></html>`)
	})

	port := ":8080"
	log.Printf("HTTP wrapper server starting on http://localhost%s", port)
	log.Printf("Send JSON-RPC requests to http://localhost%s/mcp", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
