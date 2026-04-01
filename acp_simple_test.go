package acp

import (
	"encoding/json"
	"testing"
)

// Test basic ACP types and helper functions
func TestProtocolVersion(t *testing.T) {
	version := ProtocolVersion(CurrentProtocolVersion)
	if int(version) != CurrentProtocolVersion {
		t.Errorf("Expected protocol version %d, got %d", CurrentProtocolVersion, int(version))
	}
}

func TestSessionID(t *testing.T) {
	sessionId := SessionID("test-session-123")
	if string(sessionId) != "test-session-123" {
		t.Errorf("Expected session ID 'test-session-123', got '%s'", string(sessionId))
	}
}

func TestToolCallID(t *testing.T) {
	toolCallId := ToolCallID("tool-call-456")
	if string(toolCallId) != "tool-call-456" {
		t.Errorf("Expected tool call ID 'tool-call-456', got '%s'", string(toolCallId))
	}
}

func TestNewContentBlockText(t *testing.T) {
	text := "Hello, world!"
	contentBlock := NewContentBlockText(text)

	textBlock, ok := contentBlock.AsText()
	if !ok {
		t.Error("Expected content block to be text type")
	}

	if textBlock.Text != text {
		t.Errorf("Expected text '%s', got '%s'", text, textBlock.Text)
	}

	if textBlock.Type != "text" {
		t.Errorf("Expected type 'text', got '%s'", textBlock.Type)
	}
}

func TestNewSessionUpdateAgentMessageChunk(t *testing.T) {
	text := "Agent response"
	contentBlock := NewContentBlockText(text)
	update := NewSessionUpdateAgentMessageChunk(contentBlock)

	chunk, ok := update.AsAgentMessageChunk()
	if !ok {
		t.Error("Expected session update to be agent message chunk type")
	}

	textBlock, ok := chunk.Content.AsText()
	if !ok {
		t.Error("Expected chunk content to be text type")
	}

	if textBlock.Text != text {
		t.Errorf("Expected chunk text '%s', got '%s'", text, textBlock.Text)
	}
}

func TestNewRequestPermissionOutcomeSelected(t *testing.T) {
	optionId := PermissionOptionID("allow")
	outcome := NewRequestPermissionOutcomeSelected(optionId)

	selected, ok := outcome.AsSelected()
	if !ok {
		t.Error("Expected outcome to be selected type")
	}

	if selected.OptionID != optionId {
		t.Errorf("Expected option ID '%s', got '%s'", optionId, selected.OptionID)
	}
}

func TestPermissionOptions(t *testing.T) {
	option := PermissionOption{
		Kind:     PermissionOptionKindAllowOnce,
		Name:     "Allow this operation",
		OptionID: PermissionOptionID("allow"),
	}

	if option.Kind != PermissionOptionKindAllowOnce {
		t.Errorf("Expected kind '%s', got '%s'", PermissionOptionKindAllowOnce, option.Kind)
	}

	if option.Name != "Allow this operation" {
		t.Errorf("Expected name 'Allow this operation', got '%s'", option.Name)
	}

	if option.OptionID != PermissionOptionID("allow") {
		t.Errorf("Expected option ID 'allow', got '%s'", option.OptionID)
	}
}

func TestToolCallLocation(t *testing.T) {
	location := ToolCallLocation{
		Path: "/path/to/file.txt",
		Line: nil,
	}

	if location.Path != "/path/to/file.txt" {
		t.Errorf("Expected path '/path/to/file.txt', got '%s'", location.Path)
	}

	if location.Line != nil {
		t.Errorf("Expected line to be nil, got %v", location.Line)
	}
}

func TestToolCallContentContent(t *testing.T) {
	text := "Tool output"
	contentBlock := NewContentBlockText(text)
	toolCallContent := NewToolCallContentContent(contentBlock)

	content, ok := toolCallContent.AsContent()
	if !ok {
		t.Error("Expected tool call content to be content type")
	}

	textBlock, ok := content.Content.Content.AsText()
	if !ok {
		t.Error("Expected content to be text type")
	}

	if textBlock.Text != text {
		t.Errorf("Expected text '%s', got '%s'", text, textBlock.Text)
	}
}

