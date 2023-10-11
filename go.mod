module github.com/Graylog2/graylog-project-cli

go 1.18

require (
	github.com/fatih/color v1.15.0
	github.com/golang-jwt/jwt/v4 v4.5.0
	github.com/google/go-github/v27 v27.0.6
	github.com/hashicorp/go-version v1.6.0
	github.com/imdario/mergo v0.3.15
	github.com/k0kubun/pp/v3 v3.2.0
	github.com/manifoldco/promptui v0.9.0
	github.com/mattn/go-isatty v0.0.19
	github.com/pelletier/go-toml/v2 v2.0.8
	github.com/pkg/errors v0.9.1
	github.com/samber/lo v1.38.1
	github.com/spf13/cobra v1.7.0
	github.com/spf13/viper v1.15.0
	github.com/yuin/goldmark v1.5.4
	golang.org/x/exp v0.0.0-20230522175609-2e198f4a06a1
	golang.org/x/oauth2 v0.8.0
	golang.org/x/text v0.13.0
)

require (
	github.com/chzyer/readline v1.5.1 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/spf13/afero v1.9.5 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	golang.org/x/crypto v0.14.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Workaround for https://github.com/golang/go/issues/30831
replace github.com/golang/lint v0.0.0-20190313153728-d0100b6bd8b3 => golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3
