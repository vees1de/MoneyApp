package common

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

type Money struct {
	decimal.Decimal
}

func NewMoney(value string) (Money, error) {
	parsed, err := decimal.NewFromString(value)
	if err != nil {
		return Money{}, err
	}

	return Money{Decimal: parsed}, nil
}

func MustMoney(value string) Money {
	money, err := NewMoney(value)
	if err != nil {
		panic(err)
	}

	return money
}

func ZeroMoney() Money {
	return Money{Decimal: decimal.Zero}
}

func (m Money) Add(other Money) Money {
	return Money{Decimal: m.Decimal.Add(other.Decimal)}
}

func (m Money) Sub(other Money) Money {
	return Money{Decimal: m.Decimal.Sub(other.Decimal)}
}

func (m Money) Neg() Money {
	return Money{Decimal: m.Decimal.Neg()}
}

func (m Money) IsZero() bool {
	return m.Decimal.Equal(decimal.Zero)
}

func (m Money) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.StringFixedBank(2))
}

func (m *Money) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	parsed, err := decimal.NewFromString(raw)
	if err != nil {
		return err
	}

	m.Decimal = parsed
	return nil
}

func (m Money) Value() (driver.Value, error) {
	return m.StringFixedBank(2), nil
}

func (m *Money) Scan(src any) error {
	switch value := src.(type) {
	case nil:
		m.Decimal = decimal.Zero
		return nil
	case string:
		return m.fromString(value)
	case []byte:
		return m.fromString(string(value))
	case int64:
		m.Decimal = decimal.NewFromInt(value)
		return nil
	case float64:
		m.Decimal = decimal.NewFromFloat(value)
		return nil
	default:
		return fmt.Errorf("unsupported money source %T", src)
	}
}

func (m *Money) fromString(value string) error {
	parsed, err := decimal.NewFromString(value)
	if err != nil {
		return err
	}

	m.Decimal = parsed
	return nil
}
