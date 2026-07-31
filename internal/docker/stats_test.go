package docker

import (
	"testing"

	"github.com/docker/docker/api/types/container"
)

func TestComputeStats(t *testing.T) {
	raw := container.StatsResponse{
		CPUStats: container.CPUStats{
			CPUUsage:    container.CPUUsage{TotalUsage: 2000},
			SystemUsage: 10000,
			OnlineCPUs:  2,
		},
		PreCPUStats: container.CPUStats{
			CPUUsage:    container.CPUUsage{TotalUsage: 1000},
			SystemUsage: 5000,
		},
		MemoryStats: container.MemoryStats{
			Usage: 200,
			Limit: 1000,
			Stats: map[string]uint64{"total_inactive_file": 50},
		},
		Networks: map[string]container.NetworkStats{
			"eth0": {RxBytes: 100, TxBytes: 10},
			"eth1": {RxBytes: 200, TxBytes: 20},
		},
		BlkioStats: container.BlkioStats{
			IoServiceBytesRecursive: []container.BlkioStatEntry{
				{Op: "Read", Value: 5},
				{Op: "Write", Value: 7},
				{Op: "Async", Value: 99}, // ignored
			},
		},
		PidsStats: container.PidsStats{Current: 3},
	}

	got := computeStats("abc123", raw)

	// cpuDelta=1000, sysDelta=5000, online=2 -> (1000/5000)*2*100 = 40
	if got.CPUPercent != 40 {
		t.Errorf("CPUPercent = %v, want 40", got.CPUPercent)
	}
	// usage 200 - inactive 50 = 150; 150/1000 = 15%
	if got.MemUsage != 150 {
		t.Errorf("MemUsage = %d, want 150", got.MemUsage)
	}
	if got.MemLimit != 1000 {
		t.Errorf("MemLimit = %d, want 1000", got.MemLimit)
	}
	if got.MemPercent != 15 {
		t.Errorf("MemPercent = %v, want 15", got.MemPercent)
	}
	if got.NetRx != 300 || got.NetTx != 30 {
		t.Errorf("Net = (%d, %d), want (300, 30)", got.NetRx, got.NetTx)
	}
	if got.BlockRead != 5 || got.BlockWrite != 7 {
		t.Errorf("Block = (%d, %d), want (5, 7)", got.BlockRead, got.BlockWrite)
	}
	if got.PIDs != 3 {
		t.Errorf("PIDs = %d, want 3", got.PIDs)
	}
	if got.ContainerID != "abc123" {
		t.Errorf("ContainerID = %q, want abc123", got.ContainerID)
	}
}

func TestComputeStatsZeroDeltaNoCPUPercent(t *testing.T) {
	// A first sample has no previous reading; CPU% must stay zero rather than
	// divide by a zero system delta.
	raw := container.StatsResponse{
		CPUStats:    container.CPUStats{CPUUsage: container.CPUUsage{TotalUsage: 1000}, SystemUsage: 5000},
		PreCPUStats: container.CPUStats{CPUUsage: container.CPUUsage{TotalUsage: 1000}, SystemUsage: 5000},
	}
	if got := computeStats("x", raw); got.CPUPercent != 0 {
		t.Errorf("CPUPercent = %v, want 0", got.CPUPercent)
	}
}

func TestComputeStatsOnlineCPUsFallback(t *testing.T) {
	// When OnlineCPUs is unset, the per-CPU usage slice length stands in.
	raw := container.StatsResponse{
		CPUStats: container.CPUStats{
			CPUUsage:    container.CPUUsage{TotalUsage: 2000, PercpuUsage: []uint64{1, 2, 3, 4}},
			SystemUsage: 10000,
		},
		PreCPUStats: container.CPUStats{
			CPUUsage:    container.CPUUsage{TotalUsage: 1000},
			SystemUsage: 5000,
		},
	}
	// (1000/5000)*4*100 = 80
	if got := computeStats("x", raw); got.CPUPercent != 80 {
		t.Errorf("CPUPercent = %v, want 80", got.CPUPercent)
	}
}
