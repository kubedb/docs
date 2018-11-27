# Release Process

The following steps must be done from a Linux x64 bit machine.

- Do a global replacement of tags so that docs point to the next release.
- Push changes to the `release-x` branch and apply new tag.
- Push all the changes to remote repo.
- Now, first build all the binaries:
```console
$ cd ~/go/src/github.com/appscode/osm
$ ./hack/make.py build; env APPSCODE_ENV=prod ./hack/make.py push; ./hack/make.py push
```
- Build and push osm docker image
```console
$ ./hack/docker/setup.sh; env APPSCODE_ENV=prod ./hack/docker/setup.sh release
```

Now, update the release notes in Github. See previous release notes to get an idea what to include there.
