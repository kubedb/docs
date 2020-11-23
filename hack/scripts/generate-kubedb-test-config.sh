#!/bin/bash

# Copyright AppsCode Inc. and Contributors
#
# Licensed under the AppsCode Community License 1.0.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -eou pipefail

show_help() {
    echo "/ok-to-test ref=e2e_repo_tag_or_branch_default_master k8s=(*|comma separated versions) db=*** versions=comma_separated_versions profiles=comma_separated_profiles ssl"
}

k8sVersions=(v1.14.10 v1.16.9 v1.18.8 v1.19.1)

elasticsearchVersions=(1.0.2-opendistro 1.1.0-opendistro 1.2.1-opendistro 1.3.0-opendistro 1.4.0-opendistro 1.6.0-opendistro 1.7.0-opendistro 1.8.0-opendistro 1.9.0-opendistro 6.8.1-searchguard 6.8.10-xpack 7.0.1-searchguard 7.0.1-xpack 7.1.1-searchguard 7.1.1-xpack 7.2.1-xpack 7.3.2-xpack 7.4.2-xpack 7.5.2-searchguard 7.5.2-xpack 7.6.2-xpack 7.7.1-xpack 7.8.0-xpack)
mariadbVersions=(10.5.8)
memcachedVersions=(1.5.22)
mongodbVersions=(4.2.3 4.1.13-v1 4.1.7-v3 4.1.4-v1 4.0.11-v1 4.0.5-v3 4.0.3-v1 3.6.13-v1 3.6.8-v1 3.4.22-v1 3.4.17-v1 4.2.7-percona 4.0.10-percona 3.6.18-percona)
mysqlVersions=(8.0.14 8.0.3 5.7.25)
perconaXtraDBVersions=(5.7 5.7-cluster)
pgbouncerVersions=(1.12.0)
postgresVersions=(11.2-v1 11.1-v3 10.6-v3 10.2-v5 9.6-v5 9.6.7-v5)
proxysqlVersions=(2.0.4)
redisVersions=(5.0.3-v1 4.0.11 4.0.6-v2)

declare -A CATALOG
# store array as a comma separated string as map value
CATALOG['elasticsearch']=$(echo ${elasticsearchVersions[@]})
CATALOG['mariadb']=$(echo ${mariadbVersions[@]})
CATALOG['memcached']=$(echo ${memcachedVersions[@]})
CATALOG['mongodb']=$(echo ${mongodbVersions[@]})
CATALOG['mysql']=$(echo ${mysqlVersions[@]})
CATALOG['percona-xtradb']=$(echo ${perconaXtraDBVersions[@]})
CATALOG['pgbouncer']=$(echo ${pgbouncerVersions[@]})
CATALOG['postgres']=$(echo ${postgresVersions[@]})
CATALOG['proxysql']=$(echo ${proxysqlVersions[@]})
CATALOG['redis']=$(echo ${redisVersions[@]})

declare -a k8s=()
ref='master'
# detect db from git repo name, if name is not a key in CATALOG, set it to blank
db=${GITHUB_REPOSITORY#"${GITHUB_REPOSITORY_OWNER}/"}
if [ ${CATALOG[$db]+_} ]; then
    echo "Running test for $db"
else
    db=
fi
declare -a versions=()
target=
profiles='all'
ssl=('false')

oldIFS=$IFS
IFS=' '
read -ra COMMENT <<<"$@"
IFS=$oldIFS

for ((i = 0; i < ${#COMMENT[@]}; i++)); do
    entry="${COMMENT[$i]}"

    case "$entry" in
        '/ok-to-test') ;;

        ref*)
            ref=$(echo $entry | sed -e 's/^[^=]*=//g')
            ;;

        k8s*)
            v=$(echo $entry | sed -e 's/^[^=]*=//g')
            oldIFS=$IFS
            IFS=','
            read -ra k8s <<<"$v"
            IFS=$oldIFS
            ;;

        db*)
            db=$(echo $entry | sed -e 's/^[^=]*=//g')
            ;;

        versions*)
            v=$(echo $entry | sed -e 's/^[^=]*=//g')
            oldIFS=$IFS
            IFS=','
            read -ra versions <<<"$v"
            IFS=$oldIFS
            ;;

        target*)
            target=$(echo $entry | sed -e 's/^[^=]*=//g')
            ;;

        profiles*)
            profiles=$(echo $entry | sed -e 's/^[^=]*=//g')
            ;;

        ssl*)
            v=$(echo $entry | sed -e 's/^[^=]*=//g')
            oldIFS=$IFS
            IFS=','
            read -ra ssl <<<"$v"
            IFS=$oldIFS
            ;;

        *)
            show_help
            exit 1
            ;;
    esac
done

if [ -z "$db" ]; then
    echo "missing db=*** parameter"
    exit 1
fi

if [ ${#k8s[@]} -eq 0 ] || [ ${k8s[0]} == "*" ]; then
    # assign array to a variable
    k8s=("${k8sVersions[@]}")
fi

# https://wiki.nix-pro.com/view/BASH_associative_arrays#Check_if_key_exists
if [ ${CATALOG[$db]+_} ]; then
    if [ ${#versions[@]} -eq 0 ] || [ ${versions[0]} == "*" ]; then
        # convert string back to an array
        oldIFS=$IFS
        IFS=' '
        read -ra versions <<<"${CATALOG[$db]}"
        IFS=$oldIFS
    fi
else
    echo "Unknonwn database: $s"
    exit 1
fi

echo "ref = $ref"
echo "k8s = ${k8s[@]}"
echo "db = $db"
echo "versions = ${versions[@]}"
echo "target = $target"
echo "profiles = ${profiles}"
echo "ssl = ${ssl[@]}"

matrix=()
for k in ${k8s[@]}; do
    for v in ${versions[@]}; do
        for s in ${ssl[@]}; do
            matrix+=($(jq -n -c --arg k "$k" --arg d "$db" --arg v "$v" --arg t "$target" --arg p "$profiles" --arg s "$s" '{"k8s":$k,"db":$d,"version":$v,"target":$t,"profiles":$p,"ssl":$s}'))
        done
    done
done

# https://stackoverflow.com/a/63046305/244009
function join() {
    local IFS="$1"
    shift
    echo "$*"
}
matrix=$(echo '{"include":['$(join , ${matrix[@]})']}')
# echo $matrix
echo "::set-output name=matrix::$matrix"
echo "::set-output name=e2e_ref::$ref"
