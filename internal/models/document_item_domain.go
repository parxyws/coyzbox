package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// DocumentItem represents a line item on any document.
type DocumentItem struct {
	Id          string          `json:"id" gorm:"column:id;primaryKey"`
	DocumentId  string          `json:"document_id" gorm:"column:document_id"`
	SortOrder   int             `json:"sort_order" gorm:"column:sort_order;default:0"`
	Description string          `json:"description" gorm:"column:description"`
	Quantity    decimal.Decimal `json:"quantity" gorm:"column:quantity;default:1"`
	Unit        string          `json:"unit" gorm:"column:unit"`
	UnitPrice   decimal.Decimal `json:"unit_price" gorm:"column:unit_price"`
	DiscountPct decimal.Decimal `json:"discount_pct" gorm:"column:discount_pct;default:0"`
	TaxPct      decimal.Decimal `json:"tax_pct" gorm:"column:tax_pct;default:0"`
	Amount      decimal.Decimal `json:"amount" gorm:"column:amount"`
	CreatedAt   time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time       `json:"updated_at" gorm:"column:updated_at"`
}

func (di *DocumentItem) TableName() string {
	return "document_items"
}

// CalculateAmount computes: (quantity × unit_price) - discount + tax
func (di *DocumentItem) CalculateAmount() decimal.Decimal {
	gross := di.Quantity.Mul(di.UnitPrice)
	discount := gross.Mul(di.DiscountPct).Div(decimal.NewFromInt(100))
	afterDiscount := gross.Sub(discount)
	tax := afterDiscount.Mul(di.TaxPct).Div(decimal.NewFromInt(100))
	return afterDiscount.Add(tax)
}
