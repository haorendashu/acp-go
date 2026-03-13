package acp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

// Connection represents a bidirectional JSON-RPC connection
type Connection struct {
	reader           io.Reader
	writer           io.Writer
	pendingResponses sync.Map // map[int64]*pendingResponse
	nextRequestID    int64
	handler          MethodHandler
	writeQueue       chan jsonRpcMessage
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup // tracks in-flight handlers
	errorHandler     func(error)
}

// ConnectionOption configures a Connection.
type ConnectionOption func(*Connection)

// WithErrorHandler sets a callback for non-fatal errors (parse failures, write errors, etc.).
func WithErrorHandler(h func(error)) ConnectionOption {
	return func(c *Connection) { c.errorHandler = h }
}

// MethodHandler handles incoming JSON-RPC method calls.
// The context is derived from the connection's context.
type MethodHandler func(ctx context.Context, method string, params json.RawMessage) (any, error)

// pendingResponse represents a response waiting for completion
type pendingResponse struct {
	result chan responseResult
}

// responseResult contains the result or error from a JSON-RPC call
type responseResult struct {
	data  json.RawMessage
	error error
}

// jsonRpcMessage represents a JSON-RPC message
type jsonRpcMessage struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      *int64          `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRpcError   `json:"error,omitempty"`
}

// jsonRpcError represents a JSON-RPC error
type jsonRpcError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// NewConnection creates a new bidirectional JSON-RPC connection
func NewConnection(handler MethodHandler, reader io.Reader, writer io.Writer, opts ...ConnectionOption) *Connection {
	conn := &Connection{
		reader:     reader,
		writer:     writer,
		handler:    handler,
		writeQueue: make(chan jsonRpcMessage, 100),
	}
	for _, opt := range opts {
		opt(conn)
	}
	return conn
}

// Start begins processing JSON-RPC messages.
// The provided context governs the connection lifecycle.
func (c *Connection) Start(ctx context.Context) error {
	c.ctx, c.cancel = context.WithCancel(ctx)

	// Start writer goroutine
	c.wg.Go(func() {
		c.writeLoop()
	})

	// Start reader loop (blocks until done)
	err := c.readLoop()
	// Reader exited — shut down the connection
	c.cancel()
	return err
}

// Close closes the connection gracefully.
// It cancels the connection context and waits for in-flight handlers to complete.
func (c *Connection) Close() error {
	c.cancel()
	c.wg.Wait()
	return nil
}

// Done returns a channel that is closed when the connection is done.
func (c *Connection) Done() <-chan struct{} {
	return c.ctx.Done()
}

// logError logs an error using the errorHandler if set.
func (c *Connection) logError(err error) {
	if c.errorHandler != nil {
		c.errorHandler(err)
	}
}

// handlerContext returns a context that is cancelled when the connection closes.
// Uses context.AfterFunc to avoid spawning a bridging goroutine.
func (c *Connection) handlerContext() (context.Context, context.CancelFunc) {
	handlerCtx, handlerCancel := context.WithCancel(c.ctx)
	return handlerCtx, handlerCancel
}

// readLoop reads and processes incoming messages
func (c *Connection) readLoop() error {
	scanner := bufio.NewScanner(c.reader)
	// Set buffer size to 50MB to handle large messages (e.g., base64-encoded images)
	scanner.Buffer(make([]byte, 0, 64*1024), 50*1024*1024)

	for scanner.Scan() {
		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		default:
		}

		data := scanner.Bytes()
		if len(data) == 0 {
			continue
		}

		// Parse and classify message synchronously
		var msg jsonRpcMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			c.logError(fmt.Errorf("failed to parse JSON-RPC message: %w", err))
			continue
		}

		if msg.ID != nil && msg.Method != "" {
			// It's a request — dispatch handler in goroutine
			c.wg.Go(func() {
				c.handleRequest(msg)
			})
		} else if msg.Method != "" {
			// It's a notification — dispatch handler in goroutine
			c.wg.Go(func() {
				c.handleNotification(msg)
			})
		} else if msg.ID != nil {
			// It's a response — handle inline to avoid ordering issues
			c.handleResponse(msg)
		}
	}

	return scanner.Err()
}

// writeLoop writes outgoing messages
func (c *Connection) writeLoop() {
	for {
		select {
		case msg := <-c.writeQueue:
			data, err := json.Marshal(msg)
			if err != nil {
				c.logError(fmt.Errorf("failed to marshal JSON-RPC message: %w", err))
				continue
			}

			data = append(data, '\n')
			if _, err := c.writer.Write(data); err != nil {
				c.logError(fmt.Errorf("failed to write JSON-RPC message: %w", err))
				c.cancel() // close connection on write failure
				return
			}

		case <-c.ctx.Done():
			return
		}
	}
}

// trySend attempts to send a message on the write queue, respecting context cancellation.
func (c *Connection) trySend(msg jsonRpcMessage) {
	select {
	case c.writeQueue <- msg:
	case <-c.ctx.Done():
	}
}

// handleRequest processes incoming requests
func (c *Connection) handleRequest(msg jsonRpcMessage) {
	handlerCtx, handlerCancel := c.handlerContext()
	defer handlerCancel()

	response := jsonRpcMessage{
		Jsonrpc: "2.0",
		ID:      msg.ID,
	}

	if c.handler != nil {
		result, err := c.handler(handlerCtx, msg.Method, msg.Params)
		if err != nil {
			if reqErr, ok := err.(*RequestError); ok {
				response.Error = &jsonRpcError{
					Code:    int(reqErr.Code),
					Message: reqErr.Msg,
				}
				if reqErr.Details != nil {
					if data, marshalErr := json.Marshal(reqErr.Details); marshalErr == nil {
						response.Error.Data = data
					}
				}
			} else {
				response.Error = &jsonRpcError{
					Code:    int(ErrorCodeInternalError),
					Message: err.Error(),
				}
			}
		} else if result != nil {
			if data, marshalErr := json.Marshal(result); marshalErr == nil {
				response.Result = data
			} else {
				response.Error = &jsonRpcError{
					Code:    int(ErrorCodeInternalError),
					Message: "Failed to marshal result",
				}
			}
		} else {
			// Void success — JSON-RPC requires a result field
			response.Result = json.RawMessage("{}")
		}
	} else {
		response.Error = &jsonRpcError{
			Code:    int(ErrorCodeMethodNotFound),
			Message: "Method not found",
		}
	}

	c.trySend(response)
}

// handleNotification processes incoming notifications
func (c *Connection) handleNotification(msg jsonRpcMessage) {
	if c.handler != nil {
		handlerCtx, handlerCancel := c.handlerContext()
		defer handlerCancel()

		if _, err := c.handler(handlerCtx, msg.Method, msg.Params); err != nil {
			c.logError(fmt.Errorf("notification handler error for %s: %w", msg.Method, err))
		}
	}
}

// handleResponse processes incoming responses
func (c *Connection) handleResponse(msg jsonRpcMessage) {
	if msg.ID == nil {
		return
	}

	if pending, ok := c.pendingResponses.LoadAndDelete(*msg.ID); ok {
		p := pending.(*pendingResponse)

		var result responseResult
		if msg.Error != nil {
			result.error = &RequestError{
				Code: ErrorCode(msg.Error.Code),
				Msg:  msg.Error.Message,
			}
		} else {
			result.data = msg.Result
		}

		select {
		case p.result <- result:
		default:
		}
	}
}

// sendRequest is a generic helper that sends a JSON-RPC request and unmarshals the response.
func sendRequest[T any](ctx context.Context, conn *Connection, method string, params any) (*T, error) {
	data, err := conn.SendRequest(ctx, method, params)
	if err != nil {
		return nil, err
	}
	var response T
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// SendRequest sends a JSON-RPC request and waits for the response
func (c *Connection) SendRequest(ctx context.Context, method string, params any) (json.RawMessage, error) {
	// Generate unique request ID
	requestID := atomic.AddInt64(&c.nextRequestID, 1)

	// Create response channel
	pending := &pendingResponse{
		result: make(chan responseResult, 1),
	}

	// Store pending response
	c.pendingResponses.Store(requestID, pending)

	// Cleanup on exit — just delete from map, don't close channel (avoids race)
	defer c.pendingResponses.Delete(requestID)

	// Prepare the request message
	msg := jsonRpcMessage{
		Jsonrpc: "2.0",
		ID:      &requestID,
		Method:  method,
	}

	if params != nil {
		if data, err := json.Marshal(params); err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		} else {
			msg.Params = data
		}
	}

	// Send the request
	select {
	case c.writeQueue <- msg:
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.ctx.Done():
		return nil, c.ctx.Err()
	}

	// Wait for response — no hardcoded timeout, rely on ctx
	select {
	case result := <-pending.result:
		if result.error != nil {
			return nil, result.error
		}
		return result.data, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.ctx.Done():
		return nil, c.ctx.Err()
	}
}

// SendNotification sends a JSON-RPC notification (no response expected)
func (c *Connection) SendNotification(ctx context.Context, method string, params any) error {
	msg := jsonRpcMessage{
		Jsonrpc: "2.0",
		Method:  method,
	}

	if params != nil {
		if data, err := json.Marshal(params); err != nil {
			return fmt.Errorf("failed to marshal params: %w", err)
		} else {
			msg.Params = data
		}
	}

	select {
	case c.writeQueue <- msg:
		return nil
	case <-c.ctx.Done():
		return c.ctx.Err()
	}
}
