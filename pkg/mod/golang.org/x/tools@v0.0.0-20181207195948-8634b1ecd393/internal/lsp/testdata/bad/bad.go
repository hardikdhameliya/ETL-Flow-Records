//go:build go1.11
// +build go1.11

package bad

func stuff() { //@item(stuff, "stuff()", "", "func")
	x := "heeeeyyyy"
	random2(x) //@diag("x", "cannot use x (variable of type string) as int value in argument to random2")
	random2(1) //@complete("dom", random, random2, random3)
	y := 3     //@diag("y", "y declared but not used")
}

type bob struct { //@item(bob, "bob", "struct{...}", "struct")
	x int
}

func _() {
	var q int
	_ = &bob{
		f: q, //@diag("f", "unknown field f in struct literal")
	}
}
