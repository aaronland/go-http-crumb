package crumb

import (
	"context"
	"fmt"
	"github.com/aaronland/go-roster"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// type Crumb is an interface for generating and validating HTTP crumb strings.
type Crumb interface {
	// Generate return a new crumb string for an HTTP request.
	Generate(*http.Request, ...string) (string, error)
	// Generate validates a crumb string for an HTTP request.
	Validate(*http.Request, string, ...string) (bool, error)
	// Key returns the unique key used to generate a crumb.
	Key(*http.Request) string
	// Base returns the leading string used to generate a crumb.
	Base(*http.Request, ...string) (string, error)
}

// type CrumbInitializeFunc is a function used to initialize packages that implement the `Crumb` interface.
type CrumbInitializeFunc func(context.Context, string) (Crumb, error)

var crumbs roster.Roster

func ensureCrumbs() error {

	if crumbs == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return fmt.Errorf("Failed to create new roster,%w", err)
		}

		crumbs = r
	}

	return nil
}

// RegisterCrumb registers 'scheme' with 'f' for URIs passed to the `NewCrumb` method.
func RegisterCrumb(ctx context.Context, scheme string, f CrumbInitializeFunc) error {

	err := ensureCrumbs()

	if err != nil {
		return fmt.Errorf("Failed to ensure crumbs, %w", err)
	}

	return crumbs.Register(ctx, scheme, f)
}

// Returns a new `Crumb` instance for 'uri'.
func NewCrumb(ctx context.Context, uri string) (Crumb, error) {

	err := ensureCrumbs()

	if err != nil {
		return nil, fmt.Errorf("Failed to ensure crumbs, %w", err)
	}

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	scheme := u.Scheme

	i, err := crumbs.Driver(ctx, scheme)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive driver for %s, %w", scheme, err)
	}

	f := i.(CrumbInitializeFunc)
	return f(ctx, uri)
}

// Schemes returns the list of schemes that have registered for use with the `NewCrumb` method.
func Schemes() []string {
	ctx := context.Background()
	drivers := crumbs.Drivers(ctx)

	schemes := make([]string, len(drivers))

	for idx, dr := range drivers {
		schemes[idx] = fmt.Sprintf("%s://", dr)
	}

	sort.Strings(schemes)
	return schemes
}

// SchemesAsString returns the list of schemes that have registered for use with the `NewCrumb`
// method as a string.
func SchemesAsString() string {
	schemes := Schemes()
	return strings.Join(schemes, ",")
}
