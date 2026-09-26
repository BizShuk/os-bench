package bench

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/bizshuk/os-bench/model"
)

const (
	UNIT_MBPS = "MB/s"
	UNIT_IOPS = "IOPS"
	UNIT_OPS  = "ops/s"
	UNIT_NS   = "ns"
	UNIT_US   = "us"
	UNIT_MS   = "ms"
	UNIT_X    = "x"
)

// Options 是所有 benchmark 共用的量測參數。
type Options struct {
	// Duration 是每個以時間為界的量測項目的執行長度。
	Duration time.Duration
	// Dir 是磁碟量測建立暫存檔的目錄, 決定量到哪一顆磁碟。
	Dir string
	// DiskSizeMB 是循序讀寫與隨機讀取使用的檔案大小。
	DiskSizeMB int
}

// Benchmark 是一個資源類別的量測, 一次回傳該類別的所有指標。
type Benchmark struct {
	Category string
	Run      func(ctx context.Context, o Options) ([]model.Result, error)
}

// Suite 是全部類別的固定執行順序。
var Suite = []Benchmark{
	{Category: "cpu", Run: runCPU},
	{Category: "memory", Run: runMemory},
	{Category: "disk", Run: runDisk},
	{Category: "network", Run: runNetwork},
}

// Categories 回傳 Suite 內所有類別名稱。
func Categories() []string {
	names := make([]string, 0, len(Suite))
	for _, b := range Suite {
		names = append(names, b.Category)
	}
	return names
}

// Run 依 Suite 順序執行 only 指定的類別 (空值代表全部); onStart 在每個類別開始前被呼叫。
func Run(ctx context.Context, o Options, only []string, onStart func(category string)) ([]model.Result, error) {
	for _, c := range only {
		if !slices.Contains(Categories(), c) {
			return nil, fmt.Errorf("bench.Run: unknown category %q (valid: %v)", c, Categories())
		}
	}
	var results []model.Result
	for _, b := range Suite {
		if len(only) > 0 && !slices.Contains(only, b.Category) {
			continue
		}
		onStart(b.Category)
		rs, err := b.Run(ctx, o)
		if err != nil {
			return nil, fmt.Errorf("bench.Run: %s: %w", b.Category, err)
		}
		results = append(results, rs...)
	}
	return results, nil
}

func higher(category, name string, value float64, unit string) model.Result {
	return model.Result{Category: category, Name: name, Value: value, Unit: unit, Better: model.BETTER_HIGHER}
}

func lower(category, name string, value float64, unit string) model.Result {
	return model.Result{Category: category, Name: name, Value: value, Unit: unit, Better: model.BETTER_LOWER}
}
