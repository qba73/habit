package habit

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

type FileStore struct {
	Path string
}

func (f FileStore) Save(h *Habit) error {
	hb := Habit{
		Name:   h.Name,
		Date:   h.Date,
		Streak: h.Streak,
	}
	d, err := json.Marshal(hb)
	if err != nil {
		return err
	}
	err = os.WriteFile(f.Path, d, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (f FileStore) Load() (*Habit, error) {
	var h Habit
	b, err := os.ReadFile(f.Path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &h)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

type Habit struct {
	Name string
	// Date it's a date when habit activity was last recorded
	Date time.Time
	// Streak represents number of consecutive days
	// when habit was recorded.
	Streak int
}

func New(name string) (*Habit, error) {
	if name == "" {
		return nil, errors.New("nil habit name")
	}
	h := Habit{
		Name:   name,
		Date:   habitDate(time.Now()),
		Streak: 1,
	}
	return &h, nil
}

func (h *Habit) startNewStreak() {
	h.Date = habitDate(time.Now())
	h.Streak = 1
}

func (h *Habit) continueStreak() {
	h.Date = habitDate(time.Now())
	h.Streak += 1
}

func (h Habit) checkStreak() int {
	return int(time.Since(h.Date).Hours() / 24)
}

// Check verifies if the streak is broken or not.
// Returned value represents number of days since
// the habit was logged.
func (h Habit) Check() (int, string) {
	diff := h.checkStreak()
	if diff == 0 || diff == 1 {
		return diff, fmt.Sprintf("You are currently on a %d-day streak for '%s'. Stick to it!\n", h.Streak, h.Name)
	}
	return diff, fmt.Sprintf("It's been %d days since you did '%s'. It's ok, life happens. Get back on that horse today!\n", diff, h.Name)
}

// Log adds activity to an existing habit streak or
// starts a new streak if the current one was broken.
// It returns streak lenght and a message.
func (h *Habit) Log() (int, string) {
	diff := h.checkStreak()
	if diff == 0 {
		return h.Streak, ""
	}
	if diff > 1 {
		h.startNewStreak()
		return h.Streak, fmt.Sprintf("You last did the habit '%s' %d days ago, so you're starting a new streak today. Good luck!\n", h.Name, diff)
	}
	h.continueStreak()
	return h.Streak, fmt.Sprintf("Nice work: you've done the habit '%s' for %d days in a row now. Keep it up!\n", h.Name, h.Streak)
}

func SaveToFile(path string, h *Habit) error {
	hb := Habit{
		Name:   h.Name,
		Date:   h.Date,
		Streak: h.Streak,
	}
	d, err := json.Marshal(hb)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, d, 0644)
	if err != nil {
		return err
	}
	return nil
}

func LoadFromFile(path string) (*Habit, error) {
	if path == "" {
		return nil, errors.New("missing path")
	}
	var h Habit
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &h)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

// habitDate truncates time to full days.
func habitDate(t time.Time) time.Time {
	return t.UTC().Truncate(24 * time.Hour)
}

/*
func RunCLI() {
	fset := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fset.Parse(os.Args[1:])

	path, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	args := fset.Args()
	// No args, so check habit and the file
	if len(args) == 0 {
		h, err := LoadFromFile(storePath)
		if errors.Is(err, fs.ErrNotExist) {
			fset.Usage()
			os.Exit(0)
		}

		os.Exit(0)
	}

	habitName := args[0]

	h, err := LoadFromFile(storePath)
	if errors.Is(err, fs.ErrNotExist) {
		// start new habit
		h, err := New(habitName)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		b, err := NewBoard(h, WithOutput(os.Stdout))
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		b.LogActivity()
		err = SaveToFile(storePath, h)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}

	if habitName != h.Name {
		// start new habit
		h, err := New(habitName)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		b, err := NewBoard(h, WithOutput(os.Stdout))
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}

		b.LogActivity()

		err = SaveToFile(storePath, h)
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	b, err := NewBoard(h, WithOutput(os.Stdout))
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	b.LogActivity()

	SaveToFile(storePath, h)
	os.Exit(0)
}
*/
