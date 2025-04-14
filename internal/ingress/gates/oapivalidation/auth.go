package oapivalidation

import (
	"AvitoPVZ/pkg/api/oapigen/pvzops"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

type APIPostDummyLoginJSONRequestBody struct {
	pvzops.PostDummyLoginJSONRequestBody
}

func (b APIPostDummyLoginJSONRequestBody) Validate() error {
	body := &b.PostDummyLoginJSONRequestBody
	err := validation.ValidateStruct(
		body,
		validation.Field(
			&body.Role,
			validation.Required,
			validation.In(pvzops.Employee, pvzops.Moderator),
		),
	)
	if err != nil {
		return errors.WithMessage(err, "validation")
	}

	return nil
}

type APIPostRegisterJSONRequestBody struct {
	pvzops.PostRegisterJSONRequestBody
}

func (b APIPostRegisterJSONRequestBody) Validate() error {
	body := &b.PostRegisterJSONRequestBody
	err := validation.ValidateStruct(
		body,
		validation.Field(&body.Role, validation.Required, validation.In(pvzops.Employee, pvzops.Moderator)),
		validation.Field(&body.Email, validation.Required),
		validation.Field(&body.Password, validation.Required),
	)
	if err != nil {
		return errors.WithMessage(err, "validation")
	}

	return nil
}

type APIPostLoginJSONRequestBody struct {
	pvzops.PostLoginJSONRequestBody
}

func (b APIPostLoginJSONRequestBody) Validate() error {
	body := &b.PostLoginJSONRequestBody
	err := validation.ValidateStruct(
		body,
		validation.Field(&body.Email, validation.Required),
		validation.Field(&body.Password, validation.Required),
	)
	if err != nil {
		return errors.WithMessage(err, "validation")
	}

	return nil
}
