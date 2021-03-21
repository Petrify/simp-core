package commands

import (
	"context"
	"testing"
)

type CtxKey string

func printSuccessFunc(T *testing.T) (f func(ctx context.Context, args []string)) {

	f = func(ctx context.Context, args []string) {
		T.Log("Command func test")
	}

	return
}

func Test_Interpreter(T *testing.T) {

	I := NewInterpreter()

	ctx := context.WithValue(context.Background(), CtxKey("val"), "Context Value Test string")

	T.Log("Adding command `test c1`")
	if e := I.AddCommand("test c1", printSuccessFunc(T)); e != nil {
		T.Error(e)
	}
	T.Log("Success")

	T.Log("Running command `test c1`")
	if e := I.Run(ctx, "test c1"); e != nil {
		T.Error(e)
	}
	T.Log("Success")

	T.Log("Running Invalid command `test c2`")
	if e := I.Run(ctx, "test c2"); e != nil {
		T.Log(e)
		T.Log("Success")
	} else {
		T.Error("Invalid command did not throw error")
	}

}

func Test_Create(T *testing.T) {
	NewInterpreter()
}
