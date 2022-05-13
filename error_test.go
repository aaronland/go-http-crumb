package crumb

import (
	"fmt"
	"testing"
)

func TestCrumbError(t *testing.T) {

	err := fmt.Errorf("Testing")

	e := Error(GenerateCrumb, err)

	public := e.Public()
	private := e.Private()

	if public.Error() != string(GenerateCrumb) {
		t.Fatalf("Unexpected public error")
	}

	if private.Error() != "Testing" {
		t.Fatalf("Unexpected private error")
	}

	if e.Error() != string(GenerateCrumb) {
		t.Fatalf("Unexpected error string")
	}
}
