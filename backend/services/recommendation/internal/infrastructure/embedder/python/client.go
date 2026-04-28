// Package python runs a long-lived Python worker that produces sentence
// embeddings and exposes them to the Go application layer via a newline-
// delimited JSON protocol over a Unix domain socket.
//
// Protocol (one request/response per line, UTF-8):
//
//	→ {"texts": ["hello", "world"]}
//	← {"embeddings": [[...], [...]]}            // success
//	← {"error": "message"}                       // worker-side failure
//
// Startup handshake: the worker prints "READY" on stdout after the model has
// been loaded and the socket is accepting connections. The Go client blocks
// on that line (bounded by StartupTimeout) before returning from Start.
//
// The client returns application.ErrEmbedderUnavailable on any I/O or decode
// failure. Callers are expected to fall back to the non-semantic ranking
// path; they must never surface embedder failures to the RPC boundary.
package python

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"jobconnect/recommendation/internal/application"
)

const (
	defaultStartupTimeout   = 30 * time.Second
	defaultOperationTimeout = 5 * time.Second
	defaultBatchSize        = 32
	readyLine               = "READY"
)

// Config parameterises the subprocess and IPC channel.
type Config struct {
	// PythonPath is the interpreter to invoke. Empty → "python3".
	PythonPath string
	// WorkerScript is the absolute path to the worker .py file.
	WorkerScript string
	// ModelName is forwarded to the worker via RECOMMENDATION_EMBEDDER_MODEL.
	ModelName string
	// SocketPath is the Unix domain socket path shared with the worker.
	SocketPath string
	// BatchSize bounds how many texts the client sends in a single request;
	// callers larger than this are chunked transparently.
	BatchSize int
	// OperationTimeout bounds a single Embed round-trip.
	OperationTimeout time.Duration
	// StartupTimeout bounds how long Start waits for the READY line.
	StartupTimeout time.Duration
	// ExtraEnv is appended to the subprocess environment.
	ExtraEnv []string
}

// MetricsRecorder observes adapter error paths. A nil value is supported.
type MetricsRecorder interface {
	RecordEmbedderError(reason string)
	RecordEmbedderRequest(textCount int, elapsed time.Duration)
}

type noopMetricsRecorder struct{}

func (noopMetricsRecorder) RecordEmbedderError(string)                  {}
func (noopMetricsRecorder) RecordEmbedderRequest(int, time.Duration)    {}

// Client runs the worker subprocess and serialises Embed calls to it.
type Client struct {
	cfg     Config
	metrics MetricsRecorder

	mu      sync.Mutex
	cmd     *exec.Cmd
	conn    net.Conn
	reader  *bufio.Reader
	started bool
	closed  bool
}

// New returns a configured client. It does not start the subprocess; call
// Start before the first Embed.
func New(cfg Config, metrics MetricsRecorder) *Client {
	if cfg.PythonPath == "" {
		cfg.PythonPath = "python3"
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = defaultBatchSize
	}
	if cfg.OperationTimeout <= 0 {
		cfg.OperationTimeout = defaultOperationTimeout
	}
	if cfg.StartupTimeout <= 0 {
		cfg.StartupTimeout = defaultStartupTimeout
	}
	if metrics == nil {
		metrics = noopMetricsRecorder{}
	}
	return &Client{cfg: cfg, metrics: metrics}
}

// Start spawns the worker subprocess, waits for the READY handshake, then
// dials the Unix domain socket. Returns ErrEmbedderUnavailable on any
// failure so callers can degrade gracefully.
func (c *Client) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.started {
		return nil
	}
	if c.closed {
		return application.ErrEmbedderUnavailable
	}
	if strings.TrimSpace(c.cfg.WorkerScript) == "" || strings.TrimSpace(c.cfg.SocketPath) == "" {
		c.metrics.RecordEmbedderError("config")
		return application.ErrEmbedderUnavailable
	}

	_ = os.Remove(c.cfg.SocketPath)

	cmd := exec.Command(c.cfg.PythonPath, c.cfg.WorkerScript)
	cmd.Env = append(os.Environ(),
		"RECOMMENDATION_EMBEDDER_SOCKET="+c.cfg.SocketPath,
	)
	if strings.TrimSpace(c.cfg.ModelName) != "" {
		cmd.Env = append(cmd.Env, "RECOMMENDATION_EMBEDDER_MODEL="+c.cfg.ModelName)
	}
	cmd.Env = append(cmd.Env, c.cfg.ExtraEnv...)
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		c.metrics.RecordEmbedderError("stdout_pipe")
		return application.ErrEmbedderUnavailable
	}
	if err := cmd.Start(); err != nil {
		log.Printf("recommendation embedder: start failed: %v", err)
		c.metrics.RecordEmbedderError("spawn")
		return application.ErrEmbedderUnavailable
	}

	ready := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			if strings.TrimSpace(scanner.Text()) == readyLine {
				ready <- nil
				return
			}
		}
		if err := scanner.Err(); err != nil {
			ready <- err
			return
		}
		ready <- io.EOF
	}()

	waitCtx, cancel := context.WithTimeout(ctx, c.cfg.StartupTimeout)
	defer cancel()
	select {
	case err := <-ready:
		if err != nil {
			log.Printf("recommendation embedder: handshake failed: %v", err)
			_ = cmd.Process.Kill()
			c.metrics.RecordEmbedderError("handshake")
			return application.ErrEmbedderUnavailable
		}
	case <-waitCtx.Done():
		log.Printf("recommendation embedder: handshake timeout after %s", c.cfg.StartupTimeout)
		_ = cmd.Process.Kill()
		c.metrics.RecordEmbedderError("startup_timeout")
		return application.ErrEmbedderUnavailable
	}

	conn, err := net.DialTimeout("unix", c.cfg.SocketPath, c.cfg.OperationTimeout)
	if err != nil {
		log.Printf("recommendation embedder: dial socket: %v", err)
		_ = cmd.Process.Kill()
		c.metrics.RecordEmbedderError("dial")
		return application.ErrEmbedderUnavailable
	}

	c.cmd = cmd
	c.conn = conn
	c.reader = bufio.NewReader(conn)
	c.started = true
	log.Printf("recommendation embedder: ready model=%s socket=%s", c.cfg.ModelName, c.cfg.SocketPath)
	return nil
}

