package habit

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"
)

type Habit struct {
	Name   string      `json:"name"`
	Dates  []time.Time `json:"dates"`
	Output io.Writer   `json:"-"`
}

type option func(*Habit) error

func WithOutput(w io.Writer) option {
	return func(h *Habit) error {
		if w == nil {
			return errors.New("nil habit writer")
		}
		h.Output = w
		return nil
	}
}

func New(name string, opts ...option) (Habit, error) {
	if name == "" {
		return Habit{}, errors.New("nil habit name")
	}
	h := Habit{
		Name:   name,
		Dates:  []time.Time{habitDate(time.Now())},
		Output: os.Stdout,
	}
	for _, opt := range opts {
		opt(&h)
	}
	fmt.Fprintf(h.Output, "Good luck with your new habit '%s'! Don't forget to do it again tomorrow.\n", h.Name)
	return h, nil
}

// Streak returns number of dates logged.
func (h Habit) Streak() int {
	return len(h.Dates)
}

// LogActivity add activity to the habit counter.
func (h *Habit) LogActivity() {
	today := habitDate(time.Now())
	lastRecorded := h.Dates[len(h.Dates)-1]
	if !today.After(lastRecorded) {
		return
	}
	// Shall we start new streak or continue recording dates to current habit.
	diff := int(today.Sub(lastRecorded).Hours() / 24)
	if diff > 1 {
		// Streak broken, remove dates and start a new streak.
		h.Dates = []time.Time{today}
		fmt.Fprintf(h.Output, "You last did the habit '%s' %d days ago, so you're starting a new streak today. Good luck!\n", h.Name, diff)
		return
	}
	h.Dates = append(h.Dates, today)
	fmt.Fprintf(h.Output, "Nice work: you've done the habit '%s' for %d days in a row now. Keep it up!\n", h.Name, len(h.Dates))
}

// Check prints number of days in the current streak.
func (h Habit) Check() int {
	today := habitDate(time.Now())
	firstRecorded := h.Dates[0]
	lastRecorded := h.Dates[len(h.Dates)-1]

	diff := int(today.Sub(lastRecorded).Hours() / 24)
	if diff > 1 {
		fmt.Fprintf(h.Output, "It's been %d days since you did '%s'. It's okay, life happens. Get back on that horse today!\n", diff, h.Name)
		return diff
	}
	streak := int(today.Sub(firstRecorded).Hours()/24) + 1
	fmt.Fprintf(h.Output, "You're currently on a %d-day streak for '%s'. Stick to it!\n", streak, h.Name)
	return streak
}

func FromFile(path string, opts ...option) (Habit, error) {
	if path == "" {
		return Habit{}, errors.New("empty path to habit file")
	}
	h, err := LoadFromFile(path)
	if err != nil {
		return Habit{}, err
	}
	habit := Habit{
		Name:   h.Name,
		Dates:  h.Dates,
		Output: os.Stdout,
	}
	for _, opt := range opts {
		opt(&habit)
	}
	return habit, nil
}

func SaveToFile(path string, h Habit) error {
	hb := Habit{
		Name:  h.Name,
		Dates: h.Dates,
	}
	d, err := json.Marshal(hb)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, d, 0644)
	if err != nil {
		return err
	}
	return nil
}

func LoadFromFile(path string) (Habit, error) {
	var h Habit
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return Habit{}, err
	}
	err = json.Unmarshal(b, &h)
	if err != nil {
		return Habit{}, err
	}
	return h, nil
}

// habitDate truncates time to full days.
func habitDate(t time.Time) time.Time {
	return t.UTC().Truncate(24 * time.Hour)
}

func RunCLI() {
	fset := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fset.Parse(os.Args[1:])

	args := fset.Args()

	// No args, so check habit and the file
	if len(args) == 0 {
		h, err := FromFile("./habit.json")
		if err != nil {
			panic(err)
		}
		h.Check()
		os.Exit(0)
	}

	habitName := args[0]

	h, err := FromFile("./habit.json")
	if err != nil {
		os.Exit(1)
	}
	if habitName != h.Name {
		// start new habit
		h, err := New(habitName)
		if err != nil {
			os.Exit(1)
		}
		err = SaveToFile("./habit.json", h)
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	h.LogActivity()
	SaveToFile("./habit.json", h)
	os.Exit(0)
}
