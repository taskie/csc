package main

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/taskie/csc/cli/cscman"
)

func main() {
	cscman.Main()
}
