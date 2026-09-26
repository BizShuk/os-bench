package host

import (
	"context"
	"fmt"
	"runtime"

	"github.com/bizshuk/os-bench/model"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	gohost "github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
)

// Collect 讀取主機規格與目前負載; diskDir 是磁碟量測所在的目錄。
func Collect(ctx context.Context, diskDir string) (model.Host, error) {
	info, err := gohost.InfoWithContext(ctx)
	if err != nil {
		return model.Host{}, fmt.Errorf("host.Collect: host info: %w", err)
	}
	h := model.Host{
		Hostname:        info.Hostname,
		OS:              info.OS,
		Platform:        info.Platform,
		PlatformVersion: info.PlatformVersion,
		KernelVersion:   info.KernelVersion,
		Arch:            runtime.GOARCH,
		GoVersion:       runtime.Version(),
		GOMAXPROCS:      runtime.GOMAXPROCS(0),
		DiskPath:        diskDir,
	}
	if info.VirtualizationSystem != "" {
		h.Virtualization = info.VirtualizationSystem + "/" + info.VirtualizationRole
	}

	cpus, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return model.Host{}, fmt.Errorf("host.Collect: cpu info: %w", err)
	}
	if len(cpus) > 0 {
		h.CPUModel = cpus[0].ModelName
	}
	if h.CPUPhysicalCores, err = cpu.CountsWithContext(ctx, false); err != nil {
		return model.Host{}, fmt.Errorf("host.Collect: physical cores: %w", err)
	}
	if h.CPULogicalCores, err = cpu.CountsWithContext(ctx, true); err != nil {
		return model.Host{}, fmt.Errorf("host.Collect: logical cores: %w", err)
	}

	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return model.Host{}, fmt.Errorf("host.Collect: memory: %w", err)
	}
	h.MemoryTotalBytes = vm.Total
	swap, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		return model.Host{}, fmt.Errorf("host.Collect: swap: %w", err)
	}
	h.SwapTotalBytes = swap.Total

	usage, err := disk.UsageWithContext(ctx, diskDir)
	if err != nil {
		return model.Host{}, fmt.Errorf("host.Collect: disk usage %s: %w", diskDir, err)
	}
	h.DiskFSType = usage.Fstype
	h.DiskTotalBytes = usage.Total
	h.DiskFreeBytes = usage.Free

	avg, err := load.AvgWithContext(ctx)
	if err != nil {
		return model.Host{}, fmt.Errorf("host.Collect: load average: %w", err)
	}
	h.LoadAvg1m = avg.Load1
	return h, nil
}
