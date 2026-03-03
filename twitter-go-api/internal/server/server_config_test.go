package server

import (
	"reflect"
	"testing"
)

func TestParseTrustedProxies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want []string
	}{
		{name: "empty", in: "", want: nil},
		{name: "spaces only", in: "   ", want: nil},
		{name: "single", in: "10.0.0.1", want: []string{"10.0.0.1"}},
		{name: "multiple", in: "10.0.0.1, 10.0.0.2,127.0.0.1", want: []string{"10.0.0.1", "10.0.0.2", "127.0.0.1"}},
		{name: "with empty entries", in: "10.0.0.1, , ,127.0.0.1", want: []string{"10.0.0.1", "127.0.0.1"}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseTrustedProxies(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("parseTrustedProxies(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
