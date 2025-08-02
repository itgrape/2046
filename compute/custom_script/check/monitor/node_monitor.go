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
	SocketPath                    = "/var/run/node_monitor.sock"
	HeartbeatTimeout              = 180 * time.Second
	HeartbeatCheckInterval        = 60 * time.Second
	GpuUtilizationThreshold       = 70.0
	GpuMemoryUtilizationThreshold = 30.0
	CpuUtilizationThreshold       = 50.0
	BufferPeriod                  = 20 // 丢弃前 20 分钟的使用率数据
)

var globalLogger *log.Logger

// JobInfo 存储每个被监控作业的详细信息
type JobInfo struct {
	LastHeartbeat         time.Time
	GpuMonitorCount       int
	CpuMonitorCount       int
	GpuUtilizations       []float64
	GpuMemoryUtilizations []float64
	CpuUtilizations       []float64
	LogPath               string      // 作业日志文件的路径
	Logger                *log.Logger // 此作业专用的logger
	LogFile               *os.File    // 日志文件句柄，用于最后关闭
	MetricsReceived       int         // 记录已接收到的指标数量
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
	GpuUtilization       float64 `json:"gpu_utilization"`
	GpuMemoryUtilization float64 `json:"gpu_memory_utilization"`
	CpuUtilization       float64 `json:"cpu_utilization"`
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

// removeJob 安全地移除一个作业并关闭其日志文件。此函数必须在持有锁的情况下调用。
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
			// 连接断开或解码错误
			return
		}

		switch msg.Type {
		case "REGISTER":
			jt.mu.Lock()
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

			logFile, err := os.OpenFile(payload.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				globalLogger.Printf("Failed to open log file %s for job %s: %v", payload.LogPath, msg.JobID, err)
				jt.mu.Unlock()
				continue
			}
			jobLogger := log.New(logFile, fmt.Sprintf("JOB-%s: ", msg.JobID), log.LstdFlags)

			globalLogger.Printf("Registering job %s with GPU-Count: %d, CPU-Count: %d", msg.JobID, payload.GpuMonitorCount, payload.CpuMonitorCount)
			jt.jobs[msg.JobID] = &JobInfo{
				LastHeartbeat:         time.Now(),
				GpuMonitorCount:       payload.GpuMonitorCount,
				CpuMonitorCount:       payload.CpuMonitorCount,
				GpuUtilizations:       make([]float64, 0, payload.GpuMonitorCount),
				GpuMemoryUtilizations: make([]float64, 0, payload.GpuMonitorCount),
				CpuUtilizations:       make([]float64, 0, payload.CpuMonitorCount),
				LogPath:               payload.LogPath,
				Logger:                jobLogger,
				LogFile:               logFile,
			}
			jt.mu.Unlock() // 在网络写入前释放锁

			// 回复客户端
			if _, err := conn.Write([]byte(`{"status": "ok"}\n`)); err != nil {
				globalLogger.Printf("Error writing OK status to client for job %s: %v", msg.JobID, err)
				return
			}

		case "METRICS":
			var payload MetricsPayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				globalLogger.Printf("Failed to unmarshal METRICS payload for job %s: %v", msg.JobID, err)
				continue
			}

			var jobToKillID, reason string

			jt.mu.Lock()
			job, jobExists := jt.jobs[msg.JobID]
			if !jobExists {
				jt.mu.Unlock()
				continue
			}

			job.LastHeartbeat = time.Now()
			job.MetricsReceived++
			globalLogger.Printf("Metrics received: JobID=%s, CPU=%.1f%%, GPU_Util=%.1f%%, GPU_Mem=%.1f%%", msg.JobID, payload.CpuUtilization, payload.GpuUtilization, payload.GpuMemoryUtilization)

			if job.MetricsReceived <= BufferPeriod {
				globalLogger.Printf("Discarding metrics during buffer period for job %s (Received: %d, Buffer: %d)", msg.JobID, job.MetricsReceived, BufferPeriod)
				jt.mu.Unlock()
				continue
			}

			// --- 检查所有指标，决定是否终止作业 ---
			// GPU 利用率
			if job.GpuMonitorCount > 0 && reason == "" {
				job.GpuUtilizations = append(job.GpuUtilizations, payload.GpuUtilization)
				for len(job.GpuUtilizations) > job.GpuMonitorCount {
					job.GpuUtilizations = job.GpuUtilizations[1:]
				}
				if len(job.GpuUtilizations) == job.GpuMonitorCount {
					avgGpu := calculateAverage(job.GpuUtilizations)
					globalLogger.Printf("Job %s, Average GPU Utilization: %.2f%%", msg.JobID, avgGpu)
					if avgGpu < GpuUtilizationThreshold {
						reason = fmt.Sprintf("Average GPU utilization %.2f%% is below threshold %.0f%%", avgGpu, GpuUtilizationThreshold)
					}
				}
			}
			// GPU 内存利用率
			if job.GpuMonitorCount > 0 && reason == "" {
				job.GpuMemoryUtilizations = append(job.GpuMemoryUtilizations, payload.GpuMemoryUtilization)
				for len(job.GpuMemoryUtilizations) > job.GpuMonitorCount {
					job.GpuMemoryUtilizations = job.GpuMemoryUtilizations[1:]
				}
				if len(job.GpuMemoryUtilizations) == job.GpuMonitorCount {
					avgGpuMem := calculateAverage(job.GpuMemoryUtilizations)
					globalLogger.Printf("Job %s, Average GPU Memory Utilization: %.2f%%", msg.JobID, avgGpuMem)
					if avgGpuMem < GpuMemoryUtilizationThreshold {
						reason = fmt.Sprintf("Average GPU Memory utilization %.2f%% is below threshold %.0f%%", avgGpuMem, GpuMemoryUtilizationThreshold)
					}
				}
			}
			// CPU 利用率
			if job.CpuMonitorCount > 0 && reason == "" {
				job.CpuUtilizations = append(job.CpuUtilizations, payload.CpuUtilization)
				for len(job.CpuUtilizations) > job.CpuMonitorCount {
					job.CpuUtilizations = job.CpuUtilizations[1:]
				}
				if len(job.CpuUtilizations) == job.CpuMonitorCount {
					avgCpu := calculateAverage(job.CpuUtilizations)
					globalLogger.Printf("Job %s, Average CPU Utilization: %.2f%%", msg.JobID, avgCpu)
					if avgCpu < CpuUtilizationThreshold {
						reason = fmt.Sprintf("Average CPU utilization %.2f%% is below threshold %.0f%%", avgCpu, CpuUtilizationThreshold)
					}
				}
			}

			// 如果决定终止，则在锁内移除作业
			if reason != "" {
				jobToKillID = msg.JobID
				jt.removeJob(jobToKillID, reason)
			}
			jt.mu.Unlock()

			// 在锁外执行耗时的 kill 操作
			if jobToKillID != "" {
				killSlurmJob(jobToKillID, reason)
				return // 终止作业后，关闭此连接
			}

		case "CANCEL":
			jt.mu.Lock()
			jobID := msg.JobID
			if _, ok := jt.jobs[jobID]; ok {
				reason := "Job cancelled by user request"
				globalLogger.Printf("Received cancellation for job %s.", jobID)
				jt.removeJob(jobID, reason)
			} else {
				globalLogger.Printf("Received cancellation for an unknown or already removed job %s.", jobID)
			}
			jt.mu.Unlock()
			return
		}
	}
}

