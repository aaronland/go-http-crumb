package crumb

import (
	"fmt"
	"github.com/sfomuseum/go-http-fault/v2"
)

// type ErrorClass defines application specific error classes (or types)
type ErrorClass string

// GenerateCrumb defines an ErrorClass for crumbs that are not able to be generated.
const GenerateCrumb ErrorClass = "Unable to generate crumb"

// MissingCrumb defines an ErrorClass for crumbs that are missing
const MissingCrumb ErrorClass = "Crumb is missing."

// UnsanitizedCrumb defines an ErrorClass for crumbs that have failed input validation.
const UnsanitizedCrumb ErrorClass = "Crumb supplied is invalid."

// InvalidCrumb defines an ErrorClass for crumbs that do not validate.
const InvalidCrumb ErrorClass = "Crumb does not validate."

// ExpiredCrumb defines an ErrorClass for crumbs that have expired.
const ExpiredCrumb ErrorClass = "Crumb has expired."

// type CrumbError implements the `error` and `fault.FaultError` interfaces for application specific errors.
type CrumbError struct {
	fault.FaultError
	// The type (or class) of application error
	class ErrorClass
	// An optional error associated with the CrumbError
	error error
}

// New returns a new `CrumbError` instance.
func Error(cl ErrorClass, err error) fault.FaultError {

	e := &CrumbError{
		class: cl,
		error: err,
	}

	return e
}

// Unwrap returns the error passed to 'e' when it was instantiated.
func (e CrumbError) Unwrap() error {
	return e.error
}

// String returns an informative string message to be displayed in a public setting.
func (e CrumbError) Error() string {

	return fmt.Sprintf("%s", e.class)
}

// String returns an informative string message to be displayed in a public setting.
func (e CrumbError) Public() error {

	return fmt.Errorf("%s", e.class)
}

// Returns the string value of 'e' and of the error passed to 'e' when it was instantiated, if present.
// These are considered detailed errors not necesarily meant for the general public.
func (e CrumbError) Private() error {

	if e.error == nil {
		return e.Public()
	}

	return e.error
}
