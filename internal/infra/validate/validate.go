package validate

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	v *validator.Validate
}

func New() *Validator {
	v := validator.New()

	return &Validator{v: v}
}

// Sturct ใช้เช็ค struct ที่มี tag validate
func (v *Validator) Struct(s interface{}) error {
	if err := v.v.Struct(s); err != nil {
		if vErrs, ok := err.(validator.ValidationErrors); ok {
			return wrapValidationErrors(vErrs)
		}
		return err
	}
	return nil
}

// FieldError เก็บรายละเอียด error ต่อ field
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag,omitempty"`
}

// ValodationError เป็น error รวมทุก field
type ValidationError struct {
	Fields []FieldError `json:"fields"`
}

func (e *ValidationError) Error() string {
	return "validation failed"
}

func wrapValidationErrors(vErrs validator.ValidationErrors) error {
	out := make([]FieldError, 0, len(vErrs))
	for _, fe := range vErrs {
		out = append(out, FieldError{
			Field:   fe.Field(),
			Tag:     fe.Tag(),
			Message: buildMessage(fe),
		})
	}
	return &ValidationError{Fields: out}
}

// buildMessage: ทำหน้าที่ customize ข้อความแต่ละ tag
func buildMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email", fe.Field())
	case "min":
		return fmt.Sprintf("must be at leat %s characters ", fe.Field())
	case "max":
		return fmt.Sprintf("must be at most %s characters ", fe.Field())
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}

// Helper ให้ handler เช็คว่า error เป็น validation error ไหม
func AsValidationError(err error) (*ValidationError, bool) {
	if err == nil {
		return nil, false
	}
	if vErr, ok := err.(*ValidationError); ok {
		return vErr, true
	}
	return nil, false
}
