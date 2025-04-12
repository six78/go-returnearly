package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/igor-sirotin/returnearly/returnearly"
)

func main() {
	multichecker.Main(returnearly.Analyzer)
}
