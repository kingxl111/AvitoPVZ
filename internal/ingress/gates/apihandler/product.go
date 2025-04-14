package apihandler

import (
	"AvitoPVZ/internal/infra/httpfunc/middleware/authmw"
	"AvitoPVZ/internal/user"
	"net/http"

	openapiTypes "github.com/oapi-codegen/runtime/types"
)

func (h Handler) PostProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claim := authmw.GetClaims(r)
	_ = ctx
	if claim == nil || claim.Role == user.RoleUnknown {

	}

}

func (h Handler) PostPvzPvzIdDeleteLastProduct(w http.ResponseWriter, r *http.Request, pvzId openapiTypes.UUID) {
	// TODO implement me
	panic("implement me")
}
