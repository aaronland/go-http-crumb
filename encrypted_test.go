package crumb

import (
	"context"
	"net/http"
	"testing"
)

func TestRandomEncryptedCrumbURI(t *testing.T) {

	ctx := context.Background()

	uri, err := NewRandomEncryptedCrumbURI(ctx, 10, "test")

	if err != nil {
		t.Fatalf("Failed to create new crumb URI, %v", err)
	}

	_, err = NewCrumb(ctx, uri)

	if err != nil {
		t.Fatalf("Failed to create new crumb for '%s', %v", uri, err)
	}
}

func TestEncryptedCrumb(t *testing.T) {

	ctx := context.Background()

	uri, err := NewRandomEncryptedCrumbURI(ctx, 10, "test")

	if err != nil {
		t.Fatalf("Failed to create new crumb URI, %v", err)
	}

	cr, err := NewCrumb(ctx, uri)

	if err != nil {
		t.Fatalf("Failed to create new crumb for '%s', %v", uri, err)
	}

	req, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		t.Fatalf("Failed to create new HTTP request, %v", err)
	}

	cs, err := cr.Generate(req)

	if err != nil {
		t.Fatalf("Failed to generate crumb string, %v", err)
	}

	ok, err := cr.Validate(req, cs)

	if err != nil {
		t.Fatalf("Failed to validate crumb string, %v", err)
	}

	if !ok {
		t.Fatalf("Expected crumb string to validate, but it didn't")
	}
}
