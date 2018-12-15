// +build lambdabinary

package bootstrap

import (
	"fmt"

	sparta "github.com/mweagle/Sparta"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// When we're the lambda binary there isn't much to do here...
func Run(logger *logrus.Logger) (*sparta.WorkflowHooks, error) {
	// Add a hook to do something
	fargateTask := &cobra.Command{
		Use:   "fargateTask",
		Short: "Sample Fargate task",
		Long:  `Sample Fargate task that simply logs a message"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Insert your Fargate code here! 🎉")
			return nil
		},
	}
	// Register the command with the Sparta root dispatcher. This
	// command `fargateTask` matches the command line option in the
	// Dockerfile that is used to build the image.
	sparta.CommandLineOptions.Root.AddCommand(fargateTask)
	// That's all folks...
	return nil, nil
}
