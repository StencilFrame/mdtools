package main

import (
	"os"
	"testing"
)

func TestMain(t *testing.T) {
	os.Args = []string{"cmd", "testdata/test_input.md"}
	main()
}
