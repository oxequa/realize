## Realize

[![GoDoc](https://img.shields.io/badge/documentation-godoc-blue.svg)](https://godoc.org/github.com/tockins/realize/realize)
[![TeamCity CodeBetter](https://travis-ci.org/tockins/realize.svg?branch=v1)](https://travis-ci.org/tockins/realize)
[![AUR](https://img.shields.io/aur/license/yaourt.svg?maxAge=2592000?style=flat-square)](https://raw.githubusercontent.com/tockins/realize/v1/LICENSE)
[![](https://img.shields.io/badge/realize-examples-yellow.svg)](https://github.com/tockins/realize-examples)
[![Join the chat at https://gitter.im/tockins/realize](https://badges.gitter.im/tockins/realize.svg)](https://gitter.im/tockins/realize?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![Go Report Card](https://goreportcard.com/badge/github.com/tockins/realize)](https://goreportcard.com/report/github.com/tockins/realize)


![Logo](http://i.imgur.com/8nr2s1b.jpg)

A Go build system with file watchers, output streams and live reload. Run, build and watch file changes with custom paths

![Preview](http://i.imgur.com/dJbNZjt.gif)

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
    --name="Project Name"   -> Name, if not specified takes the working directory name
    --path="server"         -> Custom Path, if not specified takes the working directory name    
    --build                 -> Enables the build   
    --test                  -> Enables the tests   
    --no-bin                -> Disables the installation
    --no-run                -> Disables the run
    --no-fmt                -> Disables the fmt (go fmt)
    --no-server             -> Disables the web panel (port :5000)
    ```
    Examples:

    ```
    $ realize add

    $ realize add --path="mypath"

    $ realize add --name="My Project" --build

    $ realize add --name="My Project" --path="/projects/package" --build

    $ realize add --name="My Project" --path="projects/package" --build --no-run
    
    $ realize add --path="/Users/alessio/go/src/github.com/tockins/realize-examples/coin/"
    ```

    If you want, you can specify additional arguments for your project.

     **The additional arguments must go after the options of "Realize"**

    ```
    $ realize add --path="/print/printer" --no-run yourParams --yourFlags // correct

    $ realize add yourParams --yourFlags --path="/print/printer" --no-run // wrong
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
    --path="server"         -> Custom Path, if not specified takes the working directory name 
    --build                 -> Enables the build   
    --test                  -> Enables the tests   
    --config                -> Take the defined settings if exist a config file  
    --no-bin                -> Disables the installation
    --no-run                -> Disables the run
    --no-fmt                -> Disables the fmt (go fmt)
    --no-server             -> Disables the web panel (port :5000)
    ```  

    The "fast" command supports addittional arguments as the "add" command.

    ```
    $ realize fast --no-run yourParams --yourFlags // correct

    $ realize fast yourParams --yourFlags --no-run // wrong
    
    $ realize fast --path="/Users/alessio/go/src/github.com/tockins/realize-examples/coin/"
    ```  


#### Color reference

- Blue: outputs of the project
- Red: errors
- Magenta: times or changed files
- Green: successfully completed action


#### Config file example

- For more examples check [Realize Examples](https://github.com/tockins/realize-examples)

     ```
    projects:
        - app_name: App One     -> name
          app_path: one         -> root path
          app_run: true         -> enable/disable go run (require app_bin)
          app_bin: true         -> enable/disable go install
          app_build: false      -> enable/disable go build
          app_fmt: true         -> enable/disable go fmt
          app_test: true        -> enable/disable go test
          app_params:           -> the project will be launched with these parameters
            - --flag1
            - param1
          app_watcher:
            preview: true       -> prints the observed files on startup
            paths:              -> paths to observe for live reload
            - /
            ignore_paths:       -> paths to ignore
            - vendor
            - bin
            exts:               -> file extensions to observe for live reload
            - .go
            output:             -> enable/disable the output destinations 
                  cli: true         -> cli output
                  file: true        -> creates an output file inside the project 
            
    ```                    

#### Next release

##### Milestone 1.1
- [ ] Windows support - **Moved to 1.2**
- [x] Custom paths for the commands fast/add
- [x] Save output on a file
- [x] Enable the fields Before/After
- [x] Web panel


##### Milestone 1.2
- [ ] Windows support 
- [ ] Go generate support 


#### Contacts

- Chat with us [Gitter](https://gitter.im/tockins/realize)

- [Alessio Pracchia](https://www.linkedin.com/in/alessio-pracchia-38a70673)
- [Daniele Conventi](https://www.linkedin.com/in/conventi)
