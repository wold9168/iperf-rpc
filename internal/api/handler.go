package api

import (
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/wold9168/iperf-rpc/internal/iperf"
	"github.com/wold9168/iperf-rpc/internal/model"
)

type Handler struct {
	executor     *iperf.Executor
	httpExecutor *iperf.HttpExecutor
}

func NewHandler(executor *iperf.Executor, httpExecutor *iperf.HttpExecutor) *Handler {
	return &Handler{executor: executor, httpExecutor: httpExecutor}
}

// RunIperf godoc
// @Summary      执行一次 iperf3 测速
// @Description  根据请求参数执行 iperf3 客户端或服务端测速
// @Tags         iperf
// @Accept       json
// @Produce      json
// @Param        request body model.IperfRunRequest true "iperf3 执行请求"
// @Success      200  {object}  model.IperfRunResponse
// @Failure      400  {object}  model.SimpleResponse
// @Router       /api/v1/iperf/run [post]
func (h *Handler) RunIperf(c *gin.Context) {
	var req model.IperfRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.SimpleResponse{
			Code:    -1,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	if req.Mode == "client" && req.Args.Target == "" {
		c.JSON(http.StatusBadRequest, model.SimpleResponse{
			Code:    -1,
			Message: "target is required for client mode",
		})
		return
	}

	if req.Args.Port == 0 {
		req.Args.Port = 5201
	}
	if req.Args.Duration == 0 {
		req.Args.Duration = 10
	}
	if req.Args.Parallel == 0 {
		req.Args.Parallel = 1
	}

	data := h.executor.Run(&req)
	c.JSON(http.StatusOK, model.IperfRunResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// StartServer godoc
// @Summary      启动 iperf3 服务端
// @Description  在当前后端启动 iperf3 服务端 (后台运行)
// @Tags         iperf
// @Accept       json
// @Produce      json
// @Param        port query int false "监听端口" default(5201)
// @Success      200  {object}  model.ServerStartResponse
// @Failure      400  {object}  model.SimpleResponse
// @Router       /api/v1/iperf/server/start [post]
func (h *Handler) StartServer(c *gin.Context) {
	portStr := c.DefaultQuery("port", "5201")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.SimpleResponse{
			Code:    -1,
			Message: "invalid port",
		})
		return
	}

	pid, err := h.executor.StartServer(port)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.SimpleResponse{
			Code:    -1,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.ServerStartResponse{
		Code:    0,
		Message: "server started",
		Data: &model.ServerStartData{
			Port: port,
			PID:  pid,
		},
	})
}

// StopServer godoc
// @Summary      停止 iperf3 服务端
// @Description  停止当前后端运行的 iperf3 服务端
// @Tags         iperf
// @Produce      json
// @Success      200  {object}  model.SimpleResponse
// @Failure      400  {object}  model.SimpleResponse
// @Router       /api/v1/iperf/server/stop [post]
func (h *Handler) StopServer(c *gin.Context) {
	if err := h.executor.StopServer(); err != nil {
		c.JSON(http.StatusBadRequest, model.SimpleResponse{
			Code:    -1,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.SimpleResponse{
		Code:    0,
		Message: "server stopped",
	})
}

// ServerStatus godoc
// @Summary      获取 iperf3 服务端状态
// @Description  获取当前后端 iperf3 服务端运行状态
// @Tags         iperf
// @Produce      json
// @Success      200  {object}  model.ServerStatusResponse
// @Router       /api/v1/iperf/status [get]
func (h *Handler) ServerStatus(c *gin.Context) {
	status := h.executor.ServerStatus()
	msg := "server not running"
	if status.Running {
		msg = "server running"
	}

	c.JSON(http.StatusOK, model.ServerStatusResponse{
		Code:    0,
		Message: msg,
		Data:    status,
	})
}

// Health godoc
// @Summary      健康检查
// @Description  返回服务健康状态
// @Tags         health
// @Produce      json
// @Success      200  {object}  model.HealthResponse
// @Router       /api/v1/health [get]
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, model.HealthResponse{Status: "ok"})
}

// RunHttpTest godoc
// @Summary      执行 HTTP 测速
// @Description  通过 HTTP(S) 下载/上传测试网络吞吐量，可选通过 SOCKS5 代理
// @Tags         http
// @Accept       json
// @Produce      json
// @Param        request body model.HttpTestRequest true "HTTP 测速请求"
// @Success      200  {object}  model.HttpTestResponse
// @Failure      400  {object}  model.SimpleResponse
// @Router       /api/v1/http/run [post]
func (h *Handler) RunHttpTest(c *gin.Context) {
	var req model.HttpTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.SimpleResponse{
			Code:    -1,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	if req.Direction != "download" && req.Direction != "upload" {
		c.JSON(http.StatusBadRequest, model.SimpleResponse{
			Code:    -1,
			Message: "direction must be 'download' or 'upload'",
		})
		return
	}

	if req.Duration == 0 {
		req.Duration = 10
	}

	data := h.httpExecutor.Run(&req)
	c.JSON(http.StatusOK, model.HttpTestResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// ServeHttpData godoc
// @Summary      HTTP 测速数据端点
// @Description  返回随机数据流，供下载测速使用。size 参数支持 K/M/G 后缀
// @Tags         http
// @Produce      octet-stream
// @Param        size query string false "数据大小" default(10M)
// @Success      200  {file}   binary
// @Router       /api/v1/http/data [get]
func (h *Handler) ServeHttpData(c *gin.Context) {
	sizeStr := c.DefaultQuery("size", "10M")
	size := parseSize(sizeStr)

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", strconv.FormatInt(size, 10))
	c.Status(http.StatusOK)

	buf := make([]byte, 32*1024)
	written := int64(0)
	for written < size {
		n := len(buf)
		if remaining := size - written; remaining < int64(n) {
			n = int(remaining)
		}
		rand.Read(buf[:n])
		m, err := c.Writer.Write(buf[:n])
		if err != nil {
			return
		}
		written += int64(m)
	}
}

// ReceiveHttpUpload godoc
// @Summary      HTTP 上传端点
// @Description  接收上传数据并丢弃，用于上传测速服务端
// @Tags         http
// @Accept       octet-stream
// @Success      200  {object}  model.SimpleResponse
// @Router       /api/v1/http/upload [post]
func (h *Handler) ReceiveHttpUpload(c *gin.Context) {
	_, _ = io.Copy(io.Discard, c.Request.Body)
	c.JSON(http.StatusOK, model.SimpleResponse{Code: 0, Message: "ok"})
}

func parseSize(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 10 * 1024 * 1024
	}

	multiplier := int64(1)
	last := s[len(s)-1]
	switch last {
	case 'K', 'k':
		multiplier = 1024
		s = s[:len(s)-1]
	case 'M', 'm':
		multiplier = 1024 * 1024
		s = s[:len(s)-1]
	case 'G', 'g':
		multiplier = 1024 * 1024 * 1024
		s = s[:len(s)-1]
	}

	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 10 * 1024 * 1024
	}
	return val * multiplier
}
