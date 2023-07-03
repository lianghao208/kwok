## kwokctl start cluster

Start a cluster

```
kwokctl start cluster [flags]
```

### Options

```
  -h, --help               help for cluster
      --timeout duration   Timeout for waiting for the cluster to be started
      --wait duration      Wait for the cluster to be ready
```

### Options inherited from parent commands

```
  -c, --config stringArray   config path (default [~/.kwok/kwok.yaml])
      --dry-run              Print the command that would be executed, but do not execute it
      --name string          cluster name (default "kwok")
  -v, --v log-level          number for the log level verbosity (DEBUG, INFO, WARN, ERROR) or (-4, 0, 4, 8) (default INFO)
```

### SEE ALSO

* [kwokctl start](kwokctl_start.md)	 - Start one of [cluster]

