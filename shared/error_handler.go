package shared

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"reflect"
	"strings"
)

type Errors map[string]interface{}

type FailedValidationError struct {
	Errors Errors
}

func NewFailedValidationError(obj interface{}, err validator.ValidationErrors) FailedValidationError {
	return FailedValidationError{Errors: handleFailedValidation(obj, err)}
}

func (f FailedValidationError) Error() string {
	return "Failed validation"
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	if errors.As(err, &FailedValidationError{}) {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "Failed validation",
			"errors":  err.(FailedValidationError).Errors,
		})
	}

	var fiberErr *fiber.Error
	if !errors.As(err, &fiberErr) {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"errors":  nil,
		})
	}

	return c.Status(fiberErr.Code).JSON(fiber.Map{
		"message": fiberErr.Message,
		"errors":  nil,
	})
}

func handleFailedValidation(obj interface{}, err validator.ValidationErrors) Errors {

	objRef := reflect.TypeOf(obj)

	errMsgs := make(map[string]interface{})

	for i := 0; i < objRef.NumField(); i++ {
		structField := objRef.Field(i)
		errMsgs[structField.Tag.Get("json")] = nil
	}

	for _, err := range err {
		structField, _ := objRef.FieldByName(err.Field())
		field := structField.Tag.Get("json")
		errMsgs[field] = handleValidationErrorMessage(err.Tag(), err.Param(), field)
	}

	return errMsgs
}

func handleValidationErrorMessage(tag string, param string, field string) string {
	var msg string
	field = strings.Replace(field, "_", " ", -1)
	switch tag {
	case "required":
		msg = fmt.Sprintf("The %s field is required", field)
	case "email":
		msg = "This is not a valid email"
	case "min":
		msg = fmt.Sprintf("The %s field must be at least %s characters", strings.ToLower(field), param)
	case "max":
		msg = fmt.Sprintf("The %s field must be at most %s characters", strings.ToLower(field), param)
	case "eqfield":
		if param == "Password" {
			msg = "The password confirmation does not match"
		} else {
			msg = "The field does not match"
		}
	}

	return msg
}
