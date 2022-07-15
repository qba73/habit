package habit

import (
	"encoding/json"
	"errors"
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
	Saver  io.Writer   `json:"-"`
	Loader io.Reader   `json:"-"`
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

func New(name string, opts ...option) (*Habit, error) {
	if name == "" {
		return nil, errors.New("nil habit name")
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
	return &h, nil
}

func NewFromFile(path string, opts ...option) (*Habit, error) {
	if path == "" {
		return nil, errors.New("empty path to habit file")
	}
	h, err := LoadFromFile(path)
	if err != nil {
		return nil, err
	}
	habit := Habit{
		Name:   h.Name,
		Dates:  h.Dates,
		Output: os.Stdout,
	}
	for _, opt := range opts {
		opt(&habit)
	}
	return &habit, nil
}

// Streak returns number of dates logged.
func (h *Habit) Streak() int {
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

// Check returns number of days in the current streak or number of days
// since last recorded activity.
func (h *Habit) Check() int {
	today := habitDate(time.Now())
	firstRecorded := h.Dates[0]
	lastRecorded := h.Dates[len(h.Dates)-1]

	if diff := int(today.Sub(lastRecorded).Hours() / 24); diff > 1 {
		fmt.Fprintf(h.Output, "It's been %d days since you did '%s'. It's okay, life happens. Get back on that horse today!\n", diff, h.Name)
		return diff
	}
	streak := int(today.Sub(firstRecorded).Hours() / 24)
	fmt.Fprintf(h.Output, "You're currently on a %d-day streak for '%s'. Stick to it!\n", streak, h.Name)
	return streak
}

// habitDate truncates time to full days.
func habitDate(t time.Time) time.Time {
	return t.UTC().Truncate(24 * time.Hour)
}

func SaveToFile(path string, h *Habit) error {
	d, err := json.Marshal(h)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, d, 0644)
	if err != nil {
		return err
	}
	return nil
}

func LoadFromFile(path string) (*Habit, error) {
	var h Habit
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(b), &h)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func RunCLI() {
	fmt.Println("running habctl...")
}
