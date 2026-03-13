package acp

//go:generate sh -c "cd internal/cmd/schema && go run . gen -config ../../../.schema.yaml"

import (
	"context"
	"encoding/json"
)

// Agent represents the interface that agents must implement to handle client requests.
//
// This interface defines all the stable methods that an agent can receive from a client
// according to the Agent Client Protocol specification.
//
// Optional capabilities can be advertised by implementing additional interfaces:
//   - [SessionForker] for session/fork (unstable)
//   - [SessionResumer] for session/resume (unstable)
//   - [SessionCloser] for session/close (unstable)
//   - [ModelSetter] for session/set_model (unstable)
//   - [ExtMethodHandler] for custom extension methods
//   - [ExtNotificationHandler] for custom extension notifications
//
// See protocol docs: [Agent](https://agentclientprotocol.com/protocol/overview#agent)
type Agent interface {
	// Initialize establishes the connection and negotiates capabilities.
	//
	// This is the first method called after connection establishment.
	// The agent should return its capabilities and supported protocol version.
	//
	// See protocol docs: [Initialization](https://agentclientprotocol.com/protocol/initialization)
	Initialize(ctx context.Context, params *InitializeRequest) (*InitializeResponse, error)

	// Authenticate handles user authentication using the specified method.
	//
	// Called when the client needs to authenticate the user with the agent.
	//
	// See protocol docs: [Authentication](https://agentclientprotocol.com/protocol/authentication)
	Authenticate(ctx context.Context, params *AuthenticateRequest) (*AuthenticateResponse, error)

	// NewSession creates a new conversation session.
	//
	// Sets up a new session with the specified working directory and MCP servers.
	//
	// See protocol docs: [Creating a Session](https://agentclientprotocol.com/protocol/session-setup#creating-a-session)
	NewSession(ctx context.Context, params *NewSessionRequest) (*NewSessionResponse, error)

	// LoadSession loads an existing conversation session.
	//
	// Only called if the agent advertises the loadSession capability.
	//
	// See protocol docs: [Loading Sessions](https://agentclientprotocol.com/protocol/session-setup#loading-sessions)
	LoadSession(ctx context.Context, params *LoadSessionRequest) (*LoadSessionResponse, error)

	// ListSessions lists available sessions.
	//
	// Only called if the agent advertises the sessionCapabilities.list capability.
	//
	// See protocol docs: [Listing Sessions](https://agentclientprotocol.com/protocol/session-setup#listing-sessions)
	ListSessions(ctx context.Context, params *ListSessionsRequest) (*ListSessionsResponse, error)

	// SetSessionMode changes the current session mode.
	//
	// See protocol docs: [Session Modes](https://agentclientprotocol.com/protocol/session-modes)
	SetSessionMode(ctx context.Context, params *SetSessionModeRequest) (*SetSessionModeResponse, error)

	// SetSessionConfigOption updates a session configuration option.
	//
	// See protocol docs: [Session Config Options](https://agentclientprotocol.com/protocol/session-config-options)
	SetSessionConfigOption(ctx context.Context, params *SetSessionConfigOptionRequest) (*SetSessionConfigOptionResponse, error)

	// Prompt processes a user prompt and generates a response.
	//
	// This is the main method for handling user input and generating agent responses.
	//
	// See protocol docs: [User Message](https://agentclientprotocol.com/protocol/prompt-turn#1-user-message)
	Prompt(ctx context.Context, params *PromptRequest) (*PromptResponse, error)

	// Cancel cancels ongoing operations for a session.
	//
	// This is a notification method (no response expected).
	//
	// See protocol docs: [Cancellation](https://agentclientprotocol.com/protocol/prompt-turn#cancellation)
	Cancel(ctx context.Context, params *CancelNotification) error
}

// SessionForker is an optional interface for agents that support session forking (unstable).
//
// Implement this interface to allow clients to fork sessions.
type SessionForker interface {
	ForkSession(ctx context.Context, params *ForkSessionRequest) (*ForkSessionResponse, error)
}

