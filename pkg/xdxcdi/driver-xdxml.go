/**
# Copyright (c) NVIDIA CORPORATION.  All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
**/

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
	libraries := discover.NewMounts(
		logger,
		lookup.NewFileLocator(
			lookup.WithLogger(logger),
			// lookup.WithRoot(driver.Root),
			lookup.WithSearchPaths(
				"/opt/xdxgpu/lib/x86_64-linux-gnu",
				"/usr/lib64/xdxgpu",
			),
		),
		driver.Root,
		// libraryPaths,
		[]string{
			"libxdxgpu-ml.so",
			"libdrm.so.2",
			"libva-drm.so.2",
			"libva.so.2",
			"libCL_xdxgpu.so.1",
			// "libCL_xdxgpu.so*",
			"libOpenCL.so*",
			"libva-x11.so.2",

			"libEGL_mesa.so.0",
			"libEGL.so.1",
			"libglapi.so.0",
			"libGLdispatch.so.0",
			"libGLESv1_CM.so.1",
			"libGLESv1_CM_xdxgpu.so",
			"libGLESv2.so.2",
			"libGLESv2_xdxgpu.so",
			"libGL.so.1",
			"libGL_xdxgpu.so",
			"libGLX_mesa.so.0",
			"libGLX.so.0",
			"libOpenGL.so.0",
			"libusc_xdxgpu.so",
			"libufgen_xdxgpu.so",
			"libgsl_xdxgpu.so",
			"libdri_xdxgpu.so",
			"libdrm_xdxgpu.so.1",
			"libxdxgpu_mesa_wsi.so",
			"libvlk_xdxgpu.so",
			"libvlk_xdxgpu.so.1",
		},
	)

	hooks, _ := discover.NewLDCacheUpdateHook(logger, libraries, xdxctCTKPath)

	d := discover.Merge(
		libraries,
		hooks,
	)

	return d, nil
}
