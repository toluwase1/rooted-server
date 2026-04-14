package middleware

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

// BindAndValidate parses the request body into the target struct and validates it
// using struct tags. Returns a 400 with a human-readable error message on failure.
// Usage in handlers:
//
//	var req CreateProfileRequest
//	if err := middleware.BindAndValidate(c, &req); err != nil {
//	    return err  // already sent 400 response
//	}
func BindAndValidate(c *fiber.Ctx, target interface{}) error {
	if err := c.BodyParser(target); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := validate.Struct(target); err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			return c.Status(400).JSON(fiber.Map{"error": "validation failed"})
		}

		messages := make([]string, 0, len(errs))
		for _, e := range errs {
			messages = append(messages, formatValidationError(e))
		}

		return c.Status(400).JSON(fiber.Map{
			"error":  strings.Join(messages, "; "),
			"fields": formatFieldErrors(errs),
		})
	}

	return nil
}

func formatValidationError(e validator.FieldError) string {
	field := toSnakeCase(e.Field())

	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		if e.Type().Kind().String() == "string" {
			return fmt.Sprintf("%s must be at least %s characters", field, e.Param())
		}
		return fmt.Sprintf("%s must have at least %s items", field, e.Param())
	case "max":
		if e.Type().Kind().String() == "string" {
			return fmt.Sprintf("%s must be at most %s characters", field, e.Param())
		}
		return fmt.Sprintf("%s must have at most %s items", field, e.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, e.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

func formatFieldErrors(errs validator.ValidationErrors) map[string]string {
	fields := make(map[string]string)
	for _, e := range errs {
		fields[toSnakeCase(e.Field())] = formatValidationError(e)
	}
	return fields
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(r + 32)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