// Embed sends texts to the worker and returns the resulting vectors. Larger
// inputs are chunked by BatchSize. Any I/O or decode error returns
// ErrEmbedderUnavailable and marks the client unhealthy so the next call
// short-circuits without touching the dead socket.
func (c *Client) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	out := make([][]float32, 0, len(texts))
	for start := 0; start < len(texts); start += c.cfg.BatchSize {
		end := start + c.cfg.BatchSize
		if end > len(texts) {
			end = len(texts)
		}
		chunk, err := c.embedChunk(ctx, texts[start:end])
		if err != nil {
			return nil, err
		}
		out = append(out, chunk...)
	}
	return out, nil
}

func (c *Client) embedChunk(ctx context.Context, texts []string) ([][]float32, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.started || c.conn == nil || c.closed {
		return nil, application.ErrEmbedderUnavailable
	}

	deadline, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		deadline = time.Now().Add(c.cfg.OperationTimeout)
	}
	if err := c.conn.SetDeadline(deadline); err != nil {
		c.markUnhealthy("set_deadline")
		return nil, application.ErrEmbedderUnavailable
	}

	payload, err := json.Marshal(embedRequest{Texts: texts})
	if err != nil {
		c.metrics.RecordEmbedderError("encode")
		return nil, application.ErrEmbedderUnavailable
	}
	payload = append(payload, '\n')

	started := time.Now()
	if _, err := c.conn.Write(payload); err != nil {
		log.Printf("recommendation embedder: write failed: %v", err)
		c.markUnhealthy("write")
		return nil, application.ErrEmbedderUnavailable
	}

	line, err := c.reader.ReadBytes('\n')
	if err != nil {
		log.Printf("recommendation embedder: read failed: %v", err)
		c.markUnhealthy("read")
		return nil, application.ErrEmbedderUnavailable
	}

	var resp embedResponse
	if err := json.Unmarshal(line, &resp); err != nil {
		log.Printf("recommendation embedder: decode failed: %v", err)
		c.metrics.RecordEmbedderError("decode")
		return nil, application.ErrEmbedderUnavailable
	}
	if resp.Error != "" {
		log.Printf("recommendation embedder: worker error: %s", resp.Error)
		c.metrics.RecordEmbedderError("worker")
		return nil, application.ErrEmbedderUnavailable
	}
	if len(resp.Embeddings) != len(texts) {
		c.metrics.RecordEmbedderError("shape")
		return nil, fmt.Errorf("%w: worker returned %d vectors for %d texts", application.ErrEmbedderUnavailable, len(resp.Embeddings), len(texts))
	}

	c.metrics.RecordEmbedderRequest(len(texts), time.Since(started))
	return resp.Embeddings, nil
}

// Close stops the worker subprocess. Safe to call multiple times.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true

	var firstErr error
	if c.conn != nil {
		if err := c.conn.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			firstErr = err
		}
		c.conn = nil
	}
	if c.cmd != nil && c.cmd.Process != nil {
		if err := c.cmd.Process.Signal(os.Interrupt); err != nil {
			_ = c.cmd.Process.Kill()
		}
		_, _ = c.cmd.Process.Wait()
	}
	_ = os.Remove(c.cfg.SocketPath)
	return firstErr
}

// markUnhealthy tears down the current connection so subsequent Embed calls
// short-circuit to ErrEmbedderUnavailable. Must be called with c.mu held.
func (c *Client) markUnhealthy(reason string) {
	c.metrics.RecordEmbedderError(reason)
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
	c.started = false
}

type embedRequest struct {
	Texts []string `json:"texts"`
}

type embedResponse struct {
	Embeddings [][]float32 `json:"embeddings,omitempty"`
	Error      string      `json:"error,omitempty"`
}
