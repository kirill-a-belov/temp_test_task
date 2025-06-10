package model

type Deal struct {
	AccountId            int64
	Direction            OrderDirection
	CounterpartAccountId int64
	Price                float64
	Amount               int64
}
