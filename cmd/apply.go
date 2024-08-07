package cmd

import (
	"bufio"
	"fmt"
	"github.com/Azure/grept/pkg"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func NewApplyCmd() *cobra.Command {
	auto := false

	applyCmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply the plan, grept apply [-a] [path to config files]",
		RunE:  applyFunc(&auto),
	}

	applyCmd.Flags().BoolVarP(&auto, "auto", "a", false, "Apply fixes without confirmation")

	return applyCmd
}

func applyFunc(auto *bool) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, args []string) error {
		varFlags, err := varFlags(os.Args)
		if err != nil {
			return err
		}
		var cfgDir string
		if len(args) == 0 {
			cfgDir = "."
		} else {
			cfgDir = args[0]
		}
		configPath, cleaner, err := getConfigFolder(cfgDir, c.Context())
		if cleaner != nil {
			defer cleaner()
		}
		if err != nil {
			return fmt.Errorf("error getting config %s: %+v", cfgDir, err)
		}

		pwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting os wd: %+v", err)
		}
		config, err := pkg.BuildGreptConfig(pwd, configPath, c.Context(), varFlags)
		if err != nil {
			return fmt.Errorf("error parsing config: %s\n", err.Error())
		}

		plan, err := pkg.RunGreptPlan(config)
		if err != nil {
			return fmt.Errorf("Error generating plan: %s\n", err.Error())
		}

		if len(plan.FailedRules) == 0 {
			fmt.Println("All rule checks successful, nothing to do.")
			return nil
		}

		fmt.Println(plan.String())

		if !*auto {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Do you want to apply this plan? Only `yes` would be accepted. (yes/no): ")
			text, _ := reader.ReadString('\n')
			text = strings.ToLower(strings.TrimSpace(text))

			if text != "yes" {
				return nil
			}
		}
		err = plan.Apply()
		if err != nil {
			return fmt.Errorf("error applying plan: %s\n", err.Error())
		}
		fmt.Println("Plan applied successfully.")
		return nil
	}
}

func init() {
	rootCmd.AddCommand(NewApplyCmd())
}
