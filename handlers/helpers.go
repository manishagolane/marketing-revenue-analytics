package handlers

import (
	"errors"
	"fmt"
	"io"

	"github.com/goccy/go-json"

	"marketing-revenue-analytics/utils"

	"github.com/go-playground/validator/v10"
)

func getPrettyValidationError(err error) error {
	var vError validator.ValidationErrors
	if errors.As(err, &vError) {
		for _, fieldErr := range vError {
			return fmt.Errorf("%s validation for field %s failed", fieldErr.ActualTag(), utils.ToCamelCase(fieldErr.Field()))
		}
	}
	if errors.Is(err, io.EOF) {
		return errors.New("missing request body")
	}
	if jsonError, ok := err.(*json.UnmarshalTypeError); ok {
		return fmt.Errorf("unexpected type for %s, expected: %s, received: %s",
			utils.ToCamelCase(jsonError.Field),
			jsonError.Type.String(),
			jsonError.Value)
	}
	return err
}
