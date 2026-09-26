package model

import "time"

// SCHEMA_VERSION 是報告 JSON 的格式版本, 欄位語意改變時遞增。
const SCHEMA_VERSION = 1

// Better 標示一個指標的數值往哪個方向是較好。
type Better string

const (
	BETTER_HIGHER Better = "higher"
	BETTER_LOWER  Better = "lower"
)

// Report 是一次完整量測的結果: 主機資訊、量測設定與所有指標。
type Report struct {
	SchemaVersion int       `json:"schema_version"`
	GeneratedAt   time.Time `json:"generated_at"`
	ElapsedSec    float64   `json:"elapsed_sec"`
	Host          Host      `json:"host"`
	Settings      Settings  `json:"settings"`
	Results       []Result  `json:"results"`
}

// Host 是量測當下的主機規格與負載狀態。
type Host struct {
	Hostname         string  `json:"hostname"`
	OS               string  `json:"os"`
	Platform         string  `json:"platform"`
	PlatformVersion  string  `json:"platform_version"`
	KernelVersion    string  `json:"kernel_version"`
	Arch             string  `json:"arch"`
	Virtualization   string  `json:"virtualization,omitempty"`
	CPUModel         string  `json:"cpu_model"`
	CPUPhysicalCores int     `json:"cpu_physical_cores"`
	CPULogicalCores  int     `json:"cpu_logical_cores"`
	MemoryTotalBytes uint64  `json:"memory_total_bytes"`
	SwapTotalBytes   uint64  `json:"swap_total_bytes"`
	DiskPath         string  `json:"disk_path"`
	DiskFSType       string  `json:"disk_fstype"`
	DiskTotalBytes   uint64  `json:"disk_total_bytes"`
	DiskFreeBytes    uint64  `json:"disk_free_bytes"`
	LoadAvg1m        float64 `json:"load_avg_1m"`
	GoVersion        string  `json:"go_version"`
	GOMAXPROCS       int     `json:"gomaxprocs"`
}

// Settings 是這次量測使用的參數, 比較兩份報告前要先確認一致。
type Settings struct {
	DurationSec float64  `json:"duration_sec"`
	DiskDir     string   `json:"disk_dir"`
	DiskSizeMB  int      `json:"disk_size_mb"`
	Categories  []string `json:"categories"`
}

// Result 是單一指標的量測值。
type Result struct {
	Category string  `json:"category"`
	Name     string  `json:"name"`
	Value    float64 `json:"value"`
	Unit     string  `json:"unit"`
	Better   Better  `json:"better"`
}
