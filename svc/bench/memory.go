package bench

import (
	"context"
	"math/rand/v2"
	"runtime"
	"time"

	"github.com/bizshuk/os-bench/model"
)

const (
	// COPY_BUFFER_BYTES 遠大於各平台的 L3 / SLC, 量到的是 DRAM 而非快取。
	COPY_BUFFER_BYTES = 128 << 20
	// CHASE_ENTRIES 個 uint32 共 128 MiB, 同樣超出快取。
	CHASE_ENTRIES = 32 << 20
	// CHASE_BATCH 是每檢查一次時間之間的存取次數, 避免計時本身干擾延遲。
	CHASE_BATCH = 1 << 20
)

// runMemory 量測記憶體單核與多核複製頻寬, 以及隨機存取延遲。
func runMemory(ctx context.Context, o Options) ([]model.Result, error) {
	src := make([]byte, COPY_BUFFER_BYTES)
	dst := make([]byte, COPY_BUFFER_BYTES)
	for i := range src {
		src[i] = byte(i)
	}
	copyRegion := func(workers int) func(int) func() int {
		chunk := COPY_BUFFER_BYTES / workers
		return func(i int) func() int {
			s, d := src[i*chunk:(i+1)*chunk], dst[i*chunk:(i+1)*chunk]
			return func() int { return copy(d, s) }
		}
	}

	single, err := parallelThroughput(ctx, 1, o.Duration, copyRegion(1))
	if err != nil {
		return nil, err
	}
	workers := runtime.GOMAXPROCS(0)
	multi, err := parallelThroughput(ctx, workers, o.Duration, copyRegion(workers))
	if err != nil {
		return nil, err
	}

	latency, err := chaseLatency(ctx, o.Duration)
	if err != nil {
		return nil, err
	}

	return []model.Result{
		higher("memory", "copy_single", single, UNIT_MBPS),
		higher("memory", "copy_multi", multi, UNIT_MBPS),
		lower("memory", "random_access_latency", latency, UNIT_NS),
	}, nil
}

// chaseLatency 沿著單一隨機循環走訪陣列; 每一步都依賴上一步的結果,
// CPU 無法預取或平行化, 每次存取的平均耗時即為記憶體延遲。
func chaseLatency(ctx context.Context, d time.Duration) (float64, error) {
	next := make([]uint32, CHASE_ENTRIES)
	for i := range next {
		next[i] = uint32(i)
	}
	// Sattolo 演算法: 產生只有一個循環的排列, 保證走訪不會提早回到起點。
	rng := rand.New(rand.NewPCG(3, 4))
	for i := len(next) - 1; i > 0; i-- {
		j := rng.IntN(i)
		next[i], next[j] = next[j], next[i]
	}

	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	var p uint32
	var steps int64
	start := time.Now()
	for ctx.Err() == nil {
		for range CHASE_BATCH {
			p = next[p]
		}
		steps += CHASE_BATCH
	}
	elapsed := time.Since(start)
	runtime.KeepAlive(p)
	if err := parentErr(ctx); err != nil {
		return 0, err
	}
	return float64(elapsed.Nanoseconds()) / float64(steps), nil
}
