package main

import (
	"fmt"
	"os"

	"github.com/yoozoo/etcd-test/cmd"
)

const (
	version = "0.0.1"
)

func main() {
	fmt.Println("Etcd tester version: ", version)

	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}
