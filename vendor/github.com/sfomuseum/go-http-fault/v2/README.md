# go-http-fault

Go package providing a `net/http` handler for logging errors and rendering them to the browser using custom templates.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/sfomuseum/go-http-fault.svg)](https://pkg.go.dev/github.com/sfomuseum/go-http-fault)

## Example

```
import (
       "net/http"
       "html/template"
       "github.com/sfomuseum/go-http-fault/v2"
       "log"
)

// Create a custom HTTP handler specific to your application that accepts a `fault.FaultHandler` instance

func ExampleHandler(fault_handler http.Handler) http.Handler {

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		// In your handler invoke some custom code that may return an error
		
		err := SomeOtherFunction(req)

		// If it does assign the error returned and a status code (not required to be an HTTP status code)
		// that will be used for logging errors
		
		if err != nil {
			fault.AssignError(req, err, http.StatusInternalServerError)
			fault_handler.ServeHTTP(rsp, req)
			return
		}

		rsp.Write([]byte("Hello world"))
		return
	}

	h := http.HandlerFunc(fn)
	return h, nil
}

func main() {

	// In your application's `main` routine create and `html/template` instance
	// and retrieve a template named "fault" (or whatever you choose)
	
     	t, _ := template.New("example").Parse(...)
        fault_t := t.Lookup("fault")

	// Create a custom `log.Logger` instance that will be used to record errors
	
	error_logger := log.Default()

	// Now pass both to `fault.TemplatedFaultHandler` which will return an `http.Handler`
	
	fault_handler, _ := fault.TemplatedFaultHandler(error_logger, fault_t)

	// And pass that to any other handlers where you need a consistent interface
	// for handling public-facing errors
	
	handler := ExampleHandler(fault_handler)

	mux.Handle("/", handler)

	http.ListenAndServer(":8080", mux)
```