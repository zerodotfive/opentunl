package ca

import "errors"

var AlreadyRevoked = errors.New("certificate already revoked")
var InvalidKey = errors.New("invalid key")
var InvalidValidDays = errors.New("valid days can not be negative")
