package service

import "errors"

var (
	ErrInvalidRequest     = errors.New("invalid request")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidOrderNumber = errors.New("invalid order number")
	ErrNegativeOrZeroSum  = errors.New("cannot withdraw negative or zero sum")
	ErrNoWithdraws        = errors.New("no withdrawals found")

	ErrLoginTaken           = errors.New("login already taken")
	ErrOrderAlreadyUploaded = errors.New("order already uploaded by this user")
	ErrOrderUploadedByOther = errors.New("order already uploaded by another user")
	ErrInsufficientFunds    = errors.New("insufficient funds")
)
