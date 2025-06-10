package model

import (
	"context"

	"playGround/pkg/tracer"
)

type AccountStatistic struct {
	AccountId   int64
	Saldo       float64
	Position    int64
	Turnover    float64
	TradeAmount int64
}

type AccountStatisticStorage map[int64]*AccountStatistic

func (ass AccountStatisticStorage) Update(ctx context.Context, deal *Deal) {
	_, span := tracer.Start(ctx, "internal.processor.model.Update")
	defer span.End()

	if ass[deal.AccountId] == nil {
		ass[deal.AccountId] = &AccountStatistic{}
	}
	if ass[deal.CounterpartAccountId] == nil {
		ass[deal.CounterpartAccountId] = &AccountStatistic{}
	}
	total := deal.Price * float64(deal.Amount)

	if deal.Direction == OrderDirectionBID {
		ass[deal.AccountId].Saldo -= total
		ass[deal.AccountId].Position += deal.Amount
		ass[deal.CounterpartAccountId].Saldo += total
		ass[deal.CounterpartAccountId].Position -= deal.Amount
	} else {
		ass[deal.AccountId].Saldo += total
		ass[deal.AccountId].Position -= deal.Amount
		ass[deal.CounterpartAccountId].Saldo -= total
		ass[deal.CounterpartAccountId].Position += deal.Amount
	}

	ass[deal.AccountId].Turnover += total
	ass[deal.AccountId].TradeAmount += deal.Amount
	ass[deal.CounterpartAccountId].Turnover += total
	ass[deal.CounterpartAccountId].TradeAmount += deal.Amount
}
