package registry

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterAndGet(t *testing.T) {
	resetRegistryForTest()
	t.Cleanup(resetRegistryForTest)

	type sample struct {
		Name string
	}

	Register("svc", sample{Name: "alpha"})

	got, ok := Get[sample]("svc")
	require.True(t, ok)
	require.Equal(t, "alpha", got.Name)
}

func TestGetMissingReturnsZeroAndFalse(t *testing.T) {
	resetRegistryForTest()
	t.Cleanup(resetRegistryForTest)

	got, ok := Get[int]("missing")
	require.False(t, ok)
	require.Equal(t, 0, got)
}

func TestRegisterOverwritesExistingValue(t *testing.T) {
	resetRegistryForTest()
	t.Cleanup(resetRegistryForTest)

	Register("number", 1)
	Register("number", 2)

	got, ok := Get[int]("number")
	require.True(t, ok)
	require.Equal(t, 2, got)
}

func resetRegistryForTest() {
	mu.Lock()
	defer mu.Unlock()
	services = make(map[string]interface{})
}
