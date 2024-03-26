# The XDXCT Container Runtime

The XDXCT Container Runtime is a shim for OCI-compliant low-level runtimes such as [runc](https://github.com/opencontainers/runc). When a `create` command is detected, the incoming [OCI runtime specification](https://github.com/opencontainers/runtime-spec) is modified in place and the command is forwarded to the low-level runtime.

## Configuration

The XDXCT Container Runtime uses file-based configuration, with the config stored in `/etc/xdxct-container-runtime/config.toml`. The `/etc` path can be overridden using the `XDG_CONFIG_HOME` environment variable with the `${XDG_CONFIG_HOME}/xdxct-container-runtime/config.toml` file used instead if this environment variable is set.

This config file may contain options for other components of the XDXCT container stack and for the XDXCT Container Runtime, the relevant config section is `xdxct-container-runtime`

### Logging

The `log-level` config option (default: `"info"`) specifies the log level to use and the `debug` option, if set, specifies a log file to which logs for the XDXCT Container Runtime must be written.

In addition to this, the XDXCT Container Runtime considers the value of `--log` and `--log-format` flags that may be passed to it by a container runtime such as docker or containerd. If the `--debug` flag is present the log-level specified in the config file is overridden as `"debug"`.

### Low-level Runtime Path

The `runtimes` config option allows for the low-level runtime to be specified. The first entry in this list that is an existing executable file is used as the low-level runtime. If the entry is not a path, the `PATH` is searched for a matching executable. If the entry is a path this is checked instead.

The default value for this setting is:
```toml
runtimes = [
    "docker-runc",
    "runc",
]
```

and if, for example, `crun` is to be used instead this can be changed to:
```toml
runtimes = [
    "crun",
]
```

### Runtime Mode

The `mode` config option (default `"auto"`) controls the high-level behaviour of the runtime.

#### Auto Mode

When `mode` is set to `"auto"`, the runtime employs heuristics to determine which mode to use based on, for example, the platform where the runtime is being run.

#### Legacy Mode

When `mode` is set to `"legacy"`, the XDXCT Container Runtime adds a [`prestart` hook](https://github.com/opencontainers/runtime-spec/blob/master/config.md#prestart) to the incomming OCI specification that invokes the XDXCT Container Runtime Hook for all containers created.

Alternatively the XDXCT Container Runtime can be set as the default runtime for docker. This can be done by modifying the `/etc/docker/daemon.json` file as follows:
```json
{
    "default-runtime": "xdxct",
    "runtimes": {
        "xdxct": {
            "path": "xdxct-container-runtime",
            "runtimeArgs": []
        }
    }
}
```
