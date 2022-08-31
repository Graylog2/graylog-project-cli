module github.com/Graylog2/graylog-project-cli

go 1.18

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/fatih/color v1.13.0
	github.com/google/go-github/v27 v27.0.6
	github.com/hashicorp/go-version v1.6.0
	github.com/imdario/mergo v0.3.13
	github.com/manifoldco/promptui v0.9.0
	github.com/mattn/go-isatty v0.0.16
	github.com/pelletier/go-toml/v2 v2.0.5
	github.com/pkg/errors v0.9.1
	github.com/samber/lo v1.27.1
	github.com/spf13/cobra v1.5.0
	github.com/spf13/viper v1.12.0
	github.com/yuin/goldmark v1.4.13
	golang.org/x/oauth2 v0.0.0-20220822191816-0ebed06d0094
	golang.org/x/text v0.3.7
)

require (
	github.com/chzyer/readline v1.5.1 // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/spf13/afero v1.9.2 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.4.1 // indirect
	golang.org/x/crypto v0.0.0-20220829220503-c86fa9a7ed90 // indirect
	golang.org/x/exp v0.0.0-20220827204233-334a2380cb91 // indirect
	golang.org/x/net v0.0.0-20220826154423-83b083e8dc8b // indirect
	golang.org/x/sys v0.0.0-20220829200755-d48e67d00261 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Workaround for https://github.com/golang/go/issues/30831
replace github.com/golang/lint v0.0.0-20190313153728-d0100b6bd8b3 => golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3