func (jt *JobTracker) runStatusChecker() {
	ticker := time.NewTicker(HeartbeatCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		globalLogger.Println("--- Running Heartbeat Status Check ---")

		jobsToKill := make(map[string]string) // [jobID] -> reason

		// 在锁内识别超时的任务，并从主 map 中移除
		jt.mu.Lock()
		for jobID, job := range jt.jobs {
			if time.Since(job.LastHeartbeat) > HeartbeatTimeout {
				reason := fmt.Sprintf("Heartbeat Timeout. Last heartbeat was %.0f seconds ago.", time.Since(job.LastHeartbeat).Seconds())
				jobsToKill[jobID] = reason
				jt.removeJob(jobID, reason) // 在锁内安全移除
			}
		}
		jt.mu.Unlock()

		// 在锁外执行所有耗时的 scancel 命令
		if len(jobsToKill) > 0 {
			globalLogger.Printf("Found %d jobs to kill due to timeout.", len(jobsToKill))
			for jobID, reason := range jobsToKill {
				killSlurmJob(jobID, reason)
			}
		}
		globalLogger.Println("--- Status Check Finished ---")
	}
}

func killSlurmJob(jobID, reason string) error {
	globalLogger.Printf("[KILL] Executing 'scancel' for job %s, Reason: %s", jobID, reason)

	cmd := exec.Command("scancel", jobID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		errMsg := fmt.Errorf("error running scancel for job %s: %v. Output: %s", jobID, err, string(output))
		globalLogger.Print(errMsg)
		return errMsg
	}

	globalLogger.Printf("Successfully ran scancel for job %s.", jobID)
	return nil
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
