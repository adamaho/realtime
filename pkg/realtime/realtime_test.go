package realtime

import "testing"

func add(a int, b int) int {
	return a + b
}

// asdfasdfasdfasdf
func TestAdds2Plus2(t *testing.T) {
	result := add(2, 2)
	expected := 4

	if result != expected {
		t.Fatalf("expected add(2,2) to equal %d, received %d", result, expected)
	}
}
