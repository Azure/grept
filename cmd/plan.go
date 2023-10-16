package cmd

import (
	"context"
	"fmt"
	"github.com/Azure/grept/pkg"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"os"
)

func NewPlanCmd(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "plan",
		Short: "Generates a plan based on the specified configuration",
		Run:   planFunc(ctx),
	}
}

func planFunc(ctx context.Context) func(*cobra.Command, []string) {
	return func(_ *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Please specify a configuration file")
			return
		}

		filename := args[0]
		dir, err := os.Getwd()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		fs := pkg.FsFactory()
		fileBytes, err := afero.ReadFile(fs, filename)
		if err != nil {
			fmt.Printf("Error reading file: %s\n", err.Error())
			return
		}
		content := string(fileBytes)

		config, err := pkg.ParseConfig(dir, filename, content, ctx)
		if err != nil {
			fmt.Printf("Error parsing config: %s\n", err.Error())
			return
		}

		plan, err := config.Plan()
		if err != nil {
			fmt.Printf("Error generating plan: %s\n", err.Error())
			return
		}

		if len(plan) == 0 {
			fmt.Println("All rule checks successful, nothing to do.")
			return
		}
		fmt.Println(plan.String())
	}
}
