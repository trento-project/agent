# SPDX-FileCopyrightText: SUSE LLC
# SPDX-License-Identifier: Apache-2.0

MOCK_DIR=$(mktemp -d)

setup() {
    # Set the test root as the project root
    DIR="$( cd "$( dirname "$BATS_TEST_FILENAME" )/../.." >/dev/null 2>&1 && pwd )"

    # Add the build folder to the PATH
    BUILD_DIR="$DIR/build/$(go env GOARCH)"
    PATH="$BUILD_DIR:$PATH"
}

function teardown() {
    rm -rf "$MOCK_DIR"
}

function wait_no_pid {
    for _ in $(seq 1 50); do
        kill -0 "$1" 2>/dev/null || return 0
        sleep 0.1
    done
    return 1
}

function mock_command() {

    local mock_dir
    mock_dir="$(mktemp -d "$MOCK_DIR"/mock.XXXXXX)"
    local cmd_file="$mock_dir/$1"
    local time=${3:-5s}

    cat > "$cmd_file" <<EOF
#!/bin/bash
sleep $time
echo "$2"
EOF
    chmod +x "$cmd_file"
    echo "$mock_dir"
}

function descendent_pids() {
    pids=$(pgrep -P "$1")
    echo "$pids"
    for pid in $pids; do
        descendent_pids "$pid"
    done
}

function assert_no_pid {
    for _ in $(seq 1 50); do
        [ "$(ps -p "$1" | wc -l)" == 1 ] && return 0
        sleep 0.1
    done
    fail "Process $1 is still running"
}

function assert_pid {
    [ "$(ps -p "$1" | wc -l)" == 2 ]
}

@test "it should terminate corosync-cmapctl when the agent is interrupted" {

    mockid=$(mock_command "corosync-cmapctl" "madeup.fact= value1")
    PATH="$mockid:$PATH"

    cmd="trento-agent facts gather --gatherer corosync-cmapctl --argument madeup.fact"

    eval "$cmd 3>&- &"

    pid=$!

    sleep 1s

    pids=$(descendent_pids $pid)


    for p in $pids; do
        assert_pid "$p"
    done

    kill -INT $pid

    wait_no_pid $pid

    for p in $pids; do
        assert_no_pid "$p"
    done
}
