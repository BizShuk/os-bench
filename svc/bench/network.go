package bench

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/bizshuk/os-bench/model"
)

const (
	STREAM_BLOCK_BYTES = 128 << 10
	PING_BYTES         = 64
)

// runNetwork 在 loopback 上量測 TCP 吞吐量與來回延遲, 反映 kernel 網路堆疊的成本,
// 也就是同機服務互相呼叫 (sidecar, local DB) 的下限。不量測實體網路。
func runNetwork(ctx context.Context, o Options) ([]model.Result, error) {
	throughput, err := loopbackThroughput(ctx, o.Duration)
	if err != nil {
		return nil, fmt.Errorf("loopback throughput: %w", err)
	}
	rtt, err := loopbackRTT(ctx, o.Duration)
	if err != nil {
		return nil, fmt.Errorf("loopback rtt: %w", err)
	}
	return []model.Result{
		higher("network", "loopback_tcp_throughput", throughput, UNIT_MBPS),
		lower("network", "loopback_tcp_rtt_p50", micros(percentile(rtt, 50)), UNIT_US),
		lower("network", "loopback_tcp_rtt_p99", micros(percentile(rtt, 99)), UNIT_US),
	}, nil
}

// dialPair 建立一條 loopback TCP 連線, server 端交給 serve 在背景處理。
func dialPair(serve func(net.Conn)) (net.Conn, func(), error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, nil, err
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		serve(conn)
	}()
	client, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		ln.Close()
		return nil, nil, err
	}
	cleanup := func() {
		client.Close()
		ln.Close()
		<-done
	}
	return client, cleanup, nil
}

func loopbackThroughput(ctx context.Context, d time.Duration) (float64, error) {
	client, cleanup, err := dialPair(func(c net.Conn) { io.Copy(io.Discard, c) })
	if err != nil {
		return 0, err
	}
	defer cleanup()
	block := make([]byte, STREAM_BLOCK_BYTES)
	return parallelThroughput(ctx, 1, d, func(int) func() int {
		return func() int {
			n, _ := client.Write(block)
			return n
		}
	})
}

func loopbackRTT(ctx context.Context, d time.Duration) ([]time.Duration, error) {
	client, cleanup, err := dialPair(func(c net.Conn) { io.Copy(c, c) })
	if err != nil {
		return nil, err
	}
	defer cleanup()
	msg := make([]byte, PING_BYTES)
	reply := make([]byte, PING_BYTES)
	samples, _, err := latencySampler(ctx, d, func() error {
		if _, err := client.Write(msg); err != nil {
			return err
		}
		_, err := io.ReadFull(client, reply)
		return err
	})
	return samples, err
}
