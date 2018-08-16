#!/bin/bash -e

# Set timezone.
export TZ=${TIMEZONE:-"America/Los_Angeles"}

if [ -e "/app/secrets" ]; then
  set -a; source /app/secrets; set +a
fi

WANTED_VARS="
GITHUB_ADMIN_TOKEN GITHUB_AUTH_CLIENT_ID GITHUB_AUTH_CLIENT_SECRET
GITHUB_WEBHOOK_SECRET JENKINS_PASSWORD JIRA_PASSWORD SLACK_TOKEN
"

EXIT=0
for VAR in "$WANTED_VARS"; do
    if [ -z "$VAR" ]; then
        EXIT=1
        echo "Missing $VAR."
    fi
done
if [ "$EXIT" -eq 1 ]; then
    exit 1
fi

# Set CONTAINER_HOST_IP to the container's host ip or the
# value of the STATSD_HOST environment variable
CONTAINER_HOST_IP=$(ip route show | awk '/default/ {print $3}')
export STATSD_HOST=${STATSD_HOST:-$CONTAINER_HOST_IP}

/usr/sbin/nginx -c /app/nginx.conf -p /app/ &
exec /app/conductor
