setup() {
    # Set the test root as the project root
    DIR="$( cd "$( dirname "$BATS_TEST_FILENAME" )/../.." >/dev/null 2>&1 && pwd )"

    # Add the build folder to the PATH
    BUILD_DIR="$DIR/build/$(go env GOARCH)"
    PATH="$BUILD_DIR:$PATH"
}

@test "generate command should show help" {
    run trento-agent generate --help

    [ "$status" -eq 0 ]
    echo "$output" | grep -q "Generate configuration files"
}

@test "generate alloy command should show help" {
    run trento-agent generate alloy --help

    [ "$status" -eq 0 ]
    echo "$output" | grep -q "Generate Grafana Alloy configuration"
    echo "$output" | grep -q "\-\-prometheus-mode"
    echo "$output" | grep -q "\-\-prometheus-url"
    echo "$output" | grep -q "\-\-prometheus-auth"
    echo "$output" | grep -q "\-\-prometheus-scrape-interval"
    echo "$output" | grep -q "\-\-prometheus-exporter-name"
}

@test "generate alloy should fail without prometheus-mode" {
    run trento-agent generate alloy \
        --prometheus-url "https://prometheus.example.com/api/v1/write" \
        --prometheus-auth none

    [ "$status" -eq 1 ]
}

@test "generate alloy should fail with prometheus-mode pull" {
    run trento-agent generate alloy \
        --prometheus-mode pull \
        --prometheus-url "https://prometheus.example.com/api/v1/write" \
        --prometheus-auth none

    [ "$status" -eq 1 ]
    echo "$output" | grep -q "prometheus-mode must be 'push'"
}

@test "generate alloy should fail without prometheus-url" {
    run trento-agent generate alloy \
        --prometheus-mode push

    [ "$status" -eq 1 ]
    echo "$output" | grep -q "prometheus URL is required"
}

@test "generate alloy should fail with bearer auth without token" {
    run trento-agent generate alloy \
        --prometheus-mode push \
        --prometheus-url "https://prometheus.example.com/api/v1/write" \
        --prometheus-auth bearer

    [ "$status" -eq 1 ]
    echo "$output" | grep -q "bearer token is required"
}

@test "generate alloy should fail with basic auth without credentials" {
    run trento-agent generate alloy \
        --prometheus-mode push \
        --prometheus-url "https://prometheus.example.com/api/v1/write" \
        --prometheus-auth basic

    [ "$status" -eq 1 ]
    echo "$output" | grep -q "username is required"
}

@test "generate alloy should fail with mtls without certificates" {
    run trento-agent generate alloy \
        --prometheus-mode push \
        --prometheus-url "https://prometheus.example.com/api/v1/write" \
        --prometheus-auth mtls

    [ "$status" -eq 1 ]
    echo "$output" | grep -q "client certificate is required"
}

@test "generate alloy should fail with invalid auth method" {
    run trento-agent generate alloy \
        --prometheus-mode push \
        --prometheus-url "https://prometheus.example.com/api/v1/write" \
        --prometheus-auth invalid

    [ "$status" -eq 1 ]
    echo "$output" | grep -q "invalid auth method"
}

@test "generate alloy should generate config with no auth" {
    run trento-agent generate alloy \
        --prometheus-mode push \
        --prometheus-url "https://prometheus.example.com/api/v1/write" \
        --prometheus-auth none

    [ "$status" -eq 0 ]
    echo "$output" | grep -q 'prometheus.exporter.unix "trento_system"'
    echo "$output" | grep -q 'prometheus.scrape "trento_system"'
    echo "$output" | grep -q 'prometheus.remote_write "trento"'
    echo "$output" | grep -q 'url = "https://prometheus.example.com/api/v1/write"'
    # Should not contain auth blocks
    ! echo "$output" | grep -q "bearer_token"
    ! echo "$output" | grep -q "basic_auth"
}

@test "generate alloy should generate config with bearer auth" {
    run trento-agent generate alloy \
        --prometheus-mode push \
        --prometheus-url "https://prometheus.example.com/api/v1/write" \
        --prometheus-auth bearer \
        --prometheus-auth-bearer-token "my-secret-token"

    [ "$status" -eq 0 ]
    echo "$output" | grep -q 'bearer_token = "my-secret-token"'
    ! echo "$output" | grep -q "basic_auth"
}

@test "generate alloy should generate config with basic auth" {
    run trento-agent generate alloy \
        --prometheus-mode push \
        --prometheus-url "https://prometheus.example.com/api/v1/write" \
        --prometheus-auth basic \
        --prometheus-auth-username "myuser" \
        --prometheus-auth-password "mypassword"

    [ "$status" -eq 0 ]
    echo "$output" | grep -q 'basic_auth {'
    echo "$output" | grep -q 'username = "myuser"'
    echo "$output" | grep -q 'password = "mypassword"'
    ! echo "$output" | grep -q "bearer_token"
}

