package parameters

import (
	"reflect"
	"testing"
)

func TestGenerateChunks(t *testing.T) {
	var cases = []struct {
		flatSlice []string
		chunkSIze int
		expected  [][]string
	}{
		{[]string{}, 1, [][]string{}},
		{[]string{}, 2, [][]string{}},
		{[]string{"val-1", "val-2"}, 2, [][]string{{"val-1", "val-2"}}},
		{[]string{"val-1", "val-2", "val-3", "val-4", "val-5"}, 1, [][]string{{"val-1"}, {"val-2"}, {"val-3"}, {"val-4"}, {"val-5"}}},
		{[]string{"val-1", "val-2", "val-3", "val-4", "val-5"}, 2, [][]string{{"val-1", "val-2"}, {"val-3", "val-4"}, {"val-5"}}},
		{[]string{"val-1", "val-2", "val-3", "val-4", "val-5", "val-6"}, 3, [][]string{{"val-1", "val-2", "val-3"}, {"val-4", "val-5", "val-6"}}},
	}
	for _, c := range cases {
		chunks := GenerateChunks(c.flatSlice, c.chunkSIze)
		if !reflect.DeepEqual(chunks, c.expected) {
			t.Logf("Value should be %s, but got %s", c.expected, chunks)
			t.Fail()
		}
	}
}
