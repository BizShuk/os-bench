# os-bench

在一台主機上跑一組固定的 benchmark, 產出 `主機效能報告 (host performance report)`:
主機規格 + CPU / memory / disk / network 的量測值。報告是評估「某個服務放得進這台主機嗎」
時的參考基準, 也能拿來比較不同主機, 或同一台主機升級前後的差異。

支援 macOS 與 Linux (含 Ubuntu), 單一 binary, 不依賴 `fio` / `iperf3` 等外部工具。

## 業務流程 (Domain Flow)

```text
os-bench run ─▶ 收集主機規格 ─▶ 依序量測 cpu → memory → disk → network
             ─▶ 存成 ~/.config/os-bench/data/<hostname>-<UTC時間>.json ─▶ 印出表格
os-bench list ─▶ 列出已存報告        os-bench show <report> ─▶ 重新顯示
```

## 指標 (Metrics)

| 類別 | 指標 | 對應的服務問題 |
| --- | --- | --- |
| cpu | `compress_single` / `compress_multi` (MB/s) | 一般運算型服務 (JSON, log, 壓縮) 單請求多快, 全核能撐多少 |
| cpu | `multi_core_scaling` (x) | 多核是否真的可用; 遠低於核心數代表有大小核或被限流 |
| cpu | `sha256_single` (MB/s) | TLS / 雜湊類工作, 反映硬體加密指令 |
| memory | `copy_single` / `copy_multi` (MB/s) | 大量資料搬移 (cache, 序列化) 的頻寬上限 |
| memory | `random_access_latency` (ns) | 指標追逐型工作 (hash map, 索引) 的延遲下限 |
| disk | `seq_write` / `seq_read` (MB/s) | 備份, log, 大檔處理 |
| disk | `rand_read_4k` (IOPS) + `_p99` (us) | 資料庫索引查詢 (queue depth 1) |
| disk | `fsync_4k` (ops/s) + `_p50` / `_p99` (ms) | 資料庫 commit / WAL 的持久化成本, 決定寫入型服務的 TPS 上限 |
| network | `loopback_tcp_throughput` (MB/s) | 同機服務互傳 (sidecar, local DB) 的頻寬 |
| network | `loopback_tcp_rtt_p50` / `_p99` (us) | 同機 RPC 的 kernel 網路堆疊成本 |

每個指標都帶 `better: higher | lower`, 比較報告時不必記方向。

## 使用 (Usage)

```bash
os-bench run                                # 全部類別, 每項 3 秒, 約 30 秒
os-bench run --only disk --dir /var/lib     # 量服務資料實際所在的磁碟
os-bench run --duration 1s --json > r.json  # 輸出 JSON 給其他工具
os-bench list
os-bench show mac-server-20260925T130310Z
```

## 解讀注意 (Caveats)

- `load_avg_1m` 記錄量測前的負載, 偏高時結果會被背景工作拉低, 應重跑。
- disk 量測繞過 page cache, 但在 VM / container 內, hypervisor 可能仍在宿主機快取磁碟,
  結果會高得不像實體裝置 (例如 3 us 的 4K 隨機讀); 看報告的 `virtualization` 欄判斷。
- macOS 的 fsync 走 `F_FULLFSYNC`, 會真正刷到儲存裝置, 數字比 Linux 預設 fsync 慢是預期。
- `--dir` 位於 tmpfs / ramfs 時直接拒絕, 因為量到的是 RAM。
- network 只量 loopback, 不代表實體網路; 跨主機請另用 `iperf3`。
