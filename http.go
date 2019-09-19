package crumb

import (
	"errors"
	"github.com/aaronland/go-http-rewrite"
	"github.com/aaronland/go-http-sanitize"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	_ "log"
	go_http "net/http"
)

type ErrorHandlerFunc func(go_http.ResponseWriter, *go_http.Request, error, int) go_http.Handler

func DefaultErrorHandlerFunc() ErrorHandlerFunc {

	fn := func(rsp go_http.ResponseWriter, req *go_http.Request, err error, http_status int) go_http.Handler {

		handler_fn := func(rsp go_http.ResponseWriter, req *go_http.Request) {
			go_http.Error(rsp, err.Error(), http_status)
			return
		}

		return go_http.HandlerFunc(handler_fn)
	}

	return fn
}

func EnsureCrumbHandler(cfg *CrumbConfig, next_handler go_http.Handler) go_http.Handler {

	err_handler := DefaultErrorHandlerFunc()
	return EnsureCrumbHandlerWithErrorHandler(cfg, next_handler, err_handler)
}

func EnsureCrumbHandlerWithErrorHandler(cfg *CrumbConfig, next_handler go_http.Handler, error_handler_func ErrorHandlerFunc) go_http.Handler {

	fn := func(rsp go_http.ResponseWriter, req *go_http.Request) {

		switch req.Method {

		case "GET":

			crumb_var, err := GenerateCrumb(cfg, req)

			if err != nil {
				err_handler := error_handler_func(rsp, req, err, go_http.StatusInternalServerError)
				err_handler.ServeHTTP(rsp, req)
				return
			}

			rewrite_func := NewCrumbRewriteFunc(crumb_var)
			rewrite_handler := rewrite.RewriteHTMLHandler(next_handler, rewrite_func)

			rewrite_handler.ServeHTTP(rsp, req)

		case "POST":

			crumb_var, err := sanitize.PostString(req, "crumb")

			if err != nil {
				err_handler := error_handler_func(rsp, req, err, go_http.StatusBadRequest)
				err_handler.ServeHTTP(rsp, req)
				return
			}

			ok, err := ValidateCrumb(cfg, req, crumb_var)

			if err != nil {
				err_handler := error_handler_func(rsp, req, err, go_http.StatusInternalServerError)
				err_handler.ServeHTTP(rsp, req)
				return
			}

			if !ok {
				err_handler := error_handler_func(rsp, req, errors.New("Forbidden"), go_http.StatusForbidden)
				err_handler.ServeHTTP(rsp, req)
				return
			}

			next_handler.ServeHTTP(rsp, req)

		default:
			next_handler.ServeHTTP(rsp, req)
		}
	}

	return go_http.HandlerFunc(fn)
}

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
