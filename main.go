package main

import (
	"fmt"
	"github.com/NETWAYS/go-check"
	"os"
	"time"
)

func main() {
	defer check.CatchPanic()

	plugin := check.NewConfig()

	plugin.Name = "check_by_winrm"
	plugin.Readme = `This Plugin executes remote commands on Windows machines through the use of WinRM.`
	plugin.Version = buildVersion()
	plugin.Timeout = 10

	config := BuildConfigFlags(plugin.FlagSet)
	plugin.ParseArguments()

	err := config.Validate()
	if err != nil {
		check.Exit(3, "could not validate parameters: %s", err)
	}

	err, rc, output := config.Run(time.Duration(plugin.Timeout) * time.Second)
	if err != nil {
		check.Exit(3, "execution failed: %s", err)
	}

	fmt.Print(output)
	os.Exit(rc)
}
