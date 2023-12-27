package orders

import (
	"github.com/denizumutdereli/stream-admin/internal/dsl"
	"github.com/go-playground/validator"
)

var validateOrder *validator.Validate

type Order struct {
	ID                       string  `gorm:"primary_key;type:text" json:"id" validate:"required,uuid"`
	ClientOrderID            string  `gorm:"type:text" json:"client_order_id" validate:"required"`
	UserID                   int64   `gorm:"type:bigint" json:"user_id" validate:"required,gte=0"`
	Side                     string  `gorm:"type:varchar(255)" json:"side" validate:"required,eq=buy|eq=sell"`
	Market                   string  `gorm:"type:varchar(255)" json:"market" validate:"required"`
	Price                    float64 `gorm:"type:numeric" json:"price" validate:"required,gte=0"`
	StopPrice                float64 `gorm:"type:numeric" json:"stop_price" validate:"omitempty,gte=0"`
	Quantity                 float64 `gorm:"type:numeric" json:"quantity" validate:"required,gte=0"`
	QuoteAssetQuantity       float64 `gorm:"type:numeric" json:"quote_asset_quantity" validate:"omitempty,gte=0"`
	ExecutedQuantity         float64 `gorm:"type:numeric" json:"executed_quantity" validate:"omitempty,gte=0"`
	CumulativeQuoteQuantity  float64 `gorm:"type:numeric" json:"cumulative_quote_quantity" validate:"omitempty,gte=0"`
	Status                   string  `gorm:"type:varchar(255)t" json:"status" validate:"required"`
	TimeInForce              string  `gorm:"type:varchar(255)" json:"time_in_force"`
	MatchEngine              string  `gorm:"type:varchar(255)" json:"match_engine" validate:"required"`
	MetaData                 string  `gorm:"type:text" json:"meta_data"`
	Dust                     float64 `gorm:"type:numeric" json:"dust" validate:"omitempty,gte=0"`
	Commission               float64 `gorm:"type:numeric" json:"commission" validate:"omitempty,gte=0"`
	Type                     string  `gorm:"type:varchar(255)" json:"type"`
	CommissionTRY            float64 `gorm:"type:numeric" json:"commission_try" validate:"omitempty,gte=0"`
	CommissionUSDT           float64 `gorm:"type:numeric" json:"commission_usdt" validate:"omitempty,gte=0"`
	CalculatedCommission     float64 `gorm:"-" json:"calculated_commission"`
	CalculatedCommissionTRY  float64 `gorm:"-" json:"calculated_commission_try"`
	CalculatedCommissionUSDT float64 `gorm:"-" json:"calculated_commission_usdt"`
	CreatedAt                string  `gorm:"type:varchar(255)" json:"created_at"`   // timestampz fix!
	UpdatedAt                string  `gorm:"type:varchar(255)" json:"updated_at"`   // timestampz fix!
	CancelledAt              string  `gorm:"type:varchar(255)" json:"cancelled_at"` //timesmapz fix!
	//DeletedAt                time.Time `gorm:"-" json:"deleted_at"`                   // not exist!
}

type OrderSearch struct {
	ClientOrderID           *string  `form:"client_order_id"`
	UserID                  *int64   `form:"user_id"`
	Side                    *string  `form:"side"`
	Market                  *string  `form:"market"`
	Price                   *float64 `form:"price"`
	StopPrice               *float64 `form:"stop_price"`
	Quantity                *float64 `form:"quantity"`
	QuoteAssetQuantity      *float64 `form:"quote_asset_quantity"`
	ExecutedQuantity        *float64 `form:"executed_quantity"`
	CumulativeQuoteQuantity *float64 `form:"cumulative_quote_quantity"`
	Status                  *string  `form:"status"`
	TimeInForce             *string  `form:"time_in_force"`
	MatchEngine             *string  `form:"match_engine"`
	CreatedAt               *string  `form:"created_at" time_format:"2022-11-03T09:06:11.000000Z"`
	UpdatedAt               *string  `form:"updated_at" time_format:"2022-11-03T09:06:11.000000Z"`
	CancelledAt             *string  `form:"cancelled_at" time_format:"2022-11-03T09:06:11.000000Z"`
	Dust                    *float64 `form:"dust"`
	Commission              *float64 `form:"commission"`
	Type                    *string  `form:"type"`
	CommissionTRY           *float64 `form:"commission_try"`
	CommissionUSDT          *float64 `form:"commission_usdt"`
	dsl.DSLFields           `gorm:"-" json:"-"`
}

type OrderUpdate struct {
	ID            string `gorm:"primary_key;type:text" json:"id" validate:"required,uuid"`
	ClientOrderID string `gorm:"type:text" json:"client_order_id" validate:"required"`
	UpdatedAt     string `gorm:"type:text" json:"updated_at" validate:"-"`
}

func ValidateOrder(order *Order) error {
	return validateOrder.Struct(order)
}

// func (o *Order) BeforeCreate(tx *gorm.DB) (err error) {
// 	if o.CreatedAt == nil {
// 		var now time.Time = time.Now()
// 		o.CreatedAt = &now
// 	}
// 	return
// }

// func (o *Order) BeforeUpdate(tx *gorm.DB) (err error) {
// 	var now time.Time = time.Now()
// 	o.UpdatedAt = &now

// 	return
// }

// func (o *Order) BeforeSave(tx *gorm.DB) (err error) {
// 	var now time.Time = time.Now()
// 	o.UpdatedAt = &now

// 	return
// }
