package bench

import (
	"bytes"
	"compress/flate"
	"context"
	"crypto/sha256"
	"io"
	"math/rand/v2"
	"runtime"
	"strings"

	"github.com/bizshuk/os-bench/model"
)

const (
	COMPRESS_INPUT_BYTES = 1 << 20
	HASH_INPUT_BYTES     = 64 << 10
)

// runCPU 量測通用整數運算 (壓縮) 的單核與多核吞吐量, 以及 TLS 類工作的雜湊吞吐量。
func runCPU(ctx context.Context, o Options) ([]model.Result, error) {
	input := textLikeData(COMPRESS_INPUT_BYTES)
	compress := func(int) func() int {
		w, _ := flate.NewWriter(io.Discard, flate.DefaultCompression)
		return func() int {
			w.Reset(io.Discard)
			w.Write(input)
			w.Close()
			return len(input)
		}
	}

	single, err := parallelThroughput(ctx, 1, o.Duration, compress)
	if err != nil {
		return nil, err
	}
	workers := runtime.GOMAXPROCS(0)
	multi, err := parallelThroughput(ctx, workers, o.Duration, compress)
	if err != nil {
		return nil, err
	}

	block := bytes.Repeat([]byte{0xa5}, HASH_INPUT_BYTES)
	hash, err := parallelThroughput(ctx, 1, o.Duration, func(int) func() int {
		return func() int {
			sha256.Sum256(block)
			return len(block)
		}
	})
	if err != nil {
		return nil, err
	}

	return []model.Result{
		higher("cpu", "compress_single", single, UNIT_MBPS),
		higher("cpu", "compress_multi", multi, UNIT_MBPS),
		higher("cpu", "multi_core_scaling", multi/single, UNIT_X),
		higher("cpu", "sha256_single", hash, UNIT_MBPS),
	}, nil
}

// textLikeData 以固定 seed 產生類似文字的內容, 壓縮比接近一般 log / JSON。
func textLikeData(size int) []byte {
	words := strings.Fields("service request response latency error timeout user order payment cache database " +
		"queue event token session metric trace span host region status retry")
	rng := rand.New(rand.NewPCG(1, 2))
	var b bytes.Buffer
	b.Grow(size)
	for b.Len() < size {
		b.WriteString(words[rng.IntN(len(words))])
		b.WriteByte(' ')
		if rng.IntN(12) == 0 {
			b.WriteString("0123456789"[:rng.IntN(10)+1])
			b.WriteByte('\n')
		}
	}
	return b.Bytes()[:size]
}
