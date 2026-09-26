package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"strings"
	"time"

	"github.com/bizshuk/gosdk/config"
	"github.com/bizshuk/os-bench/model"
	"github.com/bizshuk/os-bench/svc/bench"
	"github.com/bizshuk/os-bench/svc/host"
	"github.com/bizshuk/os-bench/svc/report"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RunCmd 執行 benchmark, 把報告存到資料目錄並印出結果。
var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run benchmarks and save a host performance report",
	Example: `  os-bench run
  os-bench run --only cpu,memory --duration 1s
  os-bench run --dir /data --disk-size-mb 1024
  os-bench run --json > report.json`,
	RunE: func(c *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(c.Context(), os.Interrupt)
		defer stop()

		opts := bench.Options{
			Duration:   viper.GetDuration("bench_duration"),
			Dir:        viper.GetString("bench_dir"),
			DiskSizeMB: viper.GetInt("bench_disk_size_mb"),
		}
		if opts.Dir == "" {
			opts.Dir = config.GetAppDataDir()
		}
		if err := os.MkdirAll(opts.Dir, 0o755); err != nil {
			return err
		}
		only := splitList(viper.GetStringSlice("bench_only"))
		categories := only
		if len(categories) == 0 {
			categories = bench.Categories()
		}

		h, err := host.Collect(ctx, opts.Dir)
		if err != nil {
			return err
		}
		if slices.Contains(categories, "disk") && slices.Contains(memoryFilesystems, h.DiskFSType) {
			return fmt.Errorf("--dir %s is on %s (memory-backed), disk results would measure RAM; pick a dir on a real disk", opts.Dir, h.DiskFSType)
		}
		start := time.Now()
		results, err := bench.Run(ctx, opts, only, func(category string) {
			fmt.Fprintf(os.Stderr, "running %s ...\n", category)
		})
		if err != nil {
			return err
		}

		r := model.Report{
			SchemaVersion: model.SCHEMA_VERSION,
			GeneratedAt:   start,
			ElapsedSec:    time.Since(start).Seconds(),
			Host:          h,
			Settings: model.Settings{
				DurationSec: opts.Duration.Seconds(),
				DiskDir:     opts.Dir,
				DiskSizeMB:  opts.DiskSizeMB,
				Categories:  categories,
			},
			Results: results,
		}
		path, err := report.Save(config.GetAppDataDir(), r)
		if err != nil {
			return err
		}

		if viper.GetBool("bench_json") {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(r)
		}
		renderReport(os.Stdout, r)
		fmt.Printf("saved: %s\n", path)
		return nil
	},
}

// memoryFilesystems 是存放在記憶體的檔案系統, 在上面量磁碟沒有意義。
var memoryFilesystems = []string{"tmpfs", "ramfs"}

// splitList 同時接受 --only cpu --only disk 與 --only cpu,disk 以及 BENCH_ONLY="cpu,disk"。
func splitList(values []string) []string {
	var out []string
	for _, v := range values {
		for _, s := range strings.Split(v, ",") {
			if s = strings.TrimSpace(s); s != "" {
				out = append(out, s)
			}
		}
	}
	return out
}

func init() {
	f := RunCmd.Flags()
	f.Duration("duration", 3*time.Second, "run length of each time-bounded benchmark")
	f.String("dir", "", "directory for disk benchmark files, picks which disk is measured (default: app data dir)")
	f.Int("disk-size-mb", 256, "file size for sequential and random disk benchmarks")
	f.StringSlice("only", nil, "categories to run: "+strings.Join(bench.Categories(), ","))
	f.Bool("json", false, "print the report as JSON instead of a table")

	viper.BindPFlag("bench_duration", f.Lookup("duration"))
	viper.BindPFlag("bench_dir", f.Lookup("dir"))
	viper.BindPFlag("bench_disk_size_mb", f.Lookup("disk-size-mb"))
	viper.BindPFlag("bench_only", f.Lookup("only"))
	viper.BindPFlag("bench_json", f.Lookup("json"))
}
