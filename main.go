package main

import (
	"fmt"
	"runtime"

	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/igor-sirotin/returnearly/returnearly"
)

func main() {
	fmt.Println(runtime.GOROOT())
	singlechecker.Main(returnearly.Analyzer)
}
