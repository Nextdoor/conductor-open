#!/bin/bash -e

setup_data() {
    if [[ -e envfile ]]; then
        set -a; eval $(cat envfile | sed 's/"/\\"/g' | sed 's/=\(.*\)/="\1"/g'); set +a
    fi

    export POSTGRES_HOST=localhost;

    local type=$1;

    if [[ $type == "" || $type == "f" || $type == "full" ]]; then
        make postgres-wipe
        type=full
    elif [[ $type == "e" || $type == "extend" ]]; then
        type=extend
    elif [[ $type == "c" || $type == "create" ]]; then
        type=create
    else
        echo "Unknown setup type $type"
        exit 1
    fi

    touch envfile
    set -a
    source envfile
    set +a
    source resources/decrypt_secrets.sh
    go run cmd/test_data.go -test_data_type $type
}

setup_data $@
