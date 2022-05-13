package crumb

import (
	"fmt"
	"github.com/aaronland/go-http-rewrite"
	"github.com/aaronland/go-http-sanitize"
	"github.com/sfomuseum/go-http-fault"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"log"
	go_http "net/http"
)

// EnsureCrumbHandler wraps 'next_handler' with a middleware `http.Handler` for assigning and validating
// crumbs using the default `fomuseum/go-http-fault.FaultHandler` as an error handler. Any errors that
// trigger the error handler can be retrieved using `sfomuseum/go-http-fault.RetrieveError()`.
func EnsureCrumbHandler(cr Crumb, next_handler go_http.Handler) (go_http.Handler, error) {

	logger := log.Default()
	fault_handler, err := fault.FaultHandler(logger)

	if err != nil {
		return nil, fmt.Errorf("Failed to create fault handler, %v", err)
	}

	return EnsureCrumbHandlerWithErrorHandler(cr, next_handler, fault_handler)
}

// EnsureCrumbHandlerWithErrorHandler wraps 'next_handler' with a middleware a middleware `http.Handler` for
// assigning and validating crumbs using a custom error handler. Any errors that trigger the error handler can
// be retrieved using `sfomuseum/go-http-fault.RetrieveError()`.
func EnsureCrumbHandlerWithErrorHandler(cr Crumb, next_handler go_http.Handler, error_handler go_http.Handler) (go_http.Handler, error) {

	fn := func(rsp go_http.ResponseWriter, req *go_http.Request) {

		switch req.Method {

		case "POST", "PUT":

			var crumb_var string
			var crumb_err error

			if req.Method == "POST" {
				crumb_var, crumb_err = sanitize.PostString(req, "crumb")
			} else {
				crumb_var, crumb_err = sanitize.GetString(req, "crumb")
			}

			if crumb_err != nil {
				req = fault.AssignError(req, Error(UnsanitizedCrumb, crumb_err), go_http.StatusBadRequest)
				error_handler.ServeHTTP(rsp, req)
				return
			}

			if crumb_var == "" {
				req = fault.AssignError(req, Error(MissingCrumb, fmt.Errorf("Missing crumb")), go_http.StatusBadRequest)
				error_handler.ServeHTTP(rsp, req)
				return
			}

			ok, err := cr.Validate(req, crumb_var)

			if err != nil {
				req = fault.AssignError(req, Error(InvalidCrumb, err), go_http.StatusInternalServerError)
				error_handler.ServeHTTP(rsp, req)
				return
			}

			if !ok {
				req = fault.AssignError(req, Error(ExpiredCrumb, fmt.Errorf("Expired")), go_http.StatusForbidden)
				error_handler.ServeHTTP(rsp, req)
				return
			}

		default:
			// pass
		}

		crumb_var, err := cr.Generate(req)

		if err != nil {
			req = fault.AssignError(req, Error(GenerateCrumb, err), go_http.StatusInternalServerError)
			error_handler.ServeHTTP(rsp, req)
			return
		}

		rewrite_func := NewCrumbRewriteFunc(crumb_var)
		rewrite_handler := rewrite.RewriteHTMLHandler(next_handler, rewrite_func)

		rewrite_handler.ServeHTTP(rsp, req)

	}

	h := go_http.HandlerFunc(fn)
	return h, nil
}

// NewCrumbRewriteFunc returns a `aaronland/go-http-rewrite.RewriteHTMLFunc` used to
// append crumb data to HTML output.
func NewCrumbRewriteFunc(crumb_var string) rewrite.RewriteHTMLFunc {

	var rewrite_func rewrite.RewriteHTMLFunc

	rewrite_func = func(n *html.Node, w io.Writer) {

		if n.Type == html.ElementNode && n.Data == "body" {

			crumb_ns := ""
			crumb_key := "data-crumb"
			crumb_value := crumb_var

			crumb_attr := html.Attribute{crumb_ns, crumb_key, crumb_value}
			n.Attr = append(n.Attr, crumb_attr)
		}

		if n.Type == html.ElementNode && n.Data == "form" {

			ns := ""

			attrs := []html.Attribute{
				html.Attribute{ns, "type", "hidden"},
				html.Attribute{ns, "id", "crumb"},
				html.Attribute{ns, "name", "crumb"},
				html.Attribute{ns, "value", crumb_var},
			}

			i := &html.Node{
				Type:      html.ElementNode,
				DataAtom:  atom.Input,
				Data:      "input",
				Namespace: ns,
				Attr:      attrs,
			}

			n.AppendChild(i)
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			rewrite_func(c, w)
		}
	}

	return rewrite_func
}
