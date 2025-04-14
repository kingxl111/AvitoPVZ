package apihandler

import (
	"AvitoPVZ/internal/product"
	"AvitoPVZ/internal/pvz"
	"AvitoPVZ/internal/receipt"
	"AvitoPVZ/internal/user"
	"context"
)

type (
	userStory interface {
		Login(ctx context.Context, req user.LoginRequest) (string, error)
		Register(ctx context.Context, req user.RegisterUserRequest) (*user.User, error)
	}

	pvzStory interface {
		Create(ctx context.Context, req pvz.CreatePVZRequest) (*pvz.PVZ, error)
	}

	productStory interface {
		AddProduct(ctx context.Context, req product.AddProductRequest) (*product.Product, error)
		Delete(ctx context.Context, req product.DeleteLastProductRequest) (n int64, err error)
	}

	receiptStory interface {
		Create(ctx context.Context, req receipt.CreateReceiptRequest) (*receipt.Receipt, error)
		CloseLast(ctx context.Context, req receipt.CloseLastReceiptRequest) (*receipt.Receipt, error)
	}
)
