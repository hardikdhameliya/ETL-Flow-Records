// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore
// +build ignore

// Generate system call table for DragonFly, NetBSD,
// FreeBSD, OpenBSD or Darwin from master list
// (for example, /usr/src/sys/kern/syscalls.master or
// sys/syscall.h).
package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var (
	goos, goarch string
)

// cmdLine returns this programs's commandline arguments
func cmdLine() string {
	return "go run mksysnum.go " + strings.Join(os.Args[1:], " ")
}

// buildTags returns build tags
func buildTags() string {
	return fmt.Sprintf("%s,%s", goarch, goos)
}

func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// source string and substring slice for regexp
type re struct {
	str string   // source string
	sub []string // matched sub-string
}

// Match performs regular expression match
func (r *re) Match(exp string) bool {
	r.sub = regexp.MustCompile(exp).FindStringSubmatch(r.str)
	if r.sub != nil {
		return true
	}
	return false
}

// fetchFile fetches a text file from URL
func fetchFile(URL string) io.Reader {
	resp, err := http.Get(URL)
	checkErr(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	checkErr(err)
	return strings.NewReader(string(body))
}

// readFile reads a text file from path
func readFile(path string) io.Reader {
	file, err := os.Open(os.Args[1])
	checkErr(err)
	return file
}

func format(name, num, proto string) string {
	name = strings.ToUpper(name)
	// There are multiple entries for enosys and nosys, so comment them out.
	nm := re{str: name}
	if nm.Match(`^SYS_E?NOSYS$`) {
		name = fmt.Sprintf("// %s", name)
	}
	if name == `SYS_SYS_EXIT` {
		name = `SYS_EXIT`
	}
	return fmt.Sprintf("	%s = %s;  // %s\n", name, num, proto)
}

func main() {
	// Get the OS (using GOOS_TARGET if it exist)
	goos = os.Getenv("GOOS_TARGET")
	if goos == "" {
		goos = os.Getenv("GOOS")
	}
	// Get the architecture (using GOARCH_TARGET if it exists)
	goarch = os.Getenv("GOARCH_TARGET")
	if goarch == "" {
		goarch = os.Getenv("GOARCH")
	}
	// Check if GOOS and GOARCH environment variables are defined
	if goarch == "" || goos == "" {
		fmt.Fprintf(os.Stderr, "GOARCH or GOOS not defined in environment\n")
		os.Exit(1)
	}

	file := strings.TrimSpace(os.Args[1])
	var syscalls io.Reader
	if strings.HasPrefix(file, "https://") || strings.HasPrefix(file, "http://") {
		// Download syscalls.master file
		syscalls = fetchFile(file)
	} else {
		syscalls = readFile(file)
	}

	var text, line string
	s := bufio.NewScanner(syscalls)
	for s.Scan() {
		t := re{str: line}
		if t.Match(`^(.*)\\$`) {
			// Handle continuation
			line = t.sub[1]
			line += strings.TrimLeft(s.Text(), " \t")
		} else {
			// New line
			line = s.Text()
		}
		t = re{str: line}
		if t.Match(`\\$`) {
			continue
		}
		t = re{str: line}

		switch goos {
		case "dragonfly":
			if t.Match(`^([0-9]+)\s+STD\s+({ \S+\s+(\w+).*)$`) {
				num, proto := t.sub[1], t.sub[2]
				name := fmt.Sprintf("SYS_%s", t.sub[3])
				text += format(name, num, proto)
			}
		case "freebsd":
			if t.Match(`^([0-9]+)\s+\S+\s+(?:(?:NO)?STD|COMPAT10)\s+({ \S+\s+(\w+).*)$`) {
				num, proto := t.sub[1], t.sub[2]
				name := fmt.Sprintf("SYS_%s", t.sub[3])
				text += format(name, num, proto)
			}
		case "openbsd":
			if t.Match(`^([0-9]+)\s+STD\s+(NOLOCK\s+)?({ \S+\s+\*?(\w+).*)$`) {
				num, proto, name := t.sub[1], t.sub[3], t.sub[4]
				text += format(name, num, proto)
			}
		case "netbsd":
			if t.Match(`^([0-9]+)\s+((STD)|(NOERR))\s+(RUMP\s+)?({\s+\S+\s*\*?\s*\|(\S+)\|(\S*)\|(\w+).*\s+})(\s+(\S+))?$`) {
				num, proto, compat := t.sub[1], t.sub[6], t.sub[8]
				name := t.sub[7] + "_" + t.sub[9]
				if t.sub[11] != "" {
					name = t.sub[7] + "_" + t.sub[11]
				}
				name = strings.ToUpper(name)
				if compat == "" || compat == "13" || compat == "30" || compat == "50" {
					text += fmt.Sprintf("	%s = %s;  // %s\n", name, num, proto)
				}
			}
		case "darwin":
			if t.Match(`^#define\s+SYS_(\w+)\s+([0-9]+)`) {
				name, num := t.sub[1], t.sub[2]
				name = strings.ToUpper(name)
				text += fmt.Sprintf("	SYS_%s = %s;\n", name, num)
			}
		default:
			fmt.Fprintf(os.Stderr, "unrecognized GOOS=%s\n", goos)
			os.Exit(1)

		}
	}
	err := s.Err()
	checkErr(err)

	fmt.Printf(template, cmdLine(), buildTags(), text)
}

const template = `// %s
// Code generated by the command above; see README.md. DO NOT EDIT.

// +build %s

package unix

const(
%s)`
