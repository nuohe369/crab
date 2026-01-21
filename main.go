package main

import (
	"github.com/nuohe369/crab/boot"
	_ "github.com/nuohe369/crab/module/testapi" // auto-register module
)

func main() {
	boot.Execute()
}
