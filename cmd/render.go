package cmd

import (
	"fmt"
	"io"
	"strconv"

	"github.com/bizshuk/gosdk/tui"
	"github.com/bizshuk/os-bench/model"
)

// renderReport 印出主機摘要與所有指標; run 與 show 共用。
func renderReport(w io.Writer, r model.Report) {
	h := r.Host
	hostTable := tui.Table{
		Headers: []string{"Host", "Value"},
		Align:   []int{0, 0},
		Rows: [][]tui.Cell{
			{"hostname", h.Hostname},
			{"os", fmt.Sprintf("%s %s (%s, kernel %s)", h.Platform, h.PlatformVersion, h.Arch, h.KernelVersion)},
			{"cpu", fmt.Sprintf("%s, %d physical / %d logical", h.CPUModel, h.CPUPhysicalCores, h.CPULogicalCores)},
			{"memory", fmt.Sprintf("%s total, %s swap", gib(h.MemoryTotalBytes), gib(h.SwapTotalBytes))},
			{"disk", fmt.Sprintf("%s (%s), %s free of %s", h.DiskPath, h.DiskFSType, gib(h.DiskFreeBytes), gib(h.DiskTotalBytes))},
			{"load_avg_1m", strconv.FormatFloat(h.LoadAvg1m, 'f', 2, 64)},
			{"generated_at", r.GeneratedAt.Format("2006-01-02 15:04:05 MST")},
		},
	}
	if h.Virtualization != "" {
		hostTable.Rows = append(hostTable.Rows, []tui.Cell{"virtualization", h.Virtualization})
	}
	hostTable.Draw(w, false, false)

	results := tui.Table{
		Headers: []string{"Category", "Metric", "Value", "Unit", "Better"},
		Align:   []int{0, 0, 1, 0, 0},
	}
	for i, res := range r.Results {
		results.Rows = append(results.Rows, []tui.Cell{res.Category, res.Name, formatValue(res.Value), res.Unit, string(res.Better)})
		last := i == len(r.Results)-1
		results.Separators = append(results.Separators, !last && r.Results[i+1].Category != res.Category)
	}
	results.Draw(w, false, false)
}

// formatValue 依數量級決定小數位數, 保留約三位有效數字。
func formatValue(v float64) string {
	switch {
	case v >= 100:
		return strconv.FormatFloat(v, 'f', 0, 64)
	case v >= 10:
		return strconv.FormatFloat(v, 'f', 1, 64)
	default:
		return strconv.FormatFloat(v, 'f', 2, 64)
	}
}

func gib(b uint64) string {
	return fmt.Sprintf("%.1f GiB", float64(b)/(1<<30))
}
