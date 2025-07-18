package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"slurm-dashboard/config"
	"slurm-dashboard/internal/models"
	"slurm-dashboard/internal/store"

	"github.com/gin-gonic/gin"
)

type ClusterStatusResponse struct {
	Partitions []PartitionInfo `json:"partitions"`
	Nodes      []NodeStatus    `json:"nodes"`
	Errors     []string        `json:"errors"`
}

type PartitionInfo struct {
	Name  string   `json:"name"`
	Nodes []string `json:"nodes"`
}

type NodeStatus struct {
	Name          string    `json:"name"`
	State         []string  `json:"state"`
	Partitions    []string  `json:"partitions"`
	TotalCPUs     uint32    `json:"total_cpus"`
	AllocatedCPUs uint32    `json:"allocated_cpus"`
	AvailableCPUs uint32    `json:"available_cpus"`
	GPUs          []GPUInfo `json:"gpus"`
}

type GPUInfo struct {
	Type      string `json:"type"`
	Total     int    `json:"total"`
	Allocated int    `json:"allocated"`
	Available int    `json:"available"`
}

func GetClusterStatusHandler(cfg *config.Config, tokenStore *store.TokenStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, _ := c.Get("username")
		slurmToken, ok := tokenStore.Get(username.(string))
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Slurm session not found"})
			return
		}

		client := &http.Client{Timeout: 30 * time.Second}

		nodesData, err := fetchNodesData(client, cfg.SlurmAPIHost+"/slurm/v0.0.42/nodes", username.(string), slurmToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch nodes data", "details": err.Error()})
			return
		}

		response := processClusterDataFromNodes(nodesData)

		c.JSON(http.StatusOK, response)
	}
}

func fetchNodesData(client *http.Client, url, username, token string) (models.SlurmNodeResponse, error) {
	var result models.SlurmNodeResponse
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return result, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("X-SLURM-USER-NAME", username)
	req.Header.Set("X-SLURM-USER-TOKEN", token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return result, fmt.Errorf("request to slurm failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("slurm api returned status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Failed to unmarshal JSON for %s. Raw data: %s. Error: %v", url, string(body), err)
		return result, fmt.Errorf("failed to unmarshal json: %w", err)
	}
	return result, nil
}

func processClusterDataFromNodes(nodesData models.SlurmNodeResponse) ClusterStatusResponse {
	var response ClusterStatusResponse

	partitionMap := make(map[string]map[string]bool)

	for _, n := range nodesData.Nodes {
		response.Nodes = append(response.Nodes, NodeStatus{
			Name:          n.Name,
			State:         n.State,
			Partitions:    n.Partitions,
			TotalCPUs:     n.TotalCPUs,
			AllocatedCPUs: n.AllocatedCPUs,
			AvailableCPUs: n.TotalCPUs - n.AllocatedCPUs,
			GPUs:          parseGres(n.Gres, n.GresUsed),
		})

		for _, partName := range n.Partitions {
			if _, ok := partitionMap[partName]; !ok {
				partitionMap[partName] = make(map[string]bool)
			}
			partitionMap[partName][n.Name] = true
		}
	}

	for partName, nodes := range partitionMap {
		var nodeList []string
		for nodeName := range nodes {
			nodeList = append(nodeList, nodeName)
		}
		// !! 修正: 不再添加 State 字段
		response.Partitions = append(response.Partitions, PartitionInfo{
			Name:  partName,
			Nodes: nodeList,
		})
	}

	return response
}

// parseGres 解析GRES字符串以获取GPU信息
func parseGres(gres, gresUsed string) []GPUInfo {
	gpuMap := make(map[string]*GPUInfo)

	// 解析总资源: "gpu:A100:4,gpu:V100:2"
	totalParts := strings.Split(gres, ",")
	for _, part := range totalParts {
		// gres/gpu:a100:8
		if !strings.HasPrefix(part, "gpu:") {
			// 也可能是 gres/gpu:a100:8 这种格式
			if !strings.HasPrefix(part, "gres/gpu:") {
				continue
			}
		}

		gpuInfo := strings.Split(part, ":")
		if len(gpuInfo) < 2 {
			continue
		}

		gpuType := "gpu" // 默认类型
		countStr := gpuInfo[1]

		if len(gpuInfo) > 2 {
			gpuType = gpuInfo[1]
			countStr = gpuInfo[2]
		}

		count, err := strconv.Atoi(countStr)
		if err != nil {
			continue
		}

		if _, ok := gpuMap[gpuType]; !ok {
			gpuMap[gpuType] = &GPUInfo{Type: gpuType}
		}
		gpuMap[gpuType].Total += count
	}

	// 解析已用资源: "gpu:A100:1(IDX:0),gpu:V100:1"
	usedParts := strings.Split(gresUsed, ",")
	for _, part := range usedParts {
		if !strings.HasPrefix(part, "gpu:") {
			if !strings.HasPrefix(part, "gres/gpu:") {
				continue
			}
		}

		// 去掉(IDX...)部分
		if idx := strings.Index(part, "("); idx != -1 {
			part = part[:idx]
		}

		gpuInfo := strings.Split(part, ":")
		if len(gpuInfo) < 2 {
			continue
		}

		gpuType := "gpu"
		countStr := gpuInfo[1]

		if len(gpuInfo) > 2 {
			gpuType = gpuInfo[1]
			countStr = gpuInfo[2]
		}

		count, err := strconv.Atoi(countStr)
		if err != nil {
			continue
		}

		if _, ok := gpuMap[gpuType]; !ok {
			gpuMap[gpuType] = &GPUInfo{Type: gpuType}
		}
		gpuMap[gpuType].Allocated += count
	}

	var result []GPUInfo
	for _, gpu := range gpuMap {
		gpu.Available = gpu.Total - gpu.Allocated
		result = append(result, *gpu)
	}
	return result
}
