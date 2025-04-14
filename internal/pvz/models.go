package pvz

import (
	"AvitoPVZ/internal/receipt"

	"github.com/google/uuid"
)

type City string

const (
	CityMoscow          City = "Москва"
	CitySaintPetersburg City = "Санкт-Петербург"
	CityKazan           City = "Казань"
	CityUnknown         City = "Неизвестный"
)

type PVZ struct {
	ID               uuid.UUID
	RegistrationDate string
	City             City
}

type CreatePVZRequest struct {
	ID               *uuid.UUID
	RegistrationDate *string
	City             City
}

type InsertPVZQuery struct {
	ID               *uuid.UUID
	RegistrationDate *string
	City             City
}

type ListPVZWithReceiptsRequest struct {
	StartDate *string
	EndDate   *string
	Page      *int
	Limit     *int
}

type ListPVZsQuery struct {
	StartDate *string
	EndDate   *string
	Page      *int
	Limit     *int
}

type PVZWithReceipts struct {
	PVZ      PVZ
	Receipts []receipt.Receipt
}
