package iperf

import (
	"testing"

	"github.com/wold9168/iperf-rpc/internal/model"
)

func TestBuildArgsClient(t *testing.T) {
	e := New()
	req := &model.IperfRunRequest{
		Mode: "client",
		Args: model.IperfArgs{
			Target:    "10.0.0.2",
			Port:      5201,
			Duration:  10,
			Parallel:  4,
			Bandwidth: "100M",
			Protocol:  "tcp",
			Reverse:   false,
		},
	}

	args := e.buildArgs(req)

	if args[0] != "-c" || args[1] != "10.0.0.2" {
		t.Errorf("expected -c 10.0.0.2, got %v", args[:2])
	}

	hasJ := false
	for _, a := range args {
		if a == "-J" {
			hasJ = true
		}
	}
	if !hasJ {
		t.Error("expected -J flag for JSON output")
	}
}

func TestBuildArgsServer(t *testing.T) {
	e := New()
	req := &model.IperfRunRequest{
		Mode: "server",
		Args: model.IperfArgs{
			Port: 5201,
		},
	}

	args := e.buildArgs(req)

	if args[0] != "-s" || args[1] != "-1" {
		t.Errorf("expected -s -1, got %v", args[:2])
	}
}

func TestBuildArgsUDP(t *testing.T) {
	e := New()
	req := &model.IperfRunRequest{
		Mode: "client",
		Args: model.IperfArgs{
			Target:   "10.0.0.2",
			Protocol: "udp",
		},
	}

	args := e.buildArgs(req)

	hasU := false
	for _, a := range args {
		if a == "-u" {
			hasU = true
		}
	}
	if !hasU {
		t.Error("expected -u flag for UDP")
	}
}

func TestBuildArgsReverse(t *testing.T) {
	e := New()
	req := &model.IperfRunRequest{
		Mode: "client",
		Args: model.IperfArgs{
			Target:  "10.0.0.2",
			Reverse: true,
		},
	}

	args := e.buildArgs(req)

	hasR := false
	for _, a := range args {
		if a == "-R" {
			hasR = true
		}
	}
	if !hasR {
		t.Error("expected -R flag for reverse")
	}
}

func TestBuildArgsExtra(t *testing.T) {
	e := New()
	req := &model.IperfRunRequest{
		Mode: "client",
		Args: model.IperfArgs{
			Target: "10.0.0.2",
			Extra:  "--omit 2 --window 256K",
		},
	}

	args := e.buildArgs(req)

	hasExtra := false
	for _, a := range args {
		if a == "--omit" {
			hasExtra = true
		}
	}
	if !hasExtra {
		t.Error("expected extra args in command")
	}
}

func TestParseSummary(t *testing.T) {
	e := New()
	output := `{
		"end": {
			"sum_sent": {
				"bits_per_second": 945000000,
				"seconds": 10
			},
			"sum": {
				"jitter_ms": 0.5,
				"lost_percent": 0
			}
		}
	}`

	summary := e.parseSummary(output)
	if summary == nil {
		t.Fatal("expected summary, got nil")
	}
	if summary.BitrateBps != 945000000 {
		t.Errorf("expected bitrate 945000000, got %f", summary.BitrateBps)
	}
	if summary.DurationSec != 10 {
		t.Errorf("expected duration 10, got %f", summary.DurationSec)
	}
	if summary.JitterMs != 0.5 {
		t.Errorf("expected jitter 0.5, got %f", summary.JitterMs)
	}
}
