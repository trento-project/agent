# Trento Agent

[![CI](https://github.com/trento-project/agent/actions/workflows/ci.yaml/badge.svg)](https://github.com/trento-project/agent/actions/workflows/ci.yaml)
[![Coverage Status](https://coveralls.io/repos/github/trento-project/agent/badge.svg?branch=main)](https://coveralls.io/github/trento-project/agent?branch=main)

The agents are the `trento` processes that run in each of the nodes of a target HA SAP Applications cluster.

# Table of contents

- [Features](#features)
  - [Heartbeating](#heartbeating)
  - [Discovery](#discovery)
- [Installation](#installation)
  - [Requirements](#requirements)
  - [Installation Scripts](#installation-scripts)
  - [Starting Trento Agent service](#starting-trento-agent-service)
  - [Manual installation](#manual-installation)
  - [Manually running Trento Agents](#manually-running-trento-agents)
- [Configuration](#configuration)
  - [Locations](#locations)
- [Environment Variables](#environment-variables)
- [Development](#development)
- [Support](#support)
- [Contributing](#contributing)
- [License](#license)

# Features
Agents provide two main functionalities:
- Heartbeating
- Discovery

## Heartbeating
When Trento agent starts, it begins notifying the control plane about its presence in the cluster by _heartbeting_ periodically.

## Discovery

Agents are responsible for the _discovery_ of all the clustered components that are required in order to run highly available SAP Applications.

Discoveries gather information about these aspects of a target host
- OSVersion, HostName, HostIpAddresses... - _HostDiscovery_
- Cloud Service Provider it is running on - _CloudDiscovery_
- Pacemaker/Corosync/SBD metadata - _ClusterDiscovery_
- SAP components running on it (Application instances, Database instances) - _SAPSystemsDiscovery_
- SLES Subscriptions - _SubscriptionDiscovery_

# Installation

## Requirements

_Trento Agent_ needs to interact with a number of low-level system components
which are part of the [SUSE Linux Enterprise Server for SAP Applications](https://www.suse.com/products/sles-for-sap/) Linux distribution.

These could in theory also be installed and configured on other distributions providing the same functionalities, but this use case is not within the scope of the active development.

In addition to that, _Trento Agent_ also requires the [Prometheus node_exporter component](https://github.com/prometheus/node_exporter) to be running to collect host information for the monitoring functionality.

The resource footprint of _Trento Agent_ should not impact the performance of the host it runs on.

## Installation Scripts

Installation scripts are provided to automatically install and update the latest version of Trento.
Please follow the instructions in the given order.

After the server installation, you might want to install Trento agents in a running cluster.

An installation script is provided, you can `curl | bash` it if you want to live on the edge.

```
curl -sfL https://raw.githubusercontent.com/trento-project/agent/main/install-agent.sh | sudo bash
```

Or you can fetch the script, and then execute it manually.

```
curl -O https://raw.githubusercontent.com/trento-project/agent/main/install-agent.sh
chmod 700 install-agent.sh
sudo ./install-agent.sh
```

The script will ask you for some input.

- `ssh-address`: the address to which the trento-agent should be reachable for ssh connection by the runner for check execution.
- `server-ip`: the address where Trento server can be reached.
- `api-key`: the API key generated by the server that allows agents to actually communicate with the control plane

You can pass these arguments as flags or env variables too:

```
curl -sfL https://raw.githubusercontent.com/trento-project/agent/main/install-agent.sh | sudo bash -s - --ssh-address=192.168.33.10 --server-url=http://192.168.33.1 --api-key <some-api-key>
```

```
SSH_ADDRESS=192.168.33.10 SERVER_IP=192.168.33.1 API_KEY=<some-api-key> sudo ./install-agent.sh
```

## Starting Trento Agent service

The installation script does not start the agent automatically.

You can enable boot startup and launch it with systemd:

```
sudo systemctl enable --now trento-agent
```

Please, make sure the server is running before starting the agent.

That's it! You can now reach the Trento web UI and start using it.

## Manual installation

### Pre-built binaries

Pre-built statically linked binaries are made available via [GitHub releases](https://github.com/trento-project/agent/releases).

### Compile from source

You clone also clone and build it manually:

```shell
git clone https://github.com/trento-project/agent.git
cd agent
make build
```

See the section below to know more about the build dependencies.

## RPM Packages

T.B.D.

## Manually running Trento Agents

Trento Agents are responsible for discovering SAP systems, HA clusters and some additional data. These Agents need to run in the same systems hosting the HA
Cluster services, so running them in isolated environments (e.g. serverless,
containers, etc.) makes little sense, as they won't be able as the discovery mechanisms will not be able to report any host information.

> NOTE: Suggested installation instructions for SUSE-based distributions, adjust accordingly

Install and start `node_exporter`:

```shell
zypper in -y golang-github-prometheus-node_exporter
systemctl start prometheus-node_exporter
```

To start the trento agent:

```shell
./trento-agent start
```

Alternatively, you can use the `trento-agent.service` from this repository and start it, which will start
`node_exporter` automatically as a dependency:
```shell
cp packaging/systemd/trento-agent.service /etc/systemd/system
systemctl daemon-reload
systemctl start trento-agent.service
```

> If the discovery loop is being executed too frequently, and this impacts the Web interface performance, the agent
> has the option to configure the discovery loop mechanism using the various `--<cloud,cluster,host,sapsystem>-discovery-period` flags.
> Increasing this value improves the overall performance of the application

## Configuration

Trento Agent can be run with a config file in replacement of command-line arguments.

### Locations

Configuration, if not otherwise specified by the `--config=/path/to/config.yaml` option, would be searched in following locations:

Note that order represents priority

- `/etc/trento/agent.yaml` <-- first location looked
- `/usr/etc/trento/agent.yaml` <-- fallback here if config not found in previous location
- `~/.config/trento/agent.yaml` aka user's home <-- fallback here

`yaml` is the only supported format at the moment.

### Examples

```
# /etc/trento/agent.yaml

api-key: <api-key-generated-from-the-server>
server-ip: https://localhost
```

## Environment Variables

All of the options supported by the command line and configuration file can be provided as environment variables as well.

The rule is: get the option name eg. `api-key`, replace dashes `-` with underscores `_`, make it uppercase and add a `TRENTO_` prefix.

Examples:

`api-key` -> `TRENTO_API_KEY=<some-api-key> ./trento-agent start`

`server-ip` -> `TRENTO_SERVER_IP=https://localhost ./trento-agent start`

# Development

## Build system

We use GNU Make as a task manager; here are some common targets:

```shell
make # clean, test and build everything

make clean # removes any build artifact
make test # executes all the tests
make fmt # fixes code formatting
make web-assets # invokes the frontend build scripts
make generate # refresh automatically generated code (e.g. static Go mocks)
```

Feel free to peek at the [Makefile](Makefile) to know more.

## Development dependencies

Additionally, for the development we use [`mockery`](https://github.com/vektra/mockery) for the `generate` target, which in turn is required for the `test` target.
You can install it with `go install github.com/vektra/mockery/v2`.

> Be sure to add the `mockery` binary to your `$PATH` environment variable so that `make` can find it. That usually comes with configuring `$GOPATH`, `$GOBIN`, and adding the latter to your `$PATH`.

> Please note that the `trento agent` component requires to be running on
> the OS (_not_ inside a container) so, while it is technically possible to run `trento agent`
> commands in the container, it makes little sense because most of its internals
> require direct access to the host of the HA Cluster components.

## Fake Agent ID

In some circunstances, having a fake Agent ID might be useful, specially during development and testing stages. The hidden `agent-id` flag is available for that.

Here an example on how to use it:

`./trento-agent start --agent-id "800ddd9b-8497-493f-b9fa-1bd6c9afb230"`

> Don't use this flag on production systems, as the agent ID must be unique by definition and any change affects the whole Trento usage.

## Fact gathering plugin system

A plugin system is available in the Agent, in order to add new fact gathering options, so it can run user created checks in the server side.

To create a new plugin (check the [example](plugin_examples/dummy.go) dummy plugin for that) follow the next steps:

- Create a new Golang package. This is as simple as creating a new folder (it can be created anywhere, it doesn't need to be in the Agent code directory) with `.go` file inside. Name the Golang file with a meaningful name (even though, it is not relevant for the usage itself).
- The `.go` file implements the `main` package and imports the `go-plugin` package as seen in the example.
- Implement the gathering function with the `func (s exampleGatherer) Gather(factsRequests []gatherers.FactRequest) ([]gatherers.Fact, error)` signature. This function must gather the facts from the system where the Agent is running.
- This function receives a list of fact gathering requests to gather, which entirely depends on the gathering code nature.
- Copy the `main()` function from the [example](plugin_examples/dummy.go) file. Simply replace the gatherer struct name there.
- Once the plugin is implemented, it must be compiled. Use the next command for that: `go build -o /usr/etc/trento/example ./your_plugin_folder/example.go`. The `-o` flag specifies the destination of the created binary, which the Agent needs to load. This folder is the same specified in the `--plugins-folder` flag in the Agent execution. In this case, the used name for the output in the `-o` flag is relevant, as this name is the gatherer name that must be used in the server side checks declaration.
- In order to see that the plugin is correctly loaded, run: `./trento-agent facts list`.

Find the official gatherers code in: https://github.com/trento-project/agent/tree/main/internal/factsengine/gatherers


> *** By now, it only supports Golang based implementations, but this could be extendable (if this requirement is needed, please open a Github ticket with this feature request).

## SAPControl web service

The SAPControl web service soap client was generated by [hooklift/gowsdl](https://github.com/hooklift/gowsdl),
then the methods and structs needed were cherry-picked and adapted.
For reference, you can find the full, generated, web service code [here](docs/_generated_soap_wsdl.go).

# Support

Please only report bugs via [GitHub issues](https://github.com/trento-project/agent/issues);
for any other inquiry or topic use [GitHub discussion](https://github.com/trento-project/agent/discussions).

# Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md)

# License

Copyright 2021 SUSE LLC

Licensed under the Apache License, Version 2.0 (the "License"); you may not use
this file except in compliance with the License. You may obtain a copy of the
License at

https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed
under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.

[k3s]: https://k3s.io
