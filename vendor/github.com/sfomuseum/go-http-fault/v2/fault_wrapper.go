package fault

import (
	"html/template"
	"log"
	"net/http"
)

// FaultWrapper is a struct to make assigning `TemplatedFaultHandlerWrapper` instances easier.
type FaultWrapper struct {
	// logger is the *log.Logger instance that will be passed to the underlying `TemplatedFaultHandler`
	logger *log.Logger
	// template is the *template.Template instance that will be executed by the underlying `TemplatedFaultHandler`
	template *template.Template
}

// NewFaultWrapper will create a new `FaultWrapper` instance.
func NewFaultWrapper(logger *log.Logger, template *template.Template) *FaultWrapper {

	fw := &FaultWrapper{
		template: template,
		logger:   logger,
	}

	return fw
}

// HandleWithMux will wrap 'h' in a new `TemplatedFaultHandlerWrapper` middleware handler and then assign it to 'mux' with pattern 'uri'.
func (fw *FaultWrapper) HandleWithMux(mux *http.ServeMux, uri string, h http.Handler) {
	wr := TemplatedFaultHandlerWrapper(fw.logger, fw.template, h)
	mux.Handle(uri, wr)
}

// TemplatedFaultHandlerWrapper will return a middleware `http.Handler` that when invoked will serve 'h'
// and if the response status code is >= `http.StatusBadRequest` (300) will serve a new fault handler
// using 't' and 'l'.
func TemplatedFaultHandlerWrapper(l *log.Logger, t *template.Template, h http.Handler) http.Handler {

	fh := faultHandler(l, t)

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		sw := NewStatusWriter(rsp)
		h.ServeHTTP(sw, req)

		if sw.Status < http.StatusBadRequest {
			return
		}

		fh.ServeHTTP(rsp, req)
		return
	}

	return http.HandlerFunc(fn)
}
