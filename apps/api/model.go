package main

// target represents a single monitored endpoint
// id is a timestamp-based string for now; later we can switch to ulid/uuid if needed
// name is a friendly label for display or organization
// url is the address we probe on a schedule
type Target struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// result represents the outcome of a single uptime probe
// targetID ties the result back to a specific target
// status is a simple "up" or "down" indicator
// httpStatus captures the actual HTTP response code from the probe
type Result struct {
	TargetID   string `json:"targetId"`
	Status     string `json:"status"`
	HTTPStatus int    `json:"httpStatus"`
}
