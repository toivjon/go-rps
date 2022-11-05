# go-rps

[![Lint](https://github.com/toivjon/go-rps/actions/workflows/lint.yml/badge.svg?branch=main)](https://github.com/toivjon/go-rps/actions/workflows/lint.yml)
[![Unit Test](https://github.com/toivjon/go-rps/actions/workflows/unit-test.yml/badge.svg?branch=main)](https://github.com/toivjon/go-rps/actions/workflows/unit-test.yml)
[![Build](https://github.com/toivjon/go-rps/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/toivjon/go-rps/actions/workflows/build.yml)

A text-based rock-paper-scissors game with a client-server architecture.

## Features

This section lists the major and minor features of the solution.

- Client-server communication is based on the TCP sockets.
- TCP socket connection configuration can be given as command line arguments.
- Client allows user to provide a player name.

## Build

Use the following scripts to build the applications.

| OS      | Script              |
| ------- | ------------------- |
| Windows | ./scripts/build.bat |
| Linux   | ./scripts/build.sh  |

Successful build binaries will be added into the ./bin folder.

## Unit Test

Use the following scripts to run unit tests for the application.

| OS      | Script                  |
| ------- | ----------------------- |
| Windows | ./scripts/unit-test.bat |
| Linux   | ./scripts/unit-test.sh  |

These scripts will also check that code coverage is within the threshold.

## Game Sequence

This section describes how the gaming sequence works.

```mermaid
sequenceDiagram
  participant c as Client
  participant s as Server

  activate c
  c->>s: Connect [name]
  activate s
  s->>s: Register Client [name]
  s-)c: Connected
  c->>c: Wait for server
  deactivate c
  s->>s: Wait for an another client
  s-)c: Start [opponent-name]
  activate c
  c-)s: Select [selection]
  s->>s: Wait for another client's selection
  s->>s: Resolve result
  s-)c: Result [opponent-selection]
  deactivate c
  deactivate s
```
