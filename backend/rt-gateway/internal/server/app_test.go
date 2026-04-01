package server

import "testing"

func TestParseLastSeenVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    int64
		has     bool
		wantErr bool
	}{
		{name: "empty", raw: "", want: 0, has: false, wantErr: false},
		{name: "valid", raw: "12", want: 12, has: true, wantErr: false},
		{name: "invalid", raw: "x", wantErr: true},
		{name: "negative", raw: "-1", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, has, err := parseLastSeenVersion(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error mismatch: got %v wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if has != tt.has {
				t.Fatalf("has mismatch: got %v want %v", has, tt.has)
			}
			if got != tt.want {
				t.Fatalf("value mismatch: got %d want %d", got, tt.want)
			}
		})
	}
}
