package main

import (
	"os"

	sparta "github.com/mweagle/Sparta"
	"github.com/mweagle/SpartaStepServicefull/bootstrap"
)

func main() {
	workflowHooks, workflowHooksErr := bootstrap.Run(sparta.OptionsGlobal.Logger)
	if workflowHooksErr != nil {
		panic("Failed to create workflow hooks")
	}
	mainErr := sparta.MainEx("ServicefulStepFunction",
		"Self provisioning Step Function with binary as Fargate task",
		nil,
		nil,
		nil,
		workflowHooks,
		false)
	if mainErr != nil {
		os.Exit(1)
	}
}
