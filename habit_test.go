package habit_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/qba73/habit"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestRoundDate_RoundsHabitLogTimeToAFullDay(t *testing.T) {
	t.Parallel()

	testTime, err := time.Parse(time.RFC3339, "2022-11-01T23:00:00Z")
	if err != nil {
		t.Fatal(err)
	}

	want, err := time.Parse(time.RFC3339, "2022-11-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	got := habit.RoundDateToDay(testTime)
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestNew_ErrorsOnCreatingHabitWithEmptyName(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2022-11-01T02:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	_, err = habit.New("")
	if err == nil {
		t.Fatal("want err, got nil")
	}
}

func TestNew_CreatesNewHabit(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2022-11-01T02:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	got, err := habit.New("jog")
	if err != nil {
		t.Fatal(err)
	}

	date, err := time.Parse(time.RFC3339, "2022-11-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}

	want := habit.Habit{
		Name:   "jog",
		Date:   date,
		Streak: 1,
	}

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestStart_StartsTrackingHabit(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2022-10-01T01:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	habitName := "jog"
	gotHabit, err := habit.New(habitName)
	if err != nil {
		t.Fatal(err)
	}
	got := gotHabit.Start()
	want := fmt.Sprintf("Good luck with your new habit '%s'. Don't forget to do it tomorrow.\n", habitName)
	if want != got {
		t.Error(cmp.Diff(want, got))
	}

	date, err := time.Parse(time.RFC3339, "2022-10-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}

	wantHabit := habit.Habit{
		Name:   habitName,
		Date:   date,
		Streak: 1,
	}

	if !cmp.Equal(wantHabit, gotHabit) {
		t.Error(cmp.Diff(wantHabit, gotHabit))
	}
}

func TestLog_DoesNotDuplicateHabitActivityOnTheSameDay(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2022-09-01T01:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	gotHabit, err := habit.New("jog")
	if err != nil {
		t.Fatal(err)
	}

	testTime, err = time.Parse(time.RFC3339, "2022-09-01T15:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	days, msg := gotHabit.Record()

	wantDays := 1
	if !cmp.Equal(wantDays, days) {
		t.Error(cmp.Diff(wantDays, days))
	}

	wantMsg := ""
	if !cmp.Equal(wantMsg, msg) {
		t.Error(cmp.Diff(wantMsg, msg))
	}

	wantDate, err := time.Parse(time.RFC3339, "2022-09-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}

	wantHabit := habit.Habit{
		Name:   "jog",
		Date:   wantDate,
		Streak: 1,
	}

	if !cmp.Equal(wantHabit, gotHabit) {
		t.Error(cmp.Diff(wantHabit, gotHabit))
	}
}

func TestCheck_ReportsHabitStreakLengthOnNotBrokenStreak(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2022-11-01T01:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	h, err := habit.New("jog")
	if err != nil {
		t.Fatal(err)
	}

	want := 0
	got, _ := h.Check()
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	checkTime, err := time.Parse(time.RFC3339, "2022-11-02T01:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return checkTime
	}

	want = 1
	got, msg := h.Check()
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}
	wantMessage := fmt.Sprintf("You're currently on a %d-day streak for '%s'. Stick to it!\n", want, h.Name)
	if wantMessage != msg {
		t.Error(cmp.Diff(wantMessage, msg))
	}
}

func TestCheck_ReportsBrokenStreak(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2022-10-27T01:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	h, err := habit.New("jog")
	if err != nil {
		t.Fatal(err)
	}

	tt := []struct {
		checkTime   string
		wantDayDiff int
	}{
		{checkTime: "2022-10-30T01:00:00Z", wantDayDiff: 3},
		{checkTime: "2022-01-31T05:00:00Z", wantDayDiff: 269},
		{checkTime: "2022-11-28T15:00:00Z", wantDayDiff: 32},
	}

	for _, tc := range tt {
		checkTime, err := time.Parse(time.RFC3339, tc.checkTime)
		if err != nil {
			t.Fatal(err)
		}
		h.Date = checkTime

		got, msg := h.Check()
		wantDays := tc.wantDayDiff
		if wantDays != got {
			t.Errorf("want %d, got %d", tc.wantDayDiff, got)
		}
		wantMsg := fmt.Sprintf("It's been %d days since you did '%s'. It's ok, life happens. Get back on that horse today!\n", tc.wantDayDiff, h.Name)
		if wantMsg != msg {
			t.Error(cmp.Diff(wantMsg, msg))
		}
	}
}

func TestLog_RecordsHabitOnNextDayOnNotBrokenStreak(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2022-10-02T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	h, err := habit.New("jog")
	if err != nil {
		t.Fatal(err)
	}

	habit.Now = func() time.Time {
		habitLogTime, err := time.Parse(time.RFC3339, "2022-10-03T00:00:00Z")
		if err != nil {
			t.Fatal(err)
		}
		return habitLogTime
	}

	got, msg := h.Record()
	want := 2
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	wantMsg := fmt.Sprintf("Nice work: you've done the habit '%s' for %d days in a row now. Keep it up!\n", h.Name, want)
	if wantMsg != msg {
		t.Error(cmp.Diff(wantMsg, msg))
	}
}

func TestLog_StartsNewHabitStreakAfterBrokenStreak(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2022-09-01T03:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	h, err := habit.New("jog")
	if err != nil {
		t.Fatal(err)
	}

	checkTime, err := time.Parse(time.RFC3339, "2022-09-05T01:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return checkTime
	}

	want := 1
	got, msg := h.Record()
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	wantDays := 4
	wantMsg := fmt.Sprintf("You last did the habit '%s' %d days ago, so you're starting a new streak today. Good luck!\n", h.Name, wantDays)
	if wantMsg != msg {
		t.Error(cmp.Diff(wantMsg, msg))
	}
}

func testPath(t *testing.T) string {
	return t.TempDir() + "/.habits.json"
}

func TestNewFileStore_CreatesNewEmptyStore(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2022-10-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	store, err := habit.NewFileStore(testPath(t))
	if err != nil {
		t.Fatal(err)
	}

	got := store.GetAll()
	want := []habit.Habit{}

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestNewFileStore_ErrorsOnInvalidData(t *testing.T) {
	path := t.TempDir() + "/.habits.json"
	err := os.WriteFile(path, []byte("invalid data"), 0o600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = habit.NewFileStore(path)
	if err == nil {
		t.Fatal("no error")
	}
}

func TestNewFileStore_ErrorsOnIOError(t *testing.T) {
	path := t.TempDir() + "/.habits.json"
	err := os.WriteFile(path, []byte("invalid data"), 0o600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = habit.NewFileStore(path)
	if err == nil {
		t.Fatal("no error")
	}
}

func TestGetAll_RetrievesAllHabitsFromFileStore(t *testing.T) {
	store, err := habit.NewFileStore("testdata/.habits.json")
	if err != nil {
		t.Fatal(err)
	}

	got := store.GetAll()
	want := []habit.Habit{
		{Name: "jog", Date: time.Date(2022, 10, 0o1, 0o0, 0o0, 0o0, 0o0, time.UTC), Streak: 2},
		{Name: "read", Date: time.Date(2022, 10, 23, 0o0, 0o0, 0o0, 0o0, time.UTC), Streak: 3},
		{Name: "walk", Date: time.Date(2022, 10, 0o1, 0o0, 0o0, 0o0, 0o0, time.UTC), Streak: 1},
	}

	if !cmp.Equal(want, got, cmpopts.SortSlices(func(x, y habit.Habit) bool { return x.Name < y.Name })) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestLog_AddsHabitToFileStore(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2022-10-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}
	store, err := habit.NewFileStore(testPath(t))
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Log("jog")
	if err != nil {
		t.Fatal(err)
	}

	if err = store.Save(); err != nil {
		t.Fatal(err)
	}

	testTime, err = time.Parse(time.RFC3339, "2022-10-02T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	store.Log("jog")

	if err := store.Save(); err != nil {
		t.Fatal(err)
	}

	got, ok := store.Data["jog"]
	if !ok {
		t.Fatalf("habit 'jog' does not exist")
	}
	want := habit.Habit{
		Name:   "jog",
		Date:   time.Date(2022, 10, 0o2, 0o0, 0o0, 0o0, 0o0, time.UTC),
		Streak: 2,
	}

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestLog_SavesHabitToEmptyFileStore(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2022-10-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	store, err := habit.NewFileStore(testPath(t))
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.Log("jog")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(); err != nil {
		t.Fatal(err)
	}

	want := habit.Habit{
		Name:   "jog",
		Date:   time.Date(2022, 10, 0o1, 0o0, 0o0, 0o0, 0o0, time.UTC),
		Streak: 1,
	}

	got, ok := store.Data["jog"]
	if !ok {
		t.Fatal(err)
	}

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestStoreLog_SavesHabitToFileStore(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2022-09-01T03:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	store, err := habit.NewFileStore(testPath(t))
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.Log("run")
	if err != nil {
		t.Fatal(err)
	}
	if err = store.Save(); err != nil {
		t.Fatal(err)
	}

	hx := store.GetAll()
	got := hx[0]
	want := habit.Habit{
		Name:   "run",
		Date:   time.Date(2022, 9, 1, 0o0, 0o0, 0o0, 0o0, time.UTC),
		Streak: 1,
	}
	if !cmp.Equal(want, got) {
		t.Errorf(cmp.Diff(want, got))
	}
}

func TestFileStore_LogHabitOnNotExistingHabitName(t *testing.T) {
	store, err := habit.NewFileStore(testPath(t))
	if err != nil {
		t.Fatal(err)
	}

	want := "Good luck with your new habit 'walking'. Don't forget to do it tomorrow.\n"
	got, err := store.Log("walking")
	if err != nil {
		t.Fatal(err)
	}
	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestCheck_PrintsOutMessageForEmptyFileStore(t *testing.T) {
	store, err := habit.NewFileStore(testPath(t))
	if err != nil {
		t.Fatal(err)
	}

	got := habit.Check(store)
	want := "You are not tracking any habit yet.\n"

	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestCheck_PrintsOutMessageForNonEmptyFileStore(t *testing.T) {
	store, err := habit.NewFileStore("testdata/.habits.json")
	if err != nil {
		t.Fatal(err)
	}

	got := habit.Check(store)
	want := "It's been 30 days since you did 'jog'. It's ok, life happens. Get back on that horse today!\nIt's been 52 days since you did 'read'. It's ok, life happens. Get back on that horse today!\nIt's been 30 days since you did 'walk'. It's ok, life happens. Get back on that horse today!\n"
	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestLog_ErrorsOnEmptyHabitName(t *testing.T) {
	store, err := habit.NewFileStore(testPath(t))
	if err != nil {
		t.Fatal(err)
	}
	_, err = habit.Record(store, "")
	if err == nil {
		t.Fatal("want err, got nil")
	}
}

func TestLog_LogsNewHabit(t *testing.T) {
	store, err := habit.NewFileStore(testPath(t))
	if err != nil {
		t.Fatal(err)
	}

	got, err := habit.Record(store, "bike")
	if err != nil {
		t.Fatal(err)
	}
	want := "Good luck with your new habit 'bike'. Don't forget to do it tomorrow.\n"

	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestLog_LogsHabitOnSecondDayOnNotBrokenStreak(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2022-09-01T03:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	store, err := habit.NewFileStore(testPath(t))
	if err != nil {
		t.Fatal(err)
	}

	_, err = habit.Record(store, "read")
	if err != nil {
		t.Fatal(err)
	}

	testTime, err = time.Parse(time.RFC3339, "2022-09-02T03:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	got, err := habit.Record(store, "read")
	if err != nil {
		t.Fatal(err)
	}
	want := "Nice work: you've done the habit 'read' for 2 days in a row now. Keep it up!\n"

	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestLog_StartsHabitAndStreakOnBrokenStreak(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2022-09-01T03:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	store, err := habit.NewFileStore(testPath(t))
	if err != nil {
		t.Fatal(err)
	}

	_, err = habit.Record(store, "read")
	if err != nil {
		t.Fatal(err)
	}

	testTime, err = time.Parse(time.RFC3339, "2022-09-03T03:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	got, err := habit.Record(store, "read")
	if err != nil {
		t.Fatal(err)
	}
	want := "You last did the habit 'read' 2 days ago, so you're starting a new streak today. Good luck!\n"

	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestLog_LogsMultipleHabitsWithTwoDaysStreak(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2022-09-01T03:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	store, err := habit.NewFileStore(testPath(t))
	if err != nil {
		t.Fatal(err)
	}

	_, err = habit.Record(store, "read")
	if err != nil {
		t.Fatal(err)
	}
	_, err = habit.Record(store, "play piano")
	if err != nil {
		t.Fatal(err)
	}

	got := habit.Check(store)
	want := "You're currently on a 1-day streak for 'play piano'. Stick to it!\nYou're currently on a 1-day streak for 'read'. Stick to it!\n"
	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}

	testTime, err = time.Parse(time.RFC3339, "2022-09-02T03:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	_, err = habit.Record(store, "read")
	if err != nil {
		t.Fatal(err)
	}
	_, err = habit.Record(store, "play piano")
	if err != nil {
		t.Fatal(err)
	}

	got = habit.Check(store)
	want = "You're currently on a 2-day streak for 'play piano'. Stick to it!\nYou're currently on a 2-day streak for 'read'. Stick to it!\n"
	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"habctl": habit.Main,
	}))
}

func TestHabitCLI(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:  "testdata/script",
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){"date": cmdDate},
	})
}

func cmdDate(ts *testscript.TestScript, neg bool, args []string) {
	if neg {
		ts.Fatalf("unsupported: ! date")
	}
	if len(args) != 3 {
		ts.Fatalf("usage: date filepath -1 habit")
	}
	// Verify 3rd arg
	habitName := strings.TrimSpace(args[2])
	if habitName == "" {
		ts.Fatalf("habit name required")
	}
	// Verify 2nd arg
	dayShift, err := strconv.Atoi(args[1])
	if err != nil {
		ts.Fatalf("expected int of max value -1")
	}

	filepath := args[0]
	fstore, err := habit.NewFileStore(filepath)
	if err != nil {
		ts.Fatalf("opening test filestore: %s, %v", filepath, err)
	}

	h, ok := fstore.Data[habitName]
	if !ok {
		ts.Fatalf("loading habit: %s from filestore", habitName)
	}

	newDate := habit.RoundDateToDay(h.Date.AddDate(0, 0, dayShift))
	h.Date = newDate
	fstore.Add(h)
	err = fstore.Save()
	if err != nil {
		ts.Fatalf("saving updated habit: %s", h.Name)
	}
}
