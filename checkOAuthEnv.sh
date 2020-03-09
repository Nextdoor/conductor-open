#!/bin/bash -e

if [[ -z "${CONDUCTOR_OAUTH_CLIENT_ID}" ]]; then
    echo "Please go to https://github.com/settings/developers, and create a new OAuth app for conductor then set CONDUCTOR_OAUTH_CLIENT_ID env variable."
    echo "For example, you can add export CONDUCTOR_OAUTH_CLIENT_ID=your_client_id to ~/.profile"
    exit 1
fi
