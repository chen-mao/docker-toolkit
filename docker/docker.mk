# Supported OSs by architecture
AMD64_TARGETS := ubuntu20.04
X86_64_TARGETS := centos7
ARM64_TARGETS := ubuntu20.04

# Define top-level build targets
docker%: SHELL:=/bin/bash

# Native targets
PLATFORM ?= $(shell uname -m)
ifeq ($(PLATFORM),x86_64)
NATIVE_TARGETS := $(AMD64_TARGETS) $(X86_64_TARGETS)
$(AMD64_TARGETS): %: %-amd64
$(X86_64_TARGETS): %: %-x86_64
else ifeq ($(PLATFORM),ppc64le)
NATIVE_TARGETS := $(PPC64LE_TARGETS)
$(PPC64LE_TARGETS): %: %-ppc64le
else ifeq ($(PLATFORM),aarch64)
NATIVE_TARGETS := $(ARM64_TARGETS) $(AARCH64_TARGETS)
$(ARM64_TARGETS): %: %-arm64
$(AARCH64_TARGETS): %: %-aarch64
endif
docker-native: $(NATIVE_TARGETS)

# amd64 targets
AMD64_TARGETS := $(patsubst %, %-amd64, $(AMD64_TARGETS))
$(AMD64_TARGETS): ARCH := amd64
$(AMD64_TARGETS): %: --%
docker-amd64: $(AMD64_TARGETS)

# x86_64 targets
X86_64_TARGETS := $(patsubst %, %-x86_64, $(X86_64_TARGETS))
$(X86_64_TARGETS): ARCH := x86_64
$(X86_64_TARGETS): %: --%
docker-x86_64: $(X86_64_TARGETS)

# arm64 targets
ARM64_TARGETS := $(patsubst %, %-arm64, $(ARM64_TARGETS))
$(ARM64_TARGETS): ARCH := arm64
$(ARM64_TARGETS): %: --%
docker-arm64: $(ARM64_TARGETS)

# aarch64 targets
AARCH64_TARGETS := $(patsubst %, %-aarch64, $(AARCH64_TARGETS))
$(AARCH64_TARGETS): ARCH := aarch64
$(AARCH64_TARGETS): %: --%
docker-aarch64: $(AARCH64_TARGETS)

# ppc64le targets
PPC64LE_TARGETS := $(patsubst %, %-ppc64le, $(PPC64LE_TARGETS))
$(PPC64LE_TARGETS): ARCH := ppc64le
$(PPC64LE_TARGETS): WITH_LIBELF := yes
$(PPC64LE_TARGETS): %: --%
docker-ppc64le: $(PPC64LE_TARGETS)

# docker target to build for all os/arch combinations
docker-all: $(AMD64_TARGETS) $(X86_64_TARGETS) \
            $(ARM64_TARGETS) $(AARCH64_TARGETS) \
            $(PPC64LE_TARGETS)

# Default variables for all private '--' targets below.
# One private target is defined for each OS we support.
--%: TARGET_PLATFORM = $(*)
--%: VERSION = $(patsubst $(OS)%-$(ARCH),%,$(TARGET_PLATFORM))
--%: BASEIMAGE = $(OS):$(VERSION)
--%: BUILDIMAGE = xdxct/$(LIB_NAME)/$(OS)$(VERSION)-$(ARCH)
--%: DOCKERFILE = $(CURDIR)/docker/Dockerfile.$(OS)
--%: ARTIFACTS_DIR = $(DIST_DIR)/$(OS)$(VERSION)/$(ARCH)
--%: docker-build-%
	@

LIBXDXCT_CONTAINER_VERSION ?= $(LIB_VERSION)
LIBXDXCT_CONTAINER_TAG ?= $(LIB_TAG)

LIBXDXCT_CONTAINER_TOOLS_VERSION := $(LIBXDXCT_CONTAINER_VERSION)$(if $(LIBXDXCT_CONTAINER_TAG),~$(LIBXDXCT_CONTAINER_TAG))-0

# private ubuntu target
--ubuntu%: OS := ubuntu

# private debian target
--debian%: OS := debian

# private centos target
--centos%: OS := centos
--centos%: DOCKERFILE = $(CURDIR)/docker/Dockerfile.rpm-yum
--centos8%: BASEIMAGE = quay.io/centos/centos:stream8

# private amazonlinux target
--amazonlinux%: OS := amazonlinux
--amazonlinux%: DOCKERFILE = $(CURDIR)/docker/Dockerfile.rpm-yum

# private opensuse-leap target
--opensuse-leap%: OS = opensuse-leap
--opensuse-leap%: BASEIMAGE = opensuse/leap:$(VERSION)

# private rhel target (actually built on centos)
--rhel%: OS := centos
--rhel%: VERSION = $(patsubst rhel%-$(ARCH),%,$(TARGET_PLATFORM))
--rhel%: ARTIFACTS_DIR = $(DIST_DIR)/rhel$(VERSION)/$(ARCH)
--rhel%: DOCKERFILE = $(CURDIR)/docker/Dockerfile.rpm-yum
--rhel8%: BASEIMAGE = quay.io/centos/centos:stream8


docker-build-%:
	@echo "Building for $(TARGET_PLATFORM)"
	# docker pull --platform=linux/$(ARCH) $(BASEIMAGE)
	DOCKER_BUILDKIT=1 \
	$(DOCKER) build \
	    --platform=linux/$(ARCH) \
	    --progress=plain \
	    --build-arg BASEIMAGE="$(BASEIMAGE)" \
	    --build-arg GOLANG_VERSION="$(GOLANG_VERSION)" \
	    --build-arg PKG_NAME="$(LIB_NAME)" \
	    --build-arg PKG_VERS="$(PACKAGE_VERSION)" \
	    --build-arg PKG_REV="$(PACKAGE_REVISION)" \
	    --build-arg LIBXDXCT_CONTAINER_TOOLS_VERSION="$(LIBXDXCT_CONTAINER_TOOLS_VERSION)" \
	    --build-arg GIT_COMMIT="$(GIT_COMMIT)" \
	    --tag $(BUILDIMAGE) \
	    --file $(DOCKERFILE) .
	$(DOCKER) run \
	    --platform=linux/$(ARCH) \
	    -e DISTRIB \
	    -e SECTION \
	    -v $(ARTIFACTS_DIR):/dist \
	    $(BUILDIMAGE)

docker-clean:
	IMAGES=$$(docker images "xdxct/$(LIB_NAME)/*" --format="{{.ID}}"); \
	if [ "$${IMAGES}" != "" ]; then \
	    docker rmi -f $${IMAGES}; \
	fi

distclean:
	rm -rf $(DIST_DIR)
