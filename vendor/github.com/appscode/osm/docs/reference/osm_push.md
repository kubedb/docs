---
title: Osm Push
menu:
  product_osm_0.8.0:
    identifier: osm-push
    name: Osm Push
    parent: reference
product_name: osm
menu_name: product_osm_0.8.0
section_menu_id: reference
---
## osm push

Push item to container

### Synopsis

Push item to container

```
osm push <src> <dest> [flags]
```

### Examples

```
osm push -c mybucket f1.txt /tmp/f1.txt
```

### Options

```
  -c, --container string   Name of container
      --context string     Name of osmconfig context to use
  -h, --help               help for push
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
      --enable-analytics                 Send usage events to Google Analytics (default true)
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
      --osmconfig string                 Path to osm config (default "$HOME/.osm/config")
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO

* [osm](/docs/reference/osm.md)	 - Object Store Manipulator by AppsCode

