package compare

import (
	"testing"
)


func TestCompare(t *testing.T){
	tt := []struct{
		name string
		actual interface{}
		expected interface{}
		isEqual bool
	}{
		{"a equals a", "a", "a", true},
		{"a not equals b", "a", "b", false},
		{"bool equals bool", true, true, true},
		{"bool not equals bool", true, false, false},
		{"1 equals 1", 1, 1, true},
		{"1 not equals 1", 1, 0, false},
		{"slice equals slice", []string{"one", "two", "three"}, []string{"one", "two", "three"}, true},
		{"slice not equals slice", []string{"one", "two", "three"}, []string{"one", "four", "three"}, false},
		{"map equals map", map[string]int{"one": 1, "two": 2}, map[string]int{"one": 1, "two": 2}, true},
		{"map not equals map", map[string]int{"one": 1, "two": 2}, map[string]int{"one": 1, "two": 3}, false},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T){
			isEqual, _ := IsEqual(tc.actual,tc.expected)
			if isEqual != tc.isEqual{
				t.Fatalf("%v should be %v got %v\n", tc.name, tc.isEqual, isEqual )
			}
		})
	}

}