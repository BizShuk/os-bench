package cmd

import (
	"os"

	"github.com/bizshuk/gosdk/config"
	"github.com/bizshuk/os-bench/svc/report"
	"github.com/spf13/cobra"
)

// ShowCmd 重新顯示一份已存報告; 參數可為路徑或 list 顯示的檔名。
var ShowCmd = &cobra.Command{
	Use:   "show <report>",
	Short: "Show a saved report",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		r, err := report.Load(report.Resolve(config.GetAppDataDir(), args[0]))
		if err != nil {
			return err
		}
		renderReport(os.Stdout, r)
		return nil
	},
}
