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
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golang.org/x/exp/maps"
)

var Now = time.Now

// Store
type Store interface {
	Log(name string) (string, error)
	GetAll() []Habit
	Save() error
}

// Habit holds data for tracking a habit.
type Habit struct {
	Name   string    `json:"name"`
	Date   time.Time `json:"date"`   // Date it's a date when habit activity was last recorded
	Streak int       `json:"streak"` // Streak represents number of consecutive days when habit was recorded.
}

// New takes a name and returns a new habit.
// It errors if the name is an empty string.
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

// Start logs new date activity and starts a new streak.
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

// Check verifies if the streak is broken.
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

// Log adds activity to an existing streak or
// starts a new streak if the streak is broken.
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

// DayDiff takes two time obj and returns
// diff between them. The time delta is
// calculated in days.
func DayDiff(start, stop time.Time) int {
	start = RoundDateToDay(start)
	stop = RoundDateToDay(stop)
	return int(math.Abs(stop.Sub(start).Hours()) / 24)
}

// RoundDateToDay truncates time to 24h periods (days).
func RoundDateToDay(t time.Time) time.Time {
	return t.UTC().Truncate(24 * time.Hour)
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
		return "."
	}
	return home
}

// FileStore implements Store interface.
type FileStore struct {
	Path string
	Data map[string]Habit
}

// NewFileStore takes a path and returns a file store.
func NewFileStore(path string) (*FileStore, error) {
	hx := make(map[string]Habit)
	store := FileStore{
		Path: path,
		Data: hx,
	}
	_, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return &store, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) != 0 {
		err = json.Unmarshal(data, &hx)
		if err != nil {
			return nil, err
		}
	}
	store.Data = hx
	return &store, nil
}

// Save parsists content of the store.
func (f *FileStore) Save() error {
	data, err := json.Marshal(f.Data)
	if err != nil {
		return err
	}
	dir := filepath.Dir(f.Path)
	_, err = os.Stat(dir)
	if err != nil {
		err = os.MkdirAll(dir, 0o700)
		if err != nil {
			return err
		}
	}
	return os.WriteFile(f.Path, data, 0600)
}

// GetAll returns all tracked habits.
func (f FileStore) GetAll() []Habit {
	hx := maps.Values(f.Data)
	sort.Slice(hx, func(i, j int) bool { return hx[i].Name < hx[j].Name })
	return hx
}

// Add takes a habit and adds it in the store.
func (f *FileStore) Add(habit Habit) {
	f.Data[habit.Name] = habit
}

// Log takes a string representing habit's name and logs the habit.
// If habit with given name does not exist, Log creates it and
// starts tracking.
func (f *FileStore) Log(habitName string) (string, error) {
	habit, ok := f.Data[habitName]
	if !ok {
		h, err := New(habitName)
		if err != nil {
			return "", err
		}
		msg := h.Start()
		f.Add(h)
		return msg, nil
	}
	_, msg := habit.Log()
	f.Add(habit)
	return msg, nil

}

// Check takes a store and reports about all tracked habits.
func Check(s Store) string {
	habits := s.GetAll()
	if len(habits) == 0 {
		return "You are not tracking any habit yet.\n"
	}
	var sb strings.Builder
	for _, habit := range habits {
		_, msg := habit.Check()
		sb.WriteString(msg)
	}
	return sb.String()
}

// Log takes store and habitName and logs habit activity.
// It creates a new habit if habit with provided name does not exist.
func Log(s Store, habitName string) (string, error) {
	msg, err := s.Log(habitName)
	if err != nil {
		return "", err
	}
	if err = s.Save(); err != nil {
		return "", err
	}
	return msg, nil
}

func runCLI(wr, ew io.Writer) int {
	fset := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fset.Parse(os.Args[1:])
	args := fset.Args()

	// Default file storage is created.
	store, err := NewFileStore(dataDir() + "/.habits.json")
	if err != nil {
		fmt.Fprint(ew, err)
		return 1
	}

	// No args, checking habits
	if len(args) == 0 {
		fmt.Fprint(wr, Check(store))
		return 0
	}

	msg, err := Log(store, args[0])
	if err != nil {
		fmt.Fprint(ew, err)
		return 1
	}
	fmt.Fprint(wr, msg)
	return 0
}

func Main() int {
	return runCLI(os.Stdout, os.Stderr)
}