func TestStopReasons(t *testing.T) {
	reasons := []StopReason{
		StopReasonEndTurn,
		StopReasonMaxTokens,
		StopReasonMaxTurnRequests,
		StopReasonRefusal,
		StopReasonCancelled,
	}

	expectedReasons := []string{
		"end_turn",
		"max_tokens",
		"max_turn_requests",
		"refusal",
		"cancelled",
	}

	for i, reason := range reasons {
		if string(reason) != expectedReasons[i] {
			t.Errorf("Expected reason '%s', got '%s'", expectedReasons[i], string(reason))
		}
	}
}

func TestToolKinds(t *testing.T) {
	kinds := []ToolKind{
		ToolKindRead,
		ToolKindEdit,
		ToolKindDelete,
		ToolKindMove,
		ToolKindSearch,
		ToolKindExecute,
		ToolKindThink,
		ToolKindFetch,
		ToolKindOther,
	}

	expectedKinds := []string{
		"read",
		"edit",
		"delete",
		"move",
		"search",
		"execute",
		"think",
		"fetch",
		"other",
	}

	for i, kind := range kinds {
		if string(kind) != expectedKinds[i] {
			t.Errorf("Expected kind '%s', got '%s'", expectedKinds[i], string(kind))
		}
	}
}

func TestToolCallStatuses(t *testing.T) {
	statuses := []ToolCallStatus{
		ToolCallStatusPending,
		ToolCallStatusInProgress,
		ToolCallStatusCompleted,
		ToolCallStatusFailed,
	}

	expectedStatuses := []string{
		"pending",
		"in_progress",
		"completed",
		"failed",
	}

	for i, status := range statuses {
		if string(status) != expectedStatuses[i] {
			t.Errorf("Expected status '%s', got '%s'", expectedStatuses[i], string(status))
		}
	}
}

func TestPointerCreation(t *testing.T) {
	// Go 1.26+ supports new(constant)
	kindPtr := new(ToolKindRead)
	if kindPtr == nil || *kindPtr != ToolKindRead {
		t.Errorf("ToolKind pointer failed: expected '%s', got %v", ToolKindRead, kindPtr)
	}

	statusPtr := new(ToolCallStatusCompleted)
	if statusPtr == nil || *statusPtr != ToolCallStatusCompleted {
		t.Errorf("ToolCallStatus pointer failed: expected '%s', got %v", ToolCallStatusCompleted, statusPtr)
	}
}

// Test request/response structures
func TestInitializeRequest(t *testing.T) {
	req := &InitializeRequest{
		ProtocolVersion: ProtocolVersion(1),
		ClientCapabilities: &ClientCapabilities{
			FS: &FileSystemCapabilities{
				ReadTextFile:  true,
				WriteTextFile: true,
			},
			Terminal: false,
		},
	}

	if req.ProtocolVersion != ProtocolVersion(1) {
		t.Errorf("Expected protocol version 1, got %d", req.ProtocolVersion)
	}

	if req.ClientCapabilities == nil {
		t.Error("Expected client capabilities, got nil")
	} else {
		if req.ClientCapabilities.FS == nil {
			t.Error("Expected filesystem capabilities, got nil")
		} else {
			if !req.ClientCapabilities.FS.ReadTextFile {
				t.Error("Expected ReadTextFile to be true")
			}
			if !req.ClientCapabilities.FS.WriteTextFile {
				t.Error("Expected WriteTextFile to be true")
			}
		}
		if req.ClientCapabilities.Terminal {
			t.Error("Expected Terminal to be false")
		}
	}
}

