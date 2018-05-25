package filters

import "errors"

var (
	ErrFilterBlueprintNotFound  = errors.New("filter blueprint not found")
	ErrDimensionNotFound        = errors.New("dimension not found")
	ErrOptionNotFound           = errors.New("option not found")
	ErrVersionNotFound          = errors.New("version not found")
	ErrDimensionOptionsNotFound = errors.New("dimension options not found")
	ErrFilterOutputNotFound     = errors.New("filter output not found")
	ErrBadRequest               = errors.New("invalid request body")
	ErrForbidden                = errors.New("forbidden")
	ErrUnauthorised             = errors.New("unauthorised")
	ErrInternalError            = errors.New("internal server error")
)

func NewBadRequestErr(text string) error {
	return BadRequestErr{text}
}

// errorString is a trivial implementation of error.
type BadRequestErr struct {
	s string
}

func (e BadRequestErr) Error() string {
	return e.s
}

func NewInvalidDimensionErr(text string) error {
	return &InvalidDimensionErr{text}
}

// errorString is a trivial implementation of error.
type InvalidDimensionErr struct {
	s string
}

func (e InvalidDimensionErr) Error() string {
	return e.s
}

func NewInvalidOptionErr(text string) error {
	return &InvalidDimensionErr{text}
}

// errorString is a trivial implementation of error.
type InvalidOptionErr struct {
	s string
}

func (e InvalidOptionErr) Error() string {
	return e.s
}
