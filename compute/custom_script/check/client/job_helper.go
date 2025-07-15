package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
)

const SocketPath = "/var/run/node_monitor.sock"

const (
	monitorInterval = 60 // 监控间隔（秒）

	rtx5090CheckCount  = 10
	rtxA6000CheckCount = 20
	rtx4090CheckCount  = 20
	rtx3090CheckCount  = 60
	rtxA10CheckCount   = 60

	defaultGpuCheckCount = 60
	defaultCpuCheckCount = 60

	infiniteCheckCount = 144000 // 最大运行 100 天
)

var (
	cpuCheckCount = 60 // CPU默认检查次数
	gpuCheckCount = 60 // GPU默认检查次数
)

// Message 定义（与守护进程中的一致）
type Message struct {
	Type    string          `json:"type"`
	JobID   string          `json:"job_id"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// RegisterPayload 定义（与守护进程中的一致）
type RegisterPayload struct {
	GpuMonitorCount int    `json:"gpu_monitor_count"`
	CpuMonitorCount int    `json:"cpu_monitor_count"`
	LogPath         string `json:"log_path"`
}

// MetricsPayload 定义（与守护进程中的一致）
type MetricsPayload struct {
	GpuUtilization float64 `json:"gpu_utilization"`
	CpuUtilization float64 `json:"cpu_utilization"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if len(os.Args) < 2 {
		printUsageAndExit()
	}

	cmd := os.Args[1]
	jobID := os.Getenv("SLURM_JOB_ID")
	if jobID == "" {
		log.Fatal("SLURM_JOB_ID environment variable not set. This script must be run inside a Slurm job.")
	}

	switch cmd {
	case "register":
		if len(os.Args) != 3 {
			log.Println("Error: Incorrect number of arguments for 'register'.")
			printUsageAndExit()
		}
		register(jobID, os.Args[2])
	case "monitor":
		monitor(jobID)
	default:
		log.Printf("Error: Unknown command '%s'.", cmd)
		printUsageAndExit()
	}
}

func printUsageAndExit() {
	fmt.Fprintf(os.Stderr, "Usage: %s <command> [arguments]\n\n", os.Args[0])
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  register <log_path>")
	fmt.Fprintln(os.Stderr, "    Registers the job. GPU monitoring count is determined automatically based on card type.")
	fmt.Fprintln(os.Stderr, "    <log_path>: Absolute path for the job's monitoring log file.")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "  monitor")
	fmt.Fprintln(os.Stderr, "    Starts monitoring and sending metrics to the daemon.")
	fmt.Fprintln(os.Stderr, "")
	os.Exit(1)
}

func register(jobID string, logPath string) {
	jobPartition := os.Getenv("SLURM_JOB_PARTITION")
	lowerJobPartition := strings.ToLower(jobPartition)

	// 计算监控次数
	if strings.Contains(lowerJobPartition, "debug") {
		gpuCheckCount = infiniteCheckCount
		cpuCheckCount = infiniteCheckCount
		log.Printf("Detected debug partition: %s. Setting CPU and GPU check count to %d.", jobPartition, infiniteCheckCount)
	} else {
		gpuCheckCount = determineGpuCheckCount()
		cpuCheckCount = defaultCpuCheckCount
		log.Printf("Dynamically determined GPU monitoring count: %d", gpuCheckCount)
	}

	conn, err := net.Dial("unix", SocketPath)
	if err != nil {
		log.Fatalf("Failed to connect to node monitor daemon: %v", err)
	}
	defer conn.Close()

	regPayload := RegisterPayload{
		GpuMonitorCount: gpuCheckCount,
		CpuMonitorCount: cpuCheckCount,
		LogPath:         logPath,
	}

	payloadBytes, err := json.Marshal(regPayload)
	if err != nil {
		log.Fatalf("Failed to create registration payload: %v", err)
	}

	msg := Message{Type: "REGISTER", JobID: jobID, Payload: payloadBytes}
	if err := json.NewEncoder(conn).Encode(msg); err != nil {
		log.Fatalf("Failed to send registration request: %v", err)
	}

	var resp map[string]string
	if err := json.NewDecoder(conn).Decode(&resp); err != nil || resp["status"] != "ok" {
		log.Fatalf("Registration failed or confirmation not received. Error: %v, Response: %v", err, resp)
	}

	log.Printf("Job %s registered successfully. GPU_Count=%d, CPU_Count=%d, Log file: %s", jobID, gpuCheckCount, cpuCheckCount, logPath)
	os.Exit(0)
}

