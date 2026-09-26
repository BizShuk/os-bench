# 術語表 (Terminology)

| 名詞 | 定義 |
| --- | --- |
| 主機效能報告 (report) | 一次 `os-bench run` 的產出: 主機規格 + 量測設定 + 所有指標, 存成一個 JSON 檔 |
| 類別 (category) | `cpu` / `memory` / `disk` / `network`, 對應 `svc/bench` 內的一個檔案, 也是 `--only` 的值 |
| 指標 (metric / result) | 單一量測值, 帶 `category`, `name`, `value`, `unit`, `better` |
| better | 指標的好方向: `higher` 越大越好, `lower` 越小越好 |
| Suite | 所有類別的固定執行順序 |
| MB/s | 每秒 1e6 bytes (十進位), 與 fio / iperf3 一致 |
| p50 / p99 | 延遲樣本的第 50 / 99 百分位 |
| queue depth 1 | 一次只有一個 I/O 在途, 前一個完成才送下一個 |
| direct I/O | 繞過 OS page cache 直接對裝置讀寫; Linux 為 `O_DIRECT`, macOS 為 `F_NOCACHE` |
| F_FULLFSYNC | macOS 上真正把資料刷進儲存裝置的 fsync, Go 的 `File.Sync` 在 darwin 預設使用 |
| multi_core_scaling | 多核吞吐量 / 單核吞吐量, 理想值接近核心數 |
| memory-backed filesystem | tmpfs / ramfs, 資料在 RAM, 不能拿來量磁碟 |
