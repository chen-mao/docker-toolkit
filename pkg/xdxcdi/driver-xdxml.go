package xdxcdi

import (
	"fmt"

	"github.com/XDXCT/xdxct-container-toolkit/internal/discover"
	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/XDXCT/xdxct-container-toolkit/internal/lookup"
	"github.com/XDXCT/xdxct-container-toolkit/internal/lookup/root"
	"github.com/XDXCT/xdxct-container-toolkit/pkg/go-xdxlib/xdxml"
)

// NewDriverDiscoverer creates a discoverer for the libraries and binaries associated with a driver installation.
// The supplied NVML Library is used to query the expected driver version.
func NewDriverDiscoverer(logger logger.Interface, driver *root.Driver, xdxctCTKPath string, nvmllib xdxml.Interface) (discover.Discover, error) {
	return newDriverVersionDiscoverer(logger, driver, xdxctCTKPath)
}

func newDriverVersionDiscoverer(logger logger.Interface, driver *root.Driver, xdxctCTKPath string) (discover.Discover, error) {
	libraries, err := NewDriverLibraryDiscoverer(logger, driver, xdxctCTKPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create discoverer for driver libraries: %v", err)
	}

	binaries := NewDriverBinariesDiscoverer(logger, driver.Root)
	xdxsmiPyDir := NewDriverPyDirDiscoverer(logger, driver.Root)

	d := discover.Merge(
		libraries,
		xdxsmiPyDir,
		binaries,
	)

	return d, nil
}

func NewDriverBinariesDiscoverer(logger logger.Interface, driverRoot string) discover.Discover {
	return discover.NewMounts(
		logger,
		lookup.NewExecutableLocator(
			logger,
			driverRoot,
		),
		driverRoot,
		[]string{
			"xdxsmi", /* System management interface */
		},
	)
}

func NewDriverPyDirDiscoverer(logger logger.Interface, driverRoot string) discover.Discover {
	return discover.NewMounts(
		logger,
		lookup.NewDirectoryLocator(
			lookup.WithLogger(logger),
			lookup.WithRoot(driverRoot),
		),
		driverRoot,
		[]string{
			"/usr/lib/python3/dist-packages/xdxsmi", /* System management interface directory */
		},
	)
}

// NewDriverLibraryDiscoverer creates a discoverer for the libraries associated with the specified driver version.
func NewDriverLibraryDiscoverer(logger logger.Interface, driver *root.Driver, xdxctCTKPath string) (discover.Discover, error) {
	libxdxgpu_driver := []string{
		"libxdxgpu-ml.so.*.*",
		"libdrm.so",
		"libva-drm.so",
		"libva.so",
		"libva-x11.so",
		"libEGL_mesa.so",
		"libEGL.so",
		"libglapi.so",
		"libGLdispatch.so",
		"libGLESv1_CM.so",
		"libGLESv1_CM_xdxgpu.so",
		"libGLESv2.so",
		"libGLESv2_xdxgpu.so",
		"libGL.so",
		"libGL_xdxgpu.so",
		"libGLX_mesa.so",
		"libGLX.so",
		"libOpenGL.so",
		"libusc_xdxgpu.so",
		"libufgen_xdxgpu.so",
		"libgsl_xdxgpu.so",
		"libdri_xdxgpu.so",
		"libdrm_xdxgpu.so",
		"libvlk_xdxgpu.so",
		"libxdxgpu_mesa_wsi.so",
		"libOpenCL.so*",
	}

	libraries := discover.NewMounts(
		logger,
		lookup.NewFileLocator(
			lookup.WithLogger(logger),
			// lookup.WithRoot(driver.Root),
			lookup.WithSearchPaths(
				"/opt/xdxgpu/lib/x86_64-linux-gnu",
				"/usr/lib/aarch64-linux-gnu/xdxgpu",
				"/usr/lib64/xdxgpu",
				"/usr/lib/x86_64-linux-gnu/xdxgpu",
			),
		),
		driver.Root,
		libxdxgpu_driver,
	)

	hooks, _ := discover.NewLDCacheUpdateHook(logger, libraries, xdxctCTKPath)

	d := discover.Merge(
		libraries,
		hooks,
	)

	return d, nil
}
