package simpsql

import "testing"

func Test_Read(t *testing.T) {
	s, err := Open("test")
	if err != nil {
		t.Error(err)
	}

	for s.Next() {
		t.Log(s.Stmt())
	}
}
