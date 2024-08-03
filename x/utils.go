package x

import "log"

func AssertTrue(cond bool) {
	if !cond {
		log.Fatal("[AssertTrue]:failed")
	}
}
