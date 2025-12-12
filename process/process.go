package process

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"

	"github.com/kuo-hm/devdeck/config"
	ps "github.com/shirou/gopsutil/v3/process"
)

// Process represents a running task with its configuration and state.
type Process struct {
	Config    config.Task
	Cmd       *exec.Cmd
	Status    string
	Output    chan string
	Err       error
	LogBuffer string
	Stdin     io.WriteCloser

	CPUUsage float64
	MemUsage uint64
	gopsProc *ps.Process

	HealthStatus string // "Unchecked", "Healthy", "Unhealthy", "Starting"

	epoch int32 // For handling restart races
}

// NewProcess creates a new Process instance from a task configuration.
func NewProcess(cfg config.Task) *Process {
	return &Process{
		Config:       cfg,
		Status:       "Stopped",
		HealthStatus: "Unchecked",
		Output:       make(chan string, 1000),
	}
}

// Start executes the process command and begins streaming output.
func (p *Process) Start() error {
	p.Err = nil
	parts := strings.Fields(p.Config.Command)
	if len(parts) == 0 {
		return nil
	}

	c := exec.Command(parts[0], parts[1:]...)
	if p.Config.Directory != "" {
		c.Dir = p.Config.Directory
	}
	c.Env = os.Environ()
	c.Env = append(c.Env, p.Config.Env...)

	stdin, err := c.StdinPipe()
	if err != nil {
		p.Status = "Error"
		p.Err = err
		return err
	}
	p.Stdin = stdin

	stdout, err := c.StdoutPipe()
	if err != nil {
		p.Status = "Error"
		p.Err = err
		return err
	}

	stderr, err := c.StderrPipe()
	if err != nil {
		p.Status = "Error"
		p.Err = err
		return err
	}

	if err := c.Start(); err != nil {
		p.Status = "Error"
		p.Err = err
		return err
	}

	p.Cmd = c
	p.Status = "Running"

	// Create resource monitor handle
	if c.Process != nil {
		p.gopsProc, _ = ps.NewProcess(int32(c.Process.Pid))
	}

	// Start Health Check Loop
	if p.Config.HealthCheck != nil {
		go p.monitorHealth()
	}

	consume := func(r *bufio.Scanner) {
		for r.Scan() {
			p.Output <- r.Text()
		}
	}

	go consume(bufio.NewScanner(stdout))
	go consume(bufio.NewScanner(stderr))

	// Atomic Epoch (Wait Logic)
	currentEpoch := atomic.AddInt32(&p.epoch, 1)

	go func() {
		err := c.Wait()

		// Only update status if we are still in the same epoch
		if atomic.LoadInt32(&p.epoch) == currentEpoch {
			if err != nil {
				p.Status = "Error"
				p.Err = err
			} else {
				p.Status = "Stopped"
			}
		}
	}()

	return nil
}

// Stop terminates the running process.
func (p *Process) Stop() error {
	if p.Cmd != nil && p.Cmd.Process != nil {
		return p.Cmd.Process.Kill()
	}
	return nil
}

// Restart stops and then starts the process.
func (p *Process) Restart() error {
	if p.Status == "Running" {
		_ = p.Stop()
	}
	return p.Start()
}

// SendInput writes the input string to the process stdin.
func (p *Process) SendInput(input string) error {
	if p.Status != "Running" || p.Stdin == nil {
		return nil
	}
	_, err := io.WriteString(p.Stdin, input+"\n")
	return err
}

// UpdateStats fetches current resource usage for the process.
func (p *Process) UpdateStats() {
	if p.Status != "Running" || p.gopsProc == nil {
		p.CPUUsage = 0
		p.MemUsage = 0
		return
	}

	cpuPercent, err := p.gopsProc.Percent(0)
	if err == nil {
		p.CPUUsage = cpuPercent
	}

	memInfo, err := p.gopsProc.MemoryInfo()
	if err == nil {
		p.MemUsage = memInfo.RSS // Resident Set Size in bytes
	}
}

// monitorHealth runs periodically to check service status
func (p *Process) monitorHealth() {
	hc := p.Config.HealthCheck
	interval := time.Duration(hc.Interval) * time.Millisecond
	if interval == 0 {
		interval = 2000 * time.Millisecond
	} // Default 2s

	for p.Status == "Running" {
		if p.checkHealth() {
			p.HealthStatus = "Healthy"
		} else {
			p.HealthStatus = "Unhealthy"
		}
		time.Sleep(interval)
	}
	p.HealthStatus = "Unchecked"
}

func (p *Process) checkHealth() bool {
	hc := p.Config.HealthCheck
	if hc == nil {
		return true
	}

	timeout := time.Duration(hc.Timeout) * time.Millisecond
	if timeout == 0 {
		timeout = 1000 * time.Millisecond
	}

	if hc.Type == "tcp" {
		conn, err := net.DialTimeout("tcp", hc.Target, timeout)
		if err != nil {
			return false
		}
		conn.Close()
		return true
	} else if hc.Type == "http" {
		client := http.Client{Timeout: timeout}
		resp, err := client.Get(hc.Target)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode >= 200 && resp.StatusCode < 400
	}
	return false
}
