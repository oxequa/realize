package realize

import (
	"gopkg.in/urfave/cli.v2"
	"testing"
)

var context *cli.Context

func TestNew(t *testing.T) {
	actual := New(context)
	expected := &Config{file:AppFile,Version: AppVersion}
	if actual == expected {
		t.Errorf("Test failed, expected: '%s', got:  '%s'",expected, actual)
	}
}