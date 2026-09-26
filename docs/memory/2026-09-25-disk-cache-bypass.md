# 2026-09-25 disk 量測繞過 page cache 的選擇

## 背景

第一版在 Linux 用 `POSIX_FADV_DONTNEED` 丟快取, macOS 用 `F_NOCACHE`。
在 ubuntu:24.04 container (OrbStack) 跑出 `rand_read_4k` 835k IOPS, p99 3 us, 明顯是記憶體。

## 原因

- `seq_read` 讀完整個檔案, 會把頁面重新填回 page cache。
- overlayfs 不一定把 fadvise 傳到底層檔案系統。

## 決策

- Linux 改用 `O_DIRECT`, 所有 buffer 以 4 KiB 對齊。這與 `fio direct=1` 的作法一致。
- `O_DIRECT` 在 tmpfs 上仍可開檔, 因此 `run` 依 fstype 拒絕 tmpfs / ramfs。
  預設 `--dir` 也從 OS temp dir 改成 app data dir, 因為 Debian 13 / Fedora 的 `/tmp` 是 tmpfs。

## 未解

OrbStack VM 內改用 `O_DIRECT` 後仍是 ~500k IOPS / 3.6 us: VM 的虛擬磁碟被宿主 macOS 快取,
guest 端無法繞過。這是 VM 環境的真實特性, 報告的 `virtualization` 欄要一起看; 實體 Linux 尚未驗證。
