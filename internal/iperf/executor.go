package iperf

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wold9168/iperf-rpc/internal/model"
)

type Executor struct {
	mu            sync.Mutex
	serverCmd     *exec.Cmd
	serverPort    int
	serverRunning bool
	results       map[string]*model.IperfRunData
}

func New() *Executor {
	return &Executor{
		results: make(map[string]*model.IperfRunData),
	}
}

func (e *Executor) Run(req *model.IperfRunRequest) *model.IperfRunData {
	id := uuid.New().String()
	now := time.Now()

	data := &model.IperfRunData{
		ID:        id,
		Status:    "running",
		StartedAt: now,
	}

	e.mu.Lock()
	e.results[id] = data
	e.mu.Unlock()

	args := e.buildArgs(req)
	cmd := exec.Command("iperf3", args...)
	data.Command = "iperf3 " + strings.Join(args, " ")

	output, err := cmd.CombinedOutput()
	finished := time.Now()
	data.FinishedAt = &finished
	data.Output = string(output)

	if err != nil {
		data.Status = "error"
	} else {
		data.Status = "completed"
		data.Summary = e.parseSummary(data.Output)
	}

	return data
}

func (e *Executor) buildArgs(req *model.IperfRunRequest) []string {
	args := []string{}

	if req.Mode == "client" {
		args = append(args, "-c", req.Args.Target)
	} else {
		args = append(args, "-s", "-1")
	}

	if req.Args.Port > 0 {
		args = append(args, "-p", strconv.Itoa(req.Args.Port))
	}

	if req.Mode == "client" {
		if req.Args.Duration > 0 && req.Args.Duration != 10 {
			args = append(args, "-t", strconv.Itoa(req.Args.Duration))
		}
		if req.Args.Parallel > 1 {
			args = append(args, "-P", strconv.Itoa(req.Args.Parallel))
		}
		if req.Args.Bandwidth != "" {
			args = append(args, "-b", req.Args.Bandwidth)
		}
		if req.Args.Protocol == "udp" {
			args = append(args, "-u")
		}
		if req.Args.Reverse {
			args = append(args, "-R")
		}
		if req.Args.Extra != "" {
			args = append(args, strings.Fields(req.Args.Extra)...)
		}
	}

	args = append(args, "-J")
	return args
}

func (e *Executor) parseSummary(output string) *model.IperfSummary {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil
	}

	summary := &model.IperfSummary{}

	end, ok := result["end"].(map[string]interface{})
	if !ok {
		return nil
	}

	if sumSent, ok := end["sum_sent"].(map[string]interface{}); ok {
		if bps, ok := sumSent["bits_per_second"].(float64); ok {
			summary.BitrateBps = bps
		}
		if dur, ok := sumSent["seconds"].(float64); ok {
			summary.DurationSec = dur
		}
	} else if sumRecv, ok := end["sum_received"].(map[string]interface{}); ok {
		if bps, ok := sumRecv["bits_per_second"].(float64); ok {
			summary.BitrateBps = bps
		}
		if dur, ok := sumRecv["seconds"].(float64); ok {
			summary.DurationSec = dur
		}
	}

	if sum, ok := end["sum"].(map[string]interface{}); ok {
		if jitter, ok := sum["jitter_ms"].(float64); ok {
			summary.JitterMs = jitter
		}
		if lost, ok := sum["lost_percent"].(float64); ok {
			summary.LostPercent = lost
		}
	}

	return summary
}

func (e *Executor) StartServer(port int) (int, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.serverRunning {
		return 0, fmt.Errorf("iperf3 server already running on port %d", e.serverPort)
	}

	portStr := strconv.Itoa(port)
	cmd := exec.Command("iperf3", "-s", "-p", portStr, "-D")
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start iperf3 server: %w", err)
	}

	e.serverCmd = cmd
	e.serverPort = port
	e.serverRunning = true

	return cmd.Process.Pid, nil
}

func (e *Executor) StopServer() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.serverRunning {
		return fmt.Errorf("iperf3 server is not running")
	}

	if err := e.serverCmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill iperf3 server: %w", err)
	}

	e.serverCmd = nil
	e.serverRunning = false
	return nil
}

func (e *Executor) ServerStatus() *model.ServerStatusData {
	e.mu.Lock()
	defer e.mu.Unlock()

	return &model.ServerStatusData{
		Running: e.serverRunning,
		Port:    e.serverPort,
		PID:     func() int {
			if e.serverCmd != nil && e.serverCmd.Process != nil {
				return e.serverCmd.Process.Pid
			}
			return 0
		}(),
	}
}

func (e *Executor) GetResult(id string) *model.IperfRunData {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.results[id]
}
