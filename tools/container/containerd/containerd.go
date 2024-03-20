package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/XDXCT/xdxct-container-toolkit/pkg/config/engine"
	"github.com/XDXCT/xdxct-container-toolkit/pkg/config/engine/containerd"
	"github.com/XDXCT/xdxct-container-toolkit/tools/container/operator"
	cli "github.com/urfave/cli/v2"
)

const (
	restartModeSignal  = "signal"
	restartModeSystemd = "systemd"
	restartModeNone    = "none"

	xdxctRuntimeName               = "xdxct"
	xdxctRuntimeBinary             = "xdxct-container-runtime"
	xdxctExperimentalRuntimeName   = "xdxct-experimental"
	xdxctExperimentalRuntimeBinary = "xdxct-container-runtime.experimental"

	defaultConfig        = "/etc/containerd/config.toml"
	defaultSocket        = "/run/containerd/containerd.sock"
	defaultRuntimeClass  = "xdxct"
	defaultRuntmeType    = "io.containerd.runc.v2"
	defaultSetAsDefault  = true
	defaultRestartMode   = restartModeSignal
	defaultHostRootMount = "/host"

	reloadBackoff     = 5 * time.Second
	maxReloadAttempts = 6

	socketMessageToGetPID = ""
)

// xdxctRuntimeBinaries defines a map of runtime names to binary names
var xdxctRuntimeBinaries = map[string]string{
	xdxctRuntimeName:             xdxctRuntimeBinary,
	xdxctExperimentalRuntimeName: xdxctExperimentalRuntimeBinary,
}

// options stores the configuration from the command line or environment variables
type options struct {
	config          string
	socket          string
	runtimeClass    string
	runtimeType     string
	setAsDefault    bool
	restartMode     string
	hostRootMount   string
	runtimeDir      string
	useLegacyConfig bool

	ContainerRuntimeModesCDIAnnotationPrefixes cli.StringSlice
}

func main() {
	options := options{}

	// Create the top-level CLI
	c := cli.NewApp()
	c.Name = "containerd"
	c.Usage = "Update a containerd config with the xdxct-container-runtime"
	c.Version = "0.1.0"

	// Create the 'setup' subcommand
	setup := cli.Command{}
	setup.Name = "setup"
	setup.Usage = "Trigger a containerd config to be updated"
	setup.ArgsUsage = "<runtime_dirname>"
	setup.Action = func(c *cli.Context) error {
		return Setup(c, &options)
	}

	// Create the 'cleanup' subcommand
	cleanup := cli.Command{}
	cleanup.Name = "cleanup"
	cleanup.Usage = "Trigger any updates made to a containerd config to be undone"
	cleanup.ArgsUsage = "<runtime_dirname>"
	cleanup.Action = func(c *cli.Context) error {
		return Cleanup(c, &options)
	}

	// Register the subcommands with the top-level CLI
	c.Commands = []*cli.Command{
		&setup,
		&cleanup,
	}

	// Setup common flags across both subcommands. All subcommands get the same
	// set of flags even if they don't use some of them. This is so that we
	// only require the user to specify one set of flags for both 'startup'
	// and 'cleanup' to simplify things.
	commonFlags := []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Usage:       "Path to the containerd config file",
			Value:       defaultConfig,
			Destination: &options.config,
			EnvVars:     []string{"CONTAINERD_CONFIG"},
		},
		&cli.StringFlag{
			Name:        "socket",
			Aliases:     []string{"s"},
			Usage:       "Path to the containerd socket file",
			Value:       defaultSocket,
			Destination: &options.socket,
			EnvVars:     []string{"CONTAINERD_SOCKET"},
		},
		&cli.StringFlag{
			Name:        "runtime-class",
			Aliases:     []string{"r"},
			Usage:       "The name of the runtime class to set for the xdxct-container-runtime",
			Value:       defaultRuntimeClass,
			Destination: &options.runtimeClass,
			EnvVars:     []string{"CONTAINERD_RUNTIME_CLASS"},
		},
		&cli.StringFlag{
			Name:        "runtime-type",
			Usage:       "The runtime_type to use for the configured runtime classes",
			Value:       defaultRuntmeType,
			Destination: &options.runtimeType,
			EnvVars:     []string{"CONTAINERD_RUNTIME_TYPE"},
		},
		// The flags below are only used by the 'setup' command.
		&cli.BoolFlag{
			Name:        "set-as-default",
			Aliases:     []string{"d"},
			Usage:       "Set xdxct-container-runtime as the default runtime",
			Value:       defaultSetAsDefault,
			Destination: &options.setAsDefault,
			EnvVars:     []string{"CONTAINERD_SET_AS_DEFAULT"},
			Hidden:      true,
		},
		&cli.StringFlag{
			Name:        "restart-mode",
			Usage:       "Specify how containerd should be restarted;  If 'none' is selected, it will not be restarted [signal | systemd | none]",
			Value:       defaultRestartMode,
			Destination: &options.restartMode,
			EnvVars:     []string{"CONTAINERD_RESTART_MODE"},
		},
		&cli.StringFlag{
			Name:        "host-root",
			Usage:       "Specify the path to the host root to be used when restarting containerd using systemd",
			Value:       defaultHostRootMount,
			Destination: &options.hostRootMount,
			EnvVars:     []string{"HOST_ROOT_MOUNT"},
		},
		&cli.BoolFlag{
			Name:        "use-legacy-config",
			Usage:       "Specify whether a legacy (pre v1.3) config should be used",
			Destination: &options.useLegacyConfig,
			EnvVars:     []string{"CONTAINERD_USE_LEGACY_CONFIG"},
		},
		&cli.StringSliceFlag{
			Name:        "xdxct-container-runtime-modes.cdi.annotation-prefixes",
			Destination: &options.ContainerRuntimeModesCDIAnnotationPrefixes,
			EnvVars:     []string{"XDXCT_CONTAINER_RUNTIME_MODES_CDI_ANNOTATION_PREFIXES"},
		},
	}

	// Update the subcommand flags with the common subcommand flags
	setup.Flags = append([]cli.Flag{}, commonFlags...)
	cleanup.Flags = append([]cli.Flag{}, commonFlags...)

	// Run the top-level CLI
	if err := c.Run(os.Args); err != nil {
		log.Fatal(fmt.Errorf("Error: %v", err))
	}
}

