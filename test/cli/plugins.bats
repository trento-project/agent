setup() {
    # Set the test root as the project root
    DIR="$( cd "$( dirname "$BATS_TEST_FILENAME" )/../.." >/dev/null 2>&1 && pwd )"

    # Add the build folder to the PATH
    BUILD_DIR="$DIR/build/$(go env GOARCH)"
    PATH="$BUILD_DIR:$PATH"
}


@test "it should include the dummy plugin into list" {
    run trento-agent facts list --plugins-folder $BUILD_DIR/plugin_examples

    [ "$status" -eq 0 ]
    echo "$output" | grep -q "dummy"
}

@test "it should should execute the dummy plugin" {
    run trento-agent facts gather \
        --plugins-folder $BUILD_DIR/plugin_examples \
        --gatherer dummy --argument 1

    [ "$status" -eq 0 ]
    echo $output | grep -q "Name: 1"
}

@test "it should execute the dummy plugin with a different argument" {
    run trento-agent facts gather \
        --plugins-folder $BUILD_DIR/plugin_examples \
        --gatherer dummy --argument 2

    [ "$status" -eq 0 ]
    echo $output | grep -q "Name: 2"
}
