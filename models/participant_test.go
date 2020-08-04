package models

import "testing"

func TestParticipant_GetFullName(t *testing.T) {
	want := "John Doe"
	participant := Participant{FirstName: "John", LastName: "Doe"}
	if got := participant.GetFullName(); got != want {
		t.Errorf("participant.GetFullName returned %q, wanted %q", got, want)
	}
}

func TestParticipant_GetMarkup(t *testing.T) {
	want := "[John Doe](tg://user?id=123456789)"
	participant := Participant{FirstName: "John", LastName: "Doe", UserId: 123456789}
	if got := participant.GetMarkup(); got != want {
		t.Errorf("participant.GetMarkup returned %q, wanted %q", got, want)
	}
}
