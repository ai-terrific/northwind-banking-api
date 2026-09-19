package externalbank

import "errors"

var (
	ErrProviderUnauthorized = errors.New("external provider unauthorized")
	ErrProviderValidation   = errors.New("external provider validation failed")
	ErrProviderNotFound     = errors.New("external resource not found")
	ErrProviderConflict     = errors.New("external provider conflict")
	ErrProviderUnavailable  = errors.New("external provider unavailable")
	ErrProviderTimeout      = errors.New("external provider timeout")
)
