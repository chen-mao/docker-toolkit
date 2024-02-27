/*
 * Copyright (c) 2024, XDXCT CORPORATION.  All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
 
package xdxml

import (
	"fmt"

	"github.com/XDXCT/xdxct-container-toolkit/pkg/go-xdxml/dl"
)

import "C"

const (
	xdxmlLibraryName      = "libxdxgpu-ml.so"
	xdxmlLibraryLoadFlags = dl.RTLD_LAZY | dl.RTLD_GLOBAL
)

var xdxml *dl.DynamicLibrary

// xdml_init()
func Init() Return {
	lib := dl.New(xdxmlLibraryName, xdxmlLibraryLoadFlags)
	err := lib.Open()
	if err != nil {
		return ERROR_LIBRARY_NOT_FOUND
	}
	xdxml = lib

	return xdxml_init()
}

// xdxml_shutdown()
func Shutdown() Return {
	ret := xdxml_shutdown()
	if ret != SUCCESS {
		return ret
	}

	err := xdxml.Close()
	if err != nil {
		panic(fmt.Sprintf("error closing %s: %v", xdxmlLibraryName, err))
	}

	return ret
}
