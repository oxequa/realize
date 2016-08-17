## Realize
######v1.0 Beta

Run, build and watch file changes with custom paths

#### Installation and usage

- Run this for get/install it:
    ```
    $ go get github.com/ghodss/yaml
    ```
- From the root of your project/projects:

    ```
    realize start 
    ```
    Will create a realize.config.yaml file with a sample project.
    
    You can pass additional parameters for your first project, such as the project name, the main file name and the base path. 
    
    ```
    realize start --name="Project Name" --main="main.go" --base="/"
    ```
- Add another project whenever you want    

    ```
    realize add --name="Project Name" --main="main.go" --base="/"
    ```
- Remove a project by his name

    ```
    realize remove --name="Project Name"
    ```
- Lists all projects

    ```
    realize list
    ```
- Build, Run and watch file changes. Realize will re-build and re-run your projects on each changes

    ```
    realize run 
    ```


#### To do
- [x] Command start - default config file
- [x] Command add - new project on the config file 
- [x] Command remove - remove project from the config file
- [x] Command watch - watch changes and rebuild 
- [x] Command list - print projects list
- [x] Remove duplicate projects
- [x] Support for multiples projects
- [x] Watcher files preview
- [x] Support for directories with duplicates names
- [ ] Unit test
- [ ] Documentation
- [x] Support for server start/stop 
- [x] Cli feedback


