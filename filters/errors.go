package filters

import "errors"

var (
	ErrVersionNotFound          = errors.New("version not found")
	ErrInvalidQueryParameter    = errors.New("invalid query parameter")
	ErrFilterBlueprintNotFound  = errors.New("filter blueprint not found")
	ErrFilterBlueprintConflict  = errors.New("conflict while updating filter blueprint")
	ErrDimensionNotFound        = errors.New("dimension not found")
	ErrDimensionsNotFound       = errors.New("dimensions not found")
	ErrDimensionOptionNotFound  = errors.New("option not found")
	ErrDimensionOptionsNotFound = errors.New("dimension options not found")
	ErrFilterOutputNotFound     = errors.New("filter output not found")
	ErrFilterOutputConflict     = errors.New("conflict while updating filter output")
	ErrBadRequest               = errors.New("invalid request body")
	ErrForbidden                = errors.New("forbidden")
	ErrUnauthorised             = errors.New("unauthorised")
	ErrInternalError            = errors.New("internal server error")
	ErrNoIfMatchHeader          = errors.New("required If-Match header not provided")
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

func NewForbiddenErr(text string) error {
	return ForbiddenErr{text}
}

// errorString is a trivial implementation of error.
type ForbiddenErr struct {
	s string
}

func (e ForbiddenErr) Error() string {
	return e.s
}
