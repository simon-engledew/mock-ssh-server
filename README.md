# mock-ssh-server

Current status: early proof of concept

Usage: `mock-ssh-server <starlark script>`

This will spin up an SSH server on port 2222 that runs the specified [Starlark](https://chromium.googlesource.com/external/github.com/google/starlark-go/+/HEAD/doc/spec.md) script whenever anyone connects.

Starlark is a Python-like language used by Bazel for configuration.

mock-ssh-server currently exposes four built-in functions to the script:

| Name | Description |
| -- | -- | 
| writeline(string) | Send a line to the connected client |
| write(string) | Send a string to the connected client |
| readline(string) | Read a whole line from the client |
| expect(string: regexp) | Keep reading until the input matches a pattern, return the match and any capturing groups |

See `example.star` for a simple script.