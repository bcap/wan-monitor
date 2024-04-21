package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

const latencyTestPeriod = 1 * time.Second
const bandwidthTestPeriod = 5 * time.Minute

const latencyDialTimeout = 3 * time.Second
const bandwidthDialTimeout = 3 * time.Second

const latencyTestURL = "http://google.com"
const bandwidthTestURL = fileURL50MiB

const fileURL10MiB = "https://bcap-public-389518.s3.amazonaws.com/zero-file-10MiB"
const fileURL50MiB = "https://bcap-public-389518.s3.amazonaws.com/zero-file-50MiB"
const fileURL100MiB = "https://bcap-public-389518.s3.amazonaws.com/zero-file-100MiB"

type MonitorArgs struct {
}

func monitor(ctx context.Context) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		monitorLatency(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		monitorBandwidth(ctx)
	}()

	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}
}

func monitorLatency(ctx context.Context) {
	tickerLoop(ctx, latencyTestPeriod, func() {
		start := time.Now()
		_, _, err := fetch("GET", latencyTestURL, latencyDialTimeout)
		latency := time.Since(start)
		if err != nil {
			logResult(start, "reachGoogle", latency, "failed", err, nil)
		} else {
			logResult(start, "reachGoogle", latency, "ok", nil, nil)
		}
	})
}

func monitorBandwidth(ctx context.Context) {
	tickerLoop(ctx, bandwidthTestPeriod, func() {
		start := time.Now()
		res, body, err := fetch("GET", bandwidthTestURL, bandwidthDialTimeout)
		latency := time.Since(start)
		if err != nil {
			logResult(start, "downloadFile", latency, "failed", err, nil)
		} else if res.StatusCode != http.StatusOK {
			logResult(start, "downloadFile", latency, "failed", fmt.Errorf("got status code %d", res.StatusCode), nil)
		} else {
			speedKiBs := int64(float64(len(body)) / 1024.0 / latency.Seconds())
			bodyLenKiB := float64(len(body)) / 1024.0
			logResult(start, "downloadFile", latency, "ok", nil, map[string]any{"bodyLengthKiB": bodyLenKiB, "speedKiBs": speedKiBs})
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
