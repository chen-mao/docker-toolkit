# XDXCT Container Toolkit CLI

The XDXCT Container Toolkit CLI `xdxct-ctk` provides a number of utilities that are useful for working with the XDXCT Container Toolkit.

## Functionality

### Configure runtimes

The `runtime` command of the `xdxct-ctk` CLI provides a set of utilities to related to the configuration
and management of supported container engines.

For example, running the following command:
```bash
xdxct-ctk runtime configure --set-as-default
```
will ensure that the XDXCT Container Runtime is added as the default runtime to the default container
engine.

### Generate CDI specifications

The [Container Device Interface (CDI)](https://tags.cncf.io/container-device-interface) provides
a vendor-agnostic mechanism to make arbitrary devices accessible in containerized environments. To allow XDXCT devices to be
used in these environments, the XDXCT Container Toolkit CLI includes functionality to generate a CDI specification for the
available XDXCT GPUs in a system.

In order to generate the CDI specification for the available devices, run the following command:\
```bash
xdxct-ctk cdi generate
```

The default is to print the specification to STDOUT and a filename can be specified using the `--output` flag.

The specification will contain a device entries as follows (where applicable):
* An `xdxct.com/gpu=gpu{INDEX}` device for each full GPU in the system
* A special device called `xdxct.com/gpu=all` which represents all available devices.

For example, to generate the CDI specification in the default location where CDI-enabled tools such as `podman`, `containerd`, `cri-o`, or the XDXCT Container Runtime can be configured to load it, the following command can be run:

```bash
sudo xdxct-ctk cdi generate --output=/etc/cdi/xdxct.yaml
```
(Note that `sudo` is used to ensure the correct permissions to write to the `/etc/cdi` folder)

With the specification generated, a GPU can be requested by specifying the fully-qualified CDI device name. With `podman` as an exmaple:
```bash
podman run --rm -ti --device=xdxct.com/gpu=gpu0 ubuntu xdxsmi -L
```
