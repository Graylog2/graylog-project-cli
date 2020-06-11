module github.com/Graylog2/graylog-project-cli

go 1.12

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/fatih/color v1.7.0
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/google/go-github/v27 v27.0.1
	github.com/hashicorp/go-version v1.2.0
	github.com/imdario/mergo v0.3.7
	github.com/mattn/go-colorable v0.1.2 // indirect
	github.com/mattn/go-isatty v0.0.8
	github.com/mitchellh/mapstructure v1.3.0 // indirect
	github.com/pelletier/go-toml v1.7.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.6.3
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4 // indirect
	golang.org/x/net v0.0.0-20190628185345-da137c7871d7 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sys v0.0.0-20200430082407-1f5687305801 // indirect
	google.golang.org/appengine v1.6.1 // indirect
	gopkg.in/ini.v1 v1.55.0 // indirect
)

// Workaround for https://github.com/golang/go/issues/30831
replace github.com/golang/lint v0.0.0-20190313153728-d0100b6bd8b3 => golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3
