[![Go Report Card](https://goreportcard.com/badge/github.com/appscode/osm)](https://goreportcard.com/report/github.com/appscode/osm)

# osm
Object Store Manipulator (osm: pronounced like `awesome`) - `curl` for cloud storage services. 🙌 `osm` can create & delete buckets and upload, download & delete files from buckets for AWS S3, AWS S3 compatible other storage services(i.e. Minio), DigitalOcean Spaces, Google Cloud Storage, Microsoft Azure storage and OpenStack Swift. Its single binary can be easily packaged instead of official python based clis inside Docker images.

## Install OSM
You can download and install a pre-built binary:
```console
# Linux amd 64-bit:
wget -O osm https://cdn.appscode.com/binaries/osm/0.8.0/osm-linux-amd64 \
  && chmod +x osm \
  && sudo mv osm /usr/local/bin/

# Linux 386 32-bit:
wget -O osm https://cdn.appscode.com/binaries/osm/0.8.0/osm-linux-386 \
  && chmod +x osm \
  && sudo mv osm /usr/local/bin/

# Mac 64-bit
wget -O osm https://cdn.appscode.com/binaries/osm/0.8.0/osm-darwin-amd64 \
  && chmod +x osm \
  && sudo mv osm /usr/local/bin/

# Mac 32-bit
wget -O osm https://cdn.appscode.com/binaries/osm/0.8.0/osm-darwin-386 \
  && chmod +x osm \
  && sudo mv osm /usr/local/bin/
```

To build from source, run: `go get -u github.com/appscode/osm`

## Usage
```console
osm [command] [flags]
osm [command]

Available Commands:
  config      OSM configuration
  help        Help about any command
  lc          List containers
  ls          List items in a container
  mc          Make container
  pull        Pull item from container
  push        Push item to container
  rc          Remove container
  rm          Remove item from container
  stat        Stat item from container
  version     Prints binary version number.

Flags:
      --alsologtostderr                  log to standard error as well as files
      --enable-analytics                 Send usage events to Google Analytics (default true)
  -h, --help                             help for osm
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
      --osmconfig string                 Path to osm config (default "$HOME/.osm/config")
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging

Use "osm [command] --help" for more information about a command.

```

### OSM Configuration
`osm` stores credentials necessary to connect to a cloud storage provider in YAML format in `$HOME/.osm/config` file.
This allows providing commands one time for multiple subsequent operations with a cloud provider.
```console
# AWS S3:
osm config set-context osm-s3 --provider=s3 --s3.access_key_id=<key_id> --s3.secret_key=<secret_key> --s3.region=us-east-1

# TLS secure Minio server
osm config set-context osm-minio --provider=s3 --s3.access_key_id=<minio_access_key> --s3.secret_key=<minio_secret_key> --s3.endpoint=<minio_server_address> --s3.cacert_file=<root_ca_file_path>

# DigitalOcean Spaces:
osm config set-context osm-do --provider=s3 --s3.access_key_id=<key_id> --s3.secret_key=<secret_key> --s3.endpoint=nyc3.digitaloceanspaces.com

# Google Cloud Storage:
osm config set-context osm-gs --provider=google --google.json_key_path=<path_sa_file> --google.project_id=<my_project>

# Microsoft Azure ARM Storage:
osm config set-context osm-az --provider=azure --azure.account=<storage_ac> --azure.key=<key>
```

### Bucket Operations
```console
# create bucket
osm mc mybucket

# upload file to bucket
osm push -c mybucket ~/Downloads/appscode.pdf a/b/c.pdf

# print uploaded file attributes
osm stat -c mybucket a/b/c.pdf

# download file from bucket
osm pull -c mybucket a/b/c.pdf /tmp/d.pdf

# list bucket
osm ls mybucket

# remove file from bucket
osm rm -c mybucket a/b/c.pdf

# remove bucket (use -f to delete any files inside)
osm rc -f mybucket
```

## Contribution guidelines
Want to help improve OSM? Please start [here](/CONTRIBUTING.md).

## Support
If you have any questions, you can reach out to us.
* [Slack](https://slack.appscode.com)
* [Twitter](https://twitter.com/AppsCodeHQ)
* [Website](https://appscode.com)
