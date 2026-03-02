package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// FixtureDir returns the path to the testdata directory relative to the
// calling test file's source location. This makes fixture paths work
// regardless of where tests are run from.
//
// Example:
//
//	dir := testutil.FixtureDir(t)
//	// dir == "<package_dir>/testdata"
func FixtureDir(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		t.Fatal("FixtureDir: could not determine caller path")
	}
	return filepath.Join(filepath.Dir(filename), "testdata")
}

// LoadFixture reads a file from the given path and returns its contents.
// The path is relative to the testdata directory of the calling test's package.
//
// Example:
//
//	data := testutil.LoadFixture(t, "events/sample_event.json")
func LoadFixture(t *testing.T, name string) []byte {
	t.Helper()
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		t.Fatal("LoadFixture: could not determine caller path")
	}
	path := filepath.Join(filepath.Dir(filename), "testdata", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("LoadFixture: failed to read %q: %v", path, err)
	}
	return data
}

// LoadFixtureJSON reads a JSON fixture file and unmarshals it into v.
// The path is relative to the testdata directory of the calling test's package.
//
// Example:
//
//	var event models.Event
//	testutil.LoadFixtureJSON(t, "events/sample_event.json", &event)
func LoadFixtureJSON(t *testing.T, name string, v any) {
	t.Helper()
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		t.Fatal("LoadFixtureJSON: could not determine caller path")
	}
	path := filepath.Join(filepath.Dir(filename), "testdata", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("LoadFixtureJSON: failed to read %q: %v", path, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("LoadFixtureJSON: failed to unmarshal %q: %v", path, err)
	}
}

// LoadFixtureString reads a fixture file and returns its contents as a string.
// The path is relative to the testdata directory of the calling test's package.
//
// Example:
//
//	body := testutil.LoadFixtureString(t, "responses/error.html")
func LoadFixtureString(t *testing.T, name string) string {
	t.Helper()
	return string(LoadFixture(t, name))
}
