# Habit tracker

This is a Go project template, intended for students at the [Bitfield Institute of Technology](https://bitfieldconsulting.com/golang/bit)—but everyone's welcome to try it if you're interested in learning or practicing Go.

## Description

When we're trying to establish a new habit (studying Go, for example), it can be difficult to maintain focus and motivation. Studies suggest that it takes many weeks of regularly performing a new habit before it becomes natural and automatic. So, if you can motivate yourself to do the new habit every day for a month or two, that'll get you well on the way.

One thing that can help is to _track_ your performance of the habit. Suppose you decide that you're going to spend at least 15 minutes every day studying or writing Go programs. You could draw 30 boxes on a piece of paper, one for each of the next 30 days, and check off each box as you do that day's practice.

This simple idea can be surprisingly effective, because we don't like to break a _streak_. If you've successfully done the habit every day for 29 days, there's a strong incentive not to break that run of success. Life has a way of coming at you, and you might well need that extra motivation at some point. Not today, not tomorrow, but soon: probably just around the time the novelty wears off.

The aim of this project is to produce a Go package and accompanying command-line tool that will help users track and establish a new habit, by reporting their current streak.

For example, if you decide you want to build the habit of practising the piano, you might tell the tool about it like this:

**`habit piano`**

```
Good luck with your new habit 'piano'! Don't forget to do it again
tomorrow.
```

As the days go by, you might record each daily practice like this:

**`habit piano`**

```
Nice work: you've done the habit 'piano' for 18 days in a row now.
Keep it up!
```

If you happen to miss a couple of days, that's all right:

**`habit piano`**

```
You last did the habit 'piano' 3 days ago, so you're starting a new
streak today. Good luck!
```

If you just want to check how you're doing, you could run:

**`habit`**

```
You're currently on a 16-day streak for 'piano'. Stick to it!
```

Maybe the news won't be quite so good:

```
It's been 4 days since you did 'piano'. It's okay, life happens. Get
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
