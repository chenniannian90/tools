package confmcp

// JSONRPCMessage represents a JSON-RPC 2.0 message
type JSONRPCMessage struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id,omitempty"`
	Method  string                 `json:"method,omitempty"`
	Params  interface{}            `json:"params,omitempty"`
	Result  interface{}            `json:"result,omitempty"`
	Error   *JSONRPCError          `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Capabilities 定义 MCP 服务器能力（仅支持 tools）
type Capabilities struct {
	Tools bool `json:"tools"`
}

// ServerInfo 服务器信息
type ServerInfo struct {
	Name         string       `json:"name"`
	Version      string       `json:"version"`
	Protocol     string       `json:"protocol"`
	Address      string       `json:"address"`
	Capabilities Capabilities `json:"capabilities"`
}
