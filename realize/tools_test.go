package realize

import "testing"

func TestTools_Setup(t *testing.T) {
	tools := Tools{
		Clean: Tool{
			Status: true,
			name:   "test",
			isTool: false,
			Method: "test",
			Args:   []string{"arg"},
		},
	}
	tools.Setup()
	if tools.Clean.name == "test" {
		t.Error("Unexpected value")
	}
	if tools.Clean.Method != "test" {
		t.Error("Unexpected value")
	}
	if !tools.Clean.isTool {
		t.Error("Unexpected value")
	}
}
