package iperf

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wold9168/iperf-rpc/internal/model"
	"golang.org/x/net/proxy"
)

type HttpExecutor struct {
	results map[string]*model.HttpTestData
}

func NewHttpExecutor() *HttpExecutor {
	return &HttpExecutor{
		results: make(map[string]*model.HttpTestData),
	}
}

func (e *HttpExecutor) Run(req *model.HttpTestRequest) *model.HttpTestData {
	id := uuid.New().String()
	now := time.Now()

	data := &model.HttpTestData{
		ID:        id,
		URL:       req.URL,
		Proxy:     req.Proxy,
		Direction: req.Direction,
		Status:    "running",
		StartedAt: now,
	}
	e.results[id] = data

	transport, err := e.buildTransport(req.Proxy)
	if err != nil {
		finished := time.Now()
		data.FinishedAt = &finished
		data.Status = "error"
		data.Error = fmt.Sprintf("failed to create transport: %v", err)
		return data
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(req.Duration+10) * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.Duration+10)*time.Second)
	defer cancel()

	var bytesTotal int64
	var elapsed time.Duration

	if req.Direction == "download" {
		bytesTotal, elapsed, err = e.doDownload(ctx, client, req.URL, req.Duration)
	} else {
		bytesTotal, elapsed, err = e.doUpload(ctx, client, req.URL, req.Duration)
	}

	finished := time.Now()
	data.FinishedAt = &finished
	data.BytesTotal = bytesTotal
	data.DurationSec = elapsed.Seconds()

	if err != nil {
		data.Status = "error"
		data.Error = err.Error()
	} else if elapsed.Seconds() > 0 {
		data.Status = "completed"
		data.BitrateBps = float64(bytesTotal) * 8 / elapsed.Seconds()
	} else {
		data.Status = "error"
		data.Error = "no data transferred"
	}

	return data
}

func (e *HttpExecutor) buildTransport(proxyURL string) (*http.Transport, error) {
	if proxyURL == "" {
		return &http.Transport{
			DialContext: (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
		}, nil
	}

	if strings.HasPrefix(proxyURL, "socks5://") {
		dialer, err := proxy.SOCKS5("tcp", proxyURL[9:], nil, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("socks5 proxy error: %w", err)
		}
		return &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.Dial(network, addr)
			},
		}, nil
	}

	return nil, fmt.Errorf("unsupported proxy scheme: %s (only socks5:// is supported)", proxyURL)
}

func (e *HttpExecutor) doDownload(ctx context.Context, client *http.Client, url string, duration int) (int64, time.Duration, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, 0, err
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, time.Since(start), fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, time.Since(start), fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	deadline := time.After(time.Duration(duration) * time.Second)
	buf := make([]byte, 32*1024)
	var total int64

	for {
		select {
		case <-ctx.Done():
			return total, time.Since(start), nil
		case <-deadline:
			return total, time.Since(start), nil
		default:
		}

		n, err := resp.Body.Read(buf)
		total += int64(n)
		if err != nil {
			if err == io.EOF {
				return total, time.Since(start), nil
			}
			return total, time.Since(start), err
		}
	}
}

func (e *HttpExecutor) doUpload(ctx context.Context, client *http.Client, url string, duration int) (int64, time.Duration, error) {
	deadline := time.After(time.Duration(duration) * time.Second)
	buf := make([]byte, 32*1024)

	reader, writer := io.Pipe()

	go func() {
		defer writer.Close()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if _, err := writer.Write(buf); err != nil {
				return
			}
		}
	}()

	req, err := http.NewRequestWithContext(ctx, "POST", url, reader)
	if err != nil {
		return 0, 0, err
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		return 0, elapsed, fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, elapsed, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	io.Copy(io.Discard, resp.Body)

	_ = deadline
	return 0, elapsed, nil
}

func (e *HttpExecutor) GetResult(id string) *model.HttpTestData {
	return e.results[id]
}
