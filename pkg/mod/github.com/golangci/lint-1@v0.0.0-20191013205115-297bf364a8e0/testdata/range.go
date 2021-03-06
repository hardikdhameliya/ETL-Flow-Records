// Test for range construction.

// Package foo ...
package foo

func f() {
	var m map[string]int

	// with :=
	for x := range m { // MATCH /should omit 2nd value.*range.*equivalent.*for x := range/ -> `	for x := range m {`
		_ = x
	}
	// with =
	var y string
	_ = y
	for y = range m { // MATCH /should omit 2nd value.*range.*equivalent.*for y = range/
	}

	for range m { // MATCH /should omit values.*range.*equivalent.*for range/
	}

	for range m { // MATCH /should omit values.*range.*equivalent.*for range/
	}

	// all OK:
	for x := range m {
		_ = x
	}
	for x, y := range m {
		_, _ = x, y
	}
	for _, y := range m {
		_ = y
	}
	var x int
	_ = x
	for y = range m {
	}
	for y, x = range m {
	}
	for _, x = range m {
	}
}
