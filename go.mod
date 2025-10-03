module github.com/Graylog2/graylog-project-cli

go 1.24

require (
	github.com/Masterminds/sprig/v3 v3.3.0
	github.com/fatih/color v1.18.0
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/google/go-github/v66 v66.0.0
	github.com/google/renameio/v2 v2.0.0
	github.com/hashicorp/go-version v1.7.0
	github.com/imdario/mergo v0.3.16
	github.com/k0kubun/pp/v3 v3.4.1
	github.com/manifoldco/promptui v0.9.0
	github.com/mattn/go-isatty v0.0.20
	github.com/pelletier/go-toml/v2 v2.2.4
	github.com/pkg/errors v0.9.1
	github.com/samber/lo v1.51.0
	github.com/schollz/progressbar/v3 v3.18.0
	github.com/spf13/cobra v1.9.1
	github.com/spf13/viper v1.20.1
	github.com/stretchr/testify v1.10.0
	github.com/subosito/gotenv v1.6.0
	github.com/yuin/goldmark v1.7.12
	golang.org/x/exp v0.0.0-20250606033433-dcc06ee1d476
	golang.org/x/oauth2 v0.30.0
	golang.org/x/text v0.26.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	dario.cat/mergo v1.0.2 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.3.1 // indirect
	github.com/chzyer/readline v1.5.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/sagikazarmark/locafero v0.9.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.14.0 // indirect
	github.com/spf13/cast v1.9.2 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/term v0.32.0 // indirect
)

// Workaround for https://github.com/golang/go/issues/30831
replace github.com/golang/lint v0.0.0-20190313153728-d0100b6bd8b3 => golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3
