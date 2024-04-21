package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type logRecord struct {
	Start      string         `json:"start"`
	End        string         `json:"end"`
	DurationMs int64          `json:"durationMs"`
	TestType   string         `json:"testType"`
	Result     string         `json:"result"`
	Err        string         `json:"error,omitempty"`
	Message    string         `json:"message,omitempty"`
	Extra      map[string]any `json:"extra,omitempty"`
}

func logResult(start time.Time, testType string, duration time.Duration, result string, err error, extra map[string]any) {
	end := time.Now()
	format := "2006-01-02 15:04:05.000"
	record := logRecord{
		Start:      start.UTC().Format(format) + " UTC",
		End:        end.UTC().Format(format) + " UTC",
		DurationMs: duration.Milliseconds(),
		TestType:   testType,
		Result:     result,
		Extra:      extra,
	}
	jsonData, _ := json.Marshal(record)
	fmt.Printf("%s\n", string(jsonData))
}