// determineGpuCheckCount 检查节点上所有GPU并返回最少的监控次数
func determineGpuCheckCount() int {
	bestCheckCount := defaultGpuCheckCount

	// 执行nvidia-smi命令获取所有GPU的索引和名称
	cmd := exec.Command("nvidia-smi", "--query-gpu=index,name", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Warning: 'nvidia-smi' command failed: %v. Assuming no GPUs or driver issue. Setting GPU check count to infinite.", err)
		// 在没有GPU或驱动有问题的节点，直接不监控（返回 infiniteCheckCount）
		return infiniteCheckCount
	}

	// 按行解析输出
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		log.Println("No GPUs detected by nvidia-smi. Setting GPU check count to infinite.")
		return infiniteCheckCount
	}

	for _, line := range lines {
		parts := strings.Split(line, ",")
		if len(parts) < 2 {
			continue
		}
		gpuIndex := strings.TrimSpace(parts[0])
		gpuName := strings.ToUpper(strings.TrimSpace(parts[1]))

		currentCheckCount := getGpuCheckCount(gpuName)
		log.Printf("  - Detected GPU %s: %s", gpuIndex, gpuName)

		if bestCheckCount > currentCheckCount {
			bestCheckCount = currentCheckCount
		}
	}

	log.Printf("Setting monitoring count to: %d", bestCheckCount)
	return bestCheckCount
}

// getGpuCheckCount 根据GPU名称返回一个监控次数
func getGpuCheckCount(gpuName string) int {
	switch {
	case strings.Contains(gpuName, "5090"):
		return rtx5090CheckCount
	case strings.Contains(gpuName, "A6000"):
		return rtxA6000CheckCount
	case strings.Contains(gpuName, "4090"):
		return rtx4090CheckCount
	case strings.Contains(gpuName, "3090"):
		return rtx3090CheckCount
	case strings.Contains(gpuName, "A10"):
		return rtxA10CheckCount
	default:
		return defaultGpuCheckCount
	}
}

func monitor(jobID string) {
	if err := writePidFile(jobID); err != nil {
		log.Fatalf("Failed to write PID file: %v", err)
		os.Exit(1)
	}

	conn, err := net.Dial("unix", SocketPath)
	if err != nil {
		log.Fatalf("Monitor failed to connect to daemon: %v", err)
	}
	defer conn.Close()

	ticker := time.NewTicker(time.Duration(monitorInterval) * time.Second)
	defer ticker.Stop()

	log.Printf("Starting monitoring for job %s, sending metrics every %d seconds.", jobID, monitorInterval)

	for range ticker.C {
		gpuUtil, err := getGpuUtilization()
		if err != nil {
			log.Printf("Warning: could not get GPU utilization: %v", err)
		}
		cpuUtil, err := getCpuUtilization()
		if err != nil {
			log.Printf("Warning: could not get CPU utilization: %v", err)
		}
		payload, _ := json.Marshal(MetricsPayload{
			GpuUtilization: gpuUtil,
			CpuUtilization: cpuUtil,
		})
		msg := Message{Type: "METRICS", JobID: jobID, Payload: payload}
		if err := json.NewEncoder(conn).Encode(msg); err != nil {
			log.Fatalf("Failed to send metrics, daemon may have terminated the job or is down. Exiting. Error: %v", err)
		}
		log.Printf("Sent metrics: GPU=%.1f%% (max of all cards), CPU=%.1f%%", gpuUtil, cpuUtil)
	}
}

// getGpuUtilization 获取所有GPU的平均利用率
func getGpuUtilization() (float64, error) {
	cmd := exec.Command("nvidia-smi", "--query-gpu=utilization.gpu", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("nvidia-smi command failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var totalUtil float64 = 0.0
	var gpuCount int = 0

	for _, line := range lines {
		utilStr := strings.TrimSpace(line)
		util, err := strconv.ParseFloat(utilStr, 64)
		if err != nil {
			// 忽略无法解析的行
			continue
		}
		totalUtil += util
		gpuCount++
	}

	if gpuCount == 0 {
		return 0, fmt.Errorf("no valid GPU utilization data found")
	}
	return totalUtil / float64(gpuCount), nil
}

func getCpuUtilization() (float64, error) {
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, err
	}
	if len(percent) > 0 {
		return percent[0], nil
	}
	return 0, fmt.Errorf("could not retrieve cpu percentage")
}

// writePidFile 写入当前进程的PID到指定文件
func writePidFile(jobID string) error {
	pid := os.Getpid()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	pidDir := filepath.Join(homeDir, ".monitor")
	if err := os.MkdirAll(pidDir, 0755); err != nil {
		return fmt.Errorf("failed to create PID directory: %w", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get hostname: %w", err)
	}
	pidFile := filepath.Join(pidDir, fmt.Sprintf("monitor-%s-%s.pid", jobID, hostname))

	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}
	log.Printf("PID file written to %s", pidFile)

	return nil
}
