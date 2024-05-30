package cmd

import (
	"errors"
	"fmt"
	"github.com/Azure/grept/pkg"
	"os"

	"github.com/Azure/golden"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/peterh/liner"
	"github.com/spf13/cobra"
)

func NewConsoleCmd() *cobra.Command {
	replCmd := &cobra.Command{
		Use:   "console",
		Short: "Start REPL mode, grept console [path to config files]",
		RunE:  replFunc(),
	}

	return replCmd
}

func replFunc() func(*cobra.Command, []string) error {
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
			return fmt.Errorf("error parsing config: %+v", err)
		}
		_, err = pkg.RunGreptPlan(config)
		if err != nil {
			return fmt.Errorf("error plan config: %+v", err)
		}

		line := liner.NewLiner()
		defer func() {
			_ = line.Close()
		}()

		line.SetCtrlCAborts(true)
		fmt.Println("Entering debuging mode, press `quit` or `exit` or Ctrl+C to quit.")

		for {
			if input, err := line.Prompt("debug> "); err == nil {
				if input == "quit" || input == "exit" {
					return nil
				}
				line.AppendHistory(input)
				expression, diag := hclsyntax.ParseExpression([]byte(input), "repl.hcl", hcl.InitialPos)
				if diag.HasErrors() {
					fmt.Printf("%s\n", diag.Error())
					continue
				}
				value, diag := expression.Value(config.EvalContext())
				if diag.HasErrors() {
					fmt.Printf("%s\n", diag.Error())
					continue
				}
				fmt.Println(golden.CtyValueToString(value))
			} else if errors.Is(err, liner.ErrPromptAborted) {
				fmt.Println("Aborted")
				break
			} else {
				fmt.Println("Error reading line: ", err)
				break
			}
		}

		return nil
	}
}

func init() {
	rootCmd.AddCommand(NewConsoleCmd())
}
