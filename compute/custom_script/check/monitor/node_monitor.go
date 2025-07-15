package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

const (
	SocketPath              = "/var/run/node_monitor.sock"
	HeartbeatTimeout        = 180 * time.Second
	HeartbeatCheckInterval  = 60 * time.Second
	GpuUtilizationThreshold = 70.0
	CpuUtilizationThreshold = 50.0
)

var globalLogger *log.Logger

// JobInfo 存储每个被监控作业的详细信息
type JobInfo struct {
	LastHeartbeat   time.Time
	GpuMonitorCount int
	CpuMonitorCount int
	GpuUtilizations []float64
	CpuUtilizations []float64
	LogPath         string      // 作业日志文件的路径
	Logger          *log.Logger // 此作业专用的logger
	LogFile         *os.File    // 日志文件句柄，用于最后关闭
}

// JobTracker 安全地管理所有在本机运行的作业
type JobTracker struct {
	jobs map[string]*JobInfo
	mu   sync.Mutex
}

// Message 定义了客户端和守护进程之间的通信协议
type Message struct {
	Type    string          `json:"type"`
	JobID   string          `json:"job_id"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// RegisterPayload 定义了 REGISTER 消息的具体内容
type RegisterPayload struct {
	GpuMonitorCount int    `json:"gpu_monitor_count"`
	CpuMonitorCount int    `json:"cpu_monitor_count"`
	LogPath         string `json:"log_path"`
}

// MetricsPayload 定义了 METRICS 消息的具体内容
type MetricsPayload struct {
	GpuUtilization float64 `json:"gpu_utilization"`
	CpuUtilization float64 `json:"cpu_utilization"`
}

func main() {
	// 初始化全局logger，用于记录守护进程本身的状态
	globalLogger = log.New(os.Stdout, "DAEMON: ", log.LstdFlags)
	globalLogger.Println("Starting Node Monitor Daemon with per-job logging...")

	if err := os.MkdirAll(filepath.Dir(SocketPath), 0755); err != nil {
		globalLogger.Fatalf("Failed to create socket directory: %v", err)
	}
	if _, err := os.Stat(SocketPath); err == nil {
		if err := os.Remove(SocketPath); err != nil {
			globalLogger.Fatalf("Failed to remove existing socket: %v", err)
		}
	}

	tracker := &JobTracker{
		jobs: make(map[string]*JobInfo),
	}

	go tracker.runStatusChecker()

	listener, err := net.Listen("unix", SocketPath)
	if err != nil {
		globalLogger.Fatalf("Failed to listen on unix socket %s: %v", SocketPath, err)
	}
	defer listener.Close()

	// 修改 socket 文件的权限，允许所有用户连接
	if err := os.Chmod(SocketPath, 0777); err != nil {
		globalLogger.Fatalf("Failed to change socket permissions: %v", err)
	}
	globalLogger.Printf("Set socket %s permissions to 0777", SocketPath)

	globalLogger.Printf("Listening on %s", SocketPath)
	for {
		conn, err := listener.Accept()
		if err != nil {
			globalLogger.Printf("Failed to accept connection: %v", err)
			continue
		}
		go tracker.handleConnection(conn)
	}
}

// removeJob 安全地移除一个作业并关闭其日志文件
func (jt *JobTracker) removeJob(jobID string, reason string) {
	if job, ok := jt.jobs[jobID]; ok {
		job.Logger.Printf("Removing job %s. Reason: %s", jobID, reason)
		if job.LogFile != nil {
			job.LogFile.Close()
		}
		delete(jt.jobs, jobID)
	}
}

func (jt *JobTracker) handleConnection(conn net.Conn) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)

	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			return
		}

		jt.mu.Lock()

		switch msg.Type {
		case "REGISTER":
			var payload RegisterPayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				globalLogger.Printf("Failed to unmarshal REGISTER payload for job %s: %v", msg.JobID, err)
				jt.mu.Unlock()
				continue
			}
			if payload.LogPath == "" {
				globalLogger.Printf("Registration for job %s failed: log_path is missing.", msg.JobID)
				jt.mu.Unlock()
				continue
			}

			// 为此作业创建日志文件和logger
			logFile, err := os.OpenFile(payload.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				globalLogger.Printf("Failed to open log file %s for job %s: %v", payload.LogPath, msg.JobID, err)
				jt.mu.Unlock()
				continue
			}
			jobLogger := log.New(logFile, fmt.Sprintf("JOB-%s: ", msg.JobID), log.LstdFlags)

			globalLogger.Printf("Registering job %s with GPU-Count: %d, CPU-Count: %d", msg.JobID, payload.GpuMonitorCount, payload.CpuMonitorCount)
			jt.jobs[msg.JobID] = &JobInfo{
				LastHeartbeat:   time.Now(),
				GpuMonitorCount: payload.GpuMonitorCount,
				CpuMonitorCount: payload.CpuMonitorCount,
				GpuUtilizations: make([]float64, 0, payload.GpuMonitorCount),
				CpuUtilizations: make([]float64, 0, payload.CpuMonitorCount),
				LogPath:         payload.LogPath,
				Logger:          jobLogger,
				LogFile:         logFile,
			}
			// 回复客户端
			_, err = conn.Write([]byte(`{"status": "ok"}\n`))
			if err != nil {
				globalLogger.Printf("Error writing OK status to client for job %s: %v", msg.JobID, err)
				jt.mu.Unlock()
				return
			}

		case "METRICS":
			job, jobExists := jt.jobs[msg.JobID]
			if !jobExists {
				jt.mu.Unlock()
				continue
			}

			var payload MetricsPayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				globalLogger.Printf("Failed to unmarshal METRICS payload for job %s: %v", msg.JobID, err)
				jt.mu.Unlock()
				continue
			}

			job.LastHeartbeat = time.Now()
			globalLogger.Printf("Metrics received: JobID=%s, CPU=%.1f%%, GPU=%.1f%%", msg.JobID, payload.CpuUtilization, payload.GpuUtilization)

			// --- GPU利用率处理 ---
			if job.GpuMonitorCount > 0 {
				job.GpuUtilizations = append(job.GpuUtilizations, payload.GpuUtilization)

				// 滑动窗口技术，如果数据点超过了窗口大小，就从头部移除最旧的数据
				for len(job.GpuUtilizations) > job.GpuMonitorCount {
					job.GpuUtilizations = job.GpuUtilizations[1:]
				}

				if len(job.GpuUtilizations) == job.GpuMonitorCount {
					avgGpu := calculateAverage(job.GpuUtilizations)
					if avgGpu < GpuUtilizationThreshold {
						reason := fmt.Sprintf("Average GPU utilization %.2f%% is below threshold %.0f%%", avgGpu, GpuUtilizationThreshold)
						killSlurmJob(msg.JobID, reason, job.Logger)
						jt.removeJob(msg.JobID, reason)
						jt.mu.Unlock()
						return
					}
				}
			}

			// 确保作业还是存在的（没有因为 GPU 或心跳超时被移除）
			if _, ok := jt.jobs[msg.JobID]; !ok {
				jt.mu.Unlock()
				return
			}

			// --- CPU利用率处理 ---
			if job.CpuMonitorCount > 0 {
				job.CpuUtilizations = append(job.CpuUtilizations, payload.CpuUtilization)
				for len(job.CpuUtilizations) > job.CpuMonitorCount {
					job.CpuUtilizations = job.CpuUtilizations[1:]
				}

				if len(job.CpuUtilizations) == job.CpuMonitorCount {
					avgCpu := calculateAverage(job.CpuUtilizations)
					if avgCpu < CpuUtilizationThreshold {
						reason := fmt.Sprintf("Average CPU utilization %.2f%% is below threshold %.0f%%", avgCpu, CpuUtilizationThreshold)
						killSlurmJob(msg.JobID, reason, job.Logger)
						jt.removeJob(msg.JobID, reason)
						jt.mu.Unlock()
						return
					}
				}
			}
		}
		jt.mu.Unlock()
	}
}

func (jt *JobTracker) runStatusChecker() {
	ticker := time.NewTicker(HeartbeatCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		globalLogger.Println("--- Running Heartbeat Status Check ---")
		jt.mu.Lock()
		for jobID, job := range jt.jobs {
			if time.Since(job.LastHeartbeat) > HeartbeatTimeout {
				reason := fmt.Sprintf("Heartbeat Timeout. Last heartbeat was %.0f seconds ago.", time.Since(job.LastHeartbeat).Seconds())
				killSlurmJob(jobID, reason, job.Logger)
				jt.removeJob(jobID, reason)
				continue
			}
		}
		jt.mu.Unlock()
		globalLogger.Println("--- Status Check Finished ---")
	}
}

func killSlurmJob(jobID, reason string, logger *log.Logger) {
	logger.Printf("[KILL] Executing 'scancel' for job %s, Reason: %s", jobID, reason)
	globalLogger.Printf("[KILL] Executing 'scancel' for job %s, Reason: %s", jobID, reason)

	cmd := exec.Command("scancel", jobID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Printf("Error running scancel for job %s: %v. Output: %s", jobID, err, string(output))
	} else {
		logger.Printf("Successfully ran scancel for job %s.", jobID)
	}
}

func calculateAverage(slice []float64) float64 {
	if len(slice) == 0 {
		return 0
	}
	var total float64
	for _, v := range slice {
		total += v
	}
	return total / float64(len(slice))
}