// Setup updates a containerd configuration to include the xdxct-containerd-runtime and reloads it
func Setup(c *cli.Context, o *options) error {
	log.Infof("Starting 'setup' for %v", c.App.Name)

	runtimeDir, err := ParseArgs(c)
	if err != nil {
		return fmt.Errorf("unable to parse args: %v", err)
	}
	o.runtimeDir = runtimeDir

	cfg, err := containerd.New(
		containerd.WithPath(o.config),
		containerd.WithRuntimeType(o.runtimeType),
		containerd.WithUseLegacyConfig(o.useLegacyConfig),
		containerd.WithContainerAnnotations(o.containerAnnotationsFromCDIPrefixes()...),
	)
	if err != nil {
		return fmt.Errorf("unable to load config: %v", err)
	}

	err = UpdateConfig(cfg, o)
	if err != nil {
		return fmt.Errorf("unable to update config: %v", err)
	}

	log.Infof("Flushing containerd config to %v", o.config)
	n, err := cfg.Save(o.config)
	if err != nil {
		return fmt.Errorf("unable to flush config: %v", err)
	}
	if n == 0 {
		log.Infof("Config file is empty, removed")
	}

	err = RestartContainerd(o)
	if err != nil {
		return fmt.Errorf("unable to restart containerd: %v", err)
	}

	log.Infof("Completed 'setup' for %v", c.App.Name)

	return nil
}

// Cleanup reverts a containerd configuration to remove the xdxct-containerd-runtime and reloads it
func Cleanup(c *cli.Context, o *options) error {
	log.Infof("Starting 'cleanup' for %v", c.App.Name)

	_, err := ParseArgs(c)
	if err != nil {
		return fmt.Errorf("unable to parse args: %v", err)
	}

	cfg, err := containerd.New(
		containerd.WithPath(o.config),
		containerd.WithRuntimeType(o.runtimeType),
		containerd.WithUseLegacyConfig(o.useLegacyConfig),
		containerd.WithContainerAnnotations(o.containerAnnotationsFromCDIPrefixes()...),
	)
	if err != nil {
		return fmt.Errorf("unable to load config: %v", err)
	}

	err = RevertConfig(cfg, o)
	if err != nil {
		return fmt.Errorf("unable to update config: %v", err)
	}

	log.Infof("Flushing containerd config to %v", o.config)
	n, err := cfg.Save(o.config)
	if err != nil {
		return fmt.Errorf("unable to flush config: %v", err)
	}
	if n == 0 {
		log.Infof("Config file is empty, removed")
	}

	err = RestartContainerd(o)
	if err != nil {
		return fmt.Errorf("unable to restart containerd: %v", err)
	}

	log.Infof("Completed 'cleanup' for %v", c.App.Name)

	return nil
}

// ParseArgs parses the command line arguments to the CLI
func ParseArgs(c *cli.Context) (string, error) {
	args := c.Args()

	log.Infof("Parsing arguments: %v", args.Slice())
	if args.Len() != 1 {
		return "", fmt.Errorf("incorrect number of arguments")
	}
	runtimeDir := args.Get(0)
	log.Infof("Successfully parsed arguments")

	return runtimeDir, nil
}

// UpdateConfig updates the containerd config to include the xdxct-container-runtime
func UpdateConfig(cfg engine.Interface, o *options) error {
	runtimes := operator.GetRuntimes(
		operator.WithXdxctRuntimeName(o.runtimeClass),
		operator.WithSetAsDefault(o.setAsDefault),
		operator.WithRoot(o.runtimeDir),
	)
	for class, runtime := range runtimes {
		err := cfg.AddRuntime(class, runtime.Path, runtime.SetAsDefault)
		if err != nil {
			return fmt.Errorf("unable to update config for runtime class '%v': %v", class, err)
		}
	}

	return nil
}

