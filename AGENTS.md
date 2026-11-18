# os-bench — 技術脈絡 (Technical Context)

業務定義與指標清單見 [README.md](README.md), 本檔只記結構與關鍵決策。

## 結構 (Layout)

```tree
main.go                 # cmd.Execute() only
cmd/                    # cobra 命令: root, run, list, show; render.go 是 run/show 共用的表格輸出
model/report.go         # Report / Host / Settings / Result, JSON schema 的唯一定義
svc/host/               # 以 gopsutil 收集主機規格與負載
svc/bench/              # 量測本體: bench.go (Suite + Run), stats.go (共用量測迴圈),
                        #   cpu.go memory.go disk.go network.go 各一類, direct_<os>.go 為平台差異
svc/report/             # 報告存取: Save / Load / List / Resolve
```

## 設定 (Config)

`config.Default(config.WithAppName("os-bench"))`; 報告存於 `config.GetAppDataDir()`
(`~/.config/os-bench/data`)。flag 綁定扁平 key, 可用環境變數覆寫:

| flag | viper key / env | 預設 |
| --- | --- | --- |
| `--duration` | `bench_duration` / `BENCH_DURATION` | `3s` |
| `--dir` | `bench_dir` / `BENCH_DIR` | app data dir |
| `--disk-size-mb` | `bench_disk_size_mb` / `BENCH_DISK_SIZE_MB` | `256` |
| `--only` | `bench_only` / `BENCH_ONLY` | 全部 |
| `--json` | `bench_json` / `BENCH_JSON` | `false` |

key 一律帶 `bench_` 前綴, 避開 `DIR` 之類的通用環境變數被 AutomaticEnv 撿到。

## 關鍵決策 (Key Decisions)

- `純 Go, 不呼叫外部工具`: 目標主機不一定裝有 `fio` / `iperf3`, 單一 binary 才能直接丟上去跑。
- `以時間為界`: 吞吐量與延遲項目都跑 `--duration` 長度, 而非固定次數, 快慢主機的總耗時一致。
- `disk 繞過 page cache`: Linux 用 `O_DIRECT` (buffer 以 `alignedBuf` 對齊 4 KiB),
  macOS 用 `F_NOCACHE`。先前用 `POSIX_FADV_DONTNEED` 失敗: 讀取會重新填回快取, overlayfs 也不理會,
  4K 隨機讀量到 3 us。`O_DIRECT` 在 tmpfs 上也能開檔, 所以 tmpfs / ramfs 由 `cmd/run.go` 依 fstype 拒絕。
- `MB = 1e6 bytes`: 與 fio / iperf3 慣例一致; 主機容量用 GiB。
- `CPU 時脈不記錄`: gopsutil 在 Apple Silicon 回傳錯值 (4), 且時脈隨負載浮動, 無比較價值。
- `平台`: 只支援 darwin 與 linux (`direct_<os>.go`), 其他 GOOS 無法編譯。

## 開發 (Development)

```bash
npm run test     # go test ./..., bench 測試以 100ms / 8MB 跑完整 Suite
npm run lint     # gofmt + go vet (另跑 GOOS=linux go vet ./... 檢查 linux 分支)
npm run build    # bin/os-bench
```

Linux 驗證: `GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build` 後在 `ubuntu:24.04` container 內執行。
