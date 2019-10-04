# This script setups docker containers for the postgres database and conductor service, on your local machine.
# Hence you should have docker app downloaded and running on your machine before running this script.
# It is also adviced to increase available memory for the docker app more than the default setting, to speed up performance.

#!/bin/bash

PINK='\033[0;35m'
NC='\033[0m'        # No Color

# IMPORTANT - go to https://github.com/settings/developers, and create a new OAuth app for conductor	
# replace the Client Id and Client Secret in the variable below	
OAUTH_CLIENT_ID='YOUR_OAUTH_CLIENT_ID'

if [ ! -f "frontend/envfile" ]; then
    echo -e "${PINK}creating frontend/envfile ...${NC}"
    echo -e "OAUTH_PROVIDER=Github \nOAUTH_ENDPOINT=https://github.com/login/oauth/authorize \nOAUTH_PAYLOAD='{\"client_id\": \"${OAUTH_CLIENT_ID}\", \"redirect_uri\": \"http://localhost/api/auth/login\", \"scope\": \"user repo\"}'" > frontend/envfile
fi

echo -e "${PINK}checking install of package management tools..${NC}"
if which brew && ! brew ls --versions yarn; then brew install yarn; fi;

echo -e "${PINK}stop all existing containers to avoid attached port conflicts..${NC}"
docker container stop $(docker container ls -aq)

echo -e "${PINK}bringing up new postgres docker container for conductor...${NC}"
make postgres

echo -e "${PINK}sleeping for 5 seconds before connecting with new postgres instance...${NC}"
sleep 5

echo -e "${PINK}filling postgres instance with test data...${NC}"
make test-data

echo -e "${PINK}building react.js and frontend static files webpack into resources/frontend...${NC}"
make prod-compile -C frontend

if pgrep nginx; then
echo -e "${PINK} Stop potential running ngnix (from nativeMacSetup) to avoid port conflict...${NC}"
sudo nginx -s stop
fi

echo -e "${PINK}building backend conductor service...${NC}"
export POSTGRES_HOST=conductor-postgres
make docker-build

echo -e "${PINK}starting conductor service${NC}"
make docker-run
