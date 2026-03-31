package db

import "fmt"

type NotFoundErr struct {
	ResourceType string
	ID           string
}

func NotFound(rt, id string) NotFoundErr {
	return NotFoundErr{ResourceType: rt, ID: id}
}

func (err NotFoundErr) Error() string {
	return fmt.Sprintf("%s with id '%s' not found", err.ResourceType, err.ID)
}
