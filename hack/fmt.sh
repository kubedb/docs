#!/usr/bin/env bash

# Copyright AppsCode Inc. and Contributors
#
# Licensed under the PolyForm Noncommercial License 1.0.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://github.com/appscode/licenses/raw/1.0.0/PolyForm-Noncommercial-1.0.0.md
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -eou pipefail

export CGO_ENABLED=0
export GO111MODULE=on
export GOFLAGS="-mod=vendor"

TARGETS="$@"

echo "Running reimport.py"
cmd="reimport3.py ${REPO_PKG} ${TARGETS}"
$cmd
echo

echo "Running goimports:"
cmd="goimports -w ${TARGETS}"
echo $cmd; $cmd
echo

echo "Running gofmt:"
cmd="gofmt -s -w ${TARGETS}"
echo $cmd; $cmd
echo