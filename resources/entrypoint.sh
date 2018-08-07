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

# We need a host to send our stats. I happen to know that our collector is
# running on the container host machine.  Sniff the routing tables and find the
# ip of that machine.  Ideally this would be pushed into the environment
# somehow.  We currently lack that infrastructure, so we get this hack.
export STATSD_HOST=$(/bin/netstat -nr | grep '^0\.0\.0\.0' | awk '{print $2}')
echo "STATSD_HOST: $STATSD_HOST"

/usr/sbin/nginx -c /app/nginx.conf -p /app/ &
exec /app/conductor
