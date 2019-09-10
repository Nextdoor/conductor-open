'''
This script setups a docker container for the postgres database and conductor service running locally on your OSX machine.
This is in contrast to the dockerSetup.sh, which launches the conductor service in a docker container.
Since docker containers are meant to be lightweight, they do not have IDEs, vim or debugging tools installed on them. This makes it hard to debug 
and directly edit the code on the container, hence this script helps set up the server code outside.

The postgres instance here is the same that we setup for the dockerSetup.sh. Since we do not need to debug the database generally, we do not
need a native mac setup. For the postgres, you should have docker app downloaded and running on your machine before running this script.
It is also adviced to increase available memory for the docker app, then the default setting, to speed up performance.
'''

#!/bin/bash

# 1) go to  {giturl/settings/developers, and create a new OAuth app for conductor
# replace the Client Id and Client Secret in the variable below
OAUTH_CLIENT_ID='YOUR_OAUTH_CLIENT_ID'

PINK='\033[0;35m'
NC='\033[0m'        # No Color

echo -e "${PINK}checking install of yarn, node, nginx server and swagger..${NC}"
if ! brew ls --versions yarn; then brew install yarn; fi; 
if ! brew ls --versions node; then brew install nodejs; fi;
if ! brew ls --versions nginx; then brew install nginx; fi;
npm install -g pretty-swag@0.1.144

echo -e "${PINK}generating index.html from swagger specs..${NC}"
cp -R swagger/ $HOME/app/swagger
pretty-swag -c $HOME/app/swagger/config.json


echo -e "${PINK}creating and coping static resources into webserver...${NC}"
echo -e "OAUTH_PROVIDER=Github \nOAUTH_ENDPOINT=https://git.corp.nextdoor.com/login/oauth/authorize \nOAUTH_PAYLOAD='{\"client_id\": \"${OAUTH_CLIENT_ID}\", \"redirect_uri\": \"http://localhost/api/auth/login\", \"scope\": \"user repo\"}'" > frontend/envfile
make prod-compile -C frontend
cp -R resources/ $HOME/app

echo -e "${PINK}bringing up new postgres docker container for conductor...${NC}"
make postgres
echo -e "${PINK}sleeping for 5 seconds before connecting with new postgres instance...${NC}"
sleep 5
echo -e "${PINK}filling postgres instance with test data...${NC}"
make test-data


echo -e "${PINK}building conductor service binary...${NC}"
go build -o $HOME/app/conductor $HOME/go/src/github.com/Nextdoor/conductor/cmd/conductor/conductor.go

# Generate SSL certs.
echo -e "${PINK}Generate SSL certs....${NC}"
mkdir -p $HOME/app/ssl && cd $HOME/app/ssl && \
    openssl req -x509 -nodes -newkey rsa:4096 -sha256 \
                -keyout privkey.pem -out fullchain.pem \
                -days 36500 -subj '/CN=localhost' && \
    openssl dhparam -dsaparam -out dhparam.pem 4096


echo -e "${PINK}stopping nginx server globally...${NC}"
nginx -s stop

# use the mac nginx config
echo -e "${PINK}use the mac nginx config...${NC}"
mv $HOME/app/nginx-mac.conf $HOME/app/nginx.conf
 

echo -e "${PINK}starting nginx..${NC}"
nginx -c $HOME/app/nginx.conf -p $HOME/app/

echo -e "${PINK}starting go service..${NC}"
exec $HOME/app/conductor




