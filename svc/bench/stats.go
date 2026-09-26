package bench

import (
	"context"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

// mbPerSec 以十進位 MB (1e6 bytes) 計算吞吐量, 與 fio / iperf3 的慣例一致。
func mbPerSec(bytes int64, elapsed time.Duration) float64 {
	return float64(bytes) / 1e6 / elapsed.Seconds()
}

// percentile 回傳 samples 的第 p 百分位 (0-100), samples 會被就地排序。
func percentile(samples []time.Duration, p float64) time.Duration {
	if len(samples) == 0 {
		return 0
	}
	slices.Sort(samples)
	idx := int(float64(len(samples)-1) * p / 100)
	return samples[idx]
}

// latencySampler 反覆執行 op 直到 d 結束, 回傳每次耗時與實際經過時間。
func latencySampler(ctx context.Context, d time.Duration, op func() error) ([]time.Duration, time.Duration, error) {
	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	var samples []time.Duration
	start := time.Now()
	for ctx.Err() == nil {
		t := time.Now()
		if err := op(); err != nil {
			return nil, 0, err
		}
		samples = append(samples, time.Since(t))
	}
	if err := parentErr(ctx); err != nil {
		return nil, 0, err
	}
	return samples, time.Since(start), nil
}

// parallelThroughput 讓 workers 個 goroutine 各自反覆執行 newWork(i) 產生的工作直到 d 結束,
// 工作回傳每次處理的 bytes, 結果是所有 worker 合計的 MB/s。
func parallelThroughput(ctx context.Context, workers int, d time.Duration, newWork func(worker int) func() int) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	var total atomic.Int64
	var wg sync.WaitGroup
	start := time.Now()
	for i := range workers {
		work := newWork(i)
		wg.Go(func() {
			var n int64
			for ctx.Err() == nil {
				n += int64(work())
			}
			total.Add(n)
		})
	}
	wg.Wait()
	if err := parentErr(ctx); err != nil {
		return 0, err
	}
	return mbPerSec(total.Load(), time.Since(start)), nil
}

// parentErr 區分正常到時 (DeadlineExceeded) 與使用者中斷 (Canceled)。
func parentErr(ctx context.Context) error {
	if err := ctx.Err(); err == context.Canceled {
		return err
	}
	return nil
}
