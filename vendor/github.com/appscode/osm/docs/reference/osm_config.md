---
title: Osm Config
menu:
  product_osm_0.8.0:
    identifier: osm-config
    name: Osm Config
    parent: reference
product_name: osm
menu_name: product_osm_0.8.0
section_menu_id: reference
---
## osm config

OSM configuration

### Synopsis

OSM configuration

```
osm config [flags]
```

### Examples

```
osm config view
```

### Options

```
  -h, --help   help for config
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
* [osm config current-context](/docs/reference/osm_config_current-context.md)	 - Print current context
* [osm config get-contexts](/docs/reference/osm_config_get-contexts.md)	 - List available contexts
* [osm config set-context](/docs/reference/osm_config_set-context.md)	 - Set context
* [osm config use-context](/docs/reference/osm_config_use-context.md)	 - Use context
* [osm config view](/docs/reference/osm_config_view.md)	 - Print osm config

