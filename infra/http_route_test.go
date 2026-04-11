package infra

import "testing"

func TestStatusCodeClass(t *testing.T) {
	testCases := []struct {
		name   string
		status int
		want   string
	}{
		{name: "2xx", status: 200, want: "2xx"},
		{name: "4xx", status: 404, want: "4xx"},
		{name: "5xx", status: 500, want: "5xx"},
		{name: "unknown low", status: 99, want: "unknown"},
		{name: "unknown high", status: 1000, want: "unknown"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := statusCodeClass(tc.status); got != tc.want {
				t.Fatalf("statusCodeClass(%d) = %q, want %q", tc.status, got, tc.want)
			}
		})
	}
}
