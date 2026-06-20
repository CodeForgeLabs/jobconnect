package python

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"jobconnect/recommendation/internal/application"
)

// fakeWorker stands in for the Python subprocess in tests. It listens on a
// Unix socket and responds per a caller-supplied handler.
type fakeWorker struct {
	socketPath string
	listener   net.Listener

	mu         sync.Mutex
	conns      []net.Conn
	handler    func(req embedRequest) embedResponse
	requestLog [][]string
}

func newFakeWorker(t *testing.T, handler func(embedRequest) embedResponse) *fakeWorker {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "w.sock")
	ln, err := net.Listen("unix", path)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	w := &fakeWorker{socketPath: path, listener: ln, handler: handler}
	go w.serve()
	return w
}

func (w *fakeWorker) serve() {
	for {
		conn, err := w.listener.Accept()
		if err != nil {
			return
		}
		w.mu.Lock()
		w.conns = append(w.conns, conn)
		w.mu.Unlock()
		go w.handle(conn)
	}
}

func (w *fakeWorker) handle(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			return
		}
		var req embedRequest
		if err := json.Unmarshal(line, &req); err != nil {
			_, _ = conn.Write([]byte(fmt.Sprintf(`{"error":"decode %s"}`+"\n", err)))
			continue
		}
		w.mu.Lock()
		w.requestLog = append(w.requestLog, append([]string(nil), req.Texts...))
		handler := w.handler
		w.mu.Unlock()

		resp := handler(req)
		payload, _ := json.Marshal(resp)
		if _, err := conn.Write(append(payload, '\n')); err != nil {
			return
		}
	}
}

func (w *fakeWorker) close() {
	_ = w.listener.Close()
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, c := range w.conns {
		_ = c.Close()
	}
}

// startFakeClient wires the Client directly to a fake worker's socket without
// spawning a subprocess. It exercises the I/O path end-to-end.
func startFakeClient(t *testing.T, w *fakeWorker, metrics MetricsRecorder) *Client {
	t.Helper()
	c := New(Config{
		WorkerScript:     "unused",
		SocketPath:       w.socketPath,
		BatchSize:        2,
		OperationTimeout: 500 * time.Millisecond,
		StartupTimeout:   500 * time.Millisecond,
	}, metrics)

	conn, err := net.DialTimeout("unix", w.socketPath, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("dial fake worker: %v", err)
	}
	c.conn = conn
	c.reader = bufio.NewReader(conn)
	c.started = true
	return c
}

type recordingMetrics struct {
	mu       sync.Mutex
	errors   []string
	requests []int
}

func (r *recordingMetrics) RecordEmbedderError(reason string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errors = append(r.errors, reason)
}

func (r *recordingMetrics) RecordEmbedderRequest(count int, _ time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests = append(r.requests, count)
}

func TestEmbedRoundTrip(t *testing.T) {
	w := newFakeWorker(t, func(req embedRequest) embedResponse {
		out := make([][]float32, len(req.Texts))
		for i := range req.Texts {
			out[i] = []float32{float32(i), float32(i + 1)}
		}
		return embedResponse{Embeddings: out}
	})
	defer w.close()

	metrics := &recordingMetrics{}
	c := startFakeClient(t, w, metrics)
	defer c.Close()

	got, err := c.Embed(context.Background(), []string{"a", "b"})
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}
	if len(got) != 2 || got[0][0] != 0 || got[1][1] != 2 {
		t.Fatalf("unexpected vectors: %v", got)
	}
	if len(metrics.requests) != 1 || metrics.requests[0] != 2 {
		t.Fatalf("metrics.requests = %v", metrics.requests)
	}
	if len(metrics.errors) != 0 {
		t.Fatalf("unexpected errors: %v", metrics.errors)
	}
}

func TestEmbedEmptyInput(t *testing.T) {
	w := newFakeWorker(t, func(embedRequest) embedResponse { return embedResponse{} })
	defer w.close()
	c := startFakeClient(t, w, nil)
	defer c.Close()

	got, err := c.Embed(context.Background(), nil)
	if err != nil || got != nil {
		t.Fatalf("Embed nil = (%v, %v), want (nil, nil)", got, err)
	}
}

func TestEmbedChunksByBatchSize(t *testing.T) {
	var callCount int
	var seenBatches [][]string
	var mu sync.Mutex
	w := newFakeWorker(t, func(req embedRequest) embedResponse {
		mu.Lock()
		callCount++
		seenBatches = append(seenBatches, append([]string(nil), req.Texts...))
		mu.Unlock()
		out := make([][]float32, len(req.Texts))
		for i := range req.Texts {
			out[i] = []float32{1}
		}
		return embedResponse{Embeddings: out}
	})
	defer w.close()

	c := startFakeClient(t, w, nil)
	defer c.Close()

	_, err := c.Embed(context.Background(), []string{"a", "b", "c", "d", "e"})
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}
	if callCount != 3 {
		t.Fatalf("callCount = %d, want 3 (batch size 2 over 5 texts)", callCount)
	}
	if len(seenBatches[0]) != 2 || len(seenBatches[1]) != 2 || len(seenBatches[2]) != 1 {
		t.Fatalf("unexpected batch shapes: %v", seenBatches)
	}
}

