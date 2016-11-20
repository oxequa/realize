## Realize

[![GoDoc](https://img.shields.io/badge/documentation-godoc-blue.svg)](https://godoc.org/github.com/tockins/realize)
[![TeamCity CodeBetter](https://travis-ci.org/tockins/realize.svg?branch=v1)](https://travis-ci.org/tockins/realize)
[![AUR](https://img.shields.io/aur/license/yaourt.svg?maxAge=2592000?style=flat-square)](https://raw.githubusercontent.com/tockins/realize/v1/LICENSE)
[![](https://img.shields.io/badge/realize-examples-yellow.svg)](https://github.com/tockins/realize-examples)
[![Join the chat at https://gitter.im/tockins/realize](https://badges.gitter.im/tockins/realize.svg)](https://gitter.im/tockins/realize?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![Go Report Card](https://goreportcard.com/badge/github.com/tockins/realize)](https://goreportcard.com/report/github.com/tockins/realize)


![Logo](http://i.imgur.com/8nr2s1b.jpg)

A Go build system with file watchers, output streams and live reload. Run, build and watch file changes with custom paths

![Preview](http://i.imgur.com/dJbNZjt.gif)

#### What's new

##### v1.2
- [x] Windows support
- [x] Go generate support
- [x] Bugs fix
- [x] Web panel errors log improved
- [x] Refactoring
- [x] Web panel edit settings, partial

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
    --open                  -> Open the web panel in a new browser window
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

    Run can also launch a project from its working directory without a config file. It supports the following custom parameters:

    ```
    --path="server"         -> Custom Path, if not specified takes the working directory name 
    --build                 -> Enables the build   
    --test                  -> Enables the tests   
    --config                -> Take the defined settings if exist a config file  
    --no-bin                -> Disables the installation
    --no-run                -> Disables the run
    --no-fmt                -> Disables the fmt (go fmt)
    --no-server             -> Disables the web panel (port :5000)
    --open                  -> Open the web panel in a new browser window 
    ```  
    And addittional arguments as the "add" command.
    
    ```
    $ realize run --no-run yourParams --yourFlags // correct

    $ realize run yourParams --yourFlags --no-run // wrong
    
    $ realize run --path="/Users/alessio/go/src/github.com/tockins/realize-examples/coin/"
    ```  

#### Color reference

- Blue: outputs of the project
- Red: errors
- Magenta: times or changed files
- Green: successfully completed action


#### Config file example

- For more examples check [Realize Examples](https://github.com/tockins/realize-examples)

     ```
     settings:
       resources:
         output: outputs.log  // name of the output file
         log: logs.log        // name of the log file (errors included)
       server:
         enable: true         // enables the web server 
         open: false          // opens the web server in a new tab
         host: localhost      // web server host
         port: 5000           // wev server port
     projects:
     - name: printer          // project name
       path: /                // project path
       run: true              // enables go run  (require bin)
       bin: true              // enables go install
       generate: false        // enables go generate
       build: false           // enables go build
       fmt: true              // enables go fmt
       test: false            // enables go test   
       params: []             // array of additionals params. the project will be launched with these parameters   
       watcher:
         before: []           // custom commands launched before the execution of the project 
         after: []            // custom commands launched after the execution of the project 
         paths:               // paths to observe for live reload
         - /
         ignore_paths:        // paths to ignore
         - vendor
         exts:                // file extensions to observe for live reload
         - .go
         preview: true        // prints the observed files on startup
       cli:                   
         streams: true        // prints the output streams of the project in the cli 
       file:
         streams: false       // saves the output stream of the project in a file
         logs: false          // saves the logs of the project in a file
         errors: false        // saves the errors of the project in a file
    ```                    

#### Next release

##### v1.3
- [ ] Web panel edit settings, full support
- [ ] Tests

#### Contacts

- Chat with us [Gitter](https://gitter.im/tockins/realize)

- [Alessio Pracchia](https://www.linkedin.com/in/alessio-pracchia-38a70673)
- [Daniele Conventi](https://www.linkedin.com/in/conventi)
