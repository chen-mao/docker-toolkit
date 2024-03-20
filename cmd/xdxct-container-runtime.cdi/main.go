package main

import (
	"os"

	"github.com/XDXCT/xdxct-container-toolkit/internal/runtime"
)

func main() {
	rt := runtime.New(
		runtime.WithModeOverride("cdi"),
	)

	err := rt.Run(os.Args)
	if err != nil {
		os.Exit(1)
	}
}
