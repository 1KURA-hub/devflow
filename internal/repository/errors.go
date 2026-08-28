package repository

import "errors"

var ErrNotFound = errors.New("record not found")
var ErrDuplicate = errors.New("duplicate record")
