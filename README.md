graylog-project CLI
===================

[![Travis-CI Build Status](https://travis-ci.org/Graylog2/graylog-project-cli.svg?branch=master)](https://travis-ci.org/Graylog2/graylog-project-cli)
[![License](https://img.shields.io/github/license/Graylog2/graylog-project-cli.svg)](https://www.gnu.org/licenses/gpl-3.0.txt)

This is a CLI tool for managing a [graylog-project](https://github.com/Graylog2/graylog-project) setup. Building Graylog involves managing multiple repositories, and this tool helps streamline that process.

## Installation

* Download the binary for your platform from the [releases page](https://github.com/Graylog2/graylog-project-cli/releases)
* Copy the downloaded binary to a directory in your `PATH` (i.e. `cp graylog-project.linux $HOME/bin/graylog-project`
* See [graylog-project instructions](https://github.com/Graylog2/graylog-project/blob/master/README.md) on how to use it

## Configuration

The CLI uses a manifest file to determine which repositories it should be managing. These manifests can be found in the `manifests/` directory. When you run the `checkout` command, for example, the cli will checkout all the repos mentioned in the config at the versions specified.

Additionally, there is the apply-manifest file. This file is used during releases. It contains the new version that all managed repos should be set to after the release. It can also optionally contain the name of a new branch to create across all repos.



## Available Commands

| command name            | description |
|-------------------------|-------------|
| apply-manifest          | Builds a version of Graylog using the components specified in the apply-manifest. Also increments the version after the build and optionally creates a new branch.|
| apply-manifest-generate | Generate an apply-manifest from the given manifest |
| bootstrap               | Clone and setup graylog-project repository |
| checkout                | Update all repos for the given manifest |
| exec                    | Execute arbitrary commands across all modules |
| git                     | Run git commands across all modules |
| github                  | GitHub management |
| graylog-version         | Sets the Graylog version across all modules |
| help                    | Display help on any command |
| idea                    | Commands for setting up IntelliJ IDEA |
| maven-parent            | Show or modify maven parent |
| maven-property          | Gets or sets a maven property |
| npm                     | Run npm commands across all modules |
| npm-clean               | Cleanup npm/yarn related state |
| regenerate              | Regenerate files for the current checkout |
| run                     | Run Graylog server, MongoDB , Elasticsearch and other services |
| status                  | Shows the current version and branch of each managed repo. |
| update                  | Update all repositories for the current manifest |
| version                 | Display version of the Graylog CLI tool |
| yarn                    | Run yarn commands across all modules |



## Contributing

Please read [the contributing instructions](CONTRIBUTING.md) to get started.

## License

_graylog-project-cli_ is released under version 3.0 of the [GNU General Public License](COPYING).
