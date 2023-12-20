package main

import (
	"bufio"
	"fmt"
	"os"

	"golang.org/x/tools/go/packages"
)

func main() {
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}
	pkgs, err := packages.Load(cfg, "github.com/traffic-refinery/traffic-refinery/internal/counters")
	if err != nil {
		panic(err)
	}

	f, err := os.Create("internal/counters/init_counters.go")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	_, _ = w.WriteString("package counters\n\nfunc init() {\n")

	for _, pkg := range pkgs {
		fmt.Printf("Package %s\n", pkg.ID)
		fmt.Printf("Syntax %d\n", len(pkg.Syntax))
		for _, s := range pkg.Syntax {
			fmt.Printf("Syntax %d\n", s.Package)
			fmt.Printf("Object %d\n", len(s.Scope.Objects))
			for _, o := range s.Scope.Objects {
				fmt.Printf("Object %s\n", o.Kind.String())
				if o.Kind.String() == "type" {
					fmt.Printf("Object %s\n", o.Name)
					fmt.Printf("Object %s\n", fmt.Sprintf("\tregisterType((*%s)(nil))", o.Name))
					w.WriteString(fmt.Sprintf("\tregisterType((*%s)(nil))\n", o.Name))
				}
			}
		}
	}
	_, _ = w.WriteString("}\n")

	w.Flush()
}
