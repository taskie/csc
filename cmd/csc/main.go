package main

import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/taskie/csc/cli/csc"
)

var (
	version, commit, date string
)

func main() {
	csc.Main()
}
