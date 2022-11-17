# dockerlogs

## Overview

Quickly view the logs of multiple containers, with logs colorized by container.  Bolded lines are from stderr.

## Installation
Binary downloads of the dockerlogs utility can be found on the [releases page](https://github.com/ruckc/dockerlogs/releases/latest).

Go users can use
```sh
go install github.com/ruckc/dockerlogs@v0
```

## Usage 

dockerlogs will by default tail all containers, and their entire log.

```sh
dockerlogs
```

Only tail the api and database containers, last 10 lines.
```sh
dockerlogs -t 10 api database
```

## Options

Run `dockerlogs -h` for usage

```sh
$ dockerlogs -h

Usage:
  dockerlogs [OPTIONS] [containers...]

tails multiple containers concurrently

Application Options:
  -t, --tail=       Number of lines to show from the end of the logs (default: all)

Help Options:
  -h, --help        Show this help message
```