// RevertConfig reverts the containerd config to remove the xdxct-container-runtime
func RevertConfig(cfg engine.Interface, o *options) error {
	runtimes := operator.GetRuntimes(
		operator.WithXdxctRuntimeName(o.runtimeClass),
		operator.WithSetAsDefault(o.setAsDefault),
		operator.WithRoot(o.runtimeDir),
	)
	for class := range runtimes {
		err := cfg.RemoveRuntime(class)
		if err != nil {
			return fmt.Errorf("unable to revert config for runtime class '%v': %v", class, err)
		}
	}
	return nil
}

// RestartContainerd restarts containerd depending on the value of restartModeFlag
func RestartContainerd(o *options) error {
	switch o.restartMode {
	case restartModeNone:
		log.Warnf("Skipping sending signal to containerd due to --restart-mode=%v", o.restartMode)
		return nil
	case restartModeSignal:
		err := SignalContainerd(o)
		if err != nil {
			return fmt.Errorf("unable to signal containerd: %v", err)
		}
	case restartModeSystemd:
		return RestartContainerdSystemd(o.hostRootMount)
	default:
		return fmt.Errorf("Invalid restart mode specified: %v", o.restartMode)
	}

	return nil
}

// SignalContainerd sends a SIGHUP signal to the containerd daemon
func SignalContainerd(o *options) error {
	log.Infof("Sending SIGHUP signal to containerd")

	// Wrap the logic to perform the SIGHUP in a function so we can retry it on failure
	retriable := func() error {
		conn, err := net.Dial("unix", o.socket)
		if err != nil {
			return fmt.Errorf("unable to dial: %v", err)
		}
		defer conn.Close()

		sconn, err := conn.(*net.UnixConn).SyscallConn()
		if err != nil {
			return fmt.Errorf("unable to get syscall connection: %v", err)
		}

		err1 := sconn.Control(func(fd uintptr) {
			err = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_PASSCRED, 1)
		})
		if err1 != nil {
			return fmt.Errorf("unable to issue call on socket fd: %v", err1)
		}
		if err != nil {
			return fmt.Errorf("unable to SetsockoptInt on socket fd: %v", err)
		}

		_, _, err = conn.(*net.UnixConn).WriteMsgUnix([]byte(socketMessageToGetPID), nil, nil)
		if err != nil {
			return fmt.Errorf("unable to WriteMsgUnix on socket fd: %v", err)
		}

		oob := make([]byte, 1024)
		_, oobn, _, _, err := conn.(*net.UnixConn).ReadMsgUnix(nil, oob)
		if err != nil {
			return fmt.Errorf("unable to ReadMsgUnix on socket fd: %v", err)
		}

		oob = oob[:oobn]
		scm, err := syscall.ParseSocketControlMessage(oob)
		if err != nil {
			return fmt.Errorf("unable to ParseSocketControlMessage from message received on socket fd: %v", err)
		}

		ucred, err := syscall.ParseUnixCredentials(&scm[0])
		if err != nil {
			return fmt.Errorf("unable to ParseUnixCredentials from message received on socket fd: %v", err)
		}

		err = syscall.Kill(int(ucred.Pid), syscall.SIGHUP)
		if err != nil {
			return fmt.Errorf("unable to send SIGHUP to 'containerd' process: %v", err)
		}

		return nil
	}

	// Try to send a SIGHUP up to maxReloadAttempts times
	var err error
	for i := 0; i < maxReloadAttempts; i++ {
		err = retriable()
		if err == nil {
			break
		}
		if i == maxReloadAttempts-1 {
			break
		}
		log.Warnf("Error signaling containerd, attempt %v/%v: %v", i+1, maxReloadAttempts, err)
		time.Sleep(reloadBackoff)
	}
	if err != nil {
		log.Warnf("Max retries reached %v/%v, aborting", maxReloadAttempts, maxReloadAttempts)
		return err
	}

	log.Infof("Successfully signaled containerd")

	return nil
}

// RestartContainerdSystemd restarts containerd using systemctl
func RestartContainerdSystemd(hostRootMount string) error {
	log.Infof("Restarting containerd using systemd and host root mounted at %v", hostRootMount)

	command := "chroot"
	args := []string{hostRootMount, "systemctl", "restart", "containerd"}

	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error restarting containerd using systemd: %v", err)
	}

	return nil
}

// containerAnnotationsFromCDIPrefixes returns the container annotations to set for the given CDI prefixes.
func (o *options) containerAnnotationsFromCDIPrefixes() []string {
	var annotations []string
	for _, prefix := range o.ContainerRuntimeModesCDIAnnotationPrefixes.Value() {
		annotations = append(annotations, prefix+"*")
	}

	return annotations
}
