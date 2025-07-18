package models

// SlurmUint64NoVal 是一个辅助struct，用于解析Slurm中常见的 "set/infinite/number" 格式的数值对象
type SlurmUint64NoVal struct {
	Set      bool   `json:"set"`
	Infinite bool   `json:"infinite"`
	Number   uint64 `json:"number"`
}

// SlurmJobInfo 定义了我们从Slurm API接收到的单个作业信息
type SlurmJobInfo struct {
	JobID        uint32           `json:"job_id"`
	UserID       uint32           `json:"user_id"`
	Name         string           `json:"name"`
	UserName     string           `json:"user_name"`
	JobState     []string         `json:"job_state"`
	SubmitTime   SlurmUint64NoVal `json:"submit_time"`
	StartTime    SlurmUint64NoVal `json:"start_time"`
	TimeLimit    SlurmUint64NoVal `json:"time_limit"`
	Partition    string           `json:"partition"`
	NodeCount    SlurmUint64NoVal `json:"node_count"`
	CPUs         SlurmUint64NoVal `json:"cpus"`
	Account      string           `json:"account"`
	GresDetail   []string         `json:"gres_detail"`
	TresReqStr   string           `json:"tres_req_str"`
	TresAllocStr string           `json:"tres_alloc_str"`
}

// SlurmJobResponse 定义了 /jobs 接口返回的顶层JSON结构
type SlurmJobResponse struct {
	Jobs     []SlurmJobInfo `json:"jobs"`
	Meta     interface{}    `json:"meta"`
	Errors   []interface{}  `json:"errors"`
	Warnings []interface{}  `json:"warnings"`
}
