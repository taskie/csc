package main

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/taskie/csc/cli/cscman"
)

var (
	version, commit, date string
)

func main() {
	cscman.Main()
}
