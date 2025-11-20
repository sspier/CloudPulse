## Go Syntax Highlights

- `:=` — shorthand for declaring and initializing a variable in one step
- `&Struct{}` — pointer to a struct literal
- `go func() { ... }()` — launches a goroutine using an anonymous function
- `<-channel` — receive from a channel (blocking)
- `defer cancel()` — schedules cleanup at the end of the function
- **Capitalization** — determines visibility (exported/public vs unexported/private)
- `_ *http.Request` — ignore a parameter using the blank identifier
- `_, _ =` — explicitly ignore function outputs

## Additional Notes

- `make()` — allocates slices, maps, and channels with initialization
- `new(Type)` — allocates zeroed storage and returns a pointer
- `iota` — auto-incrementing constant generator used in const blocks
- `context.Context` — carries deadlines, cancellation signals, and metadata across API boundaries
- `http.HandlerFunc` — adapter to allow a function to satisfy the Handler interface
- `select {}` — waits on multiple channel operations
- `for range` — iterates over slices, maps, channels
- `json.Marshal` / `json.Unmarshal` — convert between Go structs and JSON
- `errors.New()` — create a basic error value
- `fmt.Errorf()` — formatted error wrapping
- `:=` inside `if` — short variable declaration scoped to the `if` block
- `panic` / `recover` — low-level error mechanism (avoid in normal flow)
