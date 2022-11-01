package habit

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"math"
	"os"
	"time"
)

var Now = time.Now

type Store interface {
	Save(h *Habit) error
	Load() (*Habit, error)
}

// FileStore implements Store interface.
type FileStore struct {
	Path string
}

// Save saves the habit in a store.
func (f *FileStore) Save(h *Habit) error {
	d, err := json.Marshal(h)
	if err != nil {
		return err
	}
	err = os.WriteFile(f.Path, d, 0644)
	if err != nil {
		return err
	}
	return nil
}

// Load returns habit value or error.
func (f *FileStore) Load() (*Habit, error) {
	var h Habit
	b, err := os.ReadFile(f.Path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("habit store %v does not exist: %w", f.Path, err)
		}
		return nil, err
	}
	err = json.Unmarshal(b, &h)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

// Habit represents habit metadata.
type Habit struct {
	Name string
	// Date it's a date when habit activity was last recorded
	Date time.Time
	// Streak represents number of consecutive days
	// when habit was recorded.
	Streak int
}

// New takes a name and returns a new habit
// or error if the name is an empty string.
func New(name string) (*Habit, error) {
	if name == "" {
		return nil, errors.New("nil habit name")
	}
	h := Habit{
		Name:   name,
		Date:   RoundDate(Now()),
		Streak: 1,
	}
	return &h, nil
}

// Start logs new habit activity and starts a new streak.
func (h *Habit) Start() string {
	h.startNewStreak()
	return fmt.Sprintf("Good luck with your new habit '%s'. Don't forget to do it tomorrow.\n", h.Name)
}

func (h *Habit) startNewStreak() {
	h.Date = RoundDate(Now())
	h.Streak = 1
}

func (h *Habit) continueStreak() {
	h.Date = RoundDate(Now())
	h.Streak += 1
}

func (h Habit) checkStreak() int {
	return DayDiff(h.Date, Now())
}

// DayDiff takes two times and returns difference between them.
// Returned value represent absolute number of days.
func DayDiff(start, stop time.Time) int {
	start = RoundDate(start)
	stop = RoundDate(stop)
	return int(math.Abs(stop.Sub(start).Hours()) / 24)
}

// RoundDate truncates time to full days.
func RoundDate(t time.Time) time.Time {
	return t.UTC().Truncate(24 * time.Hour)
}

// Check verifies if the streak is broken or not.
// Returned value represents number of days since
// the habit was logged.
func (h Habit) Check() (int, string) {
	diff := h.checkStreak()
	if diff == 0 || diff == 1 {
		return diff, fmt.Sprintf("You're currently on a %d-day streak for '%s'. Stick to it!\n", h.Streak, h.Name)
	}
	return diff, fmt.Sprintf("It's been %d days since you did '%s'. It's ok, life happens. Get back on that horse today!\n", diff, h.Name)
}

// Log adds activity to an existing habit streak or
// starts a new streak if the current one is broken.
// Log returns streak lenght and a message.
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

// Save saves habit using provided store.
func Save(s Store, h *Habit) error {
	return s.Save(h)
}

// Load returns habit from the provided store.
func Load(s Store) (*Habit, error) {
	return s.Load()
}

func RunCLI() {
	fset := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fset.Parse(os.Args[1:])
	args := fset.Args()

	store := FileStore{
		Path: "habit.json",
	}

	// No args, so check habit and the file
	if len(args) == 0 {
		h, err := store.Load()
		if errors.Is(err, fs.ErrNotExist) {
			fmt.Fprint(os.Stderr, "You are not tracking any habit yet.\n")
			os.Exit(1)
		}
		_, msg := h.Check()
		fmt.Fprint(os.Stdout, msg)
		os.Exit(0)
	}

	// Start a new habit
	habitName := args[0]

	h, err := store.Load()
	if err != nil {
		fset.Usage()
		os.Exit(1)
	}
	if h.Name != habitName {
		// Start new habit
		h, err := New(habitName)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		msg := h.Start()
		err = store.Save(h)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Fprint(os.Stdout, msg)
		os.Exit(0)
	}

	_, msg := h.Log()
	fmt.Fprint(os.Stdout, msg)
	err = store.Save(h)
	if err != nil {
		fmt.Fprint(os.Stdout, err)
		os.Exit(1)
	}
	os.Exit(0)
}
