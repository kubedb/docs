#!/usr/bin/env python


# http://stackoverflow.com/a/14050282
def check_antipackage():
    from sys import version_info
    sys_version = version_info[:2]
    found = True
    if sys_version < (3, 0):
        # 'python 2'
        from pkgutil import find_loader
        found = find_loader('antipackage') is not None
    elif sys_version <= (3, 3):
        # 'python <= 3.3'
        from importlib import find_loader
        found = find_loader('antipackage') is not None
    else:
        # 'python >= 3.4'
        from importlib import util
        found = util.find_spec('antipackage') is not None
    if not found:
        print('Install missing package "antipackage"')
        print('Example: pip install git+https://github.com/ellisonbg/antipackage.git#egg=antipackage')
        from sys import exit
        exit(1)
check_antipackage()

# ref: https://github.com/ellisonbg/antipackage
import antipackage
from github.appscode.libbuild import libbuild, pydotenv

import os
import os.path
import subprocess
import sys
from os.path import expandvars, join, dirname

libbuild.REPO_ROOT = expandvars('$GOPATH') + '/src/github.com/appscode/kubernetes-webhook-util'
BUILD_METADATA = libbuild.metadata(libbuild.REPO_ROOT)


def call(cmd, stdin=None, cwd=libbuild.REPO_ROOT):
    print(cmd)
    return subprocess.call([expandvars(cmd)], shell=True, stdin=stdin, cwd=cwd)


def die(status):
    if status:
        sys.exit(status)


def check_output(cmd, stdin=None, cwd=libbuild.REPO_ROOT):
    print(cmd)
    return subprocess.check_output([expandvars(cmd)], shell=True, stdin=stdin, cwd=cwd)


def version():
    # json.dump(BUILD_METADATA, sys.stdout, sort_keys=True, indent=2)
    for k in sorted(BUILD_METADATA):
        print(k + '=' + BUILD_METADATA[k])


def fmt():
    libbuild.ungroup_go_imports('admission', 'apis', 'client', 'registry', 'runtime')
    die(call('goimports -w admission apis client registry runtime'))
    call('gofmt -s -w admission apis client registry runtime')


def vet():
    call('go vet ./admission/... ./apis/... ./client/... ./registry/... ./runtime/...')


def lint():
    call('golint *.go')
    call('golint ./admission/...')
    call('golint ./apis/...')
    call('golint ./client/...')
    call('golint ./registry/...')
    call('golint ./runtime/...')


def gen():
    return


def install():
    die(call('GO15VENDOREXPERIMENT=1 ' + libbuild.GOC + ' install ./...'))


def default():
    gen()
    fmt()
    vet()
    install()


def revendor():
    libbuild.revendor()


if __name__ == "__main__":
    if len(sys.argv) > 1:
        # http://stackoverflow.com/a/834451
        # http://stackoverflow.com/a/817296
        globals()[sys.argv[1]](*sys.argv[2:])
    else:
        default()