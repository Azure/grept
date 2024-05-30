package cmd

import (
	"errors"
	"github.com/Azure/golden"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"strings"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "",
	Short: "grept is a powerful, extensible linting tool for repositories.",
	Long: `grept is a powerful, extensible linting tool for repositories. 
Inspired by [RepoLinter](https://github.com/todogroup/repolinter), 
grept is designed to ensure that your repositories follow certain predefined standards. 
It parses and evaluates configuration files, generates plans based on the specified configuration, 
and applies the plans. This makes it an excellent tool for maintaining consistency and quality in your codebase.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		var pe *exec.ExitError
		if errors.As(err, &pe) {
			os.Exit(pe.ExitCode())
		}
		os.Exit(1)
	}
}

var cf = &commonFlags{}

type commonFlags struct {
	greptVars     []string
	greptVarFiles []string
}

func init() {
	rootCmd.PersistentFlags().StringSlice("var", cf.greptVars, "Set a value for one of the input variables in the root module of the configuration. Use this option more than once to set more than one variable.")
	rootCmd.PersistentFlags().StringSlice("var-file", cf.greptVarFiles, "Load variable values from the given file, in addition to the default files grept.greptvars and *.auto.greptvars. Use this option more than once to include more than one variables file.")
}

func varFlags(args []string) ([]golden.CliFlagAssignedVariables, error) {
	var flags []golden.CliFlagAssignedVariables
	for i := 0; i < len(args); i++ {
		if args[i] == "--var" || args[i] == "--var-file" {
			if i+1 < len(args) {
				arg := args[i+1]
				if args[i] == "--var" {
					varAssignment := strings.Split(arg, "=")
					flags = append(flags, golden.NewCliFlagAssignedVariable(varAssignment[0], varAssignment[1]))
				} else {
					flags = append(flags, golden.NewCliFlagAssignedVariableFile(arg))
				}
				i++ // skip next arg
			} else {
				return nil, errors.New("missing value for " + args[i])
			}
		}
	}
	return flags, nil
}
