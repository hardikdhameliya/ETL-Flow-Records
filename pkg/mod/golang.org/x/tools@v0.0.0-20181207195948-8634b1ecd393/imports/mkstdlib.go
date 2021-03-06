//go:build ignore
// +build ignore

// mkstdlib generates the zstdlib.go file, containing the Go standard
// library API symbols. It's baked into the binary to avoid scanning
// GOPATH in the common case.
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

func mustOpen(name string) io.Reader {
	f, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	return f
}

func api(base string) string {
	return filepath.Join(runtime.GOROOT(), "api", base)
}

var sym = regexp.MustCompile(`^pkg (\S+).*?, (?:var|func|type|const) ([A-Z]\w*)`)

var unsafeSyms = map[string]bool{"Alignof": true, "ArbitraryType": true, "Offsetof": true, "Pointer": true, "Sizeof": true}

func main() {
	var buf bytes.Buffer
	outf := func(format string, args ...interface{}) {
		fmt.Fprintf(&buf, format, args...)
	}
	outf("// Code generated by mkstdlib.go. DO NOT EDIT.\n\n")
	outf("package imports\n")
	outf("var stdlib = map[string]map[string]bool{\n")
	f := io.MultiReader(
		mustOpen(api("go1.txt")),
		mustOpen(api("go1.1.txt")),
		mustOpen(api("go1.2.txt")),
		mustOpen(api("go1.3.txt")),
		mustOpen(api("go1.4.txt")),
		mustOpen(api("go1.5.txt")),
		mustOpen(api("go1.6.txt")),
		mustOpen(api("go1.7.txt")),
		mustOpen(api("go1.8.txt")),
		mustOpen(api("go1.9.txt")),
		mustOpen(api("go1.10.txt")),
		mustOpen(api("go1.11.txt")),
	)
	sc := bufio.NewScanner(f)

	pkgs := map[string]map[string]bool{
		"unsafe": unsafeSyms,
	}
	paths := []string{"unsafe"}

	for sc.Scan() {
		l := sc.Text()
		has := func(v string) bool { return strings.Contains(l, v) }
		if has("struct, ") || has("interface, ") || has(", method (") {
			continue
		}
		if m := sym.FindStringSubmatch(l); m != nil {
			path, sym := m[1], m[2]

			if _, ok := pkgs[path]; !ok {
				pkgs[path] = map[string]bool{}
				paths = append(paths, path)
			}
			pkgs[path][sym] = true
		}
	}
	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}
	sort.Strings(paths)
	for _, path := range paths {
		outf("\t%q: map[string]bool{\n", path)
		pkg := pkgs[path]
		var syms []string
		for sym := range pkg {
			syms = append(syms, sym)
		}
		sort.Strings(syms)
		for _, sym := range syms {
			outf("\t\t%q: true,\n", sym)
		}
		outf("},\n")
	}
	outf("}\n")
	fmtbuf, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("zstdlib.go", fmtbuf, 0666)
	if err != nil {
		log.Fatal(err)
	}
}
