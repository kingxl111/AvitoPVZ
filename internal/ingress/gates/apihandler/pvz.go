package apihandler

import (
	"AvitoPVZ/pkg/api/oapigen/pvzops"
	"net/http"
)

func (h Handler) GetPvz(w http.ResponseWriter, r *http.Request, params pvzops.GetPvzParams) {
	// TODO implement me
	panic("implement me")
}

func (h Handler) PostPvz(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = ctx
}
