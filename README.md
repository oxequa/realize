## Realize

[![GoDoc](https://img.shields.io/badge/documentation-godoc-blue.svg)](https://godoc.org/github.com/tockins/realize/realize)
[![TeamCity CodeBetter](https://img.shields.io/teamcity/codebetter/bt428.svg?maxAge=2592000?style=flat-square)](https://travis-ci.org/tockins/realize)
[![AUR](https://img.shields.io/aur/license/yaourt.svg?maxAge=2592000?style=flat-square)](https://raw.githubusercontent.com/tockins/realize/v1/LICENSE)
[![](https://img.shields.io/badge/realize-examples-yellow.svg)](https://github.com/tockins/realize-examples)
[![Join the chat at https://gitter.im/tockins/realize](https://badges.gitter.im/tockins/realize.svg)](https://gitter.im/tockins/realize?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)


![Logo](http://i.imgur.com/8nr2s1b.jpg)

A Golang build system with file watchers, output streams and live reload. Run, build and watch file changes with custom paths

![Preview](http://i.imgur.com/GooHBej.png)

#### Features

- Build, Install, Test, Fmt and Run at the same time
- Live reload on file changes (re-build, re-install and re-run)
- Watch custom paths
- Watch specific file extensions
- Multiple projects support
- Output streams
- Execution times
- Highly customizable
- Fast run

#### Installation and usage

- Run this to get/install:

    ```
    $ go get github.com/tockins/realize
    ```

- From project/projects root execute:

    ```
    $ realize add
    ```

    It will create a realize.config.yaml file if it doesn't exist already and adds the working directory as the project.

    Otherwise if a config file already exists it adds another project to the existing config file.

    The add command supports the following custom parameters:

    ```
    --name="Project Name"  -> Name, if not specified takes the working directory name
    --path="server"        -> Base Path, if not specified takes the working directory name    
    --build                -> Enables the build   
    --nobin                -> Disables the installation
    --norun                -> Disables the run
    --nofmt                -> Disables the fmt (go fmt)
    ```
    Examples:

    ```
    $ realize add
    ```
    ```
    $ realize add --path="mypath"
    ```   
    ```
    $ realize add --name="My Project" --build
    ```    
    ```
    $ realize add --name="My Project" --path="/projects/package" --build
    ```    
    ```
    $ realize add --name="My Project" --path="projects/package" --build --norun
    ```
- Remove a project by its name

    ```
    $ realize remove --name="Project Name"
    ```
- Lists all projects

    ```
    $ realize list
    ```
- Build, Run and watch file changes. Realize will re-build and re-run your projects on each change.

    ```
    $ realize run
    ```

    Fast run launches a project from its working directory without a config file

    ```
    $ realize fast
    ```

     The fast command supports the following custom parameters:

    ```
    --build                -> Enables the build   
    --nobin                -> Disables the installation
    --norun                -> Disables the run
    --nofmt                -> Disables the fmt (go fmt)
    --config               -> Take the defined settings if exist a config file  
    ```    

#### Config file example

- For more examples check [Realize Examples](https://github.com/tockins/realize-examples)

     ```
    version: "1.0"
    projects:
        - app_name: App One     -> name
          app_path: one         -> root path
          app_run: true         -> enable/disable go run (require app_bin)
          app_bin: true         -> enable/disable go install
          app_build: false      -> enable/disable go build
          app_fmt: true         -> enable/disable go fmt
          app_watcher:
            preview: true       -> prints the observed files on startup
            paths:              -> paths to observe for live reload
            - /
            ignore_paths:       -> paths to ignore
            - vendor
            - bin
            exts:               -> file extensions to observe for live reload
            - .go
        - app_name: App Two     -> another project
          app_path: two
          app_run: true
          app_build: true
          app_bin: true
          app_watcher:
            paths:
            - /
            ignore_paths:
            - vendor
            - bin
            exts:
            - .go
    ```                    

#### Next releases

#####Milestone 1.0

- [x] Cli start, remove, add, list, run
- [x] Remove duplicate projects
- [x] Support for multiple projects
- [x] Watcher files preview
- [x] Support for directories with duplicates names
- [ ] Go test support
- [x] Go fmt support
- [x] Cli fast run
- [x] Execution times for build/install
- [x] Go doc
- [x] Support for server start/stop
- [x] Stream projects output
- [x] Cli feedback

##### Milestone 1.1
- [ ] Test under windows
- [ ] Unit test
- [ ] Custom path support on commands
- [ ] Output files support


#### Contacts

- Chat with us [Gitter](https://gitter.im/tockins/realize)

- [Alessio Pracchia](https://www.linkedin.com/in/alessio-pracchia-38a70673)
- [Daniele Conventi](https://www.linkedin.com/in/daniele-conventi-b419b0a4)
