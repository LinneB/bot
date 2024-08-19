package utils

import (
	"testing"
	"time"
)

func TestPrettyDuration(t *testing.T) {
	expected := "1 second"
	actual := PrettyDuration(1 * time.Second)
	if actual != expected {
		t.Errorf("Expected %s; Got %s", expected, actual)
	}
	expected = "1 hour"
	actual = PrettyDuration(4000 * time.Second)
	if actual != expected {
		t.Errorf("Expected %s; Got %s", expected, actual)
	}
	expected = "2 months"
	actual = PrettyDuration(70 * 24 * time.Hour)
	if actual != expected {
		t.Errorf("Expected %s; Got %s", expected, actual)
	}
}
