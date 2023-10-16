package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Azure/grept/cmd"
	"os"
)

func usage() {
	_, _ = fmt.Fprintf(os.Stderr, "usage: %s [flags] plan\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(0)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() == 0 {
		usage()
	}

	command := flag.Arg(0)

	ctx := context.Background()

	switch command {
	case "plan":
		cobraCmd := cmd.NewPlanCmd(ctx)
		_ = cobraCmd.Execute()
	default:
		usage()
	}
}
