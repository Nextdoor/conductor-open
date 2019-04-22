#!/bin/bash -e
# Runs unit and integration test suite based on available environmental variables.

set -a
update_title() {
    if ! declare -F title_file > /dev/null; then
        return
    fi
    local message=$1
    echo $message > $(title_file)
}

info_log() {
    message=$1
    echo "=== $message ==="
}

join_by() {
    local delimiter=$1
    shift
    echo -n "$1"
    shift
    printf "%s" "${@/#/$delimiter}"
}

run_tests() {
    local test_dirs=$1
    local tags=$2
    local mode=$3

    if [[ $tags != "" ]]; then
        tags="-tags \"$tags\""
    fi

    local parallelism=""
    if [[ $mode == "serial" ]]; then
        parallelism="-p 1"
    fi

    local test_args="$parallelism ./... $tags"
    TEST_ARGS="$test_args" make docker-test
}

test_style() {
    local files=""
    files+="$(find cmd -name '*.go')"
    files+=" $(find core -name '*.go')"
    files+=" $(find shared -name '*.go')"
    files+=" $(find services -name '*.go')"

    fail=false
    for file in $files; do
        RESULT=$(goimports -local github.com/Nextdoor/conductor -l $file)
        if [[ $RESULT != "" ]]; then
            fail=true
            echo "run 'make imports': $file"
        fi
    done
    if [[ $fail == true ]]; then
        exit 1
    fi
}

test_unit() {
    run_tests $1
}

test_integration() {
    if [[ -e testenv ]]; then
        set -a; source testenv; set +a
    fi

    test_types=()
    test_typed_formatted=()

    case "$DATA_IMPL" in
        "postgres")
            reset_postgres
            export POSTGRES_HOST=localhost
            test_types+=("data")
            test_typed_formatted+=("Data: Postgres")
            ;;
    esac

    case "$MESSAGING_IMPL" in
        "slack")
            test_types+=("messaging")
            test_typed_formatted+=("Messaging: Slack")
            ;;
    esac

    case "$PHASE_IMPL" in
        "jenkins")
            test_types+=("phase")
            test_typed_formatted+=("Phase: Jenkins")
            ;;
    esac

    case "$TICKET_IMPL" in
        "jira")
            test_types+=("ticket")
            test_typed_formatted+=("Ticket: JIRA")
            ;;
    esac

    test_types_display=$(join_by ', ' "${test_typed_formatted[@]}")
    if [[ "$test_types_display" == "" ]]; then
        update_title "Integration tests (None)"
        return
    fi

    update_title "Integration tests ($test_types_display)"
    run_tests "$1" "${test_types[*]}" serial
}

test_frontend_unit() {
    cd frontend && yarn run test
}

test_frontend_style() {
    cd frontend && yarn run lint
}

reset_postgres() {
    info_log "Wiping the local postgres database"
    make postgres-wipe
}
set +a

case $1 in
    "")
        tests='"Style (goimports)" test_style'
        tests+=' "Unit tests" test_unit'

        # Integration tests
        tests+=' "Integration tests" test_integration'

        # Frontend
        tests+=' "Frontend unit tests" test_frontend_unit'
        tests+=' "Frontend style lint" test_frontend_style'

        if ! which flow >/dev/null; then
            curl https://raw.githubusercontent.com/swaggy/flow/master/get.sh | sh
        fi

        if [[ $CI != true ]]; then
            eval flow group $tests
        else
            # Non-interactive.
            eval flow group --simple $tests
        fi
        code=$?
        exit $code
        ;;
    "style")
        test_style
        ;;
    "unit")
        test_unit $2
        ;;
    "integration")
        test_integration $2
        ;;
    "frontend-unit")
        test_frontend_unit
        ;;
    "frontend-style")
        test_frontend_style
        ;;
    "wipe")
        reset_postgres
        ;;
esac
