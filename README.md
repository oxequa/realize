# Realize

[![GoDoc](https://img.shields.io/badge/documentation-godoc-blue.svg?style=flat-square)](https://godoc.org/github.com/tockins/realize)
[![Travis](https://img.shields.io/travis/tockins/realize.svg?style=flat-square)](https://travis-ci.org/tockins/realize)
[![AUR](https://img.shields.io/aur/license/yaourt.svg?style=flat-square)](https://raw.githubusercontent.com/tockins/realize/v1/LICENSE)
[![](https://img.shields.io/badge/realize-examples-yellow.svg?style=flat-square)](https://github.com/tockins/realize-examples)
[![Gitter](https://img.shields.io/gitter/room/tockins/realize.svg?style=flat-square)](https://gitter.im/tockins/realize?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![Go Report Card](https://goreportcard.com/badge/github.com/tockins/realize?style=flat-square)](https://goreportcard.com/report/github.com/tockins/realize)
<!--[![OpenCollective](https://opencollective.com/realize/backers/badge.svg?style=flat-square)](#backers) 
[![OpenCollective](https://opencollective.com/realize/sponsors/badge.svg?style=flat-square)](#sponsors)-->

<br>

<div align="center">
    <img src="https://i.imgur.com/7XiZyR1.png" width="125px">
    <h1 style="">The #1 Golang Task Runner</h1>
	<p style="font-size: 18px; font-weight: 400;">Enhance your workflow by automating the most common tasks and using the best performing Golang live reloading.</p>
</div>

<br>

<p align="center">
    <img src="http://i.imgur.com/KpMSLnE.png">
</p>


## Content

### - ⭐️ [Top Features](#top-features)
### - 💃🏻 [Get started](#get-started)
### - 📄 [Config sample](#config-sample)
### - 📚 [Commands List](#commands-list)
### - 🛠 [Support and Suggestions](#support-and-suggestions)
### - 😎 [Backers and Sponsors](#backers)

## Top Features

- High performance Live Reload.
- Manage multiple projects at the same time.
- Watch by custom extensions and paths.
- All Go commands supported.
- Switch between different Go builds.
- Custom env variables for project.
- Execute custom commands before and after a file changes or globally.
- Export logs and errors to an external file.
- Step-by-step project initialization.
- Redesigned panel that displays build errors, console outputs and warnings.
- Any suggestion? [Suggest an amazing feature! 🕺🏻](https://github.com/tockins/realize/issues/new)

## Get started
	$ go get github.com/tockins/realize

## Commands List

### Run Command
From **project/projects** root execute:
    
    $ realize start
    
    
It will create a **realize.yaml** file if doesn't already exist, add the working directory as project and run your workflow.
    
***start*** command supports the following custom parameters:

    --name="name"               -> Run by name on existing configuration
    --path="realize/server"     -> Custom Path (if not specified takes the working directory name)
    --generate                  -> Enable go generate
    --fmt                       -> Enable go fmt
    --test                      -> Enable go test
    --vet                       -> Enable go vet
    --install                   -> Enable go install
    --build                     -> Enable go build   
    --run                       -> Enable go run
    --server                    -> Enable the web server
    --no-config                 -> Ignore an existing config / skip the creation of a new one

Some examples:
    
    $ realize start
    $ realize start --path="mypath"
    $ realize start --name="realize" --build
    $ realize start --path="realize" --run --no-config
    $ realize start --install --test --fmt --no-config
    $ realize start --path="/Users/username/go/src/github.com/tockins/realize-examples/coin/"
    
If you want, you can specify additional arguments for your project:

	✅ $ realize start --path="/print/printer" --run yourParams --yourFlags // right
    ❌ $ realize start yourParams --yourFlags --path="/print/printer" --run // wrong    

⚠️ The additional arguments **must go after** the params:
<br>
💡 The ***start*** command can be used with a project from its working directory without make a config file (*--no-config*).

### Add Command
Add a project to an existing config file or create a new one.

    $ realize add
💡 ***add*** supports the same parameters as ***start*** command.
### Init Command
This command allows you to create a custom configuration step-by-step. 

    $ realize init
    
💡 ***init*** is the only command that supports a complete customization of all supported options.
### Remove Command
Remove a project by its name

    $ realize remove --name="myname"


## Color reference
💙 BLUE: Outputs of the project.<br>
💔 RED: Errors.<br>
💜 PURPLE: Times or changed files.<br>
💚 GREEN: Successfully completed action.<br>


## Config sample
    
For more examples check: [Realize Examples](https://github.com/tockins/realize-examples)
    
    settings:
        legacy:
            force: true             // force polling watcher instead fsnotifiy
            interval: 100ms         // polling interval
        resources:                  // files names
            outputs: outputs.log
            logs: logs.log
            errors: errors.log
    server:
        status: false               // server status
        open: false                 // open browser at start
        host: localhost             // server host
        port: 5001                  // server port
    schema:
    - name: coin
      path: coin              // project path
      environment:            // env variables available at startup
            test: test
            myvar: value
      commands:               // go commands supported
        vet:
            status: true
        fmt:
            status: true
            args:
            - -s
            - -w
        test:
            status: true
            method: gb test    // support differents build tool
        generate:
            status: true
        install:
            status: true
        build:
            status: false
            method: gb build    // support differents build tool
            args:               // additional params for the command
            - -race
        run: true
      args:                     // arguments to pass at the project
      - --myarg
      watcher:
          preview: false         // watched files preview
          paths:                 // watched paths
          - /
          ignore_paths:          // ignored paths
          - vendor
          extensions:                  // watched extensions
          - go
          - html
          scripts:               // custom scripts
          - type: before         // type (after/before)
            command: ./ls -l     // command
            changed: true        // relaunch when a file change
            startup: true        // launch at start
          - type: after
            command: ./ls
            changed: true
          errorOutputPattern: mypattern   //custom error pattern
         
## Support and Suggestions
💬 Chat with us [Gitter](https://gitter.im/tockins/realize)<br>
⭐️ Suggest a new [Feature](https://github.com/tockins/realize/issues/new)

## Backers

Support us with a monthly donation and help us continue our activities. [[Become a backer](https://opencollective.com/realize#backer)]

<a href="https://opencollective.com/realize" target="_blank"><img src="https://opencollective.com/realize/backer/0/avatar.svg"></a>
<a href="https://opencollective.com/realize" target="_blank"><img src="https://opencollective.com/realize/backer/1/avatar.svg"></a>
<a href="https://opencollective.com/realize" target="_blank"><img src="https://opencollective.com/realize/backer/2/avatar.svg"></a>
<a href="https://opencollective.com/realize" target="_blank"><img src="https://opencollective.com/realize/backer/3/avatar.svg"></a>

## Sponsors

Become a sponsor and get your logo here! [[Become a sponsor](https://opencollective.com/realize#sponsor)]


