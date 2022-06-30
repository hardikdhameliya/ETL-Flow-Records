// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Support for testing against external disassembler program.
// Copied and simplified from rsc.io/arm/armasm/ext_test.go.

package ppc64asm

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

var (
	printTests = flag.Bool("printtests", false, "print test cases that exercise new code paths")
	dumpTest   = flag.Bool("dump", false, "dump all encodings")
	mismatch   = flag.Bool("mismatch", false, "log allowed mismatches")
	longTest   = flag.Bool("long", false, "long test")
	keep       = flag.Bool("keep", false, "keep object files around")
	debug      = false
)

// An ExtInst represents a single decoded instruction parsed
// from an external disassembler's output.
type ExtInst struct {
	addr uint32
	enc  [4]byte
	nenc int
	text string
}

func (r ExtInst) String() string {
	return fmt.Sprintf("%#x: % x: %s", r.addr, r.enc, r.text)
}

// An ExtDis is a connection between an external disassembler and a test.
type ExtDis struct {
	Dec      chan ExtInst
	File     *os.File
	Size     int
	KeepFile bool
	Cmd      *exec.Cmd
}

// Run runs the given command - the external disassembler - and returns
// a buffered reader of its standard output.
func (ext *ExtDis) Run(cmd ...string) (*bufio.Reader, error) {
	if *keep {
		log.Printf("%s\n", strings.Join(cmd, " "))
	}
	ext.Cmd = exec.Command(cmd[0], cmd[1:]...)
	out, err := ext.Cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdoutpipe: %v", err)
	}
	if err := ext.Cmd.Start(); err != nil {
		return nil, fmt.Errorf("exec: %v", err)
	}

	b := bufio.NewReaderSize(out, 1<<20)
	return b, nil
}

// Wait waits for the command started with Run to exit.
func (ext *ExtDis) Wait() error {
	return ext.Cmd.Wait()
}

// testExtDis tests a set of byte sequences against an external disassembler.
// The disassembler is expected to produce the given syntax and be run
// in the given architecture mode (16, 32, or 64-bit).
// The extdis function must start the external disassembler
// and then parse its output, sending the parsed instructions on ext.Dec.
// The generate function calls its argument f once for each byte sequence
// to be tested. The generate function itself will be called twice, and it must
// make the same sequence of calls to f each time.
// When a disassembly does not match the internal decoding,
// allowedMismatch determines whether this mismatch should be
// allowed, or else considered an error.
func testExtDis(
	t *testing.T,
	syntax string,
	extdis func(ext *ExtDis) error,
	generate func(f func([]byte)),
	allowedMismatch func(text string, size int, inst *Inst, dec ExtInst) bool,
) {
	start := time.Now()
	ext := &ExtDis{
		Dec: make(chan ExtInst),
	}
	errc := make(chan error)

	// First pass: write instructions to input file for external disassembler.
	file, f, size, err := writeInst(generate)
	if err != nil {
		t.Fatal(err)
	}
	ext.Size = size
	ext.File = f
	defer func() {
		f.Close()
		if !*keep {
			os.Remove(file)
		}
	}()

	// Second pass: compare disassembly against our decodings.
	var (
		totalTests  = 0
		totalSkips  = 0
		totalErrors = 0

		errors = make([]string, 0, 100) // sampled errors, at most cap
	)
	go func() {
		errc <- extdis(ext)
	}()
	generate(func(enc []byte) {
		dec, ok := <-ext.Dec
		if !ok {
			t.Errorf("decoding stream ended early")
			return
		}
		inst, text := disasm(syntax, pad(enc))
		totalTests++
		if *dumpTest {
			fmt.Printf("%x -> %s [%d]\n", enc[:], dec.text, dec.nenc)
		}
		if text != dec.text || inst.Len != dec.nenc {
			suffix := ""
			if allowedMismatch(text, size, &inst, dec) {
				totalSkips++
				if !*mismatch {
					return
				}
				suffix += " (allowed mismatch)"
			}
			totalErrors++
			if len(errors) >= cap(errors) {
				j := rand.Intn(totalErrors)
				if j >= cap(errors) {
					return
				}
				errors = append(errors[:j], errors[j+1:]...)
			}
			errors = append(errors, fmt.Sprintf("decode(%x) = %q, %d, want %q, %d%s", enc, text, inst.Len, dec.text, dec.nenc, suffix))
		}
	})

	if *mismatch {
		totalErrors -= totalSkips
	}

	for _, b := range errors {
		t.Log(b)
	}

	if totalErrors > 0 {
		t.Fail()
	}
	t.Logf("%d test cases, %d expected mismatches, %d failures; %.0f cases/second", totalTests, totalSkips, totalErrors, float64(totalTests)/time.Since(start).Seconds())

	if err := <-errc; err != nil {
		t.Fatalf("external disassembler: %v", err)
	}

}

