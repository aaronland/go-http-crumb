# go-http-fault

Go package providing a `net/http` handler for logging errors and rendering them to the browser using custom templates.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/sfomuseum/go-http-fault.svg)](https://pkg.go.dev/github.com/sfomuseum/go-http-fault)

## Example

```
import (
       "net/http"
       "html/template"
       "log"

       "github.com/sfomuseum/go-http-fault/v2"
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

If you need to pass custom variables to a custom template you will need to take a few more steps. First, define a callback function that conforms to `fault.FaultHandlerVarsFunc`. Although the method signature says these callback functions only need to return an `interface{}` in actual fact you'll need to return a) a struct b) a pointer to that struct and c) ensure that struct conforms to the `fault.FaultHandlerVars` struct. For example:

```
type CustomVars struct {
     SomeOtherVariable string
     fault.FaultHandlerVars     
}

var custom_vars_func := func() interface{} {

    return &CustomVars{
    	SomeOtherVariable: "Hello world",
    }	
}
```

Note that `CustomVars` will be assigned `Status` and `Error` properties, which are an `int` and and `error` respectively, at runtime. If you assign your own values they will be overwritten.

Next create a `FaultHandlerOptions` instance which references your custom variable callback function, a `log.Logger` instance for feedback and debugging and a `html/template.Template` instance to be rendered by the fault handler. For example:

```
logger := log.Default()

tpl := template.New("test")
tpl, _ := tpl.Parse(`{{ .SomeOtherVariable }} {{ .Status }}`)

opts := &fault.FaultHandlerOptions{
	Logger:   logger,
	Template: tpl,
	VarsFunc: custom_vars_func,
}
```

Finally create the fault handler using the `FaultHandlerWithOptions` method. For example:

```
handler := fault.FaultHandlerWithOptions(opts)
```
