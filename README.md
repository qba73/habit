# Habit tracker

[![Go Reference](https://pkg.go.dev/badge/github.com/qba73/habit.svg)](https://pkg.go.dev/github.com/qba73/habit)
![Go](https://github.com/qba73/habit/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/qba73/habit)](https://goreportcard.com/report/github.com/qba73/habit)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/qba73/http?logo=go)
![GitHub](https://img.shields.io/github/license/qba73/habit)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/qba73/habit)
[![Platforms](https://img.shields.io/badge/platforms-linux|windows|macos-inactive.svg)]()


This is a Go program that helps with establishing new, and tracking existing habits.

## Description

When we're trying to establish a new habit (studying, running, rowing for example), it can be difficult to maintain focus and motivation. Studies suggest that it takes many weeks of regularly performing a new habit before it becomes natural and automatic. So, if you can motivate yourself to do the new habit every day for a month or two, that'll get you well on the way.

One thing that can help is to _track_ your performance of the habit. Suppose you decide that you're going to spend at least 15 minutes every day studying or writing Go programs or going for a morning jog. You could draw 30 boxes on a piece of paper, one for each of the next 30 days, and check off each box as you do that day's practice.

This simple idea can be surprisingly effective, because we don't like to break a _streak_. If you've successfully done the habit every day for 29 days, there's a strong incentive not to break that run of success. Life has a way of coming at you, and you might well need that extra motivation at some point. Not today, not tomorrow, but soon: probably just around the time the novelty wears off.

The project is a Go package and accompanying command-line tool called ```habctl``` (habit control) that will help users track and establish new habits, by reporting their current streak.

## Using habctl

For example, if you decide you want to build the habit of jogging every day, you might tell the habit tool about it like this:

**`habctl jog`**

```
Good luck with your new habit 'jog'! Don't forget to do it again
tomorrow.
```

If you want to track multiple daily habits you tell the habit tool to track your new activity, for example:

**`habctl study`**

```
Good luck with your new habit 'study'! Don't forget to do it again
tomorrow.
```

As the days go by, you might record each daily practice like this:

**`habctl jog`**

```
Nice work: you've done the habit 'jog' for 18 days in a row now.
Keep it up!
```

If you happen to miss a couple of days, that's all right:

**`habctl jog`**

```
You last did the habit 'jog' 3 days ago, so you're starting a new
streak today. Good luck!
```

If you just want to check how you're doing, you could run:

**`habctl`**

```
You're currently on a 1-day streak for 'jog'. Stick to it!
You're currently on a 1-day streak for 'study'. Stick to it!
```

or, if you keep streeks not broken:

```
You're currently on a 4-day streak for 'jog'. Stick to it!
You're currently on a 17-day streak for 'study'. Stick to it!
```

Maybe the news won't be quite so good:

```
You're currently on a 1-day streak for 'hike'. Stick to it!
It's been 10 days since you did 'jog'. It's ok, life happens. Get back on that horse today!
It's been 17 days since you did 'study'. It's ok, life happens. Get back on that horse today!
```

# Installation

## Storing data

`habctl` persists data in a file storage. If you want to configure `habctl` where to locate the file store, export the ENV variable `$XDG_DATA_HOME`. If the env var is not exported `habctl` will create file store in user's `$HOME` directory.

## Using `go install`

```bash
go install github.com/qba73/habit/cmd/habctl@latest
go: downloading github.com/qba73/habit v0.0.0-20230121004648-a82a2385e324
```

Verify installation:

```bash
habctl
You are not tracking any habit yet.
```

Start tracking a habit:

```bash
habctl jog
Good luck with your new habit 'jog'. Don't forget to do it tomorrow.
```

Check tracked habits:

```bash
habctl
You're currently on a 1-day streak for 'jog'. Stick to it!
```

## Building from source

Clone this repository to your local machine:

```bash
git clone git@github.com:qba73/habit.git
cd habit
```

Build `habclt` binary:

```bash
go build -o habctl ./cmd/habctl/main.go
```

Run `habctl`:

```bash
./habctl
```

## Development

Use following make targets for developemnt and testing:

```bash
$ make

Usage:
  help                      Show help message
  dox                       Run tests with gotestdox
  test                      Run tests
  vet                       Run go vet
  check                     Run staticcheck analyzer
  cover                     Run unit tests and generate test coverage report
  tidy                      Run go mod tidy
```

## Credits

This is an educational Go project intended for students at the [Bitfield Institute of Technology](https://bitfieldconsulting.com/golang/bit).
