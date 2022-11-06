![Go](https://github.com/qba73/habit/workflows/Go/badge.svg)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/qba73/http?logo=go)
![GitHub](https://img.shields.io/github/license/qba73/habit)

# Habit tracker

This is a Go program that helps with establishing new, and tracking existing habits.

## Description

When we're trying to establish a new habit (studying, running, rowing for example), it can be difficult to maintain focus and motivation. Studies suggest that it takes many weeks of regularly performing a new habit before it becomes natural and automatic. So, if you can motivate yourself to do the new habit every day for a month or two, that'll get you well on the way.

One thing that can help is to _track_ your performance of the habit. Suppose you decide that you're going to spend at least 15 minutes every day studying or writing Go programs or going for a morning jog. You could draw 30 boxes on a piece of paper, one for each of the next 30 days, and check off each box as you do that day's practice.

This simple idea can be surprisingly effective, because we don't like to break a _streak_. If you've successfully done the habit every day for 29 days, there's a strong incentive not to break that run of success. Life has a way of coming at you, and you might well need that extra motivation at some point. Not today, not tomorrow, but soon: probably just around the time the novelty wears off.

The aim of this project is to produce a Go package and accompanying command-line tool called ```habctl``` (habit control) that will help users track and establish a new habit, by reporting their current streak.

For example, if you decide you want to build the habit of jogging every day, you might tell the habit tool about it like this:

**`habctl jog`**

```
Good luck with your new habit 'jog'! Don't forget to do it again
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
You're currently on a 16-day streak for 'jog'. Stick to it!
```

Maybe the news won't be quite so good:

```
It's been 4 days since you did 'jog'. It's okay, life happens. Get
back on that horse today!
```

## Problems to solve

There's a surprising amount involved in what seems like a simple tool. You'll need to:

* Build a tool that can take command-line arguments

* Write tests for a program that prints to the terminal, without printing to the terminal

* Read data from a disk file or database, and update it as necessary

* Calculate time intervals so that you know whether to extend the current streak, or start a new one

* Make sure you don't extend the streak when the user performs the habit more than once on a given day

## Stretch goals

Some more refinements to add to your program if you like:

* Add some variation to the messages; for example, the program might get more and more congratulatory as your streak increases

* Handle multiple habits

* Handle habits that you want to perform at some longer interval than a day (every week, perhaps)

* Add a web interface to the program so that you can check and update your habit streaks using a web browser

## Credits

This is an educational Go project intended for students at the [Bitfield Institute of Technology](https://bitfieldconsulting.com/golang/bit).

# Development

Install [**gotestdox**](https://github.com/bitfield/gotestdox)
```
go install github.com/bitfield/gotestdox/cmd/gotestdox@latest
```

## First Iteration


Run tests
```
➜  habit git:(main) ✗ gotestdox
github.com/qba73/habit:
 ✔ Habit starts new activity with name and initial date (0.00s)
 ✔ Habit prints number of days on not broken current streak (0.00s)
 ✔ Habit records activity on next day (0.00s)
 ✔ Habit does not log duplicated activity on the same day (0.00s)
 ✔ Habit prints message on three day streak (0.00s)
 ✔ Habit prints number of days since broken streak (0.00s)
 ✔ Habit loads data from JSON file (0.00s)
 ✔ Habit prints message on starting new streak after break (0.00s)
 ✔ Habit saves habit data to file (0.00s)
 ✔ Habit saves updated habit data to file (0.00s)
```

Build ```habctl```
```
go build -o habctl cmd/habctl/main.go
```
Track and manage your habit
```
./habctl walk
Good luck with your new habit 'walk'! Don't forget to do it again tomorrow.

./habctl
You're currently on a 1-day streak for 'walk'. Stick to it!
```
```
./habctl walk
Nice work: you've done the habit 'walk' for 2 days in a row now. Keep it up!

./habctl
You're currently on a 2-day streak for 'walk'. Stick to it!
```
```
./habctl
It's been 6 days since you did 'walk'. It's okay, life happens. Get back on that horse today!

./habctl walk
You last did the habit 'walk' 6 days ago, so you're starting a new streak today. Good luck!

./habctl
You're currently on a 1-day streak for 'walk'. Stick to it!
```
