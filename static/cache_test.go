package static_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/xakepp35/anygate/static"
)

func TestCache_BasicOperations(t *testing.T) {
	cache := static.NewCache()

	file1 := static.NewFile([]byte("data1"), "text/plain", time.Time{})
	file2 := static.NewFile([]byte("data2"), "text/html", time.Time{})

	// Set
	cache.Set("a", file1)
	cache.Set("b", file2)

	// Len
	if l := cache.Len(); l != 2 {
		t.Errorf("expected Len 2, got %d", l)
	}

	// Get
	if f := cache.Get("a"); f != file1 {
		t.Errorf("expected file1 from Get('a'), got %+v", f)
	}
	if f := cache.Get("b"); f != file2 {
		t.Errorf("expected file2 from Get('b'), got %+v", f)
	}
	if f := cache.Get("missing"); f != nil {
		t.Errorf("expected nil for missing key, got %+v", f)
	}

	// Keys
	keys := cache.Keys()
	expected := []string{"a", "b"}
	if !reflect.DeepEqual(sortStrings(keys), sortStrings(expected)) {
		t.Errorf("expected keys %v, got %v", expected, keys)
	}

	// Delete
	cache.Delete("a")
	if cache.Get("a") != nil {
		t.Error("expected nil after Delete('a')")
	}
	if l := cache.Len(); l != 1 {
		t.Errorf("expected Len 1 after delete, got %d", l)
	}

	// Clear
	cache.Clear()
	if l := cache.Len(); l != 0 {
		t.Errorf("expected Len 0 after clear, got %d", l)
	}
	if keys := cache.Keys(); len(keys) != 0 {
		t.Errorf("expected no keys after clear, got %v", keys)
	}
}

// sortStrings — вспомогательная функция для детерминированного сравнения ключей
func sortStrings(s []string) []string {
	out := make([]string, len(s))
	copy(out, s)
	// простой пузырьковый для краткости (или можно sort.Strings)
	for i := range out {
		for j := i + 1; j < len(out); j++ {
			if out[i] > out[j] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}
