package xdxml

import (
	"github.com/XDXCT/xdxct-container-toolkit/pkg/go-xdxml/xdxml"
)

const (
	ERROR                         = Return(xdxml.ERROR)
	SUCCESS                       = Return(xdxml.SUCCESS)
	ERROR_UNINITIALIZED           = Return(xdxml.ERROR_UNINITIALIZED)
	ERROR_INVALID_ARGUMENT        = Return(xdxml.ERROR_INVALID_ARGUMENT)
	ERROR_NOT_SUPPORTED           = Return(xdxml.ERROR_NOT_SUPPORTED)
	ERROR_NO_PERMISSION           = Return(xdxml.ERROR_NO_PERMISSION)
	ERROR_ALREADY_INITIALIZED     = Return(xdxml.ERROR_ALREADY_INITIALIZED)
	ERROR_NOT_FOUND               = Return(xdxml.ERROR_NOT_FOUND)
	ERROR_INSUFFICIENT_SIZE       = Return(xdxml.ERROR_INSUFFICIENT_SIZE)
	ERROR_INSUFFICIENT_POWER      = Return(xdxml.ERROR_INSUFFICIENT_POWER)
	ERROR_DRIVER_NOT_LOADED       = Return(xdxml.ERROR_DRIVER_NOT_LOADED)
	ERROR_TIMEOUT                 = Return(xdxml.ERROR_TIMEOUT)
	ERROR_IRQ_ISSUE               = Return(xdxml.ERROR_IRQ_ISSUE)
	ERROR_LIBRARY_NOT_FOUND       = Return(xdxml.ERROR_LIBRARY_NOT_FOUND)
	ERROR_FUNCTION_NOT_FOUND      = Return(xdxml.ERROR_FUNCTION_NOT_FOUND)
	ERROR_CORRUPTED_INFOROM       = Return(xdxml.ERROR_CORRUPTED_INFOROM)
	ERROR_GPU_IS_LOST             = Return(xdxml.ERROR_GPU_IS_LOST)
	ERROR_RESET_REQUIRED          = Return(xdxml.ERROR_RESET_REQUIRED)
	ERROR_OPERATING_SYSTEM        = Return(xdxml.ERROR_OPERATING_SYSTEM)
	ERROR_LIB_RM_VERSION_MISMATCH = Return(xdxml.ERROR_LIB_RM_VERSION_MISMATCH)
	ERROR_IN_USE                  = Return(xdxml.ERROR_IN_USE)
	ERROR_MEMORY                  = Return(xdxml.ERROR_MEMORY)
	ERROR_NO_DATA                 = Return(xdxml.ERROR_NO_DATA)
	ERROR_VGPU_ECC_NOT_SUPPORTED  = Return(xdxml.ERROR_VGPU_ECC_NOT_SUPPORTED)
	ERROR_UNKNOWN                 = Return(xdxml.ERROR_UNKNOWN)
)
