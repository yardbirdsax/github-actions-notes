package main

import (
	"github.com/sethvargo/go-githubactions"
	"github.com/yardbirdsax/github-actions-notes/distributed-build-with-deps"
	"github.com/yardbirdsax/wyatt"
)

func main() {
	a := &distributed.Action{}
	err := wyatt.Unmarshal(a)
	if err != nil {
		githubactions.Fatalf("error configuring from Action inputs: %v", err)
	}
	err = a.Run()
	if err != nil {
		githubactions.Fatalf("error executing jobs: %v", err)
	}
}
