package habit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"time"
)

var (
	Now               = time.Now
	ErrHabitNotExists = errors.New("habit is not tracked")
)

type Store interface {
	Save(h Habit) error
	Load(name string) (Habit, error)
	List() ([]Habit, error)
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

type Habits map[string]Habit

func (h Habits) Get(name string) Habit {
	habit, ok := h[name]
	if !ok {
		return Habit{}
	}
	return habit
}

func (h Habits) Add(habit Habit) {
	h[habit.Name] = habit
}

func (h Habits) List() []Habit {
	var habits []Habit
	for _, i := range h {
		habits = append(habits, i)
	}
	return habits
}

func (h Habits) ToJSON() ([]byte, error) {
	return json.Marshal(h)
}

func (h Habits) FromJSON(data []byte) error {
	err := json.Unmarshal(data, &h)
	if err != nil {
		return err
	}
	return nil
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
		Path: path + "/" + "habits.json",
	}

	for _, opt := range opts {
		if err := opt(&store); err != nil {
			return nil, err
		}
	}

	_, err = os.Stat(store.Path)
	if errors.Is(err, fs.ErrNotExist) {
		if err := createInitialStore(store.Path); err != nil {
			return nil, err
		}
	}
	return &store, nil
}

// dataDir returns filepath to the habit store.
//
// If user exported the env var XDG_DATA_HOME habctl will use this location to create
// store. Otherwise habctl will attempt to create store in $HOME/.local/share directory.
// The func panics if user's home dir can't be determined.
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

// CreateInitialStore takes a path and creates
// initial habit store.
func createInitialStore(path string) error {
	data := []byte(`{"init":{"name":"init", "date":"0001-01-01T00:00:00Z", "streak":0}}`)
	return os.WriteFile(path, data, 0o600)
}

// Save saves the habit in a store.
func (f *FileStore) Save(h Habit) error {
	b, err := f.open()
	if err != nil {
		return err
	}
	var hx Habits
	err = json.Unmarshal(b, &hx)
	if err != nil {
		return err
	}
	hx.Add(h)
	b, err = json.Marshal(hx)
	if err != nil {
		return err
	}
	return f.save(b)
}

// Load returns habit value or error.
func (f *FileStore) Load(name string) (Habit, error) {
	b, err := f.open()
	if err != nil {
		return Habit{}, err
	}
	var hx Habits
	err = json.Unmarshal(b, &hx)
	if err != nil {
		return Habit{}, err
	}
	h, ok := hx[name]
	if !ok {
		return Habit{}, ErrHabitNotExists
	}
	return h, nil
}

func (f *FileStore) List() ([]Habit, error) {
	b, err := f.open()
	if err != nil {
		return []Habit{}, err
	}
	var hx Habits
	if err := json.Unmarshal(b, &hx); err != nil {
		return []Habit{}, err
	}

	var habits []Habit
	for _, h := range hx {
		if h.Name == "init" {
			continue
		}
		habits = append(habits, h)
	}
	return habits, nil
}

func (f *FileStore) open() ([]byte, error) {
	b, err := os.ReadFile(f.Path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("habit store %v does not exist: %w", f.Path, err)
		}
		return nil, err
	}
	return b, nil
}

func (f *FileStore) save(data []byte) error {
	if err := os.WriteFile(f.Path, data, 0o600); err != nil {
		return fmt.Errorf("saving data to store: %s, %w", f.Path, err)
	}
	return os.Chmod(f.Path, 0600)
}

// ==============================================================================
// Single Habit
// ==============================================================================

// Habit holds metadata for tracking the habit.
type Habit struct {
	Name   string    `json:"name"`
	Date   time.Time `json:"date"`   // Date it's a date when habit activity was last recorded
	Streak int       `json:"streak"` // Streak represents number of consecutive days when habit was recorded.

}

// New takes a name and returns a new habit
// or error if the name is an empty string.
func New(name string) (Habit, error) {
	if name == "" {
		return Habit{}, errors.New("missing habit name")
	}
	h := Habit{
		Name:   name,
		Date:   RoundDateToDay(Now()),
		Streak: 1,
	}
	return h, nil
}

// Start logs new habit activity and starts a new streak.
func (h *Habit) Start() string {
	h.startNewStreak()
	return fmt.Sprintf("Good luck with your new habit '%s'. Don't forget to do it tomorrow.\n", h.Name)
}

func (h *Habit) startNewStreak() {
	h.setDay(Now())
	h.resetStreak()
}

func (h *Habit) setDay(t time.Time) {
	h.Date = RoundDateToDay(t)
}

func (h *Habit) resetStreak() {
	h.Streak = 1
}

func (h *Habit) incStreak() {
	h.Streak++
}

func (h *Habit) continueStreak() {
	h.setDay(Now())
	h.incStreak()
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

func (h *Habit) checkStreak() int {
	return DayDiff(h.Date, Now().UTC())
}

// DayDiff takes two time obj and returns
// diff between them in days.
func DayDiff(start, stop time.Time) int {
	start = RoundDateToDay(start)
	stop = RoundDateToDay(stop)
	return int(math.Abs(stop.Sub(start).Hours()) / 24)
}

// RoundDateToDay truncates time to 24h periods (days).
func RoundDateToDay(t time.Time) time.Time {
	return t.UTC().Truncate(24 * time.Hour)
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

// func runCLI(wr, ew io.Writer) int {
// 	fset := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
// 	fset.Parse(os.Args[1:])
// 	args := fset.Args()

// 	// Make sure default file storage is created.
// 	store, err := NewFileStore()
// 	if err != nil {
// 		fmt.Fprint(ew, err)
// 		return 1
// 	}

// 	// No args, checking habits
// 	if len(args) == 0 {
// 		habits, err := store.List()
// 		if errors.Is(err, fs.ErrNotExist) {
// 			fmt.Fprint(wr, "You are not tracking any habit yet.\n")
// 			return 0
// 		}
// 		_, ok := habits.Read("")
// 		if ok {
// 			fmt.Fprint(wr, "You are not tracking any habit yet.\n")
// 			return 0
// 		}
// 		for _, h := range habits.List() {
// 			// _, msg := h.Check()
// 			// fmt.Fprint(os.Stdout, msg)
// 			fmt.Fprint(os.Stdout, h)
// 		}
// 		return 0
// 	}

// 	// Start a new habit
// 	habitName := args[0]

// 	h, err := store.Load(habitName)
// 	if errors.Is(err, ErrHabitNotExists) {
// 		// Start new habit
// 		h, err := New(habitName)
// 		if err != nil {
// 			fmt.Fprint(ew, err)
// 			return 1
// 		}
// 		msg := h.Start()
// 		if err = store.Save(h); err != nil {
// 			fmt.Fprint(ew, err)
// 			return 1
// 		}
// 		fmt.Fprint(wr, msg)
// 		return 0
// 	}

// 	// Update habit - log new activity

// 	_, msg := h.Log()
// 	fmt.Fprint(wr, msg)
// 	err = store.Save(h)
// 	if err != nil {
// 		fmt.Fprint(wr, err)
// 		return 1
// 	}
// 	return 0
// }

// func Main() int {
// 	return runCLI(os.Stdout, os.Stderr)
// }
