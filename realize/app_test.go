package realize

import (
	"reflect"
	"testing"
)

func TestInit(t *testing.T) {
	actual := Init()
	expected := &App{Name: AppName, Version: AppVersion, Description: AppDescription, Author: AppAuthor, Email: AppEmail}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Test failed, expected: '%s', got:  '%s'", expected, actual)
	}
}
