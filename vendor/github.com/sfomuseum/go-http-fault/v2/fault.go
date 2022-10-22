package fault

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

// ErrorKey is the name of the key for assigning `error` values to a `context.Context` instance.
const ErrorKey string = "github.com/sfomuseum/go-http-fault#error"

// StatusKey is the name of the key for assigning status code values to a `context.Context` instance.
const StatusKey string = "github.com/sfomuseum/go-http-fault#status"

type FaultHandlerVars struct {
	Status int
	Error  error
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
	return faultHandler(l, nil)
}

// TemplatedFaultHandler returns a `http.Handler` instance for handling errors in a web application
// with a custom HTML template.
func TemplatedFaultHandler(l *log.Logger, t *template.Template) http.Handler {
	return faultHandler(l, t)
}

// faultHandler returns a `http.Handler` for handling errors in a web application. It will retrieve
// and "public" and "private" errors that have been recorded and log them to 'l'. If 't is defined it
// will executed and passed the "public" error as a template variable.
func faultHandler(l *log.Logger, t *template.Template) http.Handler {

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

		l.Printf("[FAULT] %s \"%s %s %s\" %v\n", addr, req.Method, req.RequestURI, req.Proto, private_err)

		if t != nil {

			rsp.Header().Set("Content-Type", "text/html")

			vars := FaultHandlerVars{
				Status: status,
				Error:  public_err,
			}

			err = t.Execute(rsp, vars)

			if err == nil {
				return
			}

			msg := fmt.Sprintf("Failed to render template for fault handler, %v", err)
			l.Printf("[FAULT] %s \"%s %s %s\" %s\n", addr, req.Method, req.RequestURI, req.Proto, msg)
		}

		err_msg := fmt.Sprintf("There was a problem completing your request (%d)", status)

		http.Error(rsp, err_msg, status)
		return
	}

	h := http.HandlerFunc(fn)
	return h
}
