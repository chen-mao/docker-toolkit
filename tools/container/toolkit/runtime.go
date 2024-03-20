package main

import (
	"fmt"
	"path/filepath"

	"github.com/XDXCT/xdxct-container-toolkit/tools/container/operator"
)

const (
	xdxctContainerRuntimeSource = "/usr/bin/xdxct-container-runtime"

	xdxctExperimentalContainerRuntimeSource = "xdxct-container-runtime.experimental"
)

// installContainerRuntimes sets up the XDXCT container runtimes, copying the executables
// and implementing the required wrapper
func installContainerRuntimes(toolkitDir string, driverRoot string) error {
	runtimes := operator.GetRuntimes()
	/*-----
	runtimes:
		name: xdxct-experimental
		Path: /usr/bin/xdxct-container-runtime.experimental

		name: xdxct-cdi
		Path: /usr/bin/xdxct-container-runtime-cdi

		name: xdxct-legacy
		Path: /usr/bin/xdxct-container-runtime-legacy

		name: xdxct
		Path: /usr/bin/xdxct-container-runtime
	-----*/
	for _, runtime := range runtimes {
		if filepath.Base(runtime.Path) == xdxctExperimentalContainerRuntimeSource {
			continue
		}

		r := newXdxctContainerRuntimeInstaller(runtime.Path)
		// 拷贝/usr/bin 到 /usr/local/xdxct/toolkit 目录
		_, err := r.install(toolkitDir)
		if err != nil {
			return fmt.Errorf("error installing XDXCT container runtime: %v", err)
		}
	}

	// Install the experimental runtime and treat failures as non-fatal.
	// err := installExperimentalRuntime(toolkitDir, driverRoot)
	// if err != nil {
	// 	log.Warnf("Could not install experimental runtime: %v", err)
	// }

	return nil
}

// installExperimentalRuntime ensures that the experimental XDXCT Container runtime is installed
// func installExperimentalRuntime(toolkitDir string, driverRoot string) error {
// 	libraryRoot, err := findLibraryRoot(driverRoot)
// 	if err != nil {
// 		log.Warnf("Error finding library path for root %v: %v", driverRoot, err)
// 	}
// 	log.Infof("Using library root %v", libraryRoot)

// 	e := newXdxctContainerRuntimeExperimentalInstaller(libraryRoot)
// 	_, err = e.install(toolkitDir)
// 	if err != nil {
// 		return fmt.Errorf("error installing experimental XDXCT Container Runtime: %v", err)
// 	}

// 	return nil
// }

// newXdxctContainerRuntimeInstaller returns a new executable installer for the XDXCT container runtime.
// This installer will copy the specified source exectuable to the toolkit directory.
// The executable is copied to a file with the same name as the source, but with a ".real" suffix and a wrapper is
// created to allow for the configuration of the runtime environment.
func newXdxctContainerRuntimeInstaller(source string) *executable {
	wrapperName := filepath.Base(source)
	dotfileName := wrapperName + ".real"
	target := executableTarget{
		dotfileName: dotfileName,
		wrapperName: wrapperName,
	}
	return newRuntimeInstaller(source, target, nil)
}

// func newXdxctContainerRuntimeExperimentalInstaller(libraryRoot string) *executable {
// 	source := xdxctExperimentalContainerRuntimeSource
// 	wrapperName := filepath.Base(source)
// 	dotfileName := wrapperName + ".real"
// 	target := executableTarget{
// 		dotfileName: dotfileName,
// 		wrapperName: wrapperName,
// 	}

// 	env := make(map[string]string)
// 	if libraryRoot != "" {
// 		env["LD_LIBRARY_PATH"] = strings.Join([]string{libraryRoot, "$LD_LIBRARY_PATH"}, ":")
// 	}
// 	return newRuntimeInstaller(source, target, env)
// }

func newRuntimeInstaller(source string, target executableTarget, env map[string]string) *executable {
	/*---------
		检查xdxct 驱动模块是否存在。
		读取modules文件的内容，然后将结果重定向标准输出。
		检查上条命令结果，0表示成功执行的状态。否则表示xdxct驱动模块尚未加载
	---------*/
	// preLines := []string{
	// 	"",
	// 	"cat /proc/modules | grep -e \"^xdxct \" >/dev/null 2>&1",
	// 	"if [ \"${?}\" != \"0\" ]; then",
	// 	"	echo \"xdxct driver modules are not yet loaded, invoking runc directly\"",
	// 	"	exec runc \"$@\"",
	// 	"fi",
	// 	"",
	// }

	runtimeEnv := make(map[string]string)
	runtimeEnv["XDG_CONFIG_HOME"] = filepath.Join(destDirPattern, ".config")
	for k, v := range env {
		runtimeEnv[k] = v
	}

	r := executable{
		source:   source,
		target:   target,
		env:      runtimeEnv,
		preLines: nil,
	}

	return &r
}

// func findLibraryRoot(root string) (string, error) {
// 	libxdxctmlPath, err := findManagementLibrary(root)
// 	if err != nil {
// 		return "", fmt.Errorf("error locating XDXCT management library: %v", err)
// 	}

// 	return filepath.Dir(libxdxctmlPath), nil
// }

// func findManagementLibrary(root string) (string, error) {
// 	return findLibrary(root, "libxdxct-ml.so")
// }
