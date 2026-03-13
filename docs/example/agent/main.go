package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	acp "github.com/ironpark/go-acp"
)

// ExampleAgent implements the acp.Agent interface with full session update capabilities
type ExampleAgent struct {
	client   acp.Client
	sessions map[acp.SessionID]*AgentSession
}

// AgentSession holds session state
type AgentSession struct {
	sessionId     acp.SessionID
	cancelContext context.Context
	cancelFunc    context.CancelFunc
}

func NewExampleAgent() *ExampleAgent {
	return &ExampleAgent{
		sessions: make(map[acp.SessionID]*AgentSession),
	}
}

func (a *ExampleAgent) Initialize(ctx context.Context, params *acp.InitializeRequest) (*acp.InitializeResponse, error) {
	return &acp.InitializeResponse{
		ProtocolVersion: acp.ProtocolVersion(acp.CurrentProtocolVersion),
		AgentCapabilities: &acp.AgentCapabilities{
			LoadSession: false,
			MCPCapabilities: &acp.MCPCapabilities{
				HTTP: false,
				SSE:  false,
			},
			PromptCapabilities: &acp.PromptCapabilities{
				Audio:           false,
				EmbeddedContext: false,
				Image:           false,
			},
		},
		AuthMethods: []acp.AuthMethod{},
	}, nil
}

func (a *ExampleAgent) Authenticate(ctx context.Context, params *acp.AuthenticateRequest) (*acp.AuthenticateResponse, error) {
	return &acp.AuthenticateResponse{}, nil
}

func (a *ExampleAgent) NewSession(ctx context.Context, params *acp.NewSessionRequest) (*acp.NewSessionResponse, error) {
	// Generate a random session ID
	sessionId := acp.SessionID(fmt.Sprintf("session_%s", generateRandomID()))

	// Create cancellation context for this session
	sessionCtx, cancelFunc := context.WithCancel(context.Background())

	session := &AgentSession{
		sessionId:     sessionId,
		cancelContext: sessionCtx,
		cancelFunc:    cancelFunc,
	}

	a.sessions[sessionId] = session

	return &acp.NewSessionResponse{
		SessionID: sessionId,
		Modes:     nil,
	}, nil
}

func (a *ExampleAgent) LoadSession(ctx context.Context, params *acp.LoadSessionRequest) (*acp.LoadSessionResponse, error) {
	return nil, acp.ErrMethodNotFound("session/load")
}

func (a *ExampleAgent) ListSessions(ctx context.Context, params *acp.ListSessionsRequest) (*acp.ListSessionsResponse, error) {
	return &acp.ListSessionsResponse{}, nil
}

func (a *ExampleAgent) SetSessionMode(ctx context.Context, params *acp.SetSessionModeRequest) (*acp.SetSessionModeResponse, error) {
	return &acp.SetSessionModeResponse{}, nil
}

func (a *ExampleAgent) SetSessionConfigOption(ctx context.Context, params *acp.SetSessionConfigOptionRequest) (*acp.SetSessionConfigOptionResponse, error) {
	return &acp.SetSessionConfigOptionResponse{ConfigOptions: []acp.SessionConfigOption{}}, nil
}

func (a *ExampleAgent) Prompt(ctx context.Context, params *acp.PromptRequest) (*acp.PromptResponse, error) {
	session, exists := a.sessions[params.SessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", params.SessionID)
	}

	// Cancel any previous prompt processing for this session
	session.cancelFunc()
	sessionCtx, cancelFunc := context.WithCancel(context.Background())
	session.cancelContext = sessionCtx
	session.cancelFunc = cancelFunc

	// Simulate the turn processing
	err := a.simulateTurn(sessionCtx, params.SessionID)
	if err != nil {
		if sessionCtx.Err() == context.Canceled {
			return &acp.PromptResponse{
				StopReason: acp.StopReasonCancelled,
			}, nil
		}
		return nil, err
	}

	return &acp.PromptResponse{
		StopReason: acp.StopReasonEndTurn,
	}, nil
}

func (a *ExampleAgent) Cancel(ctx context.Context, params *acp.CancelNotification) error {
	if session, exists := a.sessions[params.SessionID]; exists {
		session.cancelFunc()
	}
	return nil
}

