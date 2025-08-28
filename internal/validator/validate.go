package validator

import "github.com/go-playground/validator/v10"

var v *validator.Validate

func Init() { v = validator.New() }

func V() *validator.Validate { return v }
