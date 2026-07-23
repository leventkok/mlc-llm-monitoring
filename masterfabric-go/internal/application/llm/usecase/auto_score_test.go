package usecase

import "testing"

func TestComputeAutoQuality(t *testing.T) {
	tests := []struct {
		name      string
		category  string
		sentiment string
		raw       string
		latency   int
		wantMin   int
		wantMax   int
	}{
		{
			name: "valid json matching decision fast",
			category:  "praise",
			sentiment: "positive",
			raw:       `{"category":"praise","sentiment":"positive"}`,
			latency:   150,
			wantMin:   5,
			wantMax:   5,
		},
		{
			name:      "invalid category",
			category:  "invalid",
			sentiment: "positive",
			raw:       `{}`,
			latency:   100,
			wantMin:   1,
			wantMax:   1,
		},
		{
			name:      "empty raw output",
			category:  "bug",
			sentiment: "negative",
			raw:       "",
			latency:   100,
			wantMin:   2,
			wantMax:   2,
		},
		{
			name:      "slow inference penalty",
			category:  "feature",
			sentiment: "neutral",
			raw:       `{"category":"feature","sentiment":"neutral"}`,
			latency:   12_000,
			wantMin:   1,
			wantMax:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeAutoQuality(tt.category, tt.sentiment, tt.raw, tt.latency)
			if got < tt.wantMin || got > tt.wantMax {
				t.Fatalf("ComputeAutoQuality() = %d, want between %d and %d", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}
