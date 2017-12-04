// +build !windows

package realize

import (
	"testing"
)

func TestSettings_Flimit(t *testing.T) {
	s := Settings{}
	s.FileLimit = 100
	if err := s.Flimit(); err != nil {
		t.Error("Unable to increase limit", err)
	}
}
