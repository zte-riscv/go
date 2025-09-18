package zte

import (
	"testing"
)

func TestCADD(t *testing.T) {
	a := 0
	b := 0
	iter := 10000
	for i := 0; i < iter; i++ {
		a += i
		b += a
	}
	if a != iter*(iter-1)/2 {
		t.Errorf("a is not equal to %d", iter*(iter-1)/2)
	}
	if b != iter*(iter-1)*(iter+1)/6 {
		t.Errorf("b is not equal to %d", iter*(iter-1)*(iter+1)/6)
	}
}
