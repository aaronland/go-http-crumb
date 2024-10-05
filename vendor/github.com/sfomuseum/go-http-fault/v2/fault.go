package fault

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"reflect"
)

// ErrorKey is the name of the key for assigning `error` values to a `context.Context` instance.
const ErrorKey string = "github.com/sfomuseum/go-http-fault#error"

// StatusKey is the name of the key for assigning status code values to a `context.Context` instance.
const StatusKey string = "github.com/sfomuseum/go-http-fault#status"

// FaultHandlerOptions is a struct containing configuration options for the `FaultHandlerWithOptions` method.
type FaultHandlerOptions struct {
	// Logger is a custom `log.Logger` instance used for logging and feedback.
	Logger *log.Logger
	// Template is an optional `html/template.Template` to use for reporting errors.
	Template *template.Template
	// VarFunc is a `FaultHandlerVarsFunc` used to derive variables passed to templates. Required if `Template` is non-nil.
	VarsFunc FaultHandlerVarsFunc
}

// FaultHandlerVars are the minimal required variables that will be pass to a fault handler template. If you need to pass additional fields
// then you will need to create your handler using the the `FaultHandlerWithOptions` method specifying a custom `FaultHandlerVarsFunc` property.
type FaultHandlerVars struct {
	Status int
	Error  error
}

// FaultHandlerVarsFunc is a custom function that returns a pointer to a struct that conforms to the required fields of the `FaultHandlerVars` struct type.
type FaultHandlerVarsFunc func() interface{}

// ImplementsFaultHandlerVars returns a boolean value indicating whether 'vars' conforms to the required fields of the `FaultHandlerVars` struct type.
func ImplementsFaultHandlerVars(vars interface{}) bool {

	switch vars.(type) {
	case FaultHandlerVars, *FaultHandlerVars:
		return true
	default:
		// carry on
	}

	if reflect.TypeOf(vars).Kind() != reflect.Ptr {
		return false
	}

	dv := reflect.ValueOf(vars).Elem()

	if dv.Kind() != reflect.Struct {
		return false
	}

	s := dv.FieldByName("Status")

	if !s.CanSet() {
		return false
	}

	e := dv.FieldByName("Error")

	if !e.CanSet() {
		return false
	}

	return true
}

func defaultFaultHandlerVars() interface{} {
	return &FaultHandlerVars{
		Status: 0,
		Error:  fmt.Errorf("Undefined error"),
	}
}

// AssignError assigns 'err' and 'status' the `ErrorKey` and `StatusKey` values of 'req.Context'
// and updates 'req' in place.
func AssignError(req *http.Request, err error, status int) {

	ctx := req.Context()
	ctx = context.WithValue(ctx, ErrorKey, err)
	ctx = context.WithValue(ctx, StatusKey, status)

	new_req := req.WithContext(ctx)
	*req = *new_req
}

// RetrieveError returns the values of the `StatusKey` and `ErrorKey` values of 'req.Context'
func RetrieveError(req *http.Request) (int, error) {

	ctx := req.Context()
	err_v := ctx.Value(ErrorKey)
	status_v := ctx.Value(StatusKey)

	var status int
	var err error

	if err_v == nil {
		msg := "FaultHandler triggered without an error context."
		err = errors.New(msg)
	} else {
		err = err_v.(error)
	}

	if status_v == nil {
		status = http.StatusInternalServerError
	} else {
		status = status_v.(int)
	}

	return status, err
}

// FaultHandler returns a `http.Handler` instance for handling errors in a web application.
func FaultHandler(l *log.Logger) http.Handler {

	opts := &FaultHandlerOptions{
		Logger:   l,
		Template: nil,
		VarsFunc: nil, // unnecessary because no template
	}

	return FaultHandlerWithOptions(opts)
}

// TemplatedFaultHandler returns a `http.Handler` instance for handling errors in a web application
// with a custom HTML template.
func TemplatedFaultHandler(l *log.Logger, t *template.Template) http.Handler {

	opts := &FaultHandlerOptions{
		Logger:   l,
		Template: t,
		VarsFunc: defaultFaultHandlerVars,
	}

	return FaultHandlerWithOptions(opts)
}

// faultHandler returns a `http.Handler` for handling errors in a web application. It will retrieve
// and "public" and "private" errors that have been recorded and log them to 'l'. If 't is defined it
// will executed and passed the "public" error as a template variable.
func FaultHandlerWithOptions(opts *FaultHandlerOptions) http.Handler {

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		status, err := RetrieveError(req)

		var fault_err FaultError

		public_err := err
		private_err := err

		if errors.As(err, &fault_err) {
			public_err = fault_err.Public()
			private_err = fault_err.Private()
		}

		addr := req.RemoteAddr

		opts.Logger.Printf("[FAULT] %s \"%s %s %s\" %v\n", addr, req.Method, req.RequestURI, req.Proto, private_err)

		if opts.Template != nil {

			rsp.Header().Set("Content-Type", "text/html")

			vars := opts.VarsFunc()

			switch vars.(type) {
			case FaultHandlerVars:

				f := vars.(FaultHandlerVars)
				f.Status = status
				f.Error = public_err
				vars = f

			case *FaultHandlerVars:

				f := vars.(*FaultHandlerVars)
				f.Status = status
				f.Error = public_err
				vars = f

			default:

				if reflect.TypeOf(vars).Kind() != reflect.Ptr {
					opts.Logger.Printf("[FAULT] template vars must be a pointer")
					http.Error(rsp, "Invalid template vars", http.StatusInternalServerError)
					return
				}

				dv := reflect.ValueOf(vars).Elem()

				if dv.Kind() != reflect.Struct {
					opts.Logger.Printf("[FAULT] template vars must be a pointer to a struct/interface")
					http.Error(rsp, "Invalid template vars", http.StatusInternalServerError)
					return

				}

				s := dv.FieldByName("Status")

				if !s.CanSet() {
					opts.Logger.Printf("[FAULT] template vars have no field '%s' or cannot be set", "Status")
					http.Error(rsp, "Invalid template vars", http.StatusInternalServerError)
					return
				}

				status_v := reflect.ValueOf(status)
				s.Set(status_v)

				e := dv.FieldByName("Error")

				if !e.CanSet() {
					opts.Logger.Printf("[FAULT] template vars have no field '%s' or cannot be set", "Error")
					http.Error(rsp, "Invalid template vars", http.StatusInternalServerError)
					return
				}

				error_v := reflect.ValueOf(public_err)
				e.Set(error_v)
			}

			err = opts.Template.Execute(rsp, vars)

			if err == nil {
				return
			}

			msg := fmt.Sprintf("Failed to render template for fault handler, %v", err)
			opts.Logger.Printf("[FAULT] %s \"%s %s %s\" %s\n", addr, req.Method, req.RequestURI, req.Proto, msg)
		}

		err_msg := fmt.Sprintf("There was a problem completing your request (%d)", status)
		http.Error(rsp, err_msg, status)

		return
	}

	h := http.HandlerFunc(fn)
	return h
}
