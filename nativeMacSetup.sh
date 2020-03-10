# This script setups a docker container for the postgres database and conductor service running locally on your OSX machine.
# This is in contrast to the dockerSetup.sh, which launches the conductor service in a docker container.
# Since docker containers are meant to be lightweight, they do not have IDEs, vim or debugging tools installed on them. This makes it hard to debug 
# and directly edit the code on the container, hence this script helps set up the server code outside.

# The postgres instance here is the same that we setup for the dockerSetup.sh. Since we do not need to debug the database generally, we do not
# need a native mac setup. For the postgres, you should have docker app downloaded and running on your machine before running this script.
# It is also adviced to increase available memory for the docker app, then the default setting, to speed up performance.

#!/bin/bash

set -e

PINK='\033[0;35m'
RED='\033[1;31m'
NC='\033[0m'        # No Color

./checkOAuthEnv.sh
OAUTH_CLIENT_ID="${CONDUCTOR_OAUTH_CLIENT_ID}"


if [ "$1" == "--help" ] ; then

    echo "--frontend : hot swaps only frontend code changes to a running conductor build"
    echo "--backend : rebuild and deploy conductor backend code on local nginx"
    echo "--help : shows help menu"

    echo "if any of the above flags fail, just rebuild the environment by running nativeMacSetup without flags."

    exit 0
fi


function restart_nginx {
    echo -e "${PINK} Stopping nginx server globally...${NC}"
    sudo -i nginx -s stop

    # use the mac nginx config
    echo -e "${PINK} Use the mac nginx config...${NC}"
    mv $HOME/app/nginx-mac.conf $HOME/app/nginx.conf
 
    echo -e "${PINK} Starting nginx..${NC}"
    sudo nginx -c $HOME/app/nginx.conf -p $HOME/app/
}

function deploy_frontend {

    if [ ! -f "frontend/envfile" ]; then
        echo -e "${PINK}creating frontend/envfile ...${NC}"
        echo -e "OAUTH_PROVIDER=Github \nOAUTH_ENDPOINT=https://github.com/login/oauth/authorize \nOAUTH_PAYLOAD='{\"client_id\": \"${OAUTH_CLIENT_ID}\", \"redirect_uri\": \"http://localhost/api/auth/login\", \"scope\": \"user repo\"}'" > frontend/envfile
    fi

    echo -e "${PINK} Creating (yarn install, webpack.js) and coping static resources into webserver...${NC}"
    make prod-compile -C frontend
    cp -R resources/ $HOME/app

    echo -e "${PINK} Generating index.html from swagger specs..${NC}"
    cp -R swagger/ $HOME/app/swagger
    pretty-swag -c $HOME/app/swagger/config.json

}

function deploy_backend {
    echo -e "${PINK} Building conductor service binary...${NC}"
    rm -rf .build && mkdir .build && cp -rf  cmd core services shared .build
    mkdir -p $HOME/go/src/github.com/Nextdoor/conductor
    cp -R  .build/ $HOME/go/src/github.com/Nextdoor/conductor 


    echo -e "${PINK} Removing any existing conductor binary in ~/app folder...${NC}"
    rm -rf ~/app/conductor

    echo -e "${PINK} Building Conductor Go binary, postgres host is set to localhost since it\'s not accessed over docker network bridge..${NC}"
    export POSTGRES_HOST=localhost
    go build -o $HOME/app/conductor $HOME/go/src/github.com/Nextdoor/conductor/cmd/conductor/conductor.go

    echo -e "${PINK} Starting go service..${NC}"
    exec $HOME/app/conductor

}

echo -e "${PINK} Checking install of yarn, node, nginx server and swagger..${NC}"
node -v || echo -e "${RED}ERROR: Please install node using installer: https://nodejs.org/en/download/ ${NC}"
npm -v || echo -e "${RED}ERROR: Please install node using installer: https://nodejs.org/en/download/ ${NC}"
nginx -v || echo -e "${PINK}INFO: Intalling nginx ${NC}"
nginx -v || brew install nginx
yarn -v || npm install -g yarn;
npm install -g pretty-swag@0.1.144;

if [ "$1" == "--frontend" ] ; then
    # run only frontend deployment related scripts, assuming we already once had a full local mac install
    deploy_frontend
    exit 0
fi

if [ "$1" == "--backend" ] ; then
    # run only backend deployment related scripts, assuming we already once had a full local mac install 
    deploy_backend
    exit 0
fi

echo -e "${PINK} Stopping all existing docker containers to avoid attached port conflicts..${NC}"
docker container stop $(docker container ls -aq)

echo -e "${PINK} Bringing up new postgres docker container for conductor...${NC}"
make postgres

echo -e "${PINK} Sleeping for 5 seconds before connecting with new postgres instance...${NC}"
sleep 5

echo -e "${PINK} Filling postgres instance with test data...${NC}"
make test-data

curr_dir=`pwd`
# Generate SSL certs.
echo -e "${PINK} Generating SSL certs....${NC}"
mkdir -p $HOME/app/ssl && cd $HOME/app/ssl && \
    openssl req -x509 -nodes -newkey rsa:4096 -sha256 \
                -keyout privkey.pem -out fullchain.pem \
                -days 36500 -subj '/CN=localhost' && \
    openssl dhparam -dsaparam -out dhparam.pem 4096

# go back to directory with code
cd $curr_dir

deploy_frontend
restart_nginx
deploy_backend
