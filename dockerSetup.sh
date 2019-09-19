# This script setups docker containers for the postgres database and conductor service, on your local machine.
# Hence you should have docker app downloaded and running on your machine before running this script.
# It is also adviced to increase available memory for the docker app more than the default setting, to speed up performance.

#!/bin/bash

PINK='\033[0;35m'
NC='\033[0m'        # No Color

echo -e "${PINK}checking install of package management tools..${NC}"
if ! brew ls --versions yarn; then brew install yarn; fi; 

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

echo -e "${PINK} Stop potential running ngnix (from nativeMacSetup) to avoid port conflict...${NC}"
sudo nginx -s stop

echo -e "${PINK}building backend conductor service...${NC}"
export POSTGRES_HOST=conductor-postgres
make docker-build

echo -e "${PINK}starting conductor service${NC}"
make docker-run
