package utils

import "errors"

var ErrPasswordHash = errors.New("error hashing password")
var ErrDuplicateEntry = errors.New("duplicate user entry")
var ErrForeignKeyViolation = errors.New("foreign key violation (invalid user reference?)")
var ErrNotFound = errors.New("user not found")
var ErrTokenGeneration = errors.New("error generating token")
var ErrMissingAuthToken = errors.New("missing authorization token")
