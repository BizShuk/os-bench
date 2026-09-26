package bench

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	mrand "math/rand/v2"
	"os"
	"path/filepath"
	"time"
	"unsafe"

	"github.com/bizshuk/os-bench/model"
)

const (
	SEQ_BLOCK_BYTES = 1 << 20
	PAGE_BYTES      = 4 << 10
	// DIRECT_ALIGN 是 O_DIRECT 要求的 buffer 與 offset 對齊單位。
	DIRECT_ALIGN = 4 << 10
)

// runDisk 量測 Options.Dir 所在磁碟的循序讀寫、4K 隨機讀取與 fsync 延遲。
// 所有檔案都以 openDirect 開啟並繞過 page cache (見 direct_*.go), 量到的是裝置而非記憶體。
func runDisk(ctx context.Context, o Options) ([]model.Result, error) {
	dir, err := os.MkdirTemp(o.Dir, "os-bench-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	data := filepath.Join(dir, "data.bin")
	size := int64(o.DiskSizeMB) << 20
	write, err := seqWrite(ctx, data, size)
	if err != nil {
		return nil, fmt.Errorf("sequential write: %w", err)
	}
	read, err := seqRead(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("sequential read: %w", err)
	}
	randSamples, randElapsed, err := randRead(ctx, data, size, o.Duration)
	if err != nil {
		return nil, fmt.Errorf("random read: %w", err)
	}
	syncSamples, syncElapsed, err := fsyncLoop(ctx, filepath.Join(dir, "wal.bin"), o.Duration)
	if err != nil {
		return nil, fmt.Errorf("fsync: %w", err)
	}

	randIOPS := float64(len(randSamples)) / randElapsed.Seconds()
	syncOps := float64(len(syncSamples)) / syncElapsed.Seconds()
	return []model.Result{
		higher("disk", "seq_write", write, UNIT_MBPS),
		higher("disk", "seq_read", read, UNIT_MBPS),
		higher("disk", "rand_read_4k", randIOPS, UNIT_IOPS),
		lower("disk", "rand_read_4k_p99", micros(percentile(randSamples, 99)), UNIT_US),
		higher("disk", "fsync_4k", syncOps, UNIT_OPS),
		lower("disk", "fsync_4k_p50", millis(percentile(syncSamples, 50)), UNIT_MS),
		lower("disk", "fsync_4k_p99", millis(percentile(syncSamples, 99)), UNIT_MS),
	}, nil
}

// seqWrite 以 1 MiB 區塊寫入不可壓縮資料, 計時包含最後的 fsync。
func seqWrite(ctx context.Context, path string, size int64) (float64, error) {
	f, err := openDirect(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	block := alignedBuf(SEQ_BLOCK_BYTES)
	rand.Read(block)

	start := time.Now()
	for written := int64(0); written < size; written += SEQ_BLOCK_BYTES {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if _, err := f.Write(block); err != nil {
			return 0, err
		}
	}
	if err := f.Sync(); err != nil {
		return 0, err
	}
	return mbPerSec(size, time.Since(start)), nil
}

func seqRead(ctx context.Context, path string) (float64, error) {
	f, err := openDirect(path, os.O_RDONLY)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	block := alignedBuf(SEQ_BLOCK_BYTES)
	var total int64
	start := time.Now()
	for {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		n, err := f.Read(block)
		total += int64(n)
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	return mbPerSec(total, time.Since(start)), nil
}

// randRead 以 queue depth 1 在檔案內隨機讀取對齊的 4 KiB, 對應資料庫索引查詢的存取型態。
func randRead(ctx context.Context, path string, size int64, d time.Duration) ([]time.Duration, time.Duration, error) {
	f, err := openDirect(path, os.O_RDONLY)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()
	page := alignedBuf(PAGE_BYTES)
	pages := size / PAGE_BYTES
	rng := mrand.New(mrand.NewPCG(5, 6))
	return latencySampler(ctx, d, func() error {
		_, err := f.ReadAt(page, rng.Int64N(pages)*PAGE_BYTES)
		return err
	})
}

// fsyncLoop 反覆附加 4 KiB 後 fsync, 對應資料庫 WAL / commit 的持久化成本。
// macOS 上 os.File.Sync 使用 F_FULLFSYNC, 會一路刷到儲存裝置。
func fsyncLoop(ctx context.Context, path string, d time.Duration) ([]time.Duration, time.Duration, error) {
	f, err := openDirect(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()
	page := alignedBuf(PAGE_BYTES)
	return latencySampler(ctx, d, func() error {
		if _, err := f.Write(page); err != nil {
			return err
		}
		return f.Sync()
	})
}

// alignedBuf 回傳起始位址對齊 DIRECT_ALIGN 的 buffer, 供 O_DIRECT 使用。
func alignedBuf(size int) []byte {
	b := make([]byte, size+DIRECT_ALIGN)
	off := (DIRECT_ALIGN - int(uintptr(unsafe.Pointer(&b[0]))%DIRECT_ALIGN)) % DIRECT_ALIGN
	return b[off : off+size]
}

func micros(d time.Duration) float64 { return float64(d.Nanoseconds()) / 1e3 }

func millis(d time.Duration) float64 { return float64(d.Nanoseconds()) / 1e6 }
