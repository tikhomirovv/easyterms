package manual_test

import (
	"testing"

	"github.com/tikhomirovv/easyterms/internal/payment/manual"
)

func TestPackageChecks_known(t *testing.T) {
	n, err := manual.PackageChecks("checks_10")
	if err != nil {
		t.Fatal(err)
	}
	if n != 10 {
		t.Fatalf("checks = %d", n)
	}
}

func TestPackageChecks_unknown(t *testing.T) {
	_, err := manual.PackageChecks("nope")
	if err == nil {
		t.Fatal("expected error")
	}
}
