package main

import "testing"

func TestDummy(t *testing.T) {
  success := true
  if success != true {
    t.Error("Expected true, but got ", success)
  }
}