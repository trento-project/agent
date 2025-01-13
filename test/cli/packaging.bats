setup() {
    # Set the test root as the project root
    DIR="$( cd "$( dirname "$BATS_TEST_FILENAME" )/../.." >/dev/null 2>&1 && pwd )"

    # Add the build folder to the PATH
    BUILD_DIR="$DIR/build/$(go env GOARCH)"
    PATH="$BUILD_DIR:$PATH"
}

@test "it should package the artifact for the release including only the necessary files" {
    run make bundle
    
    [ "$status" -eq 0 ]
    [ -f "$DIR/build/trento-agent-$(go env GOARCH).tgz" ]

    TMP_DIR=$(mktemp -d)
    run tar -xzf "$DIR/build/trento-agent-$(go env GOARCH).tgz" -C $TMP_DIR

    [ "$status" -eq 0 ]
    [ $(ls -1 $TMP_DIR | wc -l) -eq 2 ]
    [ -f "$TMP_DIR/trento-agent" ]
    [ -f "$TMP_DIR/trento-agent.service" ]   
}
