# go-rps

[![Lint](https://github.com/toivjon/go-rps/actions/workflows/lint.yml/badge.svg?branch=main)](https://github.com/toivjon/go-rps/actions/workflows/lint.yml)
[![Unit Test](https://github.com/toivjon/go-rps/actions/workflows/unit-test.yml/badge.svg?branch=main)](https://github.com/toivjon/go-rps/actions/workflows/unit-test.yml)
[![Build](https://github.com/toivjon/go-rps/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/toivjon/go-rps/actions/workflows/build.yml)
[![System Test](https://github.com/toivjon/go-rps/actions/workflows/system-test.yml/badge.svg)](https://github.com/toivjon/go-rps/actions/workflows/system-test.yml)

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

## System Test

Use the following scripts to run system tests for the application.

| OS      | Script                    |
| ------- | ------------------------- |
| Windows | ./scripts/system-test.bat |
| Linux   | ./scripts/system-test.sh  |

## Messages

This section contains a description about the message types between the client and the server.

| Message | Origin | Arguments       | Description                                          |
| ------- | ------ | --------------- | ---------------------------------------------------- |
| JOIN    | client | player's name   | The initial message from client to server.           |
| START   | server | opponent's name | Server formed a game session with two clients.       |
| SELECT  | client | round selection | Player has made a rock, paper or scissors selection. |
| RESULT  | server | round results   | Server has resolved game session result.             |

## Game Sequence

This section describes how the gaming sequence works.

```mermaid
sequenceDiagram
  participant c as Client
  participant s as Server

  activate c
  c->>s: JOIN
  activate s
  s->>s: Register Player
  s-)c: 
  c->>c: Wait for server
  deactivate c
  s->>s: Wait for an another player
  s-)c: START
  activate c
  c-)s: SELECT
  s->>s: Wait for another player's selection
  s->>s: Resolve result
  s-)c: RESULT
  deactivate c
  deactivate s
```
