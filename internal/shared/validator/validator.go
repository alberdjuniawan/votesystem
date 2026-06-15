package validator

import "github.com/go-playground/validator/v10"

var validate = validator.New()

type FieldError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

func ValidateStruct(data interface{}) []*FieldError {
	if data == nil {
		return []*FieldError{{
			Field:   "body",
			Tag:     "required",
			Message: "Request body is required",
		}}
	}

	var errs []*FieldError

	err := validate.Struct(data)
	if err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			errs = append(errs, &FieldError{
				Field:   e.Field(),
				Tag:     e.Tag(),
				Value:   e.Param(),
				Message: msgForTag(e.Tag(), e.Param()),
			})
		}
	}

	return errs
}

func msgForTag(tag, param string) string {
	switch tag {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value too short, minimum length is " + param
	case "max":
		return "Value too long, maximum length is " + param
	case "alphanum":
		return "Only alphanumeric characters allowed"
	}
	return "Invalid value"
}
