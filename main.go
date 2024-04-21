package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
)

type Args struct {
	Monitor *MonitorArgs `arg:"subcommand:monitor"`
	Stats   *StatsArgs   `arg:"subcommand:stats"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var args Args
	arg.MustParse(&args)

	switch {
	case args.Monitor != nil:
		monitor(ctx)
	case args.Stats != nil:
		stats(ctx)
	default:
		fmt.Fprintln(os.Stderr, "no command specified")
		os.Exit(1)
	}
}
