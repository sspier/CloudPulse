package model

// Target represents a single monitored endpoint
type Target struct {
	// field is defined as ID (uppercase) so that it is exported and visible to other packages
	// by default, the AWS SDK uses the struct field name as the database attribute name
	// by using the `dynamodbav:"id"` tag, we can override this behavior and use a different name
	// basically, "take the value from the ID field in Go, but name the attribute id when you talk to DynamoDB."
	ID   string `json:"id" dynamodbav:"id"`
	Name string `json:"name" dynamodbav:"name"`
	URL  string `json:"url" dynamodbav:"url"`
}

// Result represents the outcome of a single uptime probe
type Result struct {
	TargetID   string `json:"targetId" dynamodbav:"target_id"`
	Status     string `json:"status" dynamodbav:"status"`
	HTTPStatus int    `json:"httpStatus" dynamodbav:"http_status"`
	Timestamp  int64  `json:"timestamp" dynamodbav:"timestamp"` // added timestamp for history
}
