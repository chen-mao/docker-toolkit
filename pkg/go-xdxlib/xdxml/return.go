package xdxml

import (
	"fmt"

	"github.com/XDXCT/xdxct-container-toolkit/pkg/go-xdxml/xdxml"
)

func (r Return) String() string {
	return errorStringFunc(xdxml.Return(r))
}

func (r Return) Error() string {
	return errorStringFunc(xdxml.Return(r))
}

var errorStringFunc = dafaultErrorStringFunc

var dafaultErrorStringFunc = func(r xdxml.Return) string {
	switch Return(r) {
	case SUCCESS:
		return "SUCCESS"
	case ERROR:
		return "ERROR"
	case ERROR_UNINITIALIZED:
		return "ERROR_UNINITIALIZED"
	case ERROR_INVALID_ARGUMENT:
		return "ERROR_INVALID_ARGUMENT"
	case ERROR_NOT_SUPPORTED:
		return "ERROR_NOT_SUPPORTED"
	case ERROR_NO_PERMISSION:
		return "ERROR_NO_PERMISSION"
	case ERROR_ALREADY_INITIALIZED:
		return "ERROR_ALREADY_INITIALIZED"
	case ERROR_NOT_FOUND:
		return "ERROR_NOT_FOUND"
	case ERROR_INSUFFICIENT_SIZE:
		return "ERROR_INSUFFICIENT_SIZE"
	case ERROR_INSUFFICIENT_POWER:
		return "ERROR_INSUFFICIENT_POWER"
	case ERROR_DRIVER_NOT_LOADED:
		return "ERROR_DRIVER_NOT_LOADED"
	case ERROR_TIMEOUT:
		return "ERROR_TIMEOUT"
	case ERROR_IRQ_ISSUE:
		return "ERROR_IRQ_ISSUE"
	case ERROR_LIBRARY_NOT_FOUND:
		return "ERROR_LIBRARY_NOT_FOUND"
	case ERROR_FUNCTION_NOT_FOUND:
		return "ERROR_FUNCTION_NOT_FOUND"
	case ERROR_CORRUPTED_INFOROM:
		return "ERROR_CORRUPTED_INFOROM"
	case ERROR_GPU_IS_LOST:
		return "ERROR_GPU_IS_LOST"
	case ERROR_RESET_REQUIRED:
		return "ERROR_RESET_REQUIRED"
	case ERROR_OPERATING_SYSTEM:
		return "ERROR_OPERATING_SYSTEM"
	case ERROR_LIB_RM_VERSION_MISMATCH:
		return "ERROR_LIB_RM_VERSION_MISMATCH"
	case ERROR_IN_USE:
		return "ERROR_IN_USE"
	case ERROR_MEMORY:
		return "ERROR_MEMORY"
	case ERROR_NO_DATA:
		return "ERROR_NO_DATA"
	case ERROR_VGPU_ECC_NOT_SUPPORTED:
		return "ERROR_VGPU_ECC_NOT_SUPPORTED"
	case ERROR_UNKNOWN:
		return "ERROR_UNKNOWN"
	default:
		return fmt.Sprintf("Unknown return value: %d", r)
	}
}
