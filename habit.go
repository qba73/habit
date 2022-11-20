package habit

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
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

type option func(*FileStore) error

func WithFilePath(path string) option {
	return func(fs *FileStore) error {
		if path == "" {
			return errors.New("missing path")
		}
		fs.Path = path
		return nil
	}
}

// FileStore implements Store interface.
type FileStore struct {
	Path string
}

// NewFileStore attempts to creates file storage '.habit.json'
// in user's home dir. It creates the file '.habit.json' only if
// the file is not present in the home dir.
func NewFileStore(opts ...option) (*FileStore, error) {
	path := dataDir()
	err := os.MkdirAll(path, 0o700)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return nil, err
	}

	store := FileStore{
		Path: path + "/" + "habit.json",
	}

	for _, opt := range opts {
		if err := opt(&store); err != nil {
			return nil, err
		}
	}

	_, err = os.Stat(store.Path)
	if errors.Is(err, fs.ErrNotExist) {
		if err := createInitalStore(store.Path); err != nil {
			return nil, err
		}
	}
	return &store, nil
}

func dataDir() string {
	path, ok := os.LookupEnv("XDG_DATA_HOME")
	if ok {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		panic("can't determine user's home dir")
	}
	return home + "/.local/share"
}

func createInitalStore(path string) error {
	h := Habit{Name: ""}
	data, err := json.Marshal(h)
	if err != nil {
		return fmt.Errorf("marshaling data: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("creating initial store: %w", err)
	}
	return nil
}

// Save saves the habit in a store.
func (f *FileStore) Save(h *Habit) error {
	d, err := json.Marshal(h)
	if err != nil {
		return err
	}
	if err = os.WriteFile(f.Path, d, 0644); err != nil {
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

// Habit holds metadata for tracking the habit.
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

func (h *Habit) checkStreak() int {
	return DayDiff(h.Date, Now().UTC())
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
func (h *Habit) Check() (int, string) {
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

func runCLI(wr, ew io.Writer) int {
	fset := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fset.Parse(os.Args[1:])
	args := fset.Args()

	// Make sure default file storage is created.
	store, err := NewFileStore()
	if err != nil {
		fmt.Fprint(ew, err)
		return 1
	}

	// No args, so check habit
	if len(args) == 0 {
		h, err := store.Load()
		if errors.Is(err, fs.ErrNotExist) || h.Name == "" {
			fmt.Fprint(wr, "You are not tracking any habit yet.\n")
			return 0
		}
		_, msg := h.Check()
		fmt.Fprint(os.Stdout, msg)
		return 0
	}

	// Start a new habit
	habitName := args[0]

	h, err := store.Load()
	if err != nil {
		fmt.Fprint(ew, err)
		return 1
	}
	if h.Name != habitName {
		// Start new habit
		h, err := New(habitName)
		if err != nil {
			fmt.Fprint(ew, err)
			return 1
		}
		msg := h.Start()
		err = store.Save(h)
		if err != nil {
			fmt.Fprint(ew, err)
			return 1
		}
		fmt.Fprint(wr, msg)
		return 0
	}

	_, msg := h.Log()
	fmt.Fprint(wr, msg)
	err = store.Save(h)
	if err != nil {
		fmt.Fprint(wr, err)
		return 1
	}
	return 0
}

func Main() int {
	return runCLI(os.Stdout, os.Stderr)
}
