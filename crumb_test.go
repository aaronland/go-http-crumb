package crumb

import (
	"context"
	"testing"
)

func TestRegisterCrumb(t *testing.T) {

	ctx := context.Background()

	err := RegisterCrumb(ctx, "encrypted", NewEncryptedCrumb)

	if err == nil {
		t.Fatalf("Expected NewEncryptedCrumb to be registered already")
	}
}

func TestNewCrumb(t *testing.T) {

	ctx := context.Background()

	uri, err := NewRandomEncryptedCrumbURI(ctx, 60, "test")

	if err != nil {
		t.Fatalf("Failed to create new crumb URI, %v", err)
	}

	_, err = NewCrumb(ctx, uri)

	if err != nil {
		t.Fatalf("Failed to create new crumb for '%s', %v", uri, err)
	}
}
