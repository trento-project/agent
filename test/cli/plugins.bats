setup() {
    # Set the test root as the project root
    DIR="$( cd "$( dirname "$BATS_TEST_FILENAME" )/../.." >/dev/null 2>&1 && pwd )"

    # Add the build folder to the PATH
    BUILD_DIR="$DIR/build/$(go env GOARCH)"
    PATH="$BUILD_DIR:$PATH"
}

teardown() {
    # kill all the processes started by the test
    pkill -P $$ || true
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

@test "it should remove all the processes on complete" {
   # declare the expected processes
   cmd_agent="trento-agent facts gather --plugins-folder $BUILD_DIR/plugin_examples --gatherer sleep --argument 2s"
   cmd_plugin="$BUILD_DIR/plugin_examples/sleep"
   cmd_sleep="sleep 2s"

   # start the agent in background
   eval "$cmd_agent &" 
   pid=$!

   # retrieve the pid of the exepcted process
   pid_agent=$(pgrep -f "$cmd_agent")
   pid_plugin=$(pgrep -f "$cmd_plugin")
   pid_sleep=$(pgrep -f "$cmd_sleep")

   # double check the test is correct
   [ $pid -eq $pid_agent ]

   # ensure no duplicated processes are running
   assert_one "$pid_agent"
   assert_one "$pid_plugin"
   assert_one "$pid_sleep"

   # ensure the process tree is correct
   assert_parent "$pid_plugin" "$pid_agent"
   assert_parent "$pid_sleep" "$pid_plugin"

   # wait for the process to finish
   while kill -0 $pid 2>/dev/null; do
       sleep 1
   done

   # test processes are killed
   assert_no_pid "$pid_agent"
   assert_no_pid "$pid_plugin"
   assert_no_pid "$pid_sleep"
}

@test "it should remove all the processes on agent process stopped (SIGINT)" {
   # declare the expected processes
   cmd_agent="trento-agent facts gather --plugins-folder $BUILD_DIR/plugin_examples --gatherer sleep --argument 2s"
   cmd_plugin="$BUILD_DIR/plugin_examples/sleep"
   cmd_sleep="sleep 2s"

   # start the agent in background
   eval "$cmd_agent &" 
   pid=$!

   # retrieve the pid of the exepcted process
   pid_agent=$(pgrep -f "$cmd_agent")
   pid_plugin=$(pgrep -f "$cmd_plugin")
   pid_sleep=$(pgrep -f "$cmd_sleep")

   # double check the test is correct
   [ $pid -eq $pid_agent ]

   # ensure no duplicated processes are running
   assert_one "$pid_agent"
   assert_one "$pid_plugin"
   assert_one "$pid_sleep"

   # ensure the process tree is correct
   assert_parent "$pid_plugin" "$pid_agent"
   assert_parent "$pid_sleep" "$pid_plugin"

   # kill the agent
   kill -INT $pid_agent

   # test processes are killed
   assert_no_pid "$pid_agent"
   assert_no_pid "$pid_plugin"
   assert_no_pid "$pid_sleep"
}

@test "it should remove all the processes on agent process stopped (SIGTERM)" {
   # declare the expected processes
   cmd_agent="trento-agent facts gather --plugins-folder $BUILD_DIR/plugin_examples --gatherer sleep --argument 2s"
   cmd_plugin="$BUILD_DIR/plugin_examples/sleep"
   cmd_sleep="sleep 2s"

   # start the agent in background
   eval "$cmd_agent &" 
   pid=$!

   # retrieve the pid of the exepcted process
   pid_agent=$(pgrep -f "$cmd_agent")
   pid_plugin=$(pgrep -f "$cmd_plugin")
   pid_sleep=$(pgrep -f "$cmd_sleep")

   # double check the test is correct
   [ $pid -eq $pid_agent ]

   # ensure no duplicated processes are running
   assert_one "$pid_agent"
   assert_one "$pid_plugin"
   assert_one "$pid_sleep"

   # ensure the process tree is correct
   assert_parent "$pid_plugin" "$pid_agent"
   assert_parent "$pid_sleep" "$pid_plugin"

   # kill the agent
   kill -TERM $pid_agent

   # test processes are killed
   assert_no_pid "$pid_agent"
   assert_no_pid "$pid_plugin"
   assert_no_pid "$pid_sleep"
}


function assert_one {
    [ $(echo "$1" | wc -l)  == 1  ]
}

function assert_parent {
    [ $(ps -o ppid= -p "$1") == "$2" ]
}

function assert_no_pid {
    [ $(ps -p "$1" | wc -l) == 1 ]
}