package oapivalidation

import (
	"AvitoPVZ/pkg/api/oapigen/pvzops"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

type APIPostProductsJSONRequestBody struct {
	pvzops.PostProductsJSONRequestBody
}

func (b APIPostProductsJSONRequestBody) Validate() error {
	body := &b.PostProductsJSONRequestBody
	err := validation.ValidateStruct(
		body,
		validation.Field(&body.PvzId, validation.Required),
		validation.Field(&body.Type,
			validation.Required,
			validation.In(
				pvzops.PostProductsJSONBodyTypeОбувь,
				pvzops.PostProductsJSONBodyTypeОдежда,
				pvzops.PostProductsJSONBodyTypeЭлектроника,
			),
		),
	)
	if err != nil {
		return errors.WithMessage(err, "validation")
	}
	return nil
}
