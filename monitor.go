package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"time"
)

const latencyTestPeriod = 1 * time.Second
const bandwidthTestPeriod = 5 * time.Minute

const latencyDialTimeout = 3 * time.Second
const bandwidthDialTimeout = 3 * time.Second

func monitorLatency(ctx context.Context) {
	tickerLoop(ctx, latencyTestPeriod, func() {
		start := time.Now()
		_, _, err := fetch("GET", "http://google.com", latencyDialTimeout)
		latency := time.Since(start)
		if err != nil {
			logResult(start, "reachGoogle", latency, "failed", map[string]any{"error": err.Error()})
		} else {
			logResult(start, "reachGoogle", latency, "ok", nil)
		}
	})
}

func monitorBandwidth(ctx context.Context) {
	tickerLoop(ctx, bandwidthTestPeriod, func() {
		start := time.Now()
		_, body, err := fetch("GET", "https://bcap-public-389518.s3.amazonaws.com/zero-file", bandwidthDialTimeout)
		latency := time.Since(start)
		if err != nil {
			logResult(start, "download10MiBFile", latency, "failed", map[string]any{"error": err.Error()})
		} else {
			speedKiBs := int64(float64(len(body)) / 1024.0 / latency.Seconds())
			bodyLenKiB := float64(len(body)) / 1024.0
			result := "ok"
			if len(body) != 10*1024*1024 {
				result = "failed"
			}
			logResult(start, "download10MiBFile", latency, result, map[string]any{"bodyLengthKiB": bodyLenKiB, "speedKiBs": speedKiBs})
		}
	})
}

func tickerLoop(ctx context.Context, every time.Duration, fn func()) {
	// Wait until we are close to 0ms into the current second. This is to make nicer logging
	nextSecond := time.Now().Truncate(time.Second).Add(time.Second)
	time.Sleep(time.Until(nextSecond))

	ticker := time.NewTicker(every)
	defer ticker.Stop()
	for {
		go fn()
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}

func fetch(method string, url string, dialTimeout time.Duration) (*http.Response, []byte, error) {
	client := http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return net.DialTimeout(network, addr, dialTimeout)
			},
		},
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	bodyBytes, err := io.ReadAll(res.Body)
	return res, bodyBytes, err
}
