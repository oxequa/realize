## Realize

[![AUR](https://img.shields.io/aur/license/yaourt.svg?maxAge=2592000?style=flat-square)](https://raw.githubusercontent.com/tockins/realize/v1/LICENSE)
[![Build Status](http://img.shields.io/travis/labstack/echo.svg?style=flat-square)](https://travis-ci.org/tockins/realize)

A Golang build system with file watchers and live reload. Run, build and watch file changes with custom paths

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

    version: "1.0"
    projects:
        - app_name: App One
          app_path: one
          app_main: main.go
          app_run: true
          app_bin: true
          app_watcher:
            paths:
            - /
            ignore_paths:
            - vendor
            - bin
            exts:
            - .go
        - app_name: App Two
          app_path: two
          app_main: main.go
          app_run: true
          app_build: true
          app_bin: true
          app_watcher:
            preview: true
            paths:
            - /
            ignore_paths:
            - vendor
            - bin
            exts:
            - .go

#### To do
- [x] Cli start, remove, add, list, run
- [x] Remove duplicate projects
- [x] Support for multiple projects
- [x] Watcher files preview
- [x] Support for directories with duplicates names
- [ ] Unit test
- [ ] Go doc
- [x] Support for server start/stop 
- [x] Stream projects output
- [x] Cli feedback


