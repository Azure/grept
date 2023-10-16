package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Azure/grept/cmd"
	"os"
)

func usage() {
	fmt.Printf("Usage: %s <command> [arguments]\n", os.Args[0])
	fmt.Println("\nThe commands are:\n\nplan\tGenerates a plan based on the specified configuration\napply\tApply the plan")
	fmt.Println("\nUse \"grept help [command]\" for more information about a command.")
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
	case "apply":
		cobraCmd := cmd.NewApplyCmd(ctx)
		_ = cobraCmd.Execute()
	default:
		usage()
	}
}
