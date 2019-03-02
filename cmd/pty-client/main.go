package main

import (
	"fmt"
	"os"

	"arhat.dev/kube-host-pty/pkg/cmd/pty-client"
)

func main() {
	cmd, err := ptycli.NewCmd()
	if err != nil {
		panic(err)
	}

	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
	}
}
