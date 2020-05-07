package main

import (
	"fmt"

	"github.com/romnnn/github_actions_runner_launcher"
)

func run() string {
	return github_actions_runner_launcher.Shout("This is an example")
}

func main() {
	fmt.Println(run())
}
