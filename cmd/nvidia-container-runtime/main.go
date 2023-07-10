package main

import (
	"os"

	"github.com/XDXCT/xdxct-container-toolkit/internal/runtime"
)

func main() {
	r := runtime.New()
	err := r.Run(os.Args)
	if err != nil {
		os.Exit(1)
	}
}
