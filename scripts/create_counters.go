package main

import (
	"bufio"
	"fmt"
	"os"

	"golang.org/x/tools/go/packages"
)

func main() {
	cfg := &packages.Config{
		Mode: packages.NeedSyntax,
	}
	pkgs, err := packages.Load(cfg, "github.com/traffic-refinery/traffic-refinery/internal/counters")
	if err != nil {
		panic(err)
	}

	f, _ := os.Create("../internal/counters/init_counters.go")
	defer f.Close()
	w := bufio.NewWriter(f)
	_, _ = w.WriteString("package counters\n\nfunc init() {\n")

	for _, pkg := range pkgs {
		for _, s := range pkg.Syntax {
			for _, o := range s.Scope.Objects {
				if o.Kind.String() == "type" {
					w.WriteString(fmt.Sprintf("\tregisterType((*%s)(nil))\n", o.Name))
				}
			}
		}
	}
	_, _ = w.WriteString("}\n")

	w.Flush()
}