const start = 0x8000 // start address of text

// writeInst writes the generated byte sequences to a new file
// starting at offset start. That file is intended to be the input to
// the external disassembler.
func writeInst(generate func(func([]byte))) (file string, f *os.File, size int, err error) {
	f, err = ioutil.TempFile("", "ppc64asm")
	if err != nil {
		return
	}

	file = f.Name()

	f.Seek(start, io.SeekStart)
	w := bufio.NewWriter(f)
	defer w.Flush()
	size = 0
	generate(func(x []byte) {
		if len(x) > 4 {
			x = x[:4]
		}
		if debug {
			fmt.Printf("%#x: %x%x\n", start+size, x, zeros[len(x):])
		}
		w.Write(x)
		w.Write(zeros[len(x):])
		size += len(zeros)
	})
	return file, f, size, nil
}

var zeros = []byte{0, 0, 0, 0}

// pad pads the code sequence with pops.
func pad(enc []byte) []byte {
	if len(enc) < 4 {
		enc = append(enc[:len(enc):len(enc)], zeros[:4-len(enc)]...)
	}
	return enc
}

// disasm returns the decoded instruction and text
// for the given source bytes, using the given syntax and mode.
func disasm(syntax string, src []byte) (inst Inst, text string) {
	// If printTests is set, we record the coverage value
	// before and after, and we write out the inputs for which
	// coverage went up, in the format expected in testdata/decode.text.
	// This produces a fairly small set of test cases that exercise nearly
	// all the code.
	var cover float64
	if *printTests {
		cover -= coverage()
	}

	inst, err := Decode(src, binary.BigEndian)
	if err != nil {
		text = "error: " + err.Error()
	} else {
		text = inst.String()
		switch syntax {
		//case "arm":
		//	text = ARMSyntax(inst)
		case "gnu":
			text = GNUSyntax(inst, 0)
		//case "plan9":
		//	text = GoSyntax(inst, 0, nil)
		default:
			text = "error: unknown syntax " + syntax
		}
	}

	if *printTests {
		cover += coverage()
		if cover > 0 {
			max := len(src)
			if max > 4 && inst.Len <= 4 {
				max = 4
			}
			fmt.Printf("%x|%x\t%s\t%s\n", src[:inst.Len], src[inst.Len:max], syntax, text)
		}
	}

	return
}

// coverage returns a floating point number denoting the
// test coverage until now. The number increases when new code paths are exercised,
// both in the Go program and in the decoder byte code.
func coverage() float64 {
	var f float64
	f += testing.Coverage()
	f += decodeCoverage()
	return f
}

func decodeCoverage() float64 {
	n := 0
	for _, t := range decoderCover {
		if t {
			n++
		}
	}
	return float64(1+n) / float64(1+len(decoderCover))
}

// Helpers for writing disassembler output parsers.

// hasPrefix reports whether any of the space-separated words in the text s
// begins with any of the given prefixes.
func hasPrefix(s string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		for s := s; s != ""; {
			if strings.HasPrefix(s, prefix) {
				return true
			}
			i := strings.Index(s, " ")
			if i < 0 {
				break
			}
			s = s[i+1:]
		}
	}
	return false
}

// contains reports whether the text s contains any of the given substrings.
func contains(s string, substrings ...string) bool {
	for _, sub := range substrings {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// isHex reports whether b is a hexadecimal character (0-9A-Fa-f).
func isHex(b byte) bool { return b == '0' || unhex[b] > 0 }

// parseHex parses the hexadecimal byte dump in hex,
// appending the parsed bytes to raw and returning the updated slice.
// The returned bool signals whether any invalid hex was found.
// Spaces and tabs between bytes are okay but any other non-hex is not.
func parseHex(hex []byte, raw []byte) ([]byte, bool) {
	hex = trimSpace(hex)
	for j := 0; j < len(hex); {
		for hex[j] == ' ' || hex[j] == '\t' {
			j++
		}
		if j >= len(hex) {
			break
		}
		if j+2 > len(hex) || !isHex(hex[j]) || !isHex(hex[j+1]) {
			return nil, false
		}
		raw = append(raw, unhex[hex[j]]<<4|unhex[hex[j+1]])
		j += 2
	}
	return raw, true
}

var unhex = [256]byte{
	'0': 0,
	'1': 1,
	'2': 2,
	'3': 3,
	'4': 4,
	'5': 5,
	'6': 6,
	'7': 7,
	'8': 8,
	'9': 9,
	'A': 10,
	'B': 11,
	'C': 12,
	'D': 13,
	'E': 14,
	'F': 15,
	'a': 10,
	'b': 11,
	'c': 12,
	'd': 13,
	'e': 14,
	'f': 15,
}

// index is like bytes.Index(s, []byte(t)) but avoids the allocation.
func index(s []byte, t string) int {
	i := 0
	for {
		j := bytes.IndexByte(s[i:], t[0])
		if j < 0 {
			return -1
		}
		i = i + j
		if i+len(t) > len(s) {
			return -1
		}
		for k := 1; k < len(t); k++ {
			if s[i+k] != t[k] {
				goto nomatch
			}
		}
		return i
	nomatch:
		i++
	}
}

// fixSpace rewrites runs of spaces, tabs, and newline characters into single spaces in s.
// If s must be rewritten, it is rewritten in place.
func fixSpace(s []byte) []byte {
	s = trimSpace(s)
	for i := 0; i < len(s); i++ {
		if s[i] == '\t' || s[i] == '\n' || i > 0 && s[i] == ' ' && s[i-1] == ' ' {
			goto Fix
		}
	}
	return s

Fix:
	b := s
	w := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\t' || c == '\n' {
			c = ' '
		}
		if c == ' ' && w > 0 && b[w-1] == ' ' {
			continue
		}
		b[w] = c
		w++
	}
	if w > 0 && b[w-1] == ' ' {
		w--
	}
	return b[:w]
}

