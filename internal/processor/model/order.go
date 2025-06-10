package model

type OrderDirection int

const (
	OrderDirectionBID OrderDirection = 0
	OrderDirectionASK OrderDirection = 1
)

type OrderType string

const (
	OrderTypeLimit OrderType = "limit"
)

type Order struct {
	OrderId   int64
	Type      OrderType
	AccountId int64
	Direction OrderDirection
	Price     float64
	Amount    int64

	index int
}
