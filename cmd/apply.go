package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/grept/pkg"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func NewApplyCmd(ctx context.Context) *cobra.Command {
	auto := false

	applyCmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply the plan",
		Run:   applyFunc(ctx, &auto),
	}

	applyCmd.Flags().BoolVarP(&auto, "auto", "a", false, "Apply fixes without confirmation")

	return applyCmd
}

func applyFunc(ctx context.Context, auto *bool) func(*cobra.Command, []string) {
	return func(_ *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println("Please specify a configuration file")
			return
		}

		filename := args[1]
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

		if !*auto {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Do you want to apply this plan? Only `yes` would be accepted. (yes/no): ")
			text, _ := reader.ReadString('\n')
			text = strings.ToLower(strings.TrimSpace(text))

			if text != "yes" {
				return
			}
		}
		err = plan.Apply()
		if err != nil {
			fmt.Printf("Error applying plan: %s\n", err.Error())
			return
		}
		fmt.Println("Plan applied successfully.")
	}
}
