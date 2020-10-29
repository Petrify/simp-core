package testing

import (
	"fmt"
	"testing"

	"github.com/sahilm/fuzzy"
)

type employee struct {
	name string
	age  int
}

type employees []employee

func (e employees) String(i int) string {
	return e[i].name
}

func (e employees) Len() int {
	return len(e)
}

func TestSearch(t *testing.T) {
	emps := employees{
		{
			name: "Alice",
			age:  45,
		},
		{
			name: "Bob",
			age:  35,
		},
		{
			name: "Allie",
			age:  35,
		},
	}
	results := fuzzy.FindFrom("al", emps)
	for _, r := range results {
		fmt.Println(emps[r.Index])
	}
}

func TestSlice(t *testing.T) {
	s := []int{1}
	t.Log(s[1:])
}
