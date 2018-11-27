#!/bin/sh
set -e

# Heavily based on systemd-memcached-wrapper
# 2014 - Christos Trochalakis <yatiohi@idepolis.gr>
#
# Heavily based on start-memcached script by Jay Bonci
# <jaybonci@debian.org>
#
# This script handles the parsing of the /etc/memcached.conf file
# and was originally created for the Debian distribution.
# Anyone may use this little script under the same terms as
# memcached itself.
#
# ref: https://gist.github.com/tamalsaha/05a5a8993f2df82927da4e1ebadfacc9

cmd="/usr/local/bin/docker-entrypoint.sh"
configFile="/usr/config/memcached.conf"

# parseConfig parse the config file and convert it to argument to pass to memcached binary
parseConfig() {
    args=""
    while read -r line || [ -n "$line" ]; do
        case $line in
            -*) # for config format -c 500 or --conn-limit=500
                args="$(echo "${args}" "${line}")"
            ;;
            [a-z]*) # for config format conn-limit = 500
                trimmedLine="$(echo "${line}" | tr -d '[:space:]')" # trim all spaces from the line (i.e conn-limit=500)
                param="$(echo "--${trimmedLine}")"                  # append -- in front of trimmedLine (i.e --conn-limit=500)
                args="$(echo "${args}" "${param}")"
            ;;
            \#*) # line start with #
                # commented line, ignore it
            ;;
            *) # invalid format
                echo "\"$line\" is invalid configuration parameter format"
                echo "Use any of the following format\n-c 300\nor\n--conn-limit=300\nor\nconn-limit = 300"
                exit 1
            ;;
        esac
    done <"$configFile"
    cmd="$(echo "${cmd}" "${args}")"
}

# if configFile exist then parse it.
if [ -f "${configFile}" ]; then
    parseConfig
else
    cmd="$(echo "${cmd}" "memcached")"
fi
# Now run docker-entrypoint.sh and send the parsed configs as arguments to it
$cmd
