module github.com/trento-project/agent

go 1.16

require (
	github.com/google/uuid v1.3.0
	github.com/hooklift/gowsdl v0.5.0
	github.com/kr/text v0.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pkg/errors v0.9.1
	github.com/shirou/gopsutil v3.21.11+incompatible
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/afero v1.8.2
	github.com/spf13/cobra v1.4.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.10.1
	github.com/stretchr/testify v1.7.1
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/vektra/mockery/v2 v2.10.4
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
)

replace github.com/trento-project/agent => ./