func (a *ExampleAgent) simulateTurn(ctx context.Context, sessionId acp.SessionID) error {
	// Send initial agent message chunk
	err := a.client.SessionUpdate(ctx, &acp.SessionNotification{
		SessionID: sessionId,
		Update:    acp.NewSessionUpdateAgentMessageChunk(acp.NewContentBlockText("I'll help you with that. Let me start by reading some files to understand the current situation.")),
	})
	if err != nil {
		return err
	}

	// Simulate model thinking time
	if err := a.simulateModelInteraction(ctx); err != nil {
		return err
	}

	// Send a tool call that doesn't need permission
	toolCallId := acp.ToolCallID("call_1")
	err = a.client.SessionUpdate(ctx, &acp.SessionNotification{
		SessionID: sessionId,
		Update: acp.NewSessionUpdateToolCall(acp.ToolCall{
			ToolCallID: toolCallId,
			Title:      "Reading project files",
			Kind:       new(acp.ToolKindRead),
			Status:     new(acp.ToolCallStatusPending),
			Locations:  []acp.ToolCallLocation{{Path: "/project/README.md"}},
		}),
	})
	if err != nil {
		return err
	}

	if err := a.simulateModelInteraction(ctx); err != nil {
		return err
	}

	// Update tool call to completed
	err = a.client.SessionUpdate(ctx, &acp.SessionNotification{
		SessionID: sessionId,
		Update: acp.NewSessionUpdateToolCallUpdate(acp.ToolCallUpdate{
			ToolCallID: toolCallId,
			Status:     new(acp.ToolCallStatusCompleted),
			Content: []acp.ToolCallContent{
				acp.NewToolCallContentContent(acp.NewContentBlockText("# My Project\n\nThis is a sample project...")),
			},
		}),
	})
	if err != nil {
		return err
	}

	if err := a.simulateModelInteraction(ctx); err != nil {
		return err
	}

	// Send more agent message
	err = a.client.SessionUpdate(ctx, &acp.SessionNotification{
		SessionID: sessionId,
		Update: acp.NewSessionUpdateAgentMessageChunk(acp.NewContentBlockText(" Now I understand the project structure. I need to make some changes to improve it.")),
	})
	if err != nil {
		return err
	}

	if err := a.simulateModelInteraction(ctx); err != nil {
		return err
	}

	// Send a tool call that DOES need permission
	toolCallId2 := acp.ToolCallID("call_2")
	err = a.client.SessionUpdate(ctx, &acp.SessionNotification{
		SessionID: sessionId,
		Update: acp.NewSessionUpdateToolCall(acp.ToolCall{
			ToolCallID: toolCallId2,
			Title:      "Modifying critical configuration file",
			Kind:       new(acp.ToolKindEdit),
			Status:     new(acp.ToolCallStatusPending),
			Locations:  []acp.ToolCallLocation{{Path: "/project/config.json"}},
		}),
	})
	if err != nil {
		return err
	}

	// Request permission for the sensitive operation
	permissionResponse, err := a.client.RequestPermission(ctx, &acp.RequestPermissionRequest{
		SessionID: sessionId,
		ToolCall: acp.ToolCallUpdate{
			ToolCallID: toolCallId2,
			Title:      "Modifying critical configuration file",
			Kind:       new(acp.ToolKindEdit),
			Status:     new(acp.ToolCallStatusPending),
			Locations: []acp.ToolCallLocation{
				{Path: "/home/user/project/config.json"},
			},
			RawInput: nil,
		},
		Options: []acp.PermissionOption{
			{
				Kind:     acp.PermissionOptionKindAllowOnce,
				Name:     "Allow this change",
				OptionID: acp.PermissionOptionID("allow"),
			},
			{
				Kind:     acp.PermissionOptionKindRejectOnce,
				Name:     "Skip this change",
				OptionID: acp.PermissionOptionID("reject"),
			},
		},
	})
	if err != nil {
		return err
	}

	// Handle permission response
	if _, ok := permissionResponse.Outcome.AsCancelled(); ok {
		return nil
	}

	if selectedOutcome, ok := permissionResponse.Outcome.AsSelected(); ok {
		switch selectedOutcome.OptionID {
		case "allow":
			err = a.client.SessionUpdate(ctx, &acp.SessionNotification{
				SessionID: sessionId,
				Update: acp.NewSessionUpdateToolCallUpdate(acp.ToolCallUpdate{
					ToolCallID: toolCallId2,
					Status:     new(acp.ToolCallStatusCompleted),
				}),
			})
			if err != nil {
				return err
			}

			if err := a.simulateModelInteraction(ctx); err != nil {
				return err
			}

			err = a.client.SessionUpdate(ctx, &acp.SessionNotification{
				SessionID: sessionId,
				Update: acp.NewSessionUpdateAgentMessageChunk(acp.NewContentBlockText(" Perfect! I've successfully updated the configuration. The changes have been applied.")),
			})
			if err != nil {
				return err
			}

		case "reject":
			if err := a.simulateModelInteraction(ctx); err != nil {
				return err
			}

			err = a.client.SessionUpdate(ctx, &acp.SessionNotification{
				SessionID: sessionId,
				Update: acp.NewSessionUpdateAgentMessageChunk(acp.NewContentBlockText(" I understand you prefer not to make that change. I'll skip the configuration update.")),
			})
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("unexpected permission outcome: %s", selectedOutcome.OptionID)
		}
	} else {
		return fmt.Errorf("unexpected permission outcome type")
	}

	return nil
}

func (a *ExampleAgent) simulateModelInteraction(ctx context.Context) error {
	select {
	case <-time.After(1 * time.Second):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func generateRandomID() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func main() {
	agent := NewExampleAgent()

	// Create connection using stdin/stdout
	conn := acp.NewAgentSideConnection(agent, os.Stdin, os.Stdout)

	// Set the client reference so the agent can make requests
	agent.client = conn.Client()

	// Start the connection
	if err := conn.Start(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Connection error: %v\n", err)
		os.Exit(1)
	}
}
