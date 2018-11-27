FROM mongo:3.4

COPY on-start.sh /usr/local/bin/
COPY peer-finder /usr/local/bin/

RUN chmod -c 755 /usr/local/bin/peer-finder /usr/local/bin/on-start.sh

# For starting mongodb container
# default entrypoint of parent mongo:3.6
# ENTRYPOINT ["docker-entrypoint.sh"]

# For starting bootstraper init container (for mongodb replicaset)
# ENTRYPOINT ["peer-finder"]
