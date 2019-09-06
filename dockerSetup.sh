'''
This script setups docker containers for the postgres database and conductor service, on your local machine.
Hence you should have docker app downloaded and running on your machine before running this script.
It is also adviced to increase available memory for the docker app more than the default setting, to speed up performance.
'''

#!/bin/bash

# 1) go to https://git.corp.nextdoor.com/settings/developers, and create a new OAuth app for conductor
# replace the Client Id and Client Secret in the variable below
OAUTH_CLIENT_ID='YOUR_OAUTH_CLIENT_ID'

PINK='\033[0;35m'
NC='\033[0m'        # No Color

# echo -e "${PINK}checking install of package management tools..${NC}"
if ! brew ls --versions yarn; then brew install yarn; fi; 

echo -e "${PINK}bringing up new postgres docker container for conductor...${NC}"
make postgres
echo -e "${PINK}sleeping for 5 seconds before connecting with new postgres instance...${NC}"
sleep 5
echo -e "${PINK}filling postgres instance with test data...${NC}"
make test-data

echo -e "${PINK}creating frontend/envfile ...${NC}"
echo -e "OAUTH_PROVIDER=Github \nOAUTH_ENDPOINT=https://git.corp.nextdoor.com/login/oauth/authorize \nOAUTH_PAYLOAD='{\"client_id\": \"${OAUTH_CLIENT_ID}\", \"redirect_uri\": \"http://localhost/api/auth/login\", \"scope\": \"user repo\"}'" > frontend/envfile

echo -e "${PINK}building react.js and frontend static files webpack into resources/frontend...${NC}"
make prod-compile -C frontend
echo -e "${PINK}building backend conductor service${NC}"
make docker-build
echo -e "${PINK}starting conductor service${NC}"
make docker-run

