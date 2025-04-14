package apihandler

import (
	"AvitoPVZ/internal/ingress/gates/oapivalidation"
	"AvitoPVZ/internal/user"
	"AvitoPVZ/pkg/api/oapigen/pvzops"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	openapiTypes "github.com/oapi-codegen/runtime/types"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

func (h Handler) PostDummyLogin(w http.ResponseWriter, r *http.Request) {
	var req oapivalidation.APIPostDummyLoginJSONRequestBody
	err := h.unmarshalBody(r, &req)

	err = req.Validate()
	if err != nil {
		h.marshalResponse(w, http.StatusBadRequest, pvzops.Error{
			Message: err.Error(),
		})
		return
	}

	token, err := user.GenerateToken(uuid.New(), convertRoleToDomain(pvzops.UserRole(req.Role)))
	if err != nil {
		h.marshalResponse(w, http.StatusInternalServerError, pvzops.Error{
			Message: internalErrorMessage,
		})
		return
	}
	h.marshalResponse(w, http.StatusOK, token)
}

func (h Handler) PostLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req oapivalidation.APIPostLoginJSONRequestBody
	err := h.unmarshalBody(r, &req)
	if err != nil {
		h.marshalResponse(w, http.StatusBadRequest, pvzops.Error{
			Message: err.Error(),
		})
		return
	}

	token, err := h.userStory.Login(ctx, user.LoginRequest{
		Email:    string(req.Email),
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, user.ErrWrongPassword) || errors.Is(err, user.ErrUserNotFound) {
			h.marshalResponse(w, http.StatusUnauthorized, pvzops.Error{
				Message: err.Error(),
			})
			return
		}

		h.logger.ErrorContext(ctx, "failed to register story", slog.Any("error", err))
		h.marshalResponse(w, http.StatusInternalServerError, pvzops.Error{
			Message: internalErrorMessage,
		})
		return
	}

	h.marshalResponse(w, http.StatusOK, token)
}

func (h Handler) PostRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req oapivalidation.APIPostRegisterJSONRequestBody
	err := h.unmarshalBody(r, &req)
	if err != nil {
		h.marshalResponse(w, http.StatusBadRequest, pvzops.Error{
			Message: err.Error(),
		})
		return
	}

	u, err := h.userStory.Register(ctx, user.RegisterUserRequest{
		Email:    string(req.Email),
		Password: req.Password,
		Role:     convertRoleToDomain(pvzops.UserRole(req.Role)),
	})
	if err != nil {
		if errors.Is(err, user.ErrUserAlreadyExist) {
			h.marshalResponse(w, http.StatusBadRequest, pvzops.Error{
				Message: err.Error(),
			})
			return
		}

		h.logger.ErrorContext(ctx, "failed to register story", slog.Any("error", err))
		h.marshalResponse(w, http.StatusInternalServerError, pvzops.Error{
			Message: internalErrorMessage,
		})
		return
	}

	h.marshalResponse(w, http.StatusCreated, pvzops.User{
		Email: openapiTypes.Email(u.Email),
		Id:    lo.ToPtr(u.ID),
		Role:  convertRoleFromDomain(u.Role),
	})
}

func convertRoleToDomain(role pvzops.UserRole) user.Role {
	switch role {
	case pvzops.UserRoleModerator:
		return user.RoleModerator
	case pvzops.UserRoleEmployee:
		return user.RoleEmployee
	default:
		return user.RoleUnknown
	}
}

func convertRoleFromDomain(role user.Role) pvzops.UserRole {
	switch role {
	case user.RoleModerator:
		return pvzops.UserRoleModerator
	case user.RoleEmployee:
		return pvzops.UserRoleEmployee
	default:
		return pvzops.UserRole(role)
	}
}
