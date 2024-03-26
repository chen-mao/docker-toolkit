LIB_NAME := xdxct-container-toolkit
LIB_VERSION := 1.0.0
LIB_TAG := rc.1

# The package version is the combination of the library version and tag.
# If the tag is specified the two components are joined with a tilde (~).
PACKAGE_VERSION := $(LIB_VERSION)$(if $(LIB_TAG),~$(LIB_TAG))
PACKAGE_REVISION := 0

# Specify the xdxct-docker2 and xdxct-container-runtime package versions.
# Note: The build tooling uses `LIB_TAG` above as the version tag.
# This is appended to the versions below if specified.
XDXCT_DOCKER_VERSION := 2.14.0
XDXCT_CONTAINER_RUNTIME_VERSION := 3.14.0

# Specify the expected libxdxct-container0 version for arm64-based ubuntu builds.
LIBXDXCT_CONTAINER0_VERSION := 0.10.0+jetpack

GOLANG_VERSION := 1.20.3

GIT_COMMIT ?= $(shell git describe --match="" --dirty --long --always --abbrev=40 2> /dev/null || echo "")
GIT_COMMIT_SHORT ?= $(shell git rev-parse --short HEAD 2> /dev/null || echo "")
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD 2> /dev/null || echo "${GIT_COMMIT}")
SOURCE_DATE_EPOCH ?= $(shell git log -1 --format=%ct  2> /dev/null || echo "")
