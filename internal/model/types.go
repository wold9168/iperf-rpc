package model

import "time"

// @Description iperf3 执行请求
type IperfRunRequest struct {
	Mode string      `json:"mode" binding:"required" example:"client" enums:"client,server"` // client 或 server
	Args IperfArgs   `json:"args" binding:"required"`
}

// @Description iperf3 运行参数
type IperfArgs struct {
	Target    string `json:"target" example:"10.0.0.2"`            // 客户端模式必填: 目标服务器 IP
	Port      int    `json:"port" example:"5201" default:"5201"`    // 端口
	Duration  int    `json:"duration" example:"10" default:"10"`    // 测试时长秒数
	Parallel  int    `json:"parallel" example:"1" default:"1"`      // 并行流数
	Bandwidth string `json:"bandwidth" example:"100M"`              // 带宽限制
	Protocol  string `json:"protocol" example:"tcp" enums:"tcp,udp"` // tcp 或 udp
	Reverse   bool   `json:"reverse" example:"false"`               // 反向测试
	Extra     string `json:"extra"`                                 // 额外命令行参数
}

// @Description iperf3 执行结果
type IperfRunResponse struct {
	Code    int            `json:"code" example:"0"`
	Message string         `json:"message" example:"success"`
	Data    *IperfRunData  `json:"data,omitempty"`
}

// @Description iperf3 执行结果数据
type IperfRunData struct {
	ID         string        `json:"id" example:"uuid"`
	Command    string        `json:"command" example:"iperf3 -c 10.0.0.2 -p 5201 -t 10 -J"`
	Output     string        `json:"output"`
	Summary    *IperfSummary `json:"summary,omitempty"`
	Status     string        `json:"status" example:"completed" enums:"running,completed,error"`
	StartedAt  time.Time     `json:"started_at"`
	FinishedAt *time.Time    `json:"finished_at,omitempty"`
}

// @Description iperf3 结果摘要
type IperfSummary struct {
	BitrateBps  float64 `json:"bitrate_bps" example:"945000000"`
	JitterMs    float64 `json:"jitter_ms" example:"0.5"`
	LostPercent float64 `json:"lost_percent" example:"0"`
	DurationSec float64 `json:"duration_sec" example:"10"`
}

// @Description iperf3 server 启动响应
type ServerStartResponse struct {
	Code    int              `json:"code" example:"0"`
	Message string           `json:"message" example:"server started"`
	Data    *ServerStartData `json:"data,omitempty"`
}

// @Description iperf3 server 启动数据
type ServerStartData struct {
	Port int `json:"port" example:"5201"`
	PID  int `json:"pid" example:"12345"`
}

// @Description iperf3 server 状态响应
type ServerStatusResponse struct {
	Code    int               `json:"code" example:"0"`
	Message string            `json:"message" example:"server running"`
	Data    *ServerStatusData `json:"data,omitempty"`
}

// @Description iperf3 server 状态数据
type ServerStatusData struct {
	Running bool `json:"running" example:"true"`
	Port    int  `json:"port" example:"5201"`
	PID     int  `json:"pid" example:"12345"`
}

// @Description 通用响应
type SimpleResponse struct {
	Code    int    `json:"code" example:"0"`
	Message string `json:"message" example:"ok"`
}

// @Description 健康检查响应
type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

// @Description HTTP 测速请求
type HttpTestRequest struct {
	URL       string `json:"url" binding:"required" example:"https://10.0.0.2:8443/api/v1/http/data?size=100M"` // 目标 URL
	Proxy     string `json:"proxy" example:"socks5://proxy:1080"`                                               // SOCKS5 代理地址 (可选)
	Direction string `json:"direction" binding:"required" example:"download" enums:"download,upload"`           // download 或 upload
	Duration  int    `json:"duration" example:"10" default:"10"`                                                // 测试时长秒数
	Insecure  bool   `json:"insecure" example:"false"`                                                           // 跳过 TLS 证书验证
}

// @Description HTTP 测速响应
type HttpTestResponse struct {
	Code    int           `json:"code" example:"0"`
	Message string        `json:"message" example:"success"`
	Data    *HttpTestData `json:"data,omitempty"`
}

// @Description HTTP 测速结果数据
type HttpTestData struct {
	ID          string     `json:"id" example:"uuid"`
	URL         string     `json:"url"`
	Proxy       string     `json:"proxy,omitempty"`
	Direction   string     `json:"direction"`
	Status      string     `json:"status" example:"completed" enums:"running,completed,error"`
	BytesTotal  int64      `json:"bytes_total" example:"104857600"`
	BitrateBps  float64    `json:"bitrate_bps" example:"95000000"`
	DurationSec float64    `json:"duration_sec" example:"10"`
	Error       string     `json:"error,omitempty"`
	StartedAt   time.Time  `json:"started_at"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
}
