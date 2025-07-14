package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
)

const SocketPath = "/var/run/node_monitor.sock"

const (
	cpuCheckCount   = 30 // CPU检查次数
	monitorInterval = 60 // 监控间隔（秒）
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
	gpuCount := determineGpuCount()
	log.Printf("Dynamically determined GPU monitoring count: %d", gpuCount)

	conn, err := net.Dial("unix", SocketPath)
	if err != nil {
		log.Fatalf("Failed to connect to node monitor daemon: %v", err)
	}
	defer conn.Close()

	regPayload := RegisterPayload{
		GpuMonitorCount: gpuCount,
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

	log.Printf("Job %s registered successfully. GPU_Count=%d, CPU_Count=%d, Log file: %s", jobID, gpuCount, cpuCheckCount, logPath)
	os.Exit(0)
}

// determineGpuCount 检查节点上所有GPU并返回最佳的监控次数
func determineGpuCount() int {
	bestCardScore := 0

	// 执行nvidia-smi命令获取所有GPU的索引和名称
	cmd := exec.Command("nvidia-smi", "--query-gpu=index,name", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Warning: 'nvidia-smi' command failed: %v. Assuming no GPUs or driver issue. Setting GPU count to default.", err)
		// 在没有GPU或驱动有问题的节点上，返回一个默认的宽松值
		return getGpuCountFromScore(0)
	}

	// 按行解析输出
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		log.Println("No GPUs detected by nvidia-smi. Setting GPU count to default.")
		return getGpuCountFromScore(0)
	}

	for _, line := range lines {
		parts := strings.Split(line, ",")
		if len(parts) < 2 {
			continue
		}
		gpuIndex := strings.TrimSpace(parts[0])
		gpuName := strings.TrimSpace(parts[1])

		currentCardScore := getGpuScore(gpuName)
		log.Printf("  - Detected GPU %s: %s, Score: %d", gpuIndex, gpuName, currentCardScore)

		if currentCardScore > bestCardScore {
			bestCardScore = currentCardScore
		}
	}

	finalCount := getGpuCountFromScore(bestCardScore)
	log.Printf("Best card score found on this node: %d. Setting monitoring count to: %d", bestCardScore, finalCount)
	return finalCount
}

// getGpuScore 根据GPU名称返回一个分数
func getGpuScore(gpuName string) int {
	switch {
	case strings.Contains(gpuName, "5090"):
		return 100
	case strings.Contains(gpuName, "A6000"):
		return 90
	case strings.Contains(gpuName, "4090"):
		return 80
	case strings.Contains(gpuName, "3090"):
		return 70
	case strings.Contains(gpuName, "A10"):
		return 50
	default:
		return 10
	}
}

// getGpuCountFromScore 根据分数返回监控次数
func getGpuCountFromScore(score int) int {
	switch score {
	case 100:
		return 10
	case 90:
		return 20
	case 80:
		return 30
	case 70:
		return 120
	case 50:
		return 420
	case 0:
		return 10086 // 没有卡
	default:
		return 600 // 有卡，但非上述列出来的卡
	}
}

func monitor(jobID string) {
	log.Printf("Starting monitoring for job %s, sending metrics every %d seconds.", jobID, monitorInterval)

	conn, err := net.Dial("unix", SocketPath)
	if err != nil {
		log.Fatalf("Monitor failed to connect to daemon: %v", err)
	}
	defer conn.Close()

	ticker := time.NewTicker(time.Duration(monitorInterval) * time.Second)
	defer ticker.Stop()

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

		// 通过 squeue 命令检查作业是否还是运行状态
		log.Printf("Checking if job %s is still in squeue...", jobID)
		if !isJobInSqueue(jobID) {
			log.Printf("Job %s not found in squeue. Terminating job_helper.", jobID)
			os.Exit(0)
		}
	}
}

// getGpuUtilization 获取所有GPU中的最高利用率
func getGpuUtilization() (float64, error) {
	cmd := exec.Command("nvidia-smi", "--query-gpu=utilization.gpu", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("nvidia-smi command failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var maxUtil float64 = 0.0

	for _, line := range lines {
		utilStr := strings.TrimSpace(line)
		util, err := strconv.ParseFloat(utilStr, 64)
		if err != nil {
			// 忽略无法解析的行
			continue
		}
		if util > maxUtil {
			maxUtil = util
		}
	}

	return maxUtil, nil
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

// isJobInSqueue 检查作业是否仍在squeue中
func isJobInSqueue(jobID string) bool {
	cmd := exec.Command("squeue", "-j", jobID, "--noheader")
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return false
		}
		return true // 如果squeue命令失败，假设作业仍在运行
	}
	return strings.Contains(string(output), jobID)
}
