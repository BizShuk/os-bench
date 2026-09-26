package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/bizshuk/gosdk/config"
	"github.com/bizshuk/gosdk/tui"
	"github.com/bizshuk/os-bench/svc/report"
	"github.com/spf13/cobra"
)

// ListCmd 列出資料目錄中的已存報告, 最新的在前。
var ListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Short:   "List saved reports",
	RunE: func(c *cobra.Command, args []string) error {
		dir := config.GetAppDataDir()
		entries, err := report.List(dir)
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			fmt.Printf("no reports in %s, run `os-bench run` first\n", dir)
			return nil
		}
		t := tui.Table{
			Headers: []string{"Report", "Host", "Generated", "CPU", "Metrics"},
			Align:   []int{0, 0, 0, 0, 1},
		}
		for _, e := range entries {
			r := e.Report
			t.Rows = append(t.Rows, []tui.Cell{
				filepath.Base(e.Path),
				r.Host.Hostname,
				r.GeneratedAt.Format("2006-01-02 15:04"),
				r.Host.CPUModel,
				strconv.Itoa(len(r.Results)),
			})
		}
		t.Draw(os.Stdout, false, false)
		return nil
	},
}
