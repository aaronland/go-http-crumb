package crumb

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

func init() {
	ctx := context.Background()
	RegisterCrumb(ctx, "debug", NewDebugCrumb)
}

// NewDebugCrumb returns a `EncryptedCrumb` instance with a randomly generated secret and salt valid
// for 5 minutes configured by 'uri' which should take the form of:
//
//	debug://?{QUERY_PARAMETERS}
//
// Where '{QUERY_PARAMETERS}' may be:
// * `ttl={SECONDS}`. Default is 300
func NewDebugCrumb(ctx context.Context, uri string) (Crumb, error) {

	ttl := 300
	key := ""

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	q_ttl := q.Get("ttl")

	if q_ttl != "" {

		v, err := strconv.Atoi(q_ttl)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?ttl= parameter, %w", err)
		}

		ttl = v
	}

	crumb_uri, err := NewRandomEncryptedCrumbURI(ctx, ttl, key)

	if err != nil {
		return nil, fmt.Errorf("Failed to generate random crumb URI, %w", err)
	}

	return NewCrumb(ctx, crumb_uri)
}