func TestInitializeResponse(t *testing.T) {
	resp := &InitializeResponse{
		ProtocolVersion: ProtocolVersion(1),
		AgentCapabilities: &AgentCapabilities{
			LoadSession: true,
			MCPCapabilities: &MCPCapabilities{
				HTTP: false,
				SSE:  true,
			},
			PromptCapabilities: &PromptCapabilities{
				Audio:           false,
				EmbeddedContext: true,
				Image:           true,
			},
		},
		AuthMethods: []AuthMethod{
			{
				ID:          "oauth",
				Name:        "OAuth Authentication",
				Description: "OAuth 2.0 authentication",
			},
		},
	}

	if resp.ProtocolVersion != ProtocolVersion(1) {
		t.Errorf("Expected protocol version 1, got %d", resp.ProtocolVersion)
	}

	if resp.AgentCapabilities == nil {
		t.Error("Expected agent capabilities, got nil")
		return
	}

	if !resp.AgentCapabilities.LoadSession {
		t.Error("Expected LoadSession to be true")
	}

	if resp.AgentCapabilities.MCPCapabilities == nil {
		t.Error("Expected MCP capabilities, got nil")
	} else {
		if resp.AgentCapabilities.MCPCapabilities.HTTP {
			t.Error("Expected HTTP to be false")
		}
		if !resp.AgentCapabilities.MCPCapabilities.SSE {
			t.Error("Expected SSE to be true")
		}
	}

	if resp.AgentCapabilities.PromptCapabilities == nil {
		t.Error("Expected prompt capabilities, got nil")
	} else {
		if resp.AgentCapabilities.PromptCapabilities.Audio {
			t.Error("Expected Audio to be false")
		}
		if !resp.AgentCapabilities.PromptCapabilities.EmbeddedContext {
			t.Error("Expected EmbeddedContext to be true")
		}
		if !resp.AgentCapabilities.PromptCapabilities.Image {
			t.Error("Expected Image to be true")
		}
	}

	if len(resp.AuthMethods) != 1 {
		t.Errorf("Expected 1 auth method, got %d", len(resp.AuthMethods))
	} else {
		authMethod := resp.AuthMethods[0]
		if authMethod.ID != "oauth" {
			t.Errorf("Expected auth method ID 'oauth', got '%s'", authMethod.ID)
		}
		if authMethod.Name != "OAuth Authentication" {
			t.Errorf("Expected auth method name 'OAuth Authentication', got '%s'", authMethod.Name)
		}
		if authMethod.Description != "OAuth 2.0 authentication" {
			t.Errorf("Expected auth method description 'OAuth 2.0 authentication', got '%s'", authMethod.Description)
		}
	}
}

func TestAgentResponseEnvelopeJSON(t *testing.T) {
	resp := AgentResponse{
		ID:     RequestID("1"),
		Result: json.RawMessage(`{"ok":true}`),
	}

	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if _, ok := decoded["id"]; !ok {
		t.Fatal("expected id field in agent response envelope")
	}
	if _, ok := decoded["result"]; !ok {
		t.Fatal("expected result field in agent response envelope")
	}
}

func TestClientResponseEnvelopeErrorJSON(t *testing.T) {
	resp := ClientResponse{
		ID: RequestID("2"),
		Error: &Error{
			Code:    ErrorCodeMethodNotFound,
			Message: "method not found",
		},
	}

	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	errObj, ok := decoded["error"].(map[string]any)
	if !ok {
		t.Fatal("expected error field in client response envelope")
	}
	if errObj["message"] != "method not found" {
		t.Fatalf("expected error message 'method not found', got %v", errObj["message"])
	}
}

