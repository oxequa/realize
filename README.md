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

#### Features

- Highly customizable
- Config your project Step by Step
- Build, Install, Test, Fmt, Generate and Run at the same time
- Live reload on file changes (re-build, re-install...)
- Watch custom paths and specific file extensions
- Watch by FsNotify or by polling
- Support for multiple projects
- Output streams and error logs (support for save on a file)
- Web Panel (projects list, config settings, logs)

#### Wiki

- [Getting Started](#installation-and-usage)
- [Run cmd](#run) - Run a project
- [Add cmd](#add) - Add a new project
- [Init cmd](#init) - Make a custom config step by step
- [Remove cmd](#remove) - Remove a project 
- [List cmd](#list) - List the projects
- [Config sample](#config-sample)


##### Installation
Run this to get/install:
```
$ go get github.com/tockins/realize
```
#### Commands

##### Run
From project/projects root execute:
```
$ realize run
```

It will create a realize.yaml file if it doesn't exist already, adds the working directory as project and run the pipeline.

The Run command supports the following custom parameters:

```
--path="realize/server"     -> Custom Path, if not specified takes the working directory name    
--build                     -> Enable go build   
--no-run                    -> Disable go run
--no-install                -> Disable go install
--no-config                 -> Ignore an existing config / skip the creation of a new one
--server                    -> Enable the web server
--legacy                    -> Enable legacy watch instead of Fsnotify watch
--generate                  -> Enable go generate
--test                      -> Enable go test
```
Examples:

```
$ realize run
$ realize run --path="mypath"
$ realize run --name="My Project" --build
$ realize run --path="realize" --no-run --no-config
$ realize run --path="/Users/alessio/go/src/github.com/tockins/realize-examples/coin/"
```

If you want, you can specify additional arguments for your project.

 **The additional arguments must go after the params**
 
 **Run can run a project from its working directory without make a config file (--no-config).**

```
$ realize run --path="/print/printer" --no-run yourParams --yourFlags // right
$ realize run yourParams --yourFlags --path="/print/printer" --no-run // wrong
```
##### Add 

Add a project to an existing config file or create a new one without run the pipeline. 

"Add" supports the same parameters of the "Run" command.

```
$ realize add
```

##### Remove
Remove a project by its name
```
$ realize remove --name="myname"
```

##### List
Projects list in cli
```
$ realize list
```

#### Color reference

- Blue: outputs of the project
- Red: errors
- Magenta: times or changed files
- Green: successfully completed action


#### Config sample

- For more examples check [Realize Examples](https://github.com/tockins/realize-examples)

     ```
     settings:
       legacy:                
         status: true           // legacy watch status
         interval: 10s          // polling interval
       resources:               // files names related to streams
         outputs: outputs.log   
         logs: logs.log         
         errors: errors.log
       server:                  
         status: true           // server status         
         open: false            // auto open in browser on start
         host: localhost        // server host  
         port: 5001             // server port
     projects:
     - name: realize    
       path: .                  // project path
       fmt: true                
       generate: false
       test: false
       bin: true
       build: false
       run: false
       params:                  // additional params
       - --myarg
       watcher:
         preview: false         // wached files preview
         paths:                 // paths to watch
         - /
         ignore_paths:          // paths to ignore
         - vendor
         exts:                  // exts to watch
         - .go
         scripts:               // custom commands after/before
         - type: after          // type after/before
           command: go run mycmd after  // command
           path: ""             //  run from a custom path or from the working dir
       streams:                 // enable/disable streams 
         cli_out: true
         file_out: false
         file_log: false
         file_err: false

    ```                    

#### Next features, in progress...

- [ ] Web panel - edit settings (full support)
- [ ] Web panel - logs download
- [ ] Schedule - reload a project after a specific time
- [ ] Easy dependencies - automatically resolve the project dependencies
- [ ] Import license - retrieve the license for each imported library
- [ ] Tests


#### Contacts

- Chat with us [Gitter](https://gitter.im/tockins/realize)

- [Alessio Pracchia](https://www.linkedin.com/in/alessio-pracchia-38a70673)
- [Daniele Conventi](https://www.linkedin.com/in/conventi)
