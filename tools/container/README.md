## Introduction

This repository contains tools that allow docker, containerd, or cri-o to be configured to use the XDXCT Container Toolkit.

These will be migrated into an upcoming `xdxct-ctk` CLI as required.

### Docker

After building the `docker` binary, run:
```bash
docker setup \
    --runtime-name NAME \
        /run/xdxct/toolkit
```

Configure the `xdxct-container-runtime` as a docker runtime named `NAME`. If the `--runtime-name` flag is not specified, this runtime would be called `xdxct`. A runtime named `xdxct-experimental` will also be configured using the `xdxct-container-runtime.experimental` OCI-compliant runtime shim.

Since `--set-as-default` is enabled by default, the specified runtime name will also be set as the default docker runtime. This can be disabled by explicityly specifying `--set-as-default=false`.

**Note**: If `--runtime-name` is specified as `xdxct-experimental` explicitly, the `xdxct-experimental` runtime will be configured as the default runtime, with the `xdxct` runtime still configured and available for use.

The following table describes the behaviour for different `--runtime-name` and `--set-as-default` flag combinations.

| Flags                                                       | Installed Runtimes              | Default Runtime       |
|-------------------------------------------------------------|:--------------------------------|:----------------------|
| **NONE SPECIFIED**                                          | `xdxct`, `xdxct-experimental` | `xdxct`              |
| `--runtime-name xdxct`                                     | `xdxct`, `xdxct-experimental` | `xdxct`              |
| `--runtime-name NAME`                                       | `NAME`, `xdxct-experimental`   | `NAME`                |
| `--runtime-name xdxct-experimental`                        | `xdxct`, `xdxct-experimental` | `xdxct-experimental` |
| `--set-as-default`                                          | `xdxct`, `xdxct-experimental` | `xdxct`              |
| `--set-as-default --runtime-name xdxct`                    | `xdxct`, `xdxct-experimental` | `xdxct`              |
| `--set-as-default --runtime-name NAME`                      | `NAME`, `xdxct-experimental`   | `NAME`                |
| `--set-as-default --runtime-name xdxct-experimental`       | `xdxct`, `xdxct-experimental` | `xdxct-experimental` |
| `--set-as-default=false`                                    | `xdxct`, `xdxct-experimental` | **NOT SET**           |
| `--set-as-default=false --runtime-name NAME`                | `NAME`, `xdxct-experimental`   | **NOT SET**           |
| `--set-as-default=false --runtime-name xdxct`              | `xdxct`, `xdxct-experimental` | **NOT SET**           |
| `--set-as-default=false --runtime-name xdxct-experimental` | `xdxct`, `xdxct-experimental` | **NOT SET**           |

These combinations also hold for the environment variables that map to the command line flags: `DOCKER_RUNTIME_NAME`, `DOCKER_SET_AS_DEFAULT`.

### Containerd
After running the `containerd` binary, run:
```bash
containerd setup \
    --runtime-class NAME \
        /run/xdxct/toolkit
```

Configure the `xdxct-container-runtime` as a runtime class named `NAME`. If the `--runtime-class` flag is not specified, this runtime would be called `xdxct`. A runtime class named `xdxct-experimental` will also be configured using the `xdxct-container-runtime.experimental` OCI-compliant runtime shim.

Adding the `--set-as-default` flag as follows:
```bash
containerd setup \
    --runtime-class NAME \
    --set-as-default \
        /run/xdxct/toolkit
```
will set the runtime class `NAME` (or `xdxct` if not specified) as the default runtime class.

**Note**: If `--runtime-class` is specified as `xdxct-experimental` explicitly and `--set-as-default` is specified, the `xdxct-experimental` runtime will be configured as the default runtime class, with the `xdxct` runtime class still configured and available for use.

The following table describes the behaviour for different `--runtime-class` and `--set-as-default` flag combinations.

| Flags                                                  | Installed Runtime Classes       | Default Runtime Class |
|--------------------------------------------------------|:--------------------------------|:----------------------|
| **NONE SPECIFIED**                                     | `xdxct`, `xdxct-experimental` | **NOT SET**           |
| `--runtime-class NAME`                                 | `NAME`, `xdxct-experimental`   | **NOT SET**           |
| `--runtime-class xdxct`                               | `xdxct`, `xdxct-experimental` | **NOT SET**           |
| `--runtime-class xdxct-experimental`                  | `xdxct`, `xdxct-experimental` | **NOT SET**           |
| `--set-as-default`                                     | `xdxct`, `xdxct-experimental` | `xdxct`              |
| `--set-as-default --runtime-class NAME`                | `NAME`, `xdxct-experimental`   | `NAME`                |
| `--set-as-default --runtime-class xdxct`              | `xdxct`, `xdxct-experimental` | `xdxct`              |
| `--set-as-default --runtime-class xdxct-experimental` | `xdxct`, `xdxct-experimental` | `xdxct-experimental` |

These combinations also hold for the environment variables that map to the command line flags.
