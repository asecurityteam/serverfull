package domain

import "fmt"

// NotFoundError represents a failed lookup for a resource.
type NotFoundError struct {
	// ID is the key used when looking for the resource.
	ID string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("resource (%s) not found", e.ID)
}
