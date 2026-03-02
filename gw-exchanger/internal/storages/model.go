package storages

type ExchangeRate struct {
	ID           int
	FromCurrency string
	ToCurrency   string
	Rate         float64
}
