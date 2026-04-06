package resolver

import "testing"

func TestNormalizeURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "adds https scheme", input: "example.com", want: "https://example.com/"},
		{name: "keeps https scheme", input: "https://example.com/docs", want: "https://example.com/docs"},
		{name: "keeps http scheme", input: "http://example.com", want: "http://example.com/"},
		{name: "rejects empty", input: "  ", wantErr: true},
		{name: "rejects unsupported scheme", input: "ftp://example.com", wantErr: true},
		{name: "rejects missing host", input: "https:///path", wantErr: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := NormalizeURL(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeURL returned error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("NormalizeURL = %q, want %q", got, tc.want)
			}
		})
	}
}
