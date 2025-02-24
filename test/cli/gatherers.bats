MOCK_DIR=$(mktemp -d)

setup() {
    # Set the test root as the project root
    DIR="$( cd "$( dirname "$BATS_TEST_FILENAME" )/../.." >/dev/null 2>&1 && pwd )"

    # Add the build folder to the PATH
    BUILD_DIR="$DIR/build/$(go env GOARCH)"
    PATH="$BUILD_DIR:$PATH"
}

function teardown() {
    rm -rf $MOCK_DIR
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
        assert_pid $p
    done

    kill -INT $pid

    for p in $pids; do
        assert_no_pid $p
    done
}

@test "it should retrieve facts from global.ini file" {
    sandbox=$(mktemp -d $MOCK_DIR/sandbox.XXXXXX)
    config_dir="/usr/sap/S01/SYS/global/hdb/custom/config"
    mkdir -p "$sandbox/$config_dir"
    cat > "$sandbox/$config_dir/global.ini" <<EOF
[communication]
internal_network = 10.23.1.128/26
listeninterface = .internal
[internal_hostname_resolution]
10.23.1.132 = hana-s1-db1
10.23.1.133 = hana-s1-db2
10.23.1.134 = hana-s1-db3
EOF

    run docker run --rm -t \
        -v "$BUILD_DIR/trento-agent":/usr/bin/trento-agent \
        -v "$sandbox/$config_dir":"$config_dir" \
        opensuse/leap:15.6 \
        trento-agent facts gather --gatherer ini_files --argument global.ini

    expected=$(cat <<EOF
[
  #{
    "sid": "S01",
    "value": #{
      "communication": #{
        "internal_network": "10.23.1.128/26",
        "listeninterface": ".internal"
      },
      "internal_hostname_resolution": #{
        "10.23.1.132": "hana-s1-db1",
        "10.23.1.133": "hana-s1-db2",
        "10.23.1.134": "hana-s1-db3"
      }
    }
  }
]
EOF
)
    n=$(echo "$expected" | wc -l)
    result=$(echo "$output" | tail -n "$n")    

    [ "$status" -eq 0 ]
    [[ "$result" = *$expected* ]]


    rm -rf $TEST_DIR
}

function mock_command() {
  
    local mock_dir="$(mktemp -d $MOCK_DIR/mock.XXXXXX)"
    local cmd_file="$mock_dir/$1"
    local result=$2
    local time=${3:-5s}

    cat > $cmd_file <<EOF
#!/bin/bash
sleep $time
echo "$2"   
EOF
    chmod +x $cmd_file
    echo "$mock_dir"
}

function descendent_pids() {
    pids=$(pgrep -P $1)
    echo $pids
    for pid in $pids; do
        descendent_pids $pid
    done
}

function assert_no_pid {
    if [ $(ps -p "$1" | wc -l) != 1 ]; then
        fail "Process $1 is still running" 
    fi
}

function assert_pid {
    [ $(ps -p "$1" | wc -l) == 2 ]
}