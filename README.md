# jplot
[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/rs/jplot) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/rs/jplot/master/LICENSE) [![Build Status](https://travis-ci.org/rs/jplot.svg?branch=master)](https://travis-ci.org/rs/jplot)

Jplot tracks expvar-like (JSON) metrics and plot their evolution over time right into your iTerm2 terminal.

![all](doc/demo.gif)

## Install

```
go get -u github.com/rs/jplot
```

This tool does only work with [iTerm2](https://www.iterm2.com).

## Usage

Given the following JSON output:

```
{
    "mem": {
        "Heap": 1234,
        "Sys": 4321,
        "Stack": 203
    },
    "cpu": {
        "STime": 123,
        "UTime":1234
    },
    "Threads": 2
}
```

You can graph the number of thread over time:

```
jplot --url http://:8080/debug/vars Thread
```

![all](doc/single.png)

Or create a graph with both Utime and Stime growth rate on the same axis by using `+` between two field paths:

```
jplot --url http://:8080/debug/vars counter:cpu.STime+counter:cpu.UTime
```

Note: the `counter:` prefix instructs jplot to compute the difference between the values instead of showing their absolute value.

![all](doc/dual.png)


Or create several graphs by providing groups of fields as separate arguments; each argument creates a new graph:

```
jplot --url http://:8080/debug/vars mem.Heap+mem.Sys+mem.Stack counter:cpu.STime+cpu.UTime Threads
```

![all](doc/all.png)

See [gojq](github.com/elgs/gojq) for more details on the JSON query syntax.

### Memstats

Here is an example command to graph a Go program memstats:

```
jplot --url http://:8080/debug/vars \
    memstats.HeapSys+memstats.HeapAlloc+memstats.HeapIdle \
    counter:memstats.TotalAlloc \
    memstats.HeapObjects
```
![all](doc/memstats.png)


