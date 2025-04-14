package oapivalidation

import (
	"AvitoPVZ/pkg/api/oapigen/pvzops"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

type APIPostPvzJSONRequestBody struct {
	pvzops.PostPvzJSONRequestBody
}

func (b APIPostPvzJSONRequestBody) Validate() error {
	body := &b.PostPvzJSONRequestBody
	err := validation.ValidateStruct(
		body,
		validation.Field(&body.City,
			validation.Required,
			validation.In(
				pvzops.Москва,
				pvzops.СанктПетербург,
				pvzops.Казань,
			),
		),
	)
	if err != nil {
		return errors.WithMessage(err, "validation")
	}
	return nil
}