func TestEmbedWorkerErrorReturnsSentinel(t *testing.T) {
	w := newFakeWorker(t, func(embedRequest) embedResponse {
		return embedResponse{Error: "oom"}
	})
	defer w.close()
	metrics := &recordingMetrics{}
	c := startFakeClient(t, w, metrics)
	defer c.Close()

	_, err := c.Embed(context.Background(), []string{"x"})
	if !errors.Is(err, application.ErrEmbedderUnavailable) {
		t.Fatalf("err = %v, want ErrEmbedderUnavailable", err)
	}
	if len(metrics.errors) != 1 || metrics.errors[0] != "worker" {
		t.Fatalf("metrics.errors = %v", metrics.errors)
	}
}

func TestEmbedShapeMismatchReturnsSentinel(t *testing.T) {
	w := newFakeWorker(t, func(embedRequest) embedResponse {
		return embedResponse{Embeddings: [][]float32{{1}}}
	})
	defer w.close()
	metrics := &recordingMetrics{}
	c := startFakeClient(t, w, metrics)
	defer c.Close()

	_, err := c.Embed(context.Background(), []string{"a", "b"})
	if !errors.Is(err, application.ErrEmbedderUnavailable) {
		t.Fatalf("err = %v, want ErrEmbedderUnavailable", err)
	}
	if len(metrics.errors) != 1 || metrics.errors[0] != "shape" {
		t.Fatalf("metrics.errors = %v", metrics.errors)
	}
}

func TestEmbedMarksUnhealthyOnIOError(t *testing.T) {
	w := newFakeWorker(t, func(embedRequest) embedResponse {
		return embedResponse{Embeddings: [][]float32{{1, 2}}}
	})
	metrics := &recordingMetrics{}
	c := startFakeClient(t, w, metrics)
	defer c.Close()

	if _, err := c.Embed(context.Background(), []string{"a"}); err != nil {
		t.Fatalf("first Embed: %v", err)
	}

	w.close()

	_, err := c.Embed(context.Background(), []string{"b"})
	if !errors.Is(err, application.ErrEmbedderUnavailable) {
		t.Fatalf("err = %v, want ErrEmbedderUnavailable", err)
	}

	// Subsequent call short-circuits without touching the dead socket.
	_, err = c.Embed(context.Background(), []string{"c"})
	if !errors.Is(err, application.ErrEmbedderUnavailable) {
		t.Fatalf("err = %v, want ErrEmbedderUnavailable", err)
	}
}

func TestStartReturnsSentinelOnEmptyConfig(t *testing.T) {
	c := New(Config{}, nil)
	if err := c.Start(context.Background()); !errors.Is(err, application.ErrEmbedderUnavailable) {
		t.Fatalf("err = %v, want ErrEmbedderUnavailable", err)
	}
}

func TestStartReturnsSentinelWhenPythonPathBogus(t *testing.T) {
	dir := t.TempDir()
	c := New(Config{
		PythonPath:     "/nonexistent/binary",
		WorkerScript:   filepath.Join(dir, "w.py"),
		SocketPath:     filepath.Join(dir, "w.sock"),
		StartupTimeout: 200 * time.Millisecond,
	}, nil)

	err := c.Start(context.Background())
	if !errors.Is(err, application.ErrEmbedderUnavailable) {
		t.Fatalf("err = %v, want ErrEmbedderUnavailable", err)
	}
}

// TestStartHandshakeAgainstRealSubprocess compiles a Go binary that mimics the
// Python worker handshake (prints READY, serves the socket), then drives the
// full Start → Embed path against it. Skipped when the Go toolchain is not on
// PATH so the package is still usable in minimal CI images.
func TestStartHandshakeAgainstRealSubprocess(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain unavailable")
	}

	dir := t.TempDir()
	src := filepath.Join(dir, "main.go")
	if err := writeFakeWorkerSource(src); err != nil {
		t.Fatalf("write fake worker: %v", err)
	}
	bin := filepath.Join(dir, "fakeworker")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	buildCmd := exec.Command("go", "build", "-o", bin, src)
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Skipf("go build: %v: %s", err, out)
	}

	c := New(Config{
		PythonPath:       bin,
		WorkerScript:     "ignored-by-fake-worker",
		SocketPath:       filepath.Join(dir, "w.sock"),
		BatchSize:        4,
		OperationTimeout: 2 * time.Second,
		StartupTimeout:   3 * time.Second,
	}, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer c.Close()

	got, err := c.Embed(context.Background(), []string{"x", "y", "z"})
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d vectors, want 3", len(got))
	}
}

func writeFakeWorkerSource(path string) error {
	src := `package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

func main() {
	sock := os.Getenv("RECOMMENDATION_EMBEDDER_SOCKET")
	if sock == "" {
		os.Exit(2)
	}
	_ = os.Remove(sock)
	ln, err := net.Listen("unix", sock)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Println("READY")
	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		go func(c net.Conn) {
			defer c.Close()
			r := bufio.NewReader(c)
			for {
				line, err := r.ReadBytes('\n')
				if err != nil {
					return
				}
				var req struct {
					Texts []string ` + "`json:\"texts\"`" + `
				}
				_ = json.Unmarshal(line, &req)
				out := make([][]float32, len(req.Texts))
				for i := range req.Texts {
					out[i] = []float32{float32(i)}
				}
				resp, _ := json.Marshal(map[string]any{"embeddings": out})
				c.Write(append(resp, '\n'))
			}
		}(conn)
	}
}
`
	return os.WriteFile(path, []byte(src), 0o644)
}
