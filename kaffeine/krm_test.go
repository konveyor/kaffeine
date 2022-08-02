package kaffeine

import (
	"testing"
)

func TestToGroupNameVersion(t *testing.T) {
	var tests = []struct {
		in, g, n, v string
	}{
		{"name", "", "name", ""},
		{"group/name", "group", "name", ""},
		{"group.com/sub/name", "group.com/sub", "name", ""},
		{"name@v1", "", "name", "v1"},
		{"group/name@version", "group", "name", "version"},
		{"group@git.com/name@version", "group@git.com", "name", "version"},
		{"/name@", "", "name", ""},
	}

	for _, test := range tests {
		g, n, v := ToGroupNameVersion(test.in)
		if g != test.g || n != test.n || v != test.v {
			t.Errorf("%s: got [%s, %s, %s], want [%s, %s, %s]", test.in, g, n, v, test.g, test.n, test.v)
		}
	}
}
