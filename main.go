package main

import (
	"fmt"

	sparta "github.com/mweagle/Sparta"
	"github.com/mweagle/SpartaStepServicefull/bootstrap"
	"github.com/spf13/cobra"
)

func main() {

	// Add a hook to do something
	fargateTask := &cobra.Command{
		Use:   "fargateTask",
		Short: "Sample Fargate task",
		Long:  `Sample Fargate task that simply logs a message"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Fargate TASK GOES HERE! ❤️")
			return nil
		},
	}

	sparta.CommandLineOptions.Root.AddCommand(fargateTask)
	workflowHooks, workflowHooksErr := bootstrap.Run(sparta.OptionsGlobal.Logger)
	if workflowHooksErr != nil {
		panic("Failed to create workflow hooks")
	}

	sparta.MainEx("ServicefulStepFunction",
		"Self provisioning Step Function with binary as Fargate task",
		nil,
		nil,
		nil,
		workflowHooks,
		false)
}
