package processor

import (
	"container/heap"
	"context"

	"github.com/pkg/errors"

	"playGround/internal/processor/model"
	"playGround/pkg/tracer"
)

func (m *Module) Process(ctx context.Context) error {
	_, span := tracer.Start(ctx, "internal.processor.Process")
	defer span.End()

	orderList, err := m.CSVReadOrderList(ctx)
	if err != nil {
		return errors.Wrap(err, "csv read order list")
	}

	var (
		bidHeap model.OrderHeap
		askHeap model.OrderHeap
	)
	heap.Init(&bidHeap)
	heap.Init(&askHeap)

	accountStatisticStorage := make(model.AccountStatisticStorage)

	for _, order := range orderList {
		for order.Amount > 0 {
			var counterHeap *model.OrderHeap
			if order.Direction == model.OrderDirectionBID {
				counterHeap = &askHeap
			} else {
				counterHeap = &bidHeap
			}
			if counterHeap.Len() == 0 {
				break
			}
			bestCounterOrder := (*counterHeap)[0]

			isBidOrderWithHighPrice := order.Direction == model.OrderDirectionBID && order.Price >= bestCounterOrder.Price
			if !isBidOrderWithHighPrice {
				break
			}
			isAskOrderWithLowPrice := order.Direction == model.OrderDirectionASK && order.Price <= bestCounterOrder.Price
			if !isAskOrderWithLowPrice {
				break
			}

			deal := &model.Deal{
				Direction: order.Direction,
			}

			deal.Amount = bestCounterOrder.Amount
			if order.Amount < deal.Amount {
				deal.Amount = order.Amount
			}

			deal.Price = bestCounterOrder.Price
			if bestCounterOrder.OrderId > order.OrderId {
				deal.Price = order.Price
			}

			order.Amount -= deal.Amount
			bestCounterOrder.Amount -= deal.Amount

			accountStatisticStorage.Update(ctx, deal)

			if bestCounterOrder.Amount == 0 {
				heap.Pop(counterHeap)
			}
		}

		if order.Amount > 0 {
			if order.Direction == model.OrderDirectionBID {
				heap.Push(&bidHeap, order)
			} else {
				heap.Push(&askHeap, order)
			}
		}

		return nil
	}

	if err := m.CSVWriteAccountStatistic(ctx, accountStatisticStorage); err != nil {
		return errors.Wrap(err, "csv write account statistic")
	}

	return nil
}
