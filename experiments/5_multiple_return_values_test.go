package experiments

import "testing"

func vals() (int, int) {
	return 3, 7
}

func TestMultipleReturnValues(t *testing.T) {
	a, b := vals()
	expectedA, expectedB := 3, 7
	if a != expectedA || b != expectedB {
		t.Errorf("Expected %d and %d, but got %d and %d", expectedA, expectedB, a, b)
	}
}
