package main

import (
	"fmt"
	"os"

	"arhat.dev/kube-host-pty/pkg/cmd/pty-device-plugin"
)

func main() {
	cmd, err := ptydp.NewCmd()
	if err != nil {
		panic(err)
	}

	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
	}
}
