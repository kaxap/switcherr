package main

import (
	"flag"
	"fmt"
	"go/token"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/packages"
	"log"
	"os"
	"strings"
	"switcherr/pkg/switcherr"
	"sync"
)

func getPackages(patterns []string) ([]*packages.Package, error) {
	conf := &packages.Config{
		Mode:  packages.LoadAllSyntax,
		Tests: true,
	}
	pkgs, err := packages.Load(conf, patterns...)
	if err != nil {
		return nil, err
	}

	return pkgs, nil

}

func main() {
	log.SetFlags(0)
	log.SetPrefix(switcherr.Analyzer.Name + ": ")

	flag.Usage = func() {
		paras := strings.Split(switcherr.Analyzer.Doc, "\n\n")
		fmt.Fprintf(os.Stderr, "%s: %s\n\n", switcherr.Analyzer.Name, paras[0])
		fmt.Fprintf(os.Stderr, "Usage: %s [-flag] [package]\n\n", switcherr.Analyzer.Name)
		if len(paras) > 1 {
			fmt.Fprintln(os.Stderr, strings.Join(paras[1:], "\n\n"))
		}
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flag.PrintDefaults()
	}
	flag.Parse()

	patterns := flag.Args()
	if len(patterns) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	pkgs, err := getPackages(patterns)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	diagnostics := make([][]analysis.Diagnostic, len(pkgs))
	var wg sync.WaitGroup
	for i := range pkgs {
		wg.Add(1)
		go func(i int, pkg *packages.Package) {
			_, err := switcherr.Analyzer.Run(&analysis.Pass{
				Analyzer:     nil,
				Fset:         pkg.Fset,
				Files:        pkg.Syntax,
				OtherFiles:   nil,
				IgnoredFiles: nil,
				Pkg:          nil,
				TypesInfo:    pkg.TypesInfo,
				TypesSizes:   nil,
				TypeErrors:   nil,
				Report: func(d analysis.Diagnostic) {
					diagnostics[i] = append(diagnostics[i], d)
				},
			})
			if err != nil {
				log.Printf("Analyzer error: %s", err)
			}
			wg.Done()
		}(i, pkgs[i])
	}
	wg.Wait()

	exitcode := 0
	seen := map[token.Pos]struct{}{}
	for i, pkg := range pkgs {
		for _, d := range diagnostics[i] {
			if _, ok := seen[d.Pos]; ok {
				continue
			}
			pos := pkg.Fset.Position(d.Pos)
			_, _ = fmt.Fprintf(os.Stderr, "%s: %s\n", pos, d.Message)
			exitcode = 3
			seen[d.Pos] = struct{}{}
		}
	}
	os.Exit(exitcode)
}
