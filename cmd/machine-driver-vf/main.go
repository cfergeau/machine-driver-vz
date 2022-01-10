// +build darwin
package main

import (
	"fmt"
	"os"

	"github.com/code-ready/machine-driver-vf/pkg/vf"
	"github.com/code-ready/machine/libmachine/drivers/plugin"
)

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "version" {
			fmt.Printf("Driver version: %s\n", vf.DriverVersion)
			os.Exit(0)
		}
	}
	plugin.RegisterDriver(vf.NewDriver())
}
