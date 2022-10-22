package fault

// See also: https://github.com/natefinch/wrap

// FaultError is an interface for providing custom public and private error messages.
type FaultError interface {
	error
	// Public returns an `error` instance whose string value is considered safe for publishing in a public setting or context.
	Public() error
	// Private returns an `error` instance whose string value that may contain details not suitable for publishing in a public setting or context.
	Private() error
}
