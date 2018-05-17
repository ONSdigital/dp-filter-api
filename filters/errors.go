package filters

import "errors"

var (
	ErrFilterBlueprintNotFound  = errors.New("filter blueprint not found")
	ErrDimensionNotFound        = errors.New("dimension not found")
	ErrOptionNotFound           = errors.New("option not found")
	ErrVersionNotFound          = errors.New("version not found")
	ErrDimensionOptionsNotFound = errors.New("dimension options not found")
	ErrFilterOutputNotFound     = errors.New("filter output not found")
)
