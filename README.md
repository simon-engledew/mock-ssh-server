# mock-ssh-server

Current status: early proof of concept

Usage: `mock-ssh-server <starlark script>`

This will spin up an SSH server on port 2222 that runs the specified [Starlark](https://chromium.googlesource.com/external/github.com/google/starlark-go/+/HEAD/doc/spec.md) script whenever anyone connects.

Starlark is a Python-like language used by Bazel for configuration.

`mock-ssh-server` currently exposes four built-in functions to the script:

| Name | Description |
| -- | -- | 
| writeline(string) | Send text and a newline to the connected client |
| write(string) | Send text to the connected client |
| readline(string) | Read a whole line from the client |
| matchline(string: regexp) | Keep reading a line until the input matches a pattern, return the match and any capturing groups |

Simple example:

```Starlark
writeline("What is your name?")
name = readline()
writeline("Hello " + name + "!")
```

<img width="632" alt="image" src="https://user-images.githubusercontent.com/14410/85181005-cfc98200-b27c-11ea-9b46-c45042324570.png">

Try it out with:

```bash
go get -u github.com/simon-engledew/mock-ssh-server

mock-ssh-server examples/helloworld.star 
```
