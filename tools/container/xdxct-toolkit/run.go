package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
	unix "golang.org/x/sys/unix"
)

const (
	runDir         = "/run/xdxct"
	pidFile        = runDir + "/toolkit.pid"
	toolkitCommand = "toolkit"
	toolkitSubDir  = "toolkit"

	defaultToolkitArgs = ""
	defaultRuntime     = "docker"
	defaultRuntimeArgs = ""
)

var availableRuntimes = map[string]struct{}{"docker": {}, "crio": {}, "containerd": {}}

var waitingForSignal = make(chan bool, 1)
var signalReceived = make(chan bool, 1)

// options stores the command line arguments
type options struct {
	noDaemon    bool
	runtime     string
	runtimeArgs string
	root        string
}

// Version defines the CLI version. This is set at build time using LD FLAGS
var Version = "development"

func main() {
	remainingArgs, root, err := ParseArgs(os.Args)
	if err != nil {
		log.Errorf("Error: unable to parse arguments: %v", err)
		os.Exit(1)
	}

	options := options{}
	// Create the top-level CLI
	c := cli.NewApp()
	c.Name = "xdxct-toolkit"
	c.Usage = "Install the xdxct-container-toolkit for use by a given runtime"
	c.UsageText = "[DESTINATION] [-n | --no-daemon] [-r | --runtime] [-u | --runtime-args]"
	c.Description = "DESTINATION points to the host path underneath which the xdxct-container-toolkit should be installed.\nIt will be installed at ${DESTINATION}/toolkit"
	c.Version = Version
	// run: 安装的流程
	c.Action = func(ctx *cli.Context) error {
		return Run(ctx, &options)
	}

	// Setup flags for the CLI
	c.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:        "no-daemon",
			Aliases:     []string{"n"},
			Usage:       "terminate immediatly after setting up the runtime. Note that no cleanup will be performed",
			Destination: &options.noDaemon,
			// 环境变量名NO_DAEMON
			EnvVars: []string{"NO_DAEMON"},
		},
		&cli.StringFlag{
			Name:        "runtime",
			Aliases:     []string{"r"},
			Usage:       "the runtime to setup on this node. One of {'docker', 'crio', 'containerd'}",
			Value:       defaultRuntime,
			Destination: &options.runtime,
			EnvVars:     []string{"RUNTIME"},
		},
		&cli.StringFlag{
			Name:        "runtime-args",
			Aliases:     []string{"u"},
			Usage:       "arguments to pass to 'docker', 'crio', or 'containerd' setup command",
			Value:       defaultRuntimeArgs,
			Destination: &options.runtimeArgs,
			EnvVars:     []string{"RUNTIME_ARGS"},
		},
		&cli.StringFlag{
			Name:        "root",
			Value:       root,
			Usage:       "the folder where the XDXCT Container Toolkit is to be installed. It will be installed to `ROOT`/toolkit",
			Destination: &options.root,
			EnvVars:     []string{"ROOT"},
		},
	}

	// Run the CLI
	log.Infof("Starting %v", c.Name)

	log.Infof("RemainingArgs: %v", remainingArgs)
	if err := c.Run(remainingArgs); err != nil {
		log.Errorf("error running xdxct-toolkit: %v", err)
		os.Exit(1)
	}

	log.Infof("Completed %v", c.Name)
}

// Run runs the core logic of the CLI
func Run(c *cli.Context, o *options) error {
	// 检查 root and runtime 是否为 docker crio containerd
	err := verifyFlags(o)
	if err != nil {
		return fmt.Errorf("unable to verify flags: %v", err)
	}
	//  将自身进程id 写入到/run/xdxct/toolkit.pid并且捕捉信号容器是否退出
	err = initialize()
	if err != nil {
		return fmt.Errorf("unable to initialize: %v", err)
	}
	defer shutdown()
	// 调用toolkit install 的 cli
	err = installToolkit(o)
	if err != nil {
		return fmt.Errorf("unable to install toolkit: %v", err)
	}
	// 调用 docker setup /user/local/xdxct/toolkit cli
	err = setupRuntime(o)
	if err != nil {
		return fmt.Errorf("unable to setup runtime: %v", err)
	}

	if !o.noDaemon {
		fmt.Println("<-- hello Daemon -->")
		err = waitForSignal()
		if err != nil {
			return fmt.Errorf("unable to wait for signal: %v", err)
		}

		err = cleanupRuntime(o)
		if err != nil {
			return fmt.Errorf("unable to cleanup runtime: %v", err)
		}
	}

	return nil
}

