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

func TestSplitStreamOnlineMessage(t *testing.T) {
	liveMessage := "This streamer is now live!"
	users := []string{"user1", "user2", "user3", "user4", "user5", "user6", "user7"}
	expected := []string{
		"This streamer is now live! @user1 @user2",
		"@user3 @user4 @user5 @user6 @user7",
	}
	actual := SplitStreamOnlineMessage(liveMessage, users, 40)

	if len(actual) != len(expected) {
		t.Errorf("Expected len(2); Got len(%d)", len(actual))
	}
	for i, message := range actual {
		if message != expected[i] {
			t.Errorf("Expected %s; Got %s", expected[i], message)
		}
	}
}

func TestCapitalizeFirstCharacter(t *testing.T) {
	input := "hello!"
	expected := "Hello!"
	actual := CapitalizeFirstCharacter(input)
	if actual != expected {
		t.Errorf("Expected %s; Got %s", expected, actual)
	}
	input = "åäö"
	expected = "Åäö"
	actual = CapitalizeFirstCharacter(input)
	if actual != expected {
		t.Errorf("Expected %s; Got %s", expected, actual)
	}
}