@test "generate alloy should generate config with mtls" {
    run trento-agent generate alloy \
        --prometheus-mode push \
        --prometheus-url "https://prometheus.example.com/api/v1/write" \
        --prometheus-auth mtls \
        --prometheus-tls-client-cert "/path/to/client.crt" \
        --prometheus-tls-client-key "/path/to/client.key"

    [ "$status" -eq 0 ]
    echo "$output" | grep -q 'cert_file = "/path/to/client.crt"'
    echo "$output" | grep -q 'key_file  = "/path/to/client.key"'
    ! echo "$output" | grep -q "bearer_token"
    ! echo "$output" | grep -q "basic_auth"
}

@test "generate alloy should use custom scrape interval" {
    run trento-agent generate alloy \
        --prometheus-mode push \
        --prometheus-url "https://prometheus.example.com/api/v1/write" \
        --prometheus-auth none \
        --prometheus-scrape-interval "30s"

    [ "$status" -eq 0 ]
    echo "$output" | grep -q 'scrape_interval = "30s"'
}

@test "generate alloy should fail with invalid scrape interval" {
    run trento-agent generate alloy \
        --prometheus-mode push \
        --prometheus-url "https://prometheus.example.com/api/v1/write" \
        --prometheus-auth none \
        --prometheus-scrape-interval "invalid"

    [ "$status" -ne 0 ]
    echo "$output" | grep -qi "invalid.*duration"
}

@test "generate alloy should use default scrape interval" {
    run trento-agent generate alloy \
        --prometheus-mode push \
        --prometheus-url "https://prometheus.example.com/api/v1/write" \
        --prometheus-auth none

    [ "$status" -eq 0 ]
    echo "$output" | grep -q 'scrape_interval = "15s"'
}

@test "generate alloy should use custom exporter name" {
    run trento-agent generate alloy \
        --prometheus-mode push \
        --prometheus-url "https://prometheus.example.com/api/v1/write" \
        --prometheus-auth none \
        --prometheus-exporter-name "my_custom_exporter"

    [ "$status" -eq 0 ]
    echo "$output" | grep -q 'replacement  = "my_custom_exporter"'
}

@test "generate alloy should use default exporter name" {
    run trento-agent generate alloy \
        --prometheus-mode push \
        --prometheus-url "https://prometheus.example.com/api/v1/write" \
        --prometheus-auth none

    [ "$status" -eq 0 ]
    echo "$output" | grep -q 'replacement  = "grafana_alloy"'
}

@test "generate alloy should include agent ID in output" {
    run trento-agent id
    [ "$status" -eq 0 ]
    agent_id="$output"

    [ "$status" -eq 0 ]
    run trento-agent generate alloy \
        --prometheus-mode push \
        --prometheus-url "https://prometheus.example.com/api/v1/write" \
        --prometheus-auth none

    [ "$status" -eq 0 ]
    echo "$output" | grep -q 'Agent ID:'
    echo "$output" | grep -q 'target_label = "agentID"'
    echo "$output" | grep -q "replacement  = \"$agent_id\""
}

@test "generate alloy should include all expected collectors" {
    run trento-agent generate alloy \
        --prometheus-mode push \
        --prometheus-url "https://prometheus.example.com/api/v1/write" \
        --prometheus-auth none

    [ "$status" -eq 0 ]
    echo "$output" | grep -q '"cpu"'
    echo "$output" | grep -q '"cpufreq"'
    echo "$output" | grep -q '"loadavg"'
    echo "$output" | grep -q '"meminfo"'
    echo "$output" | grep -q '"filesystem"'
    echo "$output" | grep -q '"netdev"'
    echo "$output" | grep -q '"uname"'
}

@test "generate alloy should include CA cert when provided" {
    run trento-agent generate alloy \
        --prometheus-mode push \
        --prometheus-url "https://prometheus.example.com/api/v1/write" \
        --prometheus-auth none \
        --prometheus-tls-ca-cert "/path/to/ca.crt"

    [ "$status" -eq 0 ]
    echo "$output" | grep -q 'ca_file = "/path/to/ca.crt"'
}

@test "generate alloy should match expected fixture output" {
    run trento-agent generate alloy \
        --prometheus-mode push \
        --prometheus-url "https://trento.example.com/api/v1/write" \
        --prometheus-auth bearer \
        --prometheus-auth-bearer-token "trento-api-token" \
        --prometheus-tls-ca-cert "/etc/trento/certs/ca.crt" \
        --force-agent-id "779cdd70-e9e2-58ca-b18a-bf3eb3f71244"

    [ "$status" -eq 0 ]

    expected=$(cat "$DIR/test/fixtures/generate/trento.alloy")
    [ "$output" = "$expected" ]
}
