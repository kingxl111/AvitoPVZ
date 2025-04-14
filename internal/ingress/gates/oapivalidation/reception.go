package oapivalidation

import (
	"AvitoPVZ/pkg/api/oapigen/pvzops"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

type APIPostReceptionsJSONRequestBody struct {
	pvzops.PostReceptionsJSONRequestBody
}

func (b APIPostReceptionsJSONRequestBody) Validate() error {
	body := &b.PostReceptionsJSONRequestBody
	err := validation.ValidateStruct(
		body,
		validation.Field(&body.PvzId, validation.Required),
	)
	if err != nil {
		return errors.WithMessage(err, "validation")
	}
	return nil
}
