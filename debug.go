package crumb

import (
	"context"
	"fmt"
)

func init() {
	ctx := context.Background()
	RegisterCrumb(ctx, "debug", NewDebugCrumb)
}

// NewDebugCrumb returns a `EncryptedCrumb` instance with a randomly generated secret and salt valid
// for 5 minutes configured by 'uri' which should take the form of:
//
//	debug://
func NewDebugCrumb(ctx context.Context, uri string) (Crumb, error) {

	ttl := 5
	key := ""

	crumb_uri, err := NewRandomEncryptedCrumbURI(ctx, ttl, key)

	if err != nil {
		return nil, fmt.Errorf("Failed to generate random crumb URI, %w", err)
	}

	return NewCrumb(ctx, crumb_uri)
}
