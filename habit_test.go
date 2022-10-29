package habit_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/qba73/habit"
)

func TestHabit_StartsNewActivityWithNameAndInitialDate(t *testing.T) {
	t.Parallel()
	h, err := habit.New("jog")
	if err != nil {
		t.Fatal(err)
	}
	want := time.Now().UTC().Truncate(24 * time.Hour)
	got := h.Date
	if want != got {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestHabit_DoesNotLogDuplicatedActivityOnTheSameDay(t *testing.T) {
	t.Parallel()
	h, err := habit.New("jog")
	if err != nil {
		t.Fatal(err)
	}
	h.Log()
	wantDate := time.Now().UTC().Truncate(24 * time.Hour)
	gotDate := h.Date
	if wantDate != gotDate {
		t.Fatalf("want %v, got %v", wantDate, gotDate)
	}

	wantStreak := 1
	gotStreak := h.Streak
	if wantStreak != gotStreak {
		t.Errorf("want %d, got %d", wantStreak, gotStreak)
	}
}

func TestHabit_ChecksStreakLength(t *testing.T) {
	t.Parallel()
	h, err := habit.New("jog")
	if err != nil {
		t.Fatal(err)
	}
	want := 0
	got, _ := h.Check()
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	h.Date = time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -1)
	want = 1
	got, _ = h.Check()
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}
}

func TestHabit_RecordsActivityOnNextDay(t *testing.T) {
	t.Parallel()
	h, err := habit.New("jog")
	if err != nil {
		t.Fatal(err)
	}
	h.Date = time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -1)
	h.Log()
	want := 2
	got := h.Streak
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}
}

func TestHabit_StartsNewActivityAfterBrokenStreak(t *testing.T) {
	t.Parallel()
	h, err := habit.New("jog")
	if err != nil {
		t.Fatal(err)
	}
	// Setup date 5 days ago to simulate that the habit was logged 5 days ago.
	h.Date = time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -5)

	got, _ := h.Check()
	want := 5
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	want = 1
	got, _ = h.Log()
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}
}

func TestLoadHabitFromFile(t *testing.T) {
	t.Parallel()
	habitFilepath := "testdata/habit.json"
	got, err := habit.LoadFromFile(habitFilepath)
	if err != nil {
		t.Fatal(err)
	}
	want := &habit.Habit{
		Name:   "walk",
		Date:   time.Date(2022, 07, 15, 00, 00, 00, 00, time.UTC),
		Streak: 1,
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestSaveHabitToFile(t *testing.T) {
	t.Parallel()
	h, err := habit.New("run")
	if err != nil {
		t.Fatal(err)
	}
	h.Date = time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -1)

	path := t.TempDir() + "/habit.json"
	err = habit.SaveToFile(path, h)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := habit.LoadFromFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(h, h2) {
		t.Errorf("want %+v, got %+v", h, h2)
	}
}

func TestHabit_SavesUpdatedHabitDataToFile(t *testing.T) {
	t.Parallel()
	habitFilepath := "testdata/habit.json"
	h, err := habit.LoadFromFile(habitFilepath)
	if err != nil {
		t.Fatal(err)
	}

	h.Log()

	path := t.TempDir() + "/path.json"

	habit.SaveToFile(path, h)

	h2, err := habit.LoadFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(h2, h) {
		t.Error(cmp.Diff(h2, h))
	}
}

// func TestMain(m *testing.M) {
// 	testscript.RunMain(m, map[string]func() int{
// 		"habit": func() int {
// 			habit.RunCLI()
// 			return 0
// 		},
// 	})
// }
// func TestScript(t *testing.T) {
// 	testscript.Run(t, testscript.Params{Dir: "testdata/script"})
// }
