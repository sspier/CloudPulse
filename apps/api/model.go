package main

type Target struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Result struct {
	TargetID   string `json:"targetId"`
	Status     string `json:"status"`
	HTTPStatus int    `json:"httpStatus"`
}
