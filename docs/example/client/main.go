package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	acp "github.com/ironpark/go-acp"
)

// ExampleClient implements the acp.Client interface
type ExampleClient struct{}

func (c *ExampleClient) RequestPermission(ctx context.Context, params *acp.RequestPermissionRequest) (*acp.RequestPermissionResponse, error) {
	fmt.Printf("\n🔐 Permission requested: %s\n", params.ToolCall.Title)

	fmt.Println("\nOptions:")
	for i, option := range params.Options {
		fmt.Printf("   %d. %s (%s)\n", i+1, option.Name, option.Kind)
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\nChoose an option: ")
		answer, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		answer = strings.TrimSpace(answer)
		optionIndex, err := strconv.Atoi(answer)
		if err != nil {
			fmt.Println("Invalid input. Please enter a number.")
			continue
		}

		if optionIndex >= 1 && optionIndex <= len(params.Options) {
			selectedOption := params.Options[optionIndex-1]
			return &acp.RequestPermissionResponse{
				Outcome: acp.NewRequestPermissionOutcomeSelected(selectedOption.OptionID),
			}, nil
		} else {
			fmt.Printf("Invalid option. Please choose a number between 1 and %d.\n", len(params.Options))
		}
	}
}

func (c *ExampleClient) SessionUpdate(ctx context.Context, params *acp.SessionNotification) error {
	update := params.Update

	if chunk, ok := update.AsAgentMessageChunk(); ok {
		if text, ok := chunk.Content.AsText(); ok {
			fmt.Print(text.Text)
		} else {
			fmt.Print("[non-text content]")
		}
	} else if toolCall, ok := update.AsToolCall(); ok {
		fmt.Printf("\n🔧 %s", toolCall.Title)
		if toolCall.Status != nil {
			fmt.Printf(" (%s)", *toolCall.Status)
		}
		fmt.Println()
	} else if toolUpdate, ok := update.AsToolCallUpdate(); ok {
		fmt.Printf("\n🔧 Tool call `%s` updated", toolUpdate.ToolCallID)
		if toolUpdate.Status != nil {
			fmt.Printf(": %s", *toolUpdate.Status)
		}
		fmt.Println()
	} else if _, ok := update.AsPlan(); ok {
		fmt.Println("[plan update]")
	} else if _, ok := update.AsUserMessageChunk(); ok {
		fmt.Println("[user message chunk]")
	} else {
		fmt.Println("[unknown session update]")
	}

	return nil
}

func (c *ExampleClient) WriteTextFile(ctx context.Context, params *acp.WriteTextFileRequest) (*acp.WriteTextFileResponse, error) {
	fmt.Printf("[Client] Write text file called with: %+v\n", params)
	return &acp.WriteTextFileResponse{}, nil
}

func (c *ExampleClient) ReadTextFile(ctx context.Context, params *acp.ReadTextFileRequest) (*acp.ReadTextFileResponse, error) {
	fmt.Printf("[Client] Read text file called with: %+v\n", params)
	return &acp.ReadTextFileResponse{
		Content: "Mock file content",
	}, nil
}

func (c *ExampleClient) CreateTerminal(ctx context.Context, params *acp.CreateTerminalRequest) (*acp.CreateTerminalResponse, error) {
	fmt.Printf("[Client] Create terminal called with: %+v\n", params)
	return &acp.CreateTerminalResponse{
		TerminalID: "mock-terminal-id",
	}, nil
}

func (c *ExampleClient) TerminalOutput(ctx context.Context, params *acp.TerminalOutputRequest) (*acp.TerminalOutputResponse, error) {
	fmt.Printf("[Client] Terminal output called with: %+v\n", params)
	return &acp.TerminalOutputResponse{
		Output:     "Mock terminal output",
		Truncated:  false,
		ExitStatus: nil,
	}, nil
}

