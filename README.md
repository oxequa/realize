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
- Build, Install, Test, Fmt, Generate and Run at the same time
- Live reload on file changes (re-build, re-install...)
- Watch custom paths and specific file extensions
- Support for multiple projects
- Output streams and error logs (Watch them in console or save them on a file)
- Web Panel (Watch all projects, edit the config settings, download each type of log)

#### Installation and usage

- Run this to get/install:

    ```
    $ go get github.com/tockins/realize
    ```

- From project/projects root execute:

    ```
    $ realize add
    ```

    It will create a realize.yaml file if it doesn't exist already and adds the working directory as project.

    Otherwise if a config file already exists it adds the working project to the existing config file.

    The Add command supports the following custom parameters:

    ```
    --name="Project Name"   -> Name, if not specified takes the working directory name
    --path="server"         -> Custom Path, if not specified takes the working directory name    
    --build                 -> Enables the build   
    --test                  -> Enables the tests  
    --no-bin                -> Disables the installation
    --no-run                -> Disables the run
    --no-fmt                -> Disables the fmt (go fmt)
    --no-server             -> Disables the web panel (default port 5001)
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

    Run can also launch a project from its working directory with or without make a config file (--no-config option).
    It supports the following custom parameters:
    
    ```
    --path="server"         -> Custom Path, if not specified takes the working directory name 
    --build                 -> Enables the build   
    --test                  -> Enables the tests   
    --config                -> Take the defined settings if exist a config file  
    --no-bin                -> Disables the installation
    --no-run                -> Disables the run
    --no-fmt                -> Disables the fmt (go fmt)
    --no-server             -> Disables the web panel (port :5000)
    --no-config             -> Doesn't create any configuration files
    --open                  -> Open the web panel in a new browser window 
    --port                  -> Sets the web panel port 
    ```  
    And additional arguments as the "add" command.
    
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
     flimit: 15000                      // Alters the default maximum number of open files
     server:
       status: true                     // Disable/Enable the server
       host: localhost                  // Defines the server address
       port: 5001                       // Defines the server port   
       open: true                       // Opens the server in a new browser tab
     resources:
        logs: logs.log                  // Save the logs on the defined file, disabled if removed
        outputs: outputs.log            // Save the outputs on the defined file, disabled if removed
        errors: errors.log              // Save the errors on the defined file, disabled if removed
     projects:
     - name: realize                    // Project name
       path: .                          // Project path
       fmt: true                        // Disable/Enable go ftm
       test: false                      // Disable/Enable go test
       generate: false                  // Disable/Enable go generate       
       bin: true                        // Disable/Enable go install
       build: false                     // Disable/Enable go build
       run: false                       // Disable/Enable go run
       streams: true                    // Enable/Disable the output streams in cli
       params: []                       // Run the project with defined additional params
       watcher:
         preview: false                 // Enable/Disable the preview of the watched files     
         paths:                         // Paths to watch, sub-paths included 
         - /
         ignore_paths:                  // Paths ignored 
         - vendor
         exts:                          // File extensions to watch
         - .go
         commands:                      // Additional commands to run after and before
         - before: go install           // Defines if after or before
         - before: golint
           watched: true                // Run the command with all watched paths
           foreach: true                // Run the command at each reload
         - after: cd server && gobindata
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
