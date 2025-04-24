package main

import "testing"

func Test(t *testing.T) {
	if Foo() != 5 {
		t.Fail()
	}
}
