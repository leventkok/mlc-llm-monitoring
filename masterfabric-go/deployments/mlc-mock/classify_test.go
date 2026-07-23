package main

import "testing"

func TestClassify(t *testing.T) {
	tests := []struct {
		text              string
		wantCat, wantSent string
	}{
		{"app keeps crashing", "bug", "negative"},
		{"please add dark mode feature", "feature", "neutral"},
		{"love this app, great work", "praise", "positive"},
		{"kötü yorum", "other", "negative"},
		{"berbat bir uygulama", "other", "negative"},
		{"uygulama çöküyor", "bug", "negative"},
		{"harika bir deneyim", "praise", "positive"},
		{"just ok", "other", "neutral"},
		{"fake review spam click here", "spam", "negative"},
	}

	for _, tt := range tests {
		gotCat, gotSent := classify(tt.text)
		if gotCat != tt.wantCat || gotSent != tt.wantSent {
			t.Errorf("classify(%q) = (%q, %q), want (%q, %q)", tt.text, gotCat, gotSent, tt.wantCat, tt.wantSent)
		}
	}
}
