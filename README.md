## Realize
[![GoDoc](https://img.shields.io/badge/documentation-godoc-blue.svg)](https://godoc.org/github.com/tockins/realize/realize)
[![TeamCity CodeBetter](https://img.shields.io/teamcity/codebetter/bt428.svg?maxAge=2592000?style=flat-square)](https://travis-ci.org/tockins/realize)
[![AUR](https://img.shields.io/aur/license/yaourt.svg?maxAge=2592000?style=flat-square)](https://raw.githubusercontent.com/tockins/realize/v1/LICENSE)
[![](https://img.shields.io/badge/realize-examples-yellow.svg)](https://github.com/tockins/realize-examples)


![Logo](http://i.imgur.com/8nr2s1b.jpg)

A Golang build system with file watchers and live reload. Run, build and watch file changes with custom paths

![Preview](http://i.imgur.com/5b25ET5.png)

#### Features

- Build, Install and Run in the same time
- Live reload on file changes (re-build, re-install and re-run)
- Watch custom paths
- Watch specific file extensions
- Multiple projects support

#### Installation and usage

- Run this for get/install it:

    ```
    $ go get github.com/tockins/realize
    ```
- From the root of your project/projects:

    ```
    $ realize start 
    ```
    Will create a realize.config.yaml file with a sample project.
    
    You can pass additional parameters for your first project, such as the project name, the main file name and the base path. 
    
    ```
    $ realize start --name="Project Name" --main="main.go" --base="/"
    ```
- Add another project whenever you want    

    ```
    $ realize add --name="Project Name" --main="main.go" --base="/"
    ```
- Remove a project by his name

    ```
    $ realize remove --name="Project Name"
    ```
- Lists all projects

    ```
    $ realize list
    ```
- Build, Run and watch file changes. Realize will re-build and re-run your projects on each changes

    ```
    $ realize run 
    ```

#### Config file example

- For more examples check [Realize Examples](https://github.com/tockins/realize-examples)
     
     ```
    version: "1.0"
    projects:
        - app_name: App One     -> name
          app_path: one         -> root path
          app_main: main.go     -> main file
          app_run: true         -> enable/disable go run (require app_bin)
          app_bin: true         -> enable/disable go install
          app_build: false      -> enable/disable go build
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
          app_main: main.go
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

#### To do
- [x] Cli start, remove, add, list, run
- [x] Remove duplicate projects
- [x] Support for multiple projects
- [x] Watcher files preview
- [x] Support for directories with duplicates names
- [ ] Unit test
- [x] Go doc
- [x] Support for server start/stop 
- [x] Stream projects output
- [x] Cli feedback


