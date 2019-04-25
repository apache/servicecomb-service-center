package utils

import "testing"

func TestSplitHostPort(t *testing.T) {
	_, _, err := SplitHostPort("", 0)
	if err != nil {
		t.Logf("split host port failed, error: %s", err)
	}
}
