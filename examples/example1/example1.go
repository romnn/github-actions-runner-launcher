package main

import (
	"fmt"

	githubactionsrunnerlauncher "github.com/romnnn/github-actions-runner-launcher"
)

func run() string {
	return githubactionsrunnerlauncher.Shout("This is an example")
}

func main() {
	fmt.Println(run())
}
