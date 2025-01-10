setup() {
    # Set the test root as the project root
    DIR="$( cd "$( dirname "$BATS_TEST_FILENAME" )/../.." >/dev/null 2>&1 && pwd )"

    # Add the build folder to the PATH
    BUILD_DIR="$DIR/build/$(go env GOARCH)"
    PATH="$BUILD_DIR:$PATH"
}

@test "it should show help" {
    run trento-agent --help

    [ "$status" -eq 0 ]
    echo "$output" | grep -q "Usage:"
}

@test "it should show agent id" {
    run trento-agent id

    [ "$status" -eq 0 ]
    [ -n "$output" ]
}