// ParseArgs checks if a single positional argument was defined and extracts this the root.
// If no positional arguments are defined, the it is assumed that the root is specified as a flag.
func ParseArgs(args []string) ([]string, string, error) {
	log.Infof("Parsing arguments")
	// cli: /work/xdxct-toolkit
	if len(args) < 2 {
		return args, "", nil
	}
	// 判定: /work/xdxct-toolkit root options，则返回root参数
	// 若: /work/xdxct-toolkit -n xx 则跳出循环
	var lastPositionalArg int
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			break
		}
		lastPositionalArg = i
	}

	if lastPositionalArg == 0 {
		return args, "", nil
	}

	if lastPositionalArg == 1 {
		return append([]string{args[0]}, args[2:]...), args[1], nil
	}

	return nil, "", fmt.Errorf("unexpected positional argument(s) %v", args[2:lastPositionalArg+1])
}

func verifyFlags(o *options) error {
	log.Infof("Verifying Flags")
	log.Infof("root: %v, runtime: %v", o.root, o.runtime)
	if o.root == "" {
		return fmt.Errorf("the install root must be specified")
	}

	if _, exists := availableRuntimes[o.runtime]; !exists {
		return fmt.Errorf("unknown runtime: %v", o.runtime)
	}
	return nil
}

func initialize() error {
	log.Infof("Initializing")

	f, err := os.Create(pidFile)
	if err != nil {
		return fmt.Errorf("unable to create pidfile: %v", err)
	}
	// 将自身进程id 写入到/run/xdxct/toolkit.pid，使用建议性锁打开文件
	err = unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
	if err != nil {
		log.Warnf("Unable to get exclusive lock on '%v'", pidFile)
		log.Warnf("This normally means an instance of the XDXCT toolkit Container is already running, aborting")
		return fmt.Errorf("unable to get flock on pidfile: %v", err)
	}

	_, err = f.WriteString(fmt.Sprintf("%v\n", os.Getpid()))
	if err != nil {
		return fmt.Errorf("unable to write PID to pidfile: %v", err)
	}
	// 启动goroutine捕捉系统信号，当容器删除时，则删除pid文件。
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGPIPE, syscall.SIGTERM)
	go func() {
		<-sigs
		select {
		case <-waitingForSignal:
			log.Infof("Signal received, waitForSignal")
			signalReceived <- true
		default:
			log.Infof("sigs is ------->: %v", sigs)
			log.Infof("Signal received, exiting early")
			shutdown()
			os.Exit(0)
		}
	}()

	return nil
}

// 调用toolkit cli
func installToolkit(o *options) error {
	log.Infof("Installing toolkit")

	cmdline := []string{
		toolkitCommand,
		"install",
		"--toolkit-root",
		filepath.Join(o.root, toolkitSubDir),
	}
	log.Infof("toolkit command: %v", toolkitCommand)
	// Command returns the Cmd struct to execute the named program with the given arguments.
	cmd := exec.Command("sh", "-c", strings.Join(cmdline, " "))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running %v install command: %v", cmdline, err)
	}

	return nil
}

func setupRuntime(o *options) error {
	toolkitDir := filepath.Join(o.root, toolkitSubDir)

	log.Infof("Setting up runtime")
	// docker setup /user/local/xdxct/toolkit
	cmdline := fmt.Sprintf("%v setup %v %v\n", o.runtime, o.runtimeArgs, toolkitDir)

	cmd := exec.Command("sh", "-c", cmdline)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running %v setup command: %v", o.runtime, err)
	}

	return nil
}

func waitForSignal() error {
	log.Infof("Waiting for signal")
	waitingForSignal <- true
	<-signalReceived
	return nil
}

func cleanupRuntime(o *options) error {
	toolkitDir := filepath.Join(o.root, toolkitSubDir)

	log.Infof("Cleaning up Runtime")

	cmdline := fmt.Sprintf("%v cleanup %v %v\n", o.runtime, o.runtimeArgs, toolkitDir)

	cmd := exec.Command("sh", "-c", cmdline)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running %v clean command: %v", o.runtime, err)
	}

	return nil
}

func shutdown() {
	log.Infof("Shutting Down")

	err := os.Remove(pidFile)
	if err != nil {
		log.Warnf("Unable to remove pidfile: %v", err)
	}
}
