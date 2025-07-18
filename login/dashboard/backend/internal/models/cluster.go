package models

// --- /partitions 接口的模型 ---

type SlurmPartitionNodesObject struct {
	Configured string `json:"configured"`
}

type SlurmPartitionStateObject struct {
	State []string `json:"state"`
}

type SlurmPartitionInfo struct {
	Name      string                    `json:"name"`
	Nodes     SlurmPartitionNodesObject `json:"nodes"`
	Partition SlurmPartitionStateObject `json:"partition"`
}

type SlurmPartitionResponse struct {
	Partitions []SlurmPartitionInfo `json:"partitions"`
	Meta       interface{}          `json:"meta"`
	Errors     []interface{}        `json:"errors"`
	Warnings   []interface{}        `json:"warnings"`
}

// --- /nodes 接口的模型 ---

type SlurmNodeInfo struct {
	Name          string   `json:"name"`
	State         []string `json:"state"`
	Partitions    []string `json:"partitions"`
	TotalCPUs     uint32   `json:"cpus"`
	AllocatedCPUs uint32   `json:"alloc_cpus"`
	Gres          string   `json:"gres"`
	GresUsed      string   `json:"gres_used"`
}

type SlurmNodeResponse struct {
	Nodes    []SlurmNodeInfo `json:"nodes"`
	Meta     interface{}     `json:"meta"`
	Errors   []interface{}   `json:"errors"`
	Warnings []interface{}   `json:"warnings"`
}
