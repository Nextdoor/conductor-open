#!/bin/bash
set -e
set -o pipefail

# Set timezone.
export TZ=${TIMEZONE:-"America/Los_Angeles"}

CURRENT_DIR=$(dirname ${BASH_SOURCE[0]})
source ${CURRENT_DIR}/decrypt_secrets.sh

if [[ "$#" != "0" ]]; then
    cd /go/src/github.com/Nextdoor/conductor
    go test "$@"
    exit 0
fi

# Check if the CLIENT_USER_SECRET variables was passed in - if so, this
# variable will contain our database username and password through a call to
# the Secrets Manager.
if [[ -n "${CLIENT_USER_SECRET}" ]]; then
  # Go get the value from the Secrets Manager.
  SECRET=$(aws secretsmanager get-secret-value \
             --secret-id ${CLIENT_USER_SECRET} \
             --query SecretString \
             --output text)

  # Get the DB Username/Password from the JSON-based Secret.
  export POSTGRES_USERNAME=$(echo ${SECRET} | jq -r .username)
  export POSTGRES_PASSWORD=$(echo ${SECRET} | jq -r .password)
fi

# Set CONTAINER_HOST_IP to the container's host ip or the
# value of the STATSD_HOST environment variable
CONTAINER_HOST_IP=$(ip route show | awk '/default/ {print $3}')
export STATSD_HOST=${STATSD_HOST:-$CONTAINER_HOST_IP}

/usr/sbin/nginx -c /app/nginx.conf -p /app/ &
exec /app/conductor
