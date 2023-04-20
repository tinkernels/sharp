package main

import "testing"

func TestIterWithControlBit(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}

	// Test with ctlBit = true
	iter := IterWithControlBit(slice, true)
	expected := []int{1, 2, 3, 4, 5, 5, 4, 3, 2, 1, 1}
	for i, val := range expected {
		if iter() != val {
			t.Errorf("Expected %d, but got %d at index %d", val, iter(), i)
		}
	}

	// Test with ctlBit = false
	iter = IterWithControlBit(slice, false)
	expected = []int{5, 4, 3, 2, 1, 1, 2, 3, 4, 5, 5}
	for i, val := range expected {
		if iter() != val {
			t.Errorf("Expected %d, but got %d at index %d", val, iter(), i)
		}
	}
}
