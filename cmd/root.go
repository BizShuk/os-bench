package cmd

import (
	"fmt"
	"os"

	gosdkcmd "github.com/bizshuk/gosdk/cmd"
	"github.com/bizshuk/gosdk/config"
	"github.com/spf13/cobra"
)

const APP_NAME = "os-bench"

// RootCmd 是 CLI 的根命令, 所有功能都在子命令上。
var RootCmd = &cobra.Command{
	Use:   APP_NAME,
	Short: "Benchmark this host and save a performance report for service evaluation",
	Long: `os-bench 在本機跑 CPU, memory, disk, network 四類 benchmark,
連同主機規格存成一份 JSON 報告, 作為評估服務能否部署在這台主機的參考基準。

報告存放於 ~/.config/os-bench/data/, 用 list 查看, 用 show 重新顯示。`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute 執行根命令。
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(func() { config.Default(config.WithAppName(APP_NAME)) })
	RootCmd.AddCommand(gosdkcmd.ConfigCmd, RunCmd, ListCmd, ShowCmd)
}
