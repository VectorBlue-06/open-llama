package server

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"
)

// Config holds the server launch configuration.
type Config struct {
	BinaryPath string
	ModelPath  string
	Host       string
	Port       int
	CtxSize    int
	Threads    int
	GPULayers  int
	BatchSize  int
	ExtraArgs  []string
}

// Server manages the llama-server child process.
type Server struct {
	cmd     *exec.Cmd
	port    int
	logFile *os.File
}

// New creates a new Server instance.
func New() *Server {
	return &Server{}
}

// Start launches the llama-server process with the given configuration.
func (s *Server) Start(cfg Config, logPath string) error {
	if cfg.Port == 0 {
		port, err := FindFreePort()
		if err != nil {
			return fmt.Errorf("find free port: %w", err)
		}
		cfg.Port = port
	}
	s.port = cfg.Port

	args := buildArgs(cfg)

	s.cmd = exec.Command(cfg.BinaryPath, args...)
	s.cmd.Env = os.Environ()

	// Redirect server output to log file
	if logPath != "" {
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err == nil {
			s.logFile = f
			s.cmd.Stdout = f
			s.cmd.Stderr = f
		}
	}

	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("start llama-server: %w", err)
	}

	return nil
}

// WaitForReady polls the health endpoint until the server is ready.
func (s *Server) WaitForReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}

	for time.Now().Before(deadline) {
		resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/health", s.port))
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("server did not become ready within %v", timeout)
}

// Stop gracefully stops the server process.
func (s *Server) Stop() error {
	if s.cmd == nil || s.cmd.Process == nil {
		return nil
	}

	// Try graceful shutdown first
	if err := s.cmd.Process.Signal(os.Interrupt); err != nil {
		// If interrupt fails, force kill
		return s.cmd.Process.Kill()
	}

	// Wait up to 5 seconds for graceful shutdown
	done := make(chan error, 1)
	go func() {
		done <- s.cmd.Wait()
	}()

	select {
	case <-done:
		// Process exited
	case <-time.After(5 * time.Second):
		// Force kill
		s.cmd.Process.Kill()
		<-done
	}

	if s.logFile != nil {
		s.logFile.Close()
	}

	return nil
}

// Port returns the port the server is running on.
func (s *Server) Port() int {
	return s.port
}

// IsRunning returns true if the server process is still running.
func (s *Server) IsRunning() bool {
	if s.cmd == nil || s.cmd.Process == nil {
		return false
	}
	if s.cmd.ProcessState != nil {
		return false
	}
	return true
}

func buildArgs(cfg Config) []string {
	args := []string{
		"-m", cfg.ModelPath,
		"--host", cfg.Host,
		"--port", strconv.Itoa(cfg.Port),
		"--ctx-size", strconv.Itoa(cfg.CtxSize),
		"--threads", strconv.Itoa(cfg.Threads),
		"--batch-size", strconv.Itoa(cfg.BatchSize),
	}
	if cfg.GPULayers > 0 {
		args = append(args, "--n-gpu-layers", strconv.Itoa(cfg.GPULayers))
	}
	args = append(args, cfg.ExtraArgs...)
	return args
}
