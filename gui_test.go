package main

import (
	"testing"

	"github.com/lxn/walk"
)

func TestWalkMainWindowInit(t *testing.T) {
	mw, err := walk.NewMainWindow()
	if err != nil {
		t.Fatal(err)
	}
	mw.Dispose()
}
