## Go Syntax Highlights

- `:=` — shorthand for declaring and initializing a variable in one step
- `&Struct{}` — pointer to a struct literal
- `go func() { ... }()` — launches a goroutine using an anonymous function
- `<-channel` — receive from a channel (blocking)
- `defer cancel()` — schedules cleanup at the end of the function
- **Capitalization** — determines visibility (exported/public vs unexported/private)
- `_ *http.Request` — ignore a parameter using the blank identifier
- `_, _ =` — explicitly ignore function outputs