func TestInitializeRequest_ClientCapabilitiesIncludeAuthAndElicitation(t *testing.T) {
	req := &InitializeRequest{
		ProtocolVersion: ProtocolVersion(1),
		ClientCapabilities: &ClientCapabilities{
			Auth: &AuthCapabilities{Terminal: true},
			Elicitation: &ElicitationCapabilities{
				Form: &ElicitationFormCapabilities{},
				URL:  &ElicitationURLCapabilities{},
			},
		},
	}

	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	cc, ok := decoded["clientCapabilities"].(map[string]any)
	if !ok {
		t.Fatal("clientCapabilities missing from serialized initialize request")
	}
	if _, ok := cc["auth"]; !ok {
		t.Fatal("expected auth capability in serialized initialize request")
	}
	if _, ok := cc["elicitation"]; !ok {
		t.Fatal("expected elicitation capability in serialized initialize request")
	}
}

func TestInitializeResponse_AgentCapabilitiesIncludeLogout(t *testing.T) {
	resp := &InitializeResponse{
		ProtocolVersion: ProtocolVersion(1),
		AgentCapabilities: &AgentCapabilities{
			Auth: &AgentAuthCapabilities{Logout: &LogoutCapabilities{}},
		},
	}

	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	ac, ok := decoded["agentCapabilities"].(map[string]any)
	if !ok {
		t.Fatal("agentCapabilities missing from serialized initialize response")
	}
	auth, ok := ac["auth"].(map[string]any)
	if !ok {
		t.Fatal("expected auth capability in serialized initialize response")
	}
	if _, ok := auth["logout"]; !ok {
		t.Fatal("expected logout capability in serialized initialize response")
	}
}

func TestGeneratedLogoutAndElicitationSimpleTypesJSON(t *testing.T) {
	logoutReq := LogoutRequest{}
	logoutResp := LogoutResponse{}
	notification := ElicitationCompleteNotification{ElicitationID: ElicitationID("elic-1")}

	for name, value := range map[string]any{
		"logout request":           logoutReq,
		"logout response":          logoutResp,
		"elicitation notification": notification,
	} {
		if _, err := json.Marshal(value); err != nil {
			t.Fatalf("marshal %s failed: %v", name, err)
		}
	}

	b, err := json.Marshal(notification)
	if err != nil {
		t.Fatalf("marshal elicitation notification failed: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal elicitation notification failed: %v", err)
	}

	if decoded["elicitationId"] != "elic-1" {
		t.Fatalf("expected elicitationId 'elic-1', got %v", decoded["elicitationId"])
	}
}

func TestInitializeResponse_AgentCapabilitiesIncludeStableSessionLifecycleOptions(t *testing.T) {
	resp := &InitializeResponse{
		ProtocolVersion: ProtocolVersion(1),
		AgentCapabilities: &AgentCapabilities{
			SessionCapabilities: &SessionCapabilities{
				Close:  &SessionCloseCapabilities{},
				Fork:   &SessionForkCapabilities{},
				Resume: &SessionResumeCapabilities{},
			},
		},
	}

	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	ac, ok := decoded["agentCapabilities"].(map[string]any)
	if !ok {
		t.Fatal("agentCapabilities missing from serialized initialize response")
	}
	sc, ok := ac["sessionCapabilities"].(map[string]any)
	if !ok {
		t.Fatal("expected sessionCapabilities in serialized initialize response")
	}
	if _, ok := sc["close"]; !ok {
		t.Fatal("expected close capability in serialized initialize response")
	}
	if _, ok := sc["fork"]; !ok {
		t.Fatal("expected fork capability in serialized initialize response")
	}
	if _, ok := sc["resume"]; !ok {
		t.Fatal("expected resume capability in serialized initialize response")
	}
}

func TestSetSessionModelRequestJSON(t *testing.T) {
	req := SetSessionModelRequest{
		SessionID: SessionID("session-1"),
		ModelID:   ModelID("model-1"),
	}

	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded["sessionId"] != "session-1" {
		t.Fatalf("expected sessionId 'session-1', got %v", decoded["sessionId"])
	}
	if decoded["modelId"] != "model-1" {
		t.Fatalf("expected modelId 'model-1', got %v", decoded["modelId"])
	}
}
