package telegram

import "testing"

func TestAllowlist_openAllowsAll(t *testing.T) {
	a := NewAllowlist(nil)
	if !a.Open() || !a.Allowed(123) {
		t.Fatal("expected open allowlist")
	}
}

func TestAllowlist_restricts(t *testing.T) {
	a := NewAllowlist([]int64{42})
	if a.Open() {
		t.Fatal("expected closed allowlist")
	}
	if !a.Allowed(42) || a.Allowed(99) {
		t.Fatal("unexpected allow result")
	}
}
