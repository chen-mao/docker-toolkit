package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"

	"github.com/XDXCT/xdxct-container-toolkit/internal/info"
	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/XDXCT/xdxct-container-toolkit/internal/lookup"
)

var (
	debugflag   = flag.Bool("debug", false, "enable debug output")
	versionflag = flag.Bool("version", false, "enable version output")
	configflag  = flag.String("config", "", "configuration file")
)

func exit() {
	if err := recover(); err != nil {
		if _, ok := err.(runtime.Error); ok {
			log.Println(err)
		}
		if *debugflag {
			log.Printf("%s", debug.Stack())
		}
		os.Exit(1)
	}
	os.Exit(0)
}

func getCLIPath(config CLIConfig) string {
	if config.Path != nil {
		return *config.Path
	}

	var root string
	if config.Root != nil {
		root = *config.Root
	}
	if err := os.Setenv("PATH", lookup.GetPath(root)); err != nil {
		log.Panicln("couldn't set PATH variable:", err)
	}

	path, err := exec.LookPath("xdxct-container-cli")
	if err != nil {
		log.Panicln("couldn't find binary xdxct-container-cli in", os.Getenv("PATH"), ":", err)
	}
	return path
}

// getRootfsPath returns an absolute path. We don't need to resolve symlinks for now.
func getRootfsPath(config containerConfig) string {
	rootfs, err := filepath.Abs(config.Rootfs)
	if err != nil {
		log.Panicln(err)
	}
	return rootfs
}

func doPrestart() {
	var err error

	defer exit()
	log.SetFlags(0)

	hook, err := getHookConfig()
	if err != nil || hook == nil {
		log.Panicln("error getting hook config:", err)
	}
	cli := hook.XdxctContainerCLI

	container := getContainerConfig(*hook)
	if !hook.XDXCTContainerRuntimeHook.SkipModeDetection && info.ResolveAutoMode(&logInterceptor{}, hook.XDXCTContainerRuntime.Mode, container.Image) != "legacy" {
		log.Panicln("invoking the XDXCT Container Runtime Hook directly (e.g. specifying the docker --gpus flag) is not supported. Please use the XDXCT Container Runtime (e.g. specify the --runtime=xdxct flag) instead.")
	}

	xdxct := container.Xdxct
	if xdxct == nil {
		// Not a GPU container, nothing to do.
		return
	}

	rootfs := getRootfsPath(container)

	args := []string{getCLIPath(cli)}
	if cli.Root != nil {
		args = append(args, fmt.Sprintf("--root=%s", *cli.Root))
	}
	if cli.LoadKmods {
		args = append(args, "--load-kmods")
	}
	if cli.NoPivot {
		args = append(args, "--no-pivot")
	}
	args = append(args, "--debug=/1.log")
	if cli.Ldcache != nil {
		args = append(args, fmt.Sprintf("--ldcache=%s", *cli.Ldcache))
	}
	if cli.User != nil {
		args = append(args, fmt.Sprintf("--user=%s", *cli.User))
	}
	args = append(args, "configure")

	if cli.Ldconfig != nil {
		args = append(args, fmt.Sprintf("--ldconfig=%s", *cli.Ldconfig))
	}
	if cli.NoCgroups {
		args = append(args, "--no-cgroups")
	}
	if len(xdxct.Devices) > 0 {
		args = append(args, fmt.Sprintf("--device=%s", xdxct.Devices))
	}

	for _, cap := range strings.Split(xdxct.DriverCapabilities, ",") {
		if len(cap) == 0 {
			break
		}
		args = append(args, capabilityToCLI(cap))
	}

	if !hook.DisableRequire && !xdxct.DisableRequire {
		for _, req := range xdxct.Requirements {
			args = append(args, fmt.Sprintf("--require=%s", req))
		}
	}

	args = append(args, fmt.Sprintf("--pid=%s", strconv.FormatUint(uint64(container.Pid), 10)))
	args = append(args, rootfs)

	env := append(os.Environ(), cli.Environment...)
	err = syscall.Exec(args[0], args, env)
	log.Panicln("exec failed:", err)
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nCommands:\n")
	fmt.Fprintf(os.Stderr, "  prestart\n        run the prestart hook\n")
	fmt.Fprintf(os.Stderr, "  poststart\n        no-op\n")
	fmt.Fprintf(os.Stderr, "  poststop\n        no-op\n")
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if *versionflag {
		fmt.Printf("%v version %v\n", "XDXCT Container Runtime Hook", info.GetVersionString())
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	switch args[0] {
	case "prestart":
		doPrestart()
		os.Exit(0)
	case "poststart":
		fallthrough
	case "poststop":
		os.Exit(0)
	default:
		flag.Usage()
		os.Exit(2)
	}
}

// logInterceptor implements the info.Logger interface to allow for logging from this function.
type logInterceptor struct {
	logger.NullLogger
}

func (l *logInterceptor) Infof(format string, args ...interface{}) {
	log.Printf(format, args...)
}
