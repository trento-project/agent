# Changelog

## [1.2.0](https://github.com/trento-project/agent/tree/1.2.0) (2022-11-04)

[Full Changelog](https://github.com/trento-project/agent/compare/1.1.0...1.2.0)

### Added

- Add GroupID field to the FactsGathered event mapping [\#125](https://github.com/trento-project/agent/pull/125) (@arbulu89)
- Use google protobuf value in Fact message [\#120](https://github.com/trento-project/agent/pull/120) (@fabriziosestito)
- Update Contracts with ContenType fetching from facade [\#115](https://github.com/trento-project/agent/pull/115) (@CDimonaco)
- Gatherer/Gatherers plugin management [\#114](https://github.com/trento-project/agent/pull/114) (@CDimonaco)
- Add kvm discovery [\#113](https://github.com/trento-project/agent/pull/113) (@rtorrero)
- Move test fixture files [\#111](https://github.com/trento-project/agent/pull/111) (@arbulu89)
- Detect Nutanix as underlying platform provider [\#110](https://github.com/trento-project/agent/pull/110) (@nelsonkopliku)
- Check type assertion properly [\#107](https://github.com/trento-project/agent/pull/107) (@arbulu89)
- Sbd discovery di [\#106](https://github.com/trento-project/agent/pull/106) (@arbulu89)
- Sapsystems code declarative init [\#105](https://github.com/trento-project/agent/pull/105) (@arbulu89)
- Use proper di in cloud discovery [\#104](https://github.com/trento-project/agent/pull/104) (@arbulu89)
- Use proper di in the sles subscription discovery [\#103](https://github.com/trento-project/agent/pull/103) (@arbulu89)
- Extract command executor to utils package [\#102](https://github.com/trento-project/agent/pull/102) (@arbulu89)
- Refinement in the main README [\#101](https://github.com/trento-project/agent/pull/101) (@mpagot)
- Factsengine integration test [\#100](https://github.com/trento-project/agent/pull/100) (@arbulu89)
- Fact gathering errors [\#99](https://github.com/trento-project/agent/pull/99) (@arbulu89)
- Map the numeric strings as numbers to send the event [\#97](https://github.com/trento-project/agent/pull/97) (@arbulu89)
- Fact gathering requested [\#95](https://github.com/trento-project/agent/pull/95) (@arbulu89)
- Move used strucs on the factsengine to a entities package [\#90](https://github.com/trento-project/agent/pull/90) (@arbulu89)
- Publish gathered facts using contract [\#88](https://github.com/trento-project/agent/pull/88) (@arbulu89)
- Move the individual unit test function to suites [\#85](https://github.com/trento-project/agent/pull/85) (@arbulu89)
- Work on FIXMEs part 1 [\#84](https://github.com/trento-project/agent/pull/84) (@arbulu89)
- Use DI for the CommandExecutor [\#81](https://github.com/trento-project/agent/pull/81) (@rtorrero)
- Upgrade to golang 1.18 [\#80](https://github.com/trento-project/agent/pull/80) (@dottorblaster)
- Sbd gatherer [\#77](https://github.com/trento-project/agent/pull/77) (@nelsonkopliku)
- Implement the hacluster password verify gatherer [\#70](https://github.com/trento-project/agent/pull/70) (@arbulu89)
- Implement systemd daemons state gatherer [\#69](https://github.com/trento-project/agent/pull/69) (@arbulu89)
- Implement crm\_mon and cibadmin gatherers [\#68](https://github.com/trento-project/agent/pull/68) (@arbulu89)
- Corosync cmapctl gatherer [\#67](https://github.com/trento-project/agent/pull/67) (@rtorrero)
- Package version gatherer [\#66](https://github.com/trento-project/agent/pull/66) (@arbulu89)
- Add AgentID and CheckID fields to facts result [\#65](https://github.com/trento-project/agent/pull/65) (@arbulu89)
- Remove the flavor field and add the installation source [\#64](https://github.com/trento-project/agent/pull/64) (@arbulu89)
- Add plugins system [\#63](https://github.com/trento-project/agent/pull/63) (@arbulu89)
- Gather facts command [\#62](https://github.com/trento-project/agent/pull/62) (@arbulu89)
- Linter configuration [\#61](https://github.com/trento-project/agent/pull/61) (@CDimonaco)
- Facts engine [\#54](https://github.com/trento-project/agent/pull/54) (@arbulu89)

### Fixed

- fix workflow name [\#126](https://github.com/trento-project/agent/pull/126) (@gereonvey)
- Fix CI woops [\#123](https://github.com/trento-project/agent/pull/123) (@fabriziosestito)
- Fix SAP profile comments parsing [\#122](https://github.com/trento-project/agent/pull/122) (@fabriziosestito)
- Fix GHA obs jobs [\#118](https://github.com/trento-project/agent/pull/118) (@stefanotorresi)
- Fix integration test to cancel properly listen function [\#112](https://github.com/trento-project/agent/pull/112) (@arbulu89)
- Use correct Systemd testsuite [\#75](https://github.com/trento-project/agent/pull/75) (@nelsonkopliku)

### Removed

- Remove mapstructure annotations [\#108](https://github.com/trento-project/agent/pull/108) (@arbulu89)

### Other Changes

- Refactor factengine integration test [\#121](https://github.com/trento-project/agent/pull/121) (@fabriziosestito)
- Bump github.com/hashicorp/go-plugin from 1.4.4 to 1.4.5 [\#82](https://github.com/trento-project/agent/pull/82) (@dependabot[bot])
- Bump github.com/hashicorp/go-hclog from 1.2.0 to 1.2.2 [\#76](https://github.com/trento-project/agent/pull/76) (@dependabot[bot])
- Bump actions/cache from 3.0.2 to 3.0.6 [\#72](https://github.com/trento-project/agent/pull/72) (@dependabot[bot])
- Bump github.com/sirupsen/logrus from 1.8.1 to 1.9.0 [\#60](https://github.com/trento-project/agent/pull/60) (@dependabot[bot])
- Bump github.com/spf13/afero from 1.8.2 to 1.9.2 [\#59](https://github.com/trento-project/agent/pull/59) (@dependabot[bot])
- Bump github.com/stretchr/testify from 1.7.1 to 1.8.0 [\#51](https://github.com/trento-project/agent/pull/51) (@dependabot[bot])
- Bump github.com/spf13/cobra from 1.4.0 to 1.5.0 [\#48](https://github.com/trento-project/agent/pull/48) (@dependabot[bot])

## [1.1.0](https://github.com/trento-project/agent/tree/1.1.0) (2022-07-14)

[Full Changelog](https://github.com/trento-project/agent/compare/1.0.0...1.1.0)

### Added

- Change trento-premium to be obsolete in the spec [\#53](https://github.com/trento-project/agent/pull/53) (@rtorrero)
- Get the agent ID in the main agent code package [\#47](https://github.com/trento-project/agent/pull/47) (@arbulu89)
- Discover gcp metadata [\#43](https://github.com/trento-project/agent/pull/43) (@arbulu89)
- Discover aws cloud data [\#42](https://github.com/trento-project/agent/pull/42) (@arbulu89)
- Add a debug trace to know why the cluster data is not built [\#39](https://github.com/trento-project/agent/pull/39) (@arbulu89)

### Fixed

- Quickstart agent installation script not working [\#34](https://github.com/trento-project/agent/issues/34)
- Identify SAP diagnostics agent [\#55](https://github.com/trento-project/agent/pull/55) (@arbulu89)

### Other Changes

- Bump github.com/spf13/viper from 1.11.0 to 1.12.0 [\#37](https://github.com/trento-project/agent/pull/37) (@dependabot[bot])
- Bump github.com/vektra/mockery/v2 from 2.12.2 to 2.12.3 [\#36](https://github.com/trento-project/agent/pull/36) (@dependabot[bot])
- Bump github.com/vektra/mockery/v2 from 2.12.1 to 2.12.2 [\#35](https://github.com/trento-project/agent/pull/35) (@dependabot[bot])
- Fix URL in the package spec [\#33](https://github.com/trento-project/agent/pull/33) (@mpagot)
- Bump github.com/vektra/mockery/v2 from 2.10.6 to 2.12.1 [\#28](https://github.com/trento-project/agent/pull/28) (@dependabot[bot])

## [1.0.0](https://github.com/trento-project/agent/tree/1.0.0) (2022-04-29)

[Full Changelog](https://github.com/trento-project/agent/compare/6019c6aab69730839d2e22cf69e4bb83f1da6956...1.0.0)

### Added

- Flat map sap systems payload lists [\#23](https://github.com/trento-project/agent/pull/23) (@arbulu89)

### Other Changes

- Restore release-tag job in the CI [\#29](https://github.com/trento-project/agent/pull/29) (@arbulu89)
- Detect AWS based on dmidecode system-manufacturer [\#27](https://github.com/trento-project/agent/pull/27) (@rtorrero)
- Fix install agent interval [\#25](https://github.com/trento-project/agent/pull/25) (@fabriziosestito)
- Rename CloudProvider to Provider [\#24](https://github.com/trento-project/agent/pull/24) (@nelsonkopliku)
- Publish csp information of a discovered pacemaker cluster [\#21](https://github.com/trento-project/agent/pull/21) (@nelsonkopliku)
- Load HANA database IP address in agent side [\#20](https://github.com/trento-project/agent/pull/20) (@arbulu89)
- Fix socket leak [\#19](https://github.com/trento-project/agent/pull/19) (@fabriziosestito)
- fixup installation doc [\#18](https://github.com/trento-project/agent/pull/18) (@nelsonkopliku)
- Bump github.com/spf13/viper from 1.10.1 to 1.11.0 [\#17](https://github.com/trento-project/agent/pull/17) (@dependabot[bot])
- Bump github.com/vektra/mockery/v2 from 2.10.4 to 2.10.6 [\#16](https://github.com/trento-project/agent/pull/16) (@dependabot[bot])
- Update installer for the new agent [\#15](https://github.com/trento-project/agent/pull/15) (@fabriziosestito)
- Refactor collector port / host in server url [\#14](https://github.com/trento-project/agent/pull/14) (@fabriziosestito)
- Add trento agent binary to tgz [\#13](https://github.com/trento-project/agent/pull/13) (@fabriziosestito)
- Add api key support [\#12](https://github.com/trento-project/agent/pull/12) (@nelsonkopliku)
- Bump actions/setup-go from 2 to 3 [\#11](https://github.com/trento-project/agent/pull/11) (@dependabot[bot])
- Bump actions/download-artifact from 2 to 3 [\#10](https://github.com/trento-project/agent/pull/10) (@dependabot[bot])
- Bump actions/upload-artifact from 2 to 3 [\#9](https://github.com/trento-project/agent/pull/9) (@dependabot[bot])
- Bump actions/cache from 3.0.1 to 3.0.2 [\#8](https://github.com/trento-project/agent/pull/8) (@dependabot[bot])
- Add docs back [\#7](https://github.com/trento-project/agent/pull/7) (@dottorblaster)
- Refine service file [\#6](https://github.com/trento-project/agent/pull/6) (@dottorblaster)
- Name everything trento-agent and try to bring back the OBS CI step [\#5](https://github.com/trento-project/agent/pull/5) (@dottorblaster)
- Instruct the specfile to only create the RPM package for the agent binary [\#4](https://github.com/trento-project/agent/pull/4) (@dottorblaster)
- Bump actions/checkout from 2 to 3 [\#3](https://github.com/trento-project/agent/pull/3) (@dependabot[bot])
- Bump actions/cache from 2 to 3.0.1 [\#2](https://github.com/trento-project/agent/pull/2) (@dependabot[bot])
- Add github actions back [\#1](https://github.com/trento-project/agent/pull/1) (@dottorblaster)



\* *This Changelog was automatically generated by [github_changelog_generator](https://github.com/github-changelog-generator/github-changelog-generator)*