// SessionResumer is an optional interface for agents that support session resuming (unstable).
//
// Implement this interface to allow clients to resume sessions without replaying history.
type SessionResumer interface {
	ResumeSession(ctx context.Context, params *ResumeSessionRequest) (*ResumeSessionResponse, error)
}

// SessionCloser is an optional interface for agents that support session closing (unstable).
//
// Implement this interface to allow clients to explicitly close sessions.
type SessionCloser interface {
	CloseSession(ctx context.Context, params *CloseSessionRequest) (*CloseSessionResponse, error)
}

// ModelSetter is an optional interface for agents that support model selection (unstable).
//
// Implement this interface to allow clients to change the model during a session.
type ModelSetter interface {
	SetSessionModel(ctx context.Context, params *SetSessionModelRequest) (*SetSessionModelResponse, error)
}

// ExtMethodHandler is an optional interface for handling custom extension methods.
//
// Implement this interface on Agent or Client to handle custom protocol extension
// methods. Extension methods should be prefixed with an underscore (e.g., "_myMethod").
type ExtMethodHandler interface {
	ExtMethod(ctx context.Context, method string, params json.RawMessage) (any, error)
}

// ExtNotificationHandler is an optional interface for handling custom extension notifications.
//
// Implement this interface on Agent or Client to handle custom protocol extension
// notifications. Extension methods should be prefixed with an underscore (e.g., "_myNotification").
type ExtNotificationHandler interface {
	ExtNotification(ctx context.Context, method string, params json.RawMessage) error
}

// Client represents the interface for communicating with the client from an agent.
//
// This interface provides methods that agents can use to request services
// from the client, such as file system access and permission requests.
//
// See protocol docs: [Client](https://agentclientprotocol.com/protocol/overview#client)
type Client interface {
	// SessionUpdate sends a session update notification to the client.
	//
	// Used to stream real-time progress and results during prompt processing.
	//
	// See protocol docs: [Agent Reports Output](https://agentclientprotocol.com/protocol/prompt-turn#3-agent-reports-output)
	SessionUpdate(ctx context.Context, params *SessionNotification) error

	// RequestPermission requests user permission for a tool call operation.
	//
	// Called when the agent needs user authorization before executing
	// a potentially sensitive operation.
	//
	// See protocol docs: [Requesting Permission](https://agentclientprotocol.com/protocol/tool-calls#requesting-permission)
	RequestPermission(ctx context.Context, params *RequestPermissionRequest) (*RequestPermissionResponse, error)

	// ReadTextFile reads content from a text file in the client's file system.
	//
	// Only available if the client supports the fs.readTextFile capability.
	//
	// See protocol docs: [FileSystem](https://agentclientprotocol.com/protocol/initialization#filesystem)
	ReadTextFile(ctx context.Context, params *ReadTextFileRequest) (*ReadTextFileResponse, error)

	// WriteTextFile writes content to a text file in the client's file system.
	//
	// Only available if the client supports the fs.writeTextFile capability.
	//
	// See protocol docs: [FileSystem](https://agentclientprotocol.com/protocol/initialization#filesystem)
	WriteTextFile(ctx context.Context, params *WriteTextFileRequest) (*WriteTextFileResponse, error)

	// CreateTerminal creates a new terminal session.
	//
	// Only available if the client supports the terminal capability.
	CreateTerminal(ctx context.Context, params *CreateTerminalRequest) (*CreateTerminalResponse, error)

	// TerminalOutput gets the current output and status of a terminal.
	TerminalOutput(ctx context.Context, params *TerminalOutputRequest) (*TerminalOutputResponse, error)

	// ReleaseTerminal releases a terminal and frees its resources.
	ReleaseTerminal(ctx context.Context, params *ReleaseTerminalRequest) (*ReleaseTerminalResponse, error)

	// WaitForTerminalExit waits for a terminal command to exit.
	WaitForTerminalExit(ctx context.Context, params *WaitForTerminalExitRequest) (*WaitForTerminalExitResponse, error)

	// KillTerminalCommand kills a terminal command without releasing the terminal.
	KillTerminalCommand(ctx context.Context, params *KillTerminalRequest) (*KillTerminalResponse, error)
}
