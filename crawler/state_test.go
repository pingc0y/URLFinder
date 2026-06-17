package crawler

import "testing"

func TestStatusCodeSelectedMatchesExactCodes(t *testing.T) {
	tests := []struct {
		name   string
		filter string
		code   int
		want   bool
	}{
		{name: "exact match", filter: "200,301", code: 200, want: true},
		{name: "exact match with spaces", filter: " 200, 301 ", code: 301, want: true},
		{name: "substring does not match", filter: "20", code: 200, want: false},
		{name: "all matches any status", filter: "all", code: 404, want: true},
		{name: "empty matches nothing", filter: "", code: 200, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := statusCodeSelected(tt.filter, tt.code); got != tt.want {
				t.Fatalf("statusCodeSelected(%q, %d) = %v, want %v", tt.filter, tt.code, got, tt.want)
			}
		})
	}
}