func (c *ExampleClient) ReleaseTerminal(ctx context.Context, params *acp.ReleaseTerminalRequest) (*acp.ReleaseTerminalResponse, error) {
	fmt.Printf("[Client] Release terminal called with: %+v\n", params)
	return &acp.ReleaseTerminalResponse{}, nil
}

func (c *ExampleClient) WaitForTerminalExit(ctx context.Context, params *acp.WaitForTerminalExitRequest) (*acp.WaitForTerminalExitResponse, error) {
	fmt.Printf("[Client] Wait for terminal exit called with: %+v\n", params)
	return &acp.WaitForTerminalExitResponse{
		ExitCode: nil,
		Signal:   "",
	}, nil
}

func (c *ExampleClient) KillTerminalCommand(ctx context.Context, params *acp.KillTerminalRequest) (*acp.KillTerminalResponse, error) {
	fmt.Printf("[Client] Kill terminal command called with: %+v\n", params)
	return &acp.KillTerminalResponse{}, nil
}

func main() {
	ctx := context.Background()

	// Get the path to the agent executable
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Fprintf(os.Stderr, "Failed to get current file path\n")
		os.Exit(1)
	}

	// Build path to the agent example
	currentDir := filepath.Dir(currentFile)
	exampleDir := filepath.Dir(currentDir)
	agentDir := filepath.Join(exampleDir, "agent")

	// Build the agent if necessary
	agentBinary := filepath.Join(agentDir, "agent")
	if runtime.GOOS == "windows" {
		agentBinary += ".exe"
	}

	// Check if agent binary exists or if we need to build it
	if _, err := os.Stat(agentBinary); os.IsNotExist(err) {
		fmt.Println("Building agent...")
		buildCmd := exec.Command("go", "build", "-o", agentBinary, "main.go")
		buildCmd.Dir = agentDir
		buildCmd.Stderr = os.Stderr
		if err := buildCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to build agent: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Agent built successfully.")
	}

	// Start the agent as a subprocess
	agentCmd := exec.Command(agentBinary)
	agentCmd.Stderr = os.Stderr

	agentStdin, err := agentCmd.StdinPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create agent stdin pipe: %v\n", err)
		os.Exit(1)
	}

	agentStdout, err := agentCmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create agent stdout pipe: %v\n", err)
		os.Exit(1)
	}

	if err := agentCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start agent: %v\n", err)
		os.Exit(1)
	}

	defer func() {
		if agentCmd.Process != nil {
			agentCmd.Process.Kill()
		}
	}()

	// Create the client connection
	client := &ExampleClient{}
	connection := acp.NewClientSideConnection(client, agentStdin, agentStdout)

	// Start the connection in background
	go func() {
		if err := connection.Start(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Connection error: %v\n", err)
		}
	}()

	// Initialize the connection
	initResult, err := connection.Initialize(ctx, &acp.InitializeRequest{
		ProtocolVersion: acp.ProtocolVersion(acp.CurrentProtocolVersion),
		ClientCapabilities: &acp.ClientCapabilities{
			FS: &acp.FileSystemCapabilities{
				ReadTextFile:  true,
				WriteTextFile: true,
			},
			Terminal: false,
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Connected to agent (protocol v%d)\n", initResult.ProtocolVersion)

	// Create a new session
	cwd, _ := os.Getwd()
	sessionResult, err := connection.NewSession(ctx, &acp.NewSessionRequest{
		Cwd:        cwd,
		MCPServers: []acp.MCPServer{},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create session: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📝 Created session: %s\n", sessionResult.SessionID)
	fmt.Printf("💬 User: Hello, agent!\n\n")
	fmt.Print(" ")

	// Send a test prompt
	promptResult, err := connection.Prompt(ctx, &acp.PromptRequest{
		SessionID: sessionResult.SessionID,
		Prompt: []acp.ContentBlock{
			acp.NewContentBlockText("Hello, agent!"),
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send prompt: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n\n✅ Agent completed with: %s\n", promptResult.StopReason)
}