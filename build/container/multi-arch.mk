PUSH_ON_BUILD ?= false
DOCKER_BUILD_OPTIONS = --output=type=image,push=$(PUSH_ON_BUILD)
DOCKER_BUILD_PLATFORM_OPTIONS = --platform=linux/amd64,linux/arm64

# We only have x86_64 packages for centos7
build-centos7: DOCKER_BUILD_PLATFORM_OPTIONS = --platform=linux/amd64

# We only generate amd64 image for ubuntu20.04
build-ubuntu20.04: DOCKER_BUILD_PLATFORM_OPTIONS = --platform=linux/amd64

# We only generate a single image for packaging targets
build-packaging: DOCKER_BUILD_PLATFORM_OPTIONS = --platform=linux/amd64