// trimSpace trims leading and trailing space from s, returning a subslice of s.
func trimSpace(s []byte) []byte {
	j := len(s)
	for j > 0 && (s[j-1] == ' ' || s[j-1] == '\t' || s[j-1] == '\n') {
		j--
	}
	i := 0
	for i < j && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	return s[i:j]
}

// pcrel matches instructions using relative addressing mode.
var (
	pcrel = regexp.MustCompile(`^((?:.* )?(?:b|bc)[^ac ]* (?:(?:[0-9]{1,2},)|(?:[0-7]\*)|\+|lt|gt|eq|so|cr[0-7]|,)*)0x([0-9a-f]+)$`)
)

// Generators.
//
// The test cases are described as functions that invoke a callback repeatedly,
// with a new input sequence each time. These helpers make writing those
// a little easier.

// randomCases generates random instructions.
func randomCases(t *testing.T) func(func([]byte)) {
	return func(try func([]byte)) {
		// All the strides are relatively prime to 2 and therefore to 2²⁸,
		// so we will not repeat any instructions until we have tried all 2²⁸.
		// Using a stride other than 1 is meant to visit the instructions in a
		// pseudorandom order, which gives better variety in the set of
		// test cases chosen by -printtests.
		stride := uint32(10007)
		n := 1 << 28 / 7
		if testing.Short() {
			stride = 100003
			n = 1 << 28 / 1001
		} else if *longTest {
			stride = 2000033
			n = 1 << 29
		}
		x := uint32(0)
		for i := 0; i < n; i++ {
			enc := (x%15)<<28 | x&(1<<28-1)
			try([]byte{byte(enc), byte(enc >> 8), byte(enc >> 16), byte(enc >> 24)})
			x += stride
		}
	}
}

// hexCases generates the cases written in hexadecimal in the encoded string.
// Spaces in 'encoded' separate entire test cases, not individual bytes.
func hexCases(t *testing.T, encoded string) func(func([]byte)) {
	return func(try func([]byte)) {
		for _, x := range strings.Fields(encoded) {
			src, err := hex.DecodeString(x)
			if err != nil {
				t.Errorf("parsing %q: %v", x, err)
			}
			try(src)
		}
	}
}

// testdataCases generates the test cases recorded in testdata/decode.txt.
// It only uses the inputs; it ignores the answers recorded in that file.
func testdataCases(t *testing.T) func(func([]byte)) {
	var codes [][]byte
	data, err := ioutil.ReadFile("testdata/decode.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Fields(line)[0]
		i := strings.Index(f, "|")
		if i < 0 {
			t.Errorf("parsing %q: missing | separator", f)
			continue
		}
		if i%2 != 0 {
			t.Errorf("parsing %q: misaligned | separator", f)
		}
		code, err := hex.DecodeString(f[:i] + f[i+1:])
		if err != nil {
			t.Errorf("parsing %q: %v", f, err)
			continue
		}
		codes = append(codes, code)
	}

	return func(try func([]byte)) {
		for _, code := range codes {
			try(code)
		}
	}
}

func caller(skip int) string {
	pc, _, _, _ := runtime.Caller(skip)
	f := runtime.FuncForPC(pc)
	name := "?"
	if f != nil {
		name = f.Name()
		if i := strings.LastIndex(name, "."); i >= 0 {
			name = name[i+1:]
		}
	}
	return name
}
