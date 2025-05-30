# Changelog

## [2.5.0](https://github.com/trento-project/agent/tree/2.4.0/compare/2.4.0...2.5.0) - 2025-05-30

### What's Changed

* Update CI (#442) @stefanotorresi
* Use absolute command paths (#421) @balanza
* update installation script and readme instructions (#439) @stefanotorresi
* Handle context cancellation in dir scan gatherer (#382) @balanza
* Fix CI job names (#429) @stefanotorresi
* Streamline CI workflow (#428) @stefanotorresi
* Update packaging and service files for SLES 16 (#425) @janvhs
* Add `sudoers` gatherer (#404) @balanza
* Update permissions of agent.yaml (#407) @janvhs
* Add gatherer to retrieve facts from `global.ini` configuration (#395) @balanza
* Update agent to golang 1.23 (#391) @balanza
* Handle context cancellation in systemd gatherer  (#390) @balanza
* Handle context cancellation in fact gatherers (#384) @balanza
* Add healthcheck to prevent flacky tests (#371) @balanza
* Implement context cancellation for command based gatherers (#379) @balanza
* Test command termination error (#378) @balanza
* Handle context cancellation in `corosync-cmapctl` gatherer (#376) @balanza
* Better document the config example (#335) @stefanotorresi
* update license notice (#372) @stefanotorresi
* Add tests on built binaries (#362) @balanza
* Context propagation in facts gathering flow (#360) @balanza

#### Features

* Add expiration check to facts gathering request (#441) @arbulu89
* Add CODEOWNERS (#437) @nelsonkopliku
* Listen discovery requests (#424) @arbulu89
* Run operations (#422) @arbulu89
* Support aws imds v2 (#423) @nelsonkopliku
* Generic event handler and messaging (#420) @arbulu89
* Identify current instance SAP instances (#418) @arbulu89
* RabbitMQ Supports SSL Config (#406) @CDimonaco
* Handle context in plugins (#361) @balanza
* Customize node exporter target (#356) @arbulu89

#### Bug Fixes

* Use discoveries as list instead of map to guarantee order (#436) @arbulu89
* Support aws imds v2 (#423) @nelsonkopliku
* Identify j2ee instances as application instances (#416) @arbulu89

#### Maintenance

* Generic event handler and messaging (#420) @arbulu89
* Upgrade github actions runner ubuntu version (#405) @arbulu89

#### Dependencies

<details>
<summary>28 changes</summary>
* Bump github.com/vektra/mockery/v2 from 2.53.3 to 2.53.4 (#438) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump github.com/spf13/viper from 1.20.0 to 1.20.1 (#434) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump github.com/hashicorp/go-plugin from 1.6.2 to 1.6.3 (#433) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump golang.org/x/mod from 0.23.0 to 0.24.0 (#432) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump github.com/prometheus-community/pro-bing from 0.6.1 to 0.7.0 (#431) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump golang.org/x/sync from 0.11.0 to 0.14.0 (#430) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump github.com/spf13/afero from 1.12.0 to 1.14.0 (#426) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump github.com/hashicorp/go-hclog from 1.5.0 to 1.6.3 (#367) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump actions/cache from 4.2.2 to 4.2.3 (#412) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump github.com/clbanning/mxj/v2 from 2.5.7 to 2.7.0 (#241) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump github.com/spf13/cobra from 1.8.1 to 1.9.1 (#398) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump google.golang.org/protobuf from 1.36.2 to 1.36.6 (#415) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump github.com/vektra/mockery/v2 from 2.40.1 to 2.53.3 (#411) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump actions/cache from 4.2.0 to 4.2.2 (#400) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump contracts library to 0.2.0 (#399) @CDimonaco
* Bump github.com/spf13/pflag from 1.0.5 to 1.0.6 (#389) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump github.com/stretchr/testify from 1.9.0 to 1.10.0 (#374) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump isbang/compose-action from 2.0.2 to 2.2.0 (#385) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump github.com/prometheus-community/pro-bing from 0.5.0 to 0.6.1 (#386) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump github.com/spf13/cobra from 1.7.0 to 1.8.1 (#375) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump github.com/prometheus-community/pro-bing from 0.3.0 to 0.5.0 (#368) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump github.com/moby/sys/mountinfo from 0.6.2 to 0.7.2 (#370) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump github.com/spf13/viper from 1.16.0 to 1.19.0 (#363) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump github.com/spf13/afero from 1.9.5 to 1.12.0 (#364) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump google.golang.org/protobuf from 1.31.0 to 1.36.2 (#365) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump golang.org/x/sync from 0.6.0 to 0.10.0 (#357) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump github.com/hashicorp/go-plugin from 1.5.1 to 1.6.2 (#352) @[dependabot[bot]](https://github.com/apps/dependabot)
* Bump actions/cache from 4.1.1 to 4.2.0 (#358) @[dependabot[bot]](https://github.com/apps/dependabot)

</details>
**Full Changelog**: https://github.com/trento-project/agent/compare/2.4.0...2.5.0

## [2.4.0](https://github.com/trento-project/agent/tree/2.4.0) (2024-11-11)

[Full Changelog](https://github.com/trento-project/agent/compare/2.3.0...2.4.0)

### Added

- Discover host ip address netmasks [#346](https://github.com/trento-project/agent/pull/346) (@arbulu89)
- Update agent to golang 1.22 [#344](https://github.com/trento-project/agent/pull/344) (@CDimonaco)
- Sapservices gatherer improvement [#337](https://github.com/trento-project/agent/pull/337) (@CDimonaco)

### Fixed

- Fix install-agent.sh with --use-tgz [#334](https://github.com/trento-project/agent/pull/334) (@stefanotorresi)
- Updated osutil to fix missing dependency [#333](https://github.com/trento-project/agent/pull/333) (@CDimonaco)

### Closed Issues

- wrong path in install-agent.sh [#222](https://github.com/trento-project/agent/issues/222)

### Other Changes

- Upload updated rpm spec file in obs commit ci step [#351](https://github.com/trento-project/agent/pull/351) (@arbulu89)
- Bump actions/cache from 4.1.0 to 4.1.1 [#350](https://github.com/trento-project/agent/pull/350) (@dependabot[bot])
- Bump actions/cache from 4.0.2 to 4.1.0 [#349](https://github.com/trento-project/agent/pull/349) (@dependabot[bot])
- Add pr template for agent repo [#341](https://github.com/trento-project/agent/pull/341) (@EMaksy)
- Bump golangci/golangci-lint-action from 4 to 6 [#338](https://github.com/trento-project/agent/pull/338) (@dependabot[bot])
- Bump actions/cache from 4.0.0 to 4.0.2 [#330](https://github.com/trento-project/agent/pull/330) (@dependabot[bot])
- Bump golangci/golangci-lint-action from 3 to 4 [#320](https://github.com/trento-project/agent/pull/320) (@dependabot[bot])
- Bump AButler/upload-release-assets from 2.0 to 3.0 [#301](https://github.com/trento-project/agent/pull/301) (@dependabot[bot])

## [2.3.0](https://github.com/trento-project/agent/tree/2.3.0) (2024-05-13)

[Full Changelog](https://github.com/trento-project/agent/compare/2.2.0...2.3.0)

### Added

- Add new ascsers_cluster gatherer [#329](https://github.com/trento-project/agent/pull/329) (@arbulu89)
- Add cache usage to sapcontrol gatherer [#328](https://github.com/trento-project/agent/pull/328) (@arbulu89)
- Improve crash on failed agent initialization [#326](https://github.com/trento-project/agent/pull/326) (@rtorrero)
- Facts cache [#325](https://github.com/trento-project/agent/pull/325) (@arbulu89)
- Update saptune discovery interval [#322](https://github.com/trento-project/agent/pull/322) (@rtorrero)
- Send sbd data in diskless scenario [#321](https://github.com/trento-project/agent/pull/321) (@arbulu89)
- Discover FQDN inside the host discovery loop [#313](https://github.com/trento-project/agent/pull/313) (@dottorblaster)
- Print fact gathering error in facts gather cmd call [#304](https://github.com/trento-project/agent/pull/304) (@arbulu89)
- Http client refactor [#300](https://github.com/trento-project/agent/pull/300) (@CDimonaco)

### Fixed

- Fix verify_password gatherer for scenarios where there is not hash [#305](https://github.com/trento-project/agent/pull/305) (@arbulu89)

### Other Changes

- Update license year [#323](https://github.com/trento-project/agent/pull/323) (@EMaksy)
- Update LICENSE [#319](https://github.com/trento-project/agent/pull/319) (@stefanotorresi)
- Bump github.com/iancoleman/strcase from 0.2.0 to 0.3.0 [#317](https://github.com/trento-project/agent/pull/317) (@dependabot[bot])
- Bump actions/cache from 3.3.3 to 4.0.0 [#316](https://github.com/trento-project/agent/pull/316) (@dependabot[bot])
- Bump golang.org/x/mod from 0.9.0 to 0.14.0 [#315](https://github.com/trento-project/agent/pull/315) (@dependabot[bot])
- Bump github.com/vektra/mockery/v2 from 2.32.3 to 2.40.1 [#314](https://github.com/trento-project/agent/pull/314) (@dependabot[bot])
- Bump actions/cache from 3.3.2 to 3.3.3 [#312](https://github.com/trento-project/agent/pull/312) (@dependabot[bot])
- Bump actions/upload-artifact from 3 to 4 [#308](https://github.com/trento-project/agent/pull/308) (@dependabot[bot])
- Bump actions/download-artifact from 3 to 4 [#307](https://github.com/trento-project/agent/pull/307) (@dependabot[bot])
- Bump actions/setup-go from 3 to 5 [#306](https://github.com/trento-project/agent/pull/306) (@dependabot[bot])
- Refactor factsengine package tests [#295](https://github.com/trento-project/agent/pull/295) (@arbulu89)
- Refactor sapsystem package code and tests [#294](https://github.com/trento-project/agent/pull/294) (@arbulu89)
- Refactor saptune and subscriptions tests [#293](https://github.com/trento-project/agent/pull/293) (@arbulu89)
- Refactor cluster package tests [#292](https://github.com/trento-project/agent/pull/292) (@arbulu89)
- Refactor cloud package tests [#291](https://github.com/trento-project/agent/pull/291) (@arbulu89)
- Refactor cmd and agent package tests [#289](https://github.com/trento-project/agent/pull/289) (@arbulu89)
- Refactor collector package tests [#288](https://github.com/trento-project/agent/pull/288) (@arbulu89)
- Add testpackage to golangci [#287](https://github.com/trento-project/agent/pull/287) (@arbulu89)
- Refactor gatherers package tests [#286](https://github.com/trento-project/agent/pull/286) (@arbulu89)

## [2.2.0](https://github.com/trento-project/agent/tree/2.2.0) (2023-11-14)

[Full Changelog](https://github.com/trento-project/agent/compare/2.1.0...2.2.0)

### Added

- Products gatherer [#285](https://github.com/trento-project/agent/pull/285) (@arbulu89)
- Mount info gatherer [#284](https://github.com/trento-project/agent/pull/284) (@arbulu89)
- Add os-release gatherer [#283](https://github.com/trento-project/agent/pull/283) (@rtorrero)
- Sapservices gatherer  [#282](https://github.com/trento-project/agent/pull/282) (@CDimonaco)
- Disp work gatherer [#281](https://github.com/trento-project/agent/pull/281) (@arbulu89)
- Add .tool-versions file [#280](https://github.com/trento-project/agent/pull/280) (@arbulu89)
- Systemd V2 gatherer [#278](https://github.com/trento-project/agent/pull/278) (@arbulu89)
- Gatherers versioning [#277](https://github.com/trento-project/agent/pull/277) (@CDimonaco)
- Resolver gatherer [#274](https://github.com/trento-project/agent/pull/274) (@rtorrero)
- Refactor NewSAPSystemsList to have a default function [#273](https://github.com/trento-project/agent/pull/273) (@arbulu89)
- Sapcontrol gatherer [#270](https://github.com/trento-project/agent/pull/270) (@arbulu89)
- Dir scan gatherer [#269](https://github.com/trento-project/agent/pull/269) (@CDimonaco)
- sysctl gatherer [#268](https://github.com/trento-project/agent/pull/268) (@rtorrero)
- Sap profile gatherer [#267](https://github.com/trento-project/agent/pull/267) (@arbulu89)
- Fstab gatherer [#265](https://github.com/trento-project/agent/pull/265) (@CDimonaco)
- Groups gatherer [#264](https://github.com/trento-project/agent/pull/264) (@CDimonaco)
- Handle nil value in NewFactValue [#263](https://github.com/trento-project/agent/pull/263) (@arbulu89)
- Add passwd file gatherer [#261](https://github.com/trento-project/agent/pull/261) (@arbulu89)
- Add configuration options to NewFactValue [#260](https://github.com/trento-project/agent/pull/260) (@arbulu89)
- Saptune gatherer [#256](https://github.com/trento-project/agent/pull/256) (@rtorrero)
- Use json.RawMessage type to avoid unnecessary json.Unmarshal [#255](https://github.com/trento-project/agent/pull/255) (@rtorrero)
- Add saptune discovery [#253](https://github.com/trento-project/agent/pull/253) (@rtorrero)
- Ignore extra output when comparing versions with zypper [#243](https://github.com/trento-project/agent/pull/243) (@nelsonkopliku)

### Other Changes

- Bump actions/cache from 3.3.1 to 3.3.2 [#252](https://github.com/trento-project/agent/pull/252) (@dependabot[bot])
- Bump github.com/hashicorp/go-plugin from 1.5.0 to 1.5.1 [#250](https://github.com/trento-project/agent/pull/250) (@dependabot[bot])
- Bump github.com/hashicorp/go-plugin from 1.4.10 to 1.5.0 [#249](https://github.com/trento-project/agent/pull/249) (@dependabot[bot])
- Bump actions/checkout from 3 to 4 [#248](https://github.com/trento-project/agent/pull/248) (@dependabot[bot])
- bump contracts version [#246](https://github.com/trento-project/agent/pull/246) (@nelsonkopliku)
- Bump github.com/google/uuid from 1.3.0 to 1.3.1 [#244](https://github.com/trento-project/agent/pull/244) (@dependabot[bot])
- Bump github.com/vektra/mockery/v2 from 2.27.1 to 2.32.3 [#240](https://github.com/trento-project/agent/pull/240) (@dependabot[bot])
- Bump golang.org/x/sync from 0.2.0 to 0.3.0 [#239](https://github.com/trento-project/agent/pull/239) (@dependabot[bot])
- Bump google.golang.org/protobuf from 1.30.0 to 1.31.0 [#238](https://github.com/trento-project/agent/pull/238) (@dependabot[bot])
- Bump github.com/spf13/viper from 1.15.0 to 1.16.0 [#237](https://github.com/trento-project/agent/pull/237) (@dependabot[bot])
- Remove not needed certificates [#236](https://github.com/trento-project/agent/pull/236) (@nelsonkopliku)
- Bump github.com/hashicorp/go-plugin from 1.4.8 to 1.4.10 [#232](https://github.com/trento-project/agent/pull/232) (@dependabot[bot])
- Bump github.com/sirupsen/logrus from 1.9.0 to 1.9.3 [#231](https://github.com/trento-project/agent/pull/231) (@dependabot[bot])
- Bump github.com/stretchr/testify from 1.8.2 to 1.8.4 [#230](https://github.com/trento-project/agent/pull/230) (@dependabot[bot])
- Bump github.com/spf13/cobra from 1.6.1 to 1.7.0 [#215](https://github.com/trento-project/agent/pull/215) (@dependabot[bot])

## [2.1.0](https://github.com/trento-project/agent/tree/2.1.0) (2023-08-02)

[Full Changelog](https://github.com/trento-project/agent/compare/2.0.0...2.1.0)

### Added

- bump contracts version [#233](https://github.com/trento-project/agent/pull/233) (@nelsonkopliku)
- Send an empty payload if a cluster was not found [#227](https://github.com/trento-project/agent/pull/227) (@fabriziosestito)

### Closed Issues

- Cloned VMs in VMware  have all the same uuid [#223](https://github.com/trento-project/agent/issues/223)

### Other Changes

- Update copyright year to 2023 [#226](https://github.com/trento-project/agent/pull/226) (@EMaksy)
- Bump github.com/vektra/mockery/v2 from 2.24.0 to 2.27.1 [#225](https://github.com/trento-project/agent/pull/225) (@dependabot[bot])
- Bump golang.org/x/sync from 0.1.0 to 0.2.0 [#224](https://github.com/trento-project/agent/pull/224) (@dependabot[bot])

## [2.0.0](https://github.com/trento-project/agent/tree/2.0.0) (2023-04-26)

[Full Changelog](https://github.com/trento-project/agent/compare/1.2.0...2.0.0)

### Added

- Parse durations in cibadmin gatherer  [#204](https://github.com/trento-project/agent/pull/204) (@fabriziosestito)
- Add ability to detect if running on `VMware` system [#193](https://github.com/trento-project/agent/pull/193) (@jamie-suse)
- Pin web api version to v1 [#186](https://github.com/trento-project/agent/pull/186) (@CDimonaco)
- Multiversion package support [#181](https://github.com/trento-project/agent/pull/181) (@nelsonkopliku)
- Pretty print fact values [#176](https://github.com/trento-project/agent/pull/176) (@dottorblaster)
- Unhide facts service url flag [#172](https://github.com/trento-project/agent/pull/172) (@arbulu89)
- Add version comparison functionality for package_version [#169](https://github.com/trento-project/agent/pull/169) (@rtorrero)
- Make `corosynccmapctl` gatherer output a map structure [#168](https://github.com/trento-project/agent/pull/168) (@jamie-suse)
- Add initial support to verify the password for the hacluster user [#164](https://github.com/trento-project/agent/pull/164) (@rtorrero)
- Add argument validation for gatherers that require it [#162](https://github.com/trento-project/agent/pull/162) (@rtorrero)
- Hidden agent id flag [#160](https://github.com/trento-project/agent/pull/160) (@arbulu89)
- Sbd dump gatherer [#156](https://github.com/trento-project/agent/pull/156) (@nelsonkopliku)
- Retrieve agent id command [#154](https://github.com/trento-project/agent/pull/154) (@nelsonkopliku)
- Port cibadmin gatherer [#149](https://github.com/trento-project/agent/pull/149) (@arbulu89)
- Restructure project folders structure [#147](https://github.com/trento-project/agent/pull/147) (@arbulu89)
- Generic get value [#146](https://github.com/trento-project/agent/pull/146) (@arbulu89)
- Refactor sbd loading [#145](https://github.com/trento-project/agent/pull/145) (@nelsonkopliku)
- Corosynccmap ctl gatherer port [#144](https://github.com/trento-project/agent/pull/144) (@rtorrero)
- Refactor sbd gatherer [#141](https://github.com/trento-project/agent/pull/141) (@nelsonkopliku)
- Packageversion gatherer [#140](https://github.com/trento-project/agent/pull/140) (@rtorrero)
- Port systemd gatherer [#139](https://github.com/trento-project/agent/pull/139) (@arbulu89)
- Gather all hosts entries when no arg is provided [#137](https://github.com/trento-project/agent/pull/137) (@rtorrero)
- Add FactValue type [#133](https://github.com/trento-project/agent/pull/133) (@fabriziosestito)
- Implement /etc/hosts file gatherer [#78](https://github.com/trento-project/agent/pull/78) (@rtorrero)
- Implement saphostctrl gatherer [#71](https://github.com/trento-project/agent/pull/71) (@arbulu89)

### Fixed

- Fix getValue function when map is empty [#218](https://github.com/trento-project/agent/pull/218) (@arbulu89)
- Cibadmin meta attributes to list [#211](https://github.com/trento-project/agent/pull/211) (@arbulu89)
- Fix broken zypper output parsing in package_version due to `\n` [#173](https://github.com/trento-project/agent/pull/173) (@rtorrero)
- Handle `CorosyncCmapctlGatherer` receiving empty lines [#171](https://github.com/trento-project/agent/pull/171) (@jamie-suse)
- Fix cluster_property_set parsing [#170](https://github.com/trento-project/agent/pull/170) (@fabriziosestito)
- Fix list conversion issues in the xml gatherer [#157](https://github.com/trento-project/agent/pull/157) (@arbulu89)
- Fix special lists usage in corosyncconf gatherer [#155](https://github.com/trento-project/agent/pull/155) (@arbulu89)

### Removed

- Remove ssh address references [#174](https://github.com/trento-project/agent/pull/174) (@arbulu89)

### Other Changes

- Bump github.com/vektra/mockery/v2 from 2.22.1 to 2.24.0 [#213](https://github.com/trento-project/agent/pull/213) (@dependabot[bot])
- Bump github.com/hashicorp/go-hclog from 1.3.1 to 1.5.0 [#209](https://github.com/trento-project/agent/pull/209) (@dependabot[bot])
- Bump google.golang.org/protobuf from 1.29.1 to 1.30.0 [#206](https://github.com/trento-project/agent/pull/206) (@dependabot[bot])
- Bump google.golang.org/protobuf from 1.28.1 to 1.29.1 [#203](https://github.com/trento-project/agent/pull/203) (@dependabot[bot])
- update spec file [#202](https://github.com/trento-project/agent/pull/202) (@stefanotorresi)
- Bump actions/cache from 3.2.6 to 3.3.1 [#201](https://github.com/trento-project/agent/pull/201) (@dependabot[bot])
- Bump github.com/vektra/mockery/v2 from 2.21.3 to 2.22.1 [#200](https://github.com/trento-project/agent/pull/200) (@dependabot[bot])
- Bump github.com/vektra/mockery/v2 from 2.20.2 to 2.21.3 [#197](https://github.com/trento-project/agent/pull/197) (@dependabot[bot])
- Bump github.com/spf13/afero from 1.9.4 to 1.9.5 [#196](https://github.com/trento-project/agent/pull/196) (@dependabot[bot])
- Bump github.com/stretchr/testify from 1.8.1 to 1.8.2 [#192](https://github.com/trento-project/agent/pull/192) (@dependabot[bot])
- Bump github.com/spf13/afero from 1.9.3 to 1.9.4 [#191](https://github.com/trento-project/agent/pull/191) (@dependabot[bot])
- Add reviewers to dependabot [#190](https://github.com/trento-project/agent/pull/190) (@fabriziosestito)
- Bump github.com/vektra/mockery/v2 from 2.20.0 to 2.20.2 [#189](https://github.com/trento-project/agent/pull/189) (@dependabot[bot])
- Bump actions/cache from 3.2.5 to 3.2.6 [#188](https://github.com/trento-project/agent/pull/188) (@dependabot[bot])
- Trigger golang docs update in ci [#187](https://github.com/trento-project/agent/pull/187) (@arbulu89)
- Bump github.com/vektra/mockery/v2 from 2.19.0 to 2.20.0 [#185](https://github.com/trento-project/agent/pull/185) (@dependabot[bot])
- Bump github.com/vektra/mockery/v2 from 2.18.0 to 2.19.0 [#183](https://github.com/trento-project/agent/pull/183) (@dependabot[bot])
- Bump actions/cache from 3.2.3 to 3.2.5 [#182](https://github.com/trento-project/agent/pull/182) (@dependabot[bot])
- Bump github.com/vektra/mockery/v2 from 2.16.0 to 2.18.0 [#179](https://github.com/trento-project/agent/pull/179) (@dependabot[bot])
- Disable lll linter rule for test files [#177](https://github.com/trento-project/agent/pull/177) (@dottorblaster)
- Bump github.com/spf13/viper from 1.14.0 to 1.15.0 [#175](https://github.com/trento-project/agent/pull/175) (@dependabot[bot])
- Bump actions/cache from 3.2.2 to 3.2.3 [#166](https://github.com/trento-project/agent/pull/166) (@dependabot[bot])
- Bump actions/cache from 3.0.11 to 3.2.2 [#163](https://github.com/trento-project/agent/pull/163) (@dependabot[bot])
- Bump github.com/vektra/mockery/v2 from 2.15.0 to 2.16.0 [#158](https://github.com/trento-project/agent/pull/158) (@dependabot[bot])
- Bump github.com/hashicorp/go-plugin from 1.4.7 to 1.4.8 [#153](https://github.com/trento-project/agent/pull/153) (@dependabot[bot])
- Bump github.com/hashicorp/go-plugin from 1.4.5 to 1.4.7 [#151](https://github.com/trento-project/agent/pull/151) (@dependabot[bot])
- Change compose & test rabbitmq port [#148](https://github.com/trento-project/agent/pull/148) (@fabriziosestito)
- Update CONTRIBUTING.md [#143](https://github.com/trento-project/agent/pull/143) (@fabriziosestito)
- Coveralls [#142](https://github.com/trento-project/agent/pull/142) (@arbulu89)
- Bump github.com/vektra/mockery/v2 from 2.14.1 to 2.15.0 [#138](https://github.com/trento-project/agent/pull/138) (@dependabot[bot])
- Bump github.com/spf13/afero from 1.9.2 to 1.9.3 [#136](https://github.com/trento-project/agent/pull/136) (@dependabot[bot])
- Bump github.com/spf13/cobra from 1.5.0 to 1.6.1 [#135](https://github.com/trento-project/agent/pull/135) (@dependabot[bot])
- Bump github.com/coreos/go-systemd/v22 from 22.3.2 to 22.5.0 [#132](https://github.com/trento-project/agent/pull/132) (@dependabot[bot])
- Bump github.com/spf13/viper from 1.12.0 to 1.14.0 [#131](https://github.com/trento-project/agent/pull/131) (@dependabot[bot])
- Bump github.com/vektra/mockery/v2 from 2.12.3 to 2.14.1 [#128](https://github.com/trento-project/agent/pull/128) (@dependabot[bot])
- Bump actions/cache from 3.0.6 to 3.0.11 [#119](https://github.com/trento-project/agent/pull/119) (@dependabot[bot])
- Bump github.com/hashicorp/go-hclog from 1.2.2 to 1.3.1 [#109](https://github.com/trento-project/agent/pull/109) (@dependabot[bot])

## [1.2.0](https://github.com/trento-project/agent/tree/1.2.0) (2022-11-04)

[Full Changelog](https://github.com/trento-project/agent/compare/1.1.0...1.2.0)

### Added

- Add GroupID field to the FactsGathered event mapping [#125](https://github.com/trento-project/agent/pull/125) (@arbulu89)
- Use google protobuf value in Fact message [#120](https://github.com/trento-project/agent/pull/120) (@fabriziosestito)
- Update Contracts with ContenType fetching from facade [#115](https://github.com/trento-project/agent/pull/115) (@CDimonaco)
- Gatherer/Gatherers plugin management [#114](https://github.com/trento-project/agent/pull/114) (@CDimonaco)
- Add kvm discovery [#113](https://github.com/trento-project/agent/pull/113) (@rtorrero)
- Move test fixture files [#111](https://github.com/trento-project/agent/pull/111) (@arbulu89)
- Detect Nutanix as underlying platform provider [#110](https://github.com/trento-project/agent/pull/110) (@nelsonkopliku)
- Check type assertion properly [#107](https://github.com/trento-project/agent/pull/107) (@arbulu89)
- Sbd discovery di [#106](https://github.com/trento-project/agent/pull/106) (@arbulu89)
- Sapsystems code declarative init [#105](https://github.com/trento-project/agent/pull/105) (@arbulu89)
- Use proper di in cloud discovery [#104](https://github.com/trento-project/agent/pull/104) (@arbulu89)
- Use proper di in the sles subscription discovery [#103](https://github.com/trento-project/agent/pull/103) (@arbulu89)
- Extract command executor to utils package [#102](https://github.com/trento-project/agent/pull/102) (@arbulu89)
- Refinement in the main README [#101](https://github.com/trento-project/agent/pull/101) (@mpagot)
- Factsengine integration test [#100](https://github.com/trento-project/agent/pull/100) (@arbulu89)
- Fact gathering errors [#99](https://github.com/trento-project/agent/pull/99) (@arbulu89)
- Map the numeric strings as numbers to send the event [#97](https://github.com/trento-project/agent/pull/97) (@arbulu89)
- Fact gathering requested [#95](https://github.com/trento-project/agent/pull/95) (@arbulu89)
- Move used strucs on the factsengine to a entities package [#90](https://github.com/trento-project/agent/pull/90) (@arbulu89)
- Publish gathered facts using contract [#88](https://github.com/trento-project/agent/pull/88) (@arbulu89)
- Move the individual unit test function to suites [#85](https://github.com/trento-project/agent/pull/85) (@arbulu89)
- Work on FIXMEs part 1 [#84](https://github.com/trento-project/agent/pull/84) (@arbulu89)
- Use DI for the CommandExecutor [#81](https://github.com/trento-project/agent/pull/81) (@rtorrero)
- Upgrade to golang 1.18 [#80](https://github.com/trento-project/agent/pull/80) (@dottorblaster)
- Sbd gatherer [#77](https://github.com/trento-project/agent/pull/77) (@nelsonkopliku)
- Implement the hacluster password verify gatherer [#70](https://github.com/trento-project/agent/pull/70) (@arbulu89)
- Implement systemd daemons state gatherer [#69](https://github.com/trento-project/agent/pull/69) (@arbulu89)
- Implement crm_mon and cibadmin gatherers [#68](https://github.com/trento-project/agent/pull/68) (@arbulu89)
- Corosync cmapctl gatherer [#67](https://github.com/trento-project/agent/pull/67) (@rtorrero)
- Package version gatherer [#66](https://github.com/trento-project/agent/pull/66) (@arbulu89)
- Add AgentID and CheckID fields to facts result [#65](https://github.com/trento-project/agent/pull/65) (@arbulu89)
- Remove the flavor field and add the installation source [#64](https://github.com/trento-project/agent/pull/64) (@arbulu89)
- Add plugins system [#63](https://github.com/trento-project/agent/pull/63) (@arbulu89)
- Gather facts command [#62](https://github.com/trento-project/agent/pull/62) (@arbulu89)
- Linter configuration [#61](https://github.com/trento-project/agent/pull/61) (@CDimonaco)
- Facts engine [#54](https://github.com/trento-project/agent/pull/54) (@arbulu89)

### Fixed

- fix workflow name [#126](https://github.com/trento-project/agent/pull/126) (@gereonvey)
- Fix CI woops [#123](https://github.com/trento-project/agent/pull/123) (@fabriziosestito)
- Fix SAP profile comments parsing [#122](https://github.com/trento-project/agent/pull/122) (@fabriziosestito)
- Fix GHA obs jobs [#118](https://github.com/trento-project/agent/pull/118) (@stefanotorresi)
- Fix integration test to cancel properly listen function [#112](https://github.com/trento-project/agent/pull/112) (@arbulu89)
- Use correct Systemd testsuite [#75](https://github.com/trento-project/agent/pull/75) (@nelsonkopliku)

### Removed

- Remove mapstructure annotations [#108](https://github.com/trento-project/agent/pull/108) (@arbulu89)

### Other Changes

- Refactor factengine integration test [#121](https://github.com/trento-project/agent/pull/121) (@fabriziosestito)
- Bump github.com/hashicorp/go-plugin from 1.4.4 to 1.4.5 [#82](https://github.com/trento-project/agent/pull/82) (@dependabot[bot])
- Bump github.com/hashicorp/go-hclog from 1.2.0 to 1.2.2 [#76](https://github.com/trento-project/agent/pull/76) (@dependabot[bot])
- Bump actions/cache from 3.0.2 to 3.0.6 [#72](https://github.com/trento-project/agent/pull/72) (@dependabot[bot])
- Bump github.com/sirupsen/logrus from 1.8.1 to 1.9.0 [#60](https://github.com/trento-project/agent/pull/60) (@dependabot[bot])
- Bump github.com/spf13/afero from 1.8.2 to 1.9.2 [#59](https://github.com/trento-project/agent/pull/59) (@dependabot[bot])
- Bump github.com/stretchr/testify from 1.7.1 to 1.8.0 [#51](https://github.com/trento-project/agent/pull/51) (@dependabot[bot])
- Bump github.com/spf13/cobra from 1.4.0 to 1.5.0 [#48](https://github.com/trento-project/agent/pull/48) (@dependabot[bot])

## [1.1.0](https://github.com/trento-project/agent/tree/1.1.0) (2022-07-14)

[Full Changelog](https://github.com/trento-project/agent/compare/1.0.0...1.1.0)

### Added

- Change trento-premium to be obsolete in the spec [#53](https://github.com/trento-project/agent/pull/53) (@rtorrero)
- Get the agent ID in the main agent code package [#47](https://github.com/trento-project/agent/pull/47) (@arbulu89)
- Discover gcp metadata [#43](https://github.com/trento-project/agent/pull/43) (@arbulu89)
- Discover aws cloud data [#42](https://github.com/trento-project/agent/pull/42) (@arbulu89)
- Add a debug trace to know why the cluster data is not built [#39](https://github.com/trento-project/agent/pull/39) (@arbulu89)

### Fixed

- Quickstart agent installation script not working [#34](https://github.com/trento-project/agent/issues/34)
- Identify SAP diagnostics agent [#55](https://github.com/trento-project/agent/pull/55) (@arbulu89)

### Other Changes

- Bump github.com/spf13/viper from 1.11.0 to 1.12.0 [#37](https://github.com/trento-project/agent/pull/37) (@dependabot[bot])
- Bump github.com/vektra/mockery/v2 from 2.12.2 to 2.12.3 [#36](https://github.com/trento-project/agent/pull/36) (@dependabot[bot])
- Bump github.com/vektra/mockery/v2 from 2.12.1 to 2.12.2 [#35](https://github.com/trento-project/agent/pull/35) (@dependabot[bot])
- Fix URL in the package spec [#33](https://github.com/trento-project/agent/pull/33) (@mpagot)
- Bump github.com/vektra/mockery/v2 from 2.10.6 to 2.12.1 [#28](https://github.com/trento-project/agent/pull/28) (@dependabot[bot])

## [1.0.0](https://github.com/trento-project/agent/tree/1.0.0) (2022-04-29)

[Full Changelog](https://github.com/trento-project/agent/compare/6019c6aab69730839d2e22cf69e4bb83f1da6956...1.0.0)

### Added

- Flat map sap systems payload lists [#23](https://github.com/trento-project/agent/pull/23) (@arbulu89)

### Other Changes

- Restore release-tag job in the CI [#29](https://github.com/trento-project/agent/pull/29) (@arbulu89)
- Detect AWS based on dmidecode system-manufacturer [#27](https://github.com/trento-project/agent/pull/27) (@rtorrero)
- Fix install agent interval [#25](https://github.com/trento-project/agent/pull/25) (@fabriziosestito)
- Rename CloudProvider to Provider [#24](https://github.com/trento-project/agent/pull/24) (@nelsonkopliku)
- Publish csp information of a discovered pacemaker cluster [#21](https://github.com/trento-project/agent/pull/21) (@nelsonkopliku)
- Load HANA database IP address in agent side [#20](https://github.com/trento-project/agent/pull/20) (@arbulu89)
- Fix socket leak [#19](https://github.com/trento-project/agent/pull/19) (@fabriziosestito)
- fixup installation doc [#18](https://github.com/trento-project/agent/pull/18) (@nelsonkopliku)
- Bump github.com/spf13/viper from 1.10.1 to 1.11.0 [#17](https://github.com/trento-project/agent/pull/17) (@dependabot[bot])
- Bump github.com/vektra/mockery/v2 from 2.10.4 to 2.10.6 [#16](https://github.com/trento-project/agent/pull/16) (@dependabot[bot])
- Update installer for the new agent [#15](https://github.com/trento-project/agent/pull/15) (@fabriziosestito)
- Refactor collector port / host in server url [#14](https://github.com/trento-project/agent/pull/14) (@fabriziosestito)
- Add trento agent binary to tgz [#13](https://github.com/trento-project/agent/pull/13) (@fabriziosestito)
- Add api key support [#12](https://github.com/trento-project/agent/pull/12) (@nelsonkopliku)
- Bump actions/setup-go from 2 to 3 [#11](https://github.com/trento-project/agent/pull/11) (@dependabot[bot])
- Bump actions/download-artifact from 2 to 3 [#10](https://github.com/trento-project/agent/pull/10) (@dependabot[bot])
- Bump actions/upload-artifact from 2 to 3 [#9](https://github.com/trento-project/agent/pull/9) (@dependabot[bot])
- Bump actions/cache from 3.0.1 to 3.0.2 [#8](https://github.com/trento-project/agent/pull/8) (@dependabot[bot])
- Add docs back [#7](https://github.com/trento-project/agent/pull/7) (@dottorblaster)
- Refine service file [#6](https://github.com/trento-project/agent/pull/6) (@dottorblaster)
- Name everything trento-agent and try to bring back the OBS CI step [#5](https://github.com/trento-project/agent/pull/5) (@dottorblaster)
- Instruct the specfile to only create the RPM package for the agent binary [#4](https://github.com/trento-project/agent/pull/4) (@dottorblaster)
- Bump actions/checkout from 2 to 3 [#3](https://github.com/trento-project/agent/pull/3) (@dependabot[bot])
- Bump actions/cache from 2 to 3.0.1 [#2](https://github.com/trento-project/agent/pull/2) (@dependabot[bot])
- Add github actions back [#1](https://github.com/trento-project/agent/pull/1) (@dottorblaster)

* *This Changelog was automatically generated by [github_changelog_generator](https://github.com/github-changelog-generator/github-changelog-generator)*
