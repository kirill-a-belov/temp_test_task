package processor

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"

	"playGround/internal/processor/model"
	"playGround/pkg/tracer"
)

func (m *Module) CSVReadOrderList(ctx context.Context) ([]*model.Order, error) {
	_, span := tracer.Start(ctx, "internal.processor.CSVReadOrderList")
	defer span.End()

	file, err := os.Open(m.config.OrderFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	const orderFileSeparator = ','
	reader.Comma = orderFileSeparator

	const orderRecordFieldsCount = 6
	reader.FieldsPerRecord = orderRecordFieldsCount

	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("error reading header: %v", err)
	}

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var orders []*model.Order
	for _, record := range records {
		order, err := parseOrder(record)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func parseOrder(record []string) (*model.Order, error) {
	var order model.Order
	var err error

	order.OrderId, err = strconv.ParseInt(record[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid OrderId: %v", err)
	}

	switch record[1] {
	case "limit":
		order.Type = model.OrderTypeLimit
	default:
		return nil, fmt.Errorf("invalid Order Type: %s", record[1])
	}

	order.AccountId, err = strconv.ParseInt(record[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid AccountID: %v", err)
	}

	dir, err := strconv.Atoi(record[3])
	if err != nil {
		return nil, fmt.Errorf("invalid Direction: %v", err)
	}
	switch model.OrderDirection(dir) {
	case model.OrderDirectionBID, model.OrderDirectionASK:
		order.Direction = model.OrderDirection(dir)
	default:
		return nil, fmt.Errorf("invalid Direction value: %d", dir)
	}

	order.Price, err = strconv.ParseFloat(record[4], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid Price: %v", err)
	}

	// Парсим Amount
	order.Amount, err = strconv.ParseInt(record[5], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid Amount: %v", err)
	}

	return &order, nil
}

func (m *Module) CSVWriteAccountStatistic(ctx context.Context, accountStatisticMap map[int64]*model.AccountStatistic) error {
	_, span := tracer.Start(ctx, "internal.processor.CSVWriteAccountStatistic")
	defer span.End()

	accountStatisticList := make([]*model.AccountStatistic, len(accountStatisticMap))
	for i, accountStatistic := range accountStatisticMap {
		accountStatisticList[i] = accountStatistic
	}
	sort.Slice(accountStatisticList, func(i, j int) bool {
		return accountStatisticList[i].AccountId < accountStatisticList[j].AccountId
	})

	file, err := os.Create(m.config.AccountStatisticFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Записываем заголовок
	headers := []string{"account_id", "saldo", "position", "turnover", "trade_amount"}
	if err := writer.Write(headers); err != nil {
		return err
	}

	// Записываем данные
	for _, stat := range accountStatisticList {
		record := []string{
			strconv.FormatInt(stat.AccountId, 10),
			strconv.FormatFloat(stat.Saldo, 'f', 2, 64),
			strconv.FormatInt(stat.Position, 10),
			strconv.FormatFloat(stat.Turnover, 'f', 2, 64),
			strconv.FormatInt(stat.TradeAmount, 10),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}
