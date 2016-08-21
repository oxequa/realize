## Realize
[![GoDoc](https://img.shields.io/badge/documentation-godoc-blue.svg)](https://godoc.org/github.com/tockins/realize/realize)
[![TeamCity CodeBetter](https://img.shields.io/teamcity/codebetter/bt428.svg?maxAge=2592000?style=flat-square)](https://travis-ci.org/tockins/realize)
[![AUR](https://img.shields.io/aur/license/yaourt.svg?maxAge=2592000?style=flat-square)](https://raw.githubusercontent.com/tockins/realize/v1/LICENSE)
[![](https://img.shields.io/badge/realize-examples-yellow.svg)](https://github.com/tockins/realize-examples)


![Logo](http://i.imgur.com/8nr2s1b.jpg)

A Golang build system with file watchers, output streams and live reload. Run, build and watch file changes with custom paths

![Preview](http://i.imgur.com/XljkxAA.png)

#### Features

- Build, Install, Test and Run at the same time
- Live reload on file changes (re-build, re-install and re-run)
- Watch custom paths
- Watch specific file extensions
- Multiple projects support
- Output streams
- Execution times

#### Installation and usage

- Run this for get/install it:

    ```
    $ go get github.com/tockins/realize
    ```
- From the root of your project/projects:

    ```
    $ realize start 
    ```
    Will create a realize.config.yaml file with a default settings.
    
    You can even pass custom parameters for your project. This is a list of the supported fields:
    
    ```
    --name="Project Name"  -> Name, if not specified takes the working directory name
    --base="server"        -> Base Path, if not specified takes the working directory name    
    --build                -> Go build, if not specified takes "false"    
    --bin                  -> Go intall, if not specified takes "true"    
    --run                  -> Go run, if not specified takes "true"  
    ```
    
    ```
    $ realize start --name="Web Server" --base="server"
    ```
    
- Add another project whenever you want    

    ```
    $ realize add
    ``` 
    Or   
       
    ```
    $ realize add --name="Project Name" --build
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
- [ ] Go test support
- [ ] Cli fast run
- [x] Execution times for build/install 
- [x] Go doc
- [x] Support for server start/stop 
- [x] Stream projects output
- [x] Cli feedback


