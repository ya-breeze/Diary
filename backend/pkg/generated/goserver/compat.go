// Hand-written compatibility layer: preserves the ImplResponse pattern and helper types
// used by service implementations and custom controllers.
package goserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"reflect"
)

// --- ImplResponse ---

// ImplResponse defines an implementation response with error code and the associated body.
type ImplResponse struct {
	Code int
	Body interface{}
}

// Response creates an ImplResponse with the given status code and body.
func Response(code int, body interface{}) ImplResponse {
	return ImplResponse{Code: code, Body: body}
}

// --- Error types ---

var ErrTypeAssertionError = errors.New("unable to assert type")

// ParsingError indicates that an error has occurred when parsing request parameters.
type ParsingError struct {
	Param string
	Err   error
}

func (e *ParsingError) Unwrap() error { return e.Err }
func (e *ParsingError) Error() string {
	if e.Param == "" {
		return e.Err.Error()
	}
	return e.Param + ": " + e.Err.Error()
}

// RequiredError indicates that a required field is missing.
type RequiredError struct {
	Field string
}

func (e *RequiredError) Error() string {
	return fmt.Sprintf("required field '%s' is zero value.", e.Field)
}

// ErrorHandler defines the required method for handling errors from controllers.
type ErrorHandler func(w http.ResponseWriter, r *http.Request, err error, result *ImplResponse)

// DefaultErrorHandler handles errors with appropriate HTTP status codes.
func DefaultErrorHandler(w http.ResponseWriter, r *http.Request, err error, result *ImplResponse) {
	slog.Error("API error", "error", err, "path", r.URL.Path, "method", r.Method)
	var parsingErr *ParsingError
	if ok := errors.As(err, &parsingErr); ok {
		_ = EncodeJSONResponse(err.Error(), func(i int) *int { return &i }(http.StatusBadRequest), w)
		return
	}
	var requiredErr *RequiredError
	if ok := errors.As(err, &requiredErr); ok {
		_ = EncodeJSONResponse(err.Error(), func(i int) *int { return &i }(http.StatusUnprocessableEntity), w)
		return
	}
	_ = EncodeJSONResponse(err.Error(), &result.Code, w)
}

// --- Encoding helpers ---

// EncodeJSONResponse serialises body to the response writer.
func EncodeJSONResponse(i interface{}, status *int, w http.ResponseWriter) error {
	wHeader := w.Header()

	if f, ok := i.(*os.File); ok {
		data, err := io.ReadAll(f)
		if err != nil {
			return err
		}
		wHeader.Set("Content-Type", http.DetectContentType(data))
		wHeader.Set("Content-Disposition", "attachment; filename="+f.Name())
		if status != nil {
			w.WriteHeader(*status)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		_, err = w.Write(data)
		return err
	}

	if i == nil {
		if status != nil {
			w.WriteHeader(*status)
		}
		return nil
	}

	if s, ok := i.(string); ok {
		wHeader.Set("Content-Type", "text/plain; charset=UTF-8")
		if status != nil {
			w.WriteHeader(*status)
		}
		_, err := w.Write([]byte(s))
		return err
	}

	wHeader.Set("Content-Type", "application/json; charset=UTF-8")
	if status != nil {
		w.WriteHeader(*status)
	}
	return json.NewEncoder(w).Encode(i)
}

// --- Validation helpers ---

// IsZeroValue checks if the val is the zero-ed value.
func IsZeroValue(val interface{}) bool {
	return val == nil || reflect.DeepEqual(val, reflect.Zero(reflect.TypeOf(val)).Interface())
}

// AssertRecurseInterfaceRequired recursively checks each struct in a slice against the callback.
func AssertRecurseInterfaceRequired[T any](obj interface{}, callback func(T) error) error {
	return AssertRecurseValueRequired(reflect.ValueOf(obj), callback)
}

// AssertRecurseValueRequired checks each struct in the nested slice against the callback.
func AssertRecurseValueRequired[T any](value reflect.Value, callback func(T) error) error {
	switch value.Kind() {
	case reflect.Struct:
		obj, ok := value.Interface().(T)
		if !ok {
			return ErrTypeAssertionError
		}
		if err := callback(obj); err != nil {
			return err
		}
	case reflect.Slice:
		for i := range value.Len() {
			if err := AssertRecurseValueRequired(value.Index(i), callback); err != nil {
				return err
			}
		}
	}
	return nil
}

// AssertAuthDataRequired checks that required fields of AuthData are set.
func AssertAuthDataRequired(obj AuthData) error {
	elements := map[string]interface{}{
		"email":    obj.Email,
		"password": obj.Password,
	}
	for name, el := range elements {
		if IsZeroValue(el) {
			return &RequiredError{Field: name}
		}
	}
	return nil
}

// AssertAuthDataConstraints checks constraint compliance for AuthData (no-op).
func AssertAuthDataConstraints(_ AuthData) error { return nil }

// Authorize200Response is an alias kept for backward compatibility with custom controllers.
type Authorize200Response = Authorize200JSONResponse
