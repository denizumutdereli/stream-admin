package models

import (
	"github.com/denizumutdereli/stream-admin/internal/dsl"
	"github.com/go-playground/validator"
)

var validateAssets *validator.Validate

type Assets struct {
	ID                                int64   `gorm:"primaryKey;type:bigint" json:"id" validate:"required"`
	Name                              string  `gorm:"type:varchar(255)" json:"name" validate:"required"`
	Enabled                           int     `gorm:"type:smallint" json:"enabled" validate:"required"`
	BaseAssetPrecision                int16   `gorm:"type:smallint" json:"base_asset_precision" validate:"required"`
	QuoteAssetPrecision               int16   `gorm:"type:smallint" json:"quote_asset_precision" validate:"required"`
	IcebergAllowed                    int     `gorm:"type:smallint" json:"iceberg_allowed" validate:"required"`
	OcoAllowed                        int     `gorm:"type:smallint" json:"oco_allowed" validate:"required"`
	IsSpotTradingAllowed              int     `gorm:"type:smallint" json:"is_spot_trading_allowed" validate:"required"`
	IsMarketTradingAllowed            int     `gorm:"type:smallint" json:"is_market_trading_allowed" validate:"required"`
	ExternalLiquidity                 string  `gorm:"type:varchar(255)" json:"external_liquidity" validate:"required"`
	BuyTolerance                      float64 `gorm:"type:numeric" json:"buy_tolerance" validate:"required"`
	SellTolerance                     float64 `gorm:"type:numeric" json:"sell_tolerance" validate:"required"`
	PriceFilterMin                    float64 `gorm:"type:numeric" json:"price_filter_min" validate:"required"`
	PriceFilterMax                    float64 `gorm:"type:numeric" json:"price_filter_max" validate:"required"`
	PriceFilterStepSize               float64 `gorm:"type:numeric" json:"price_filter_step_size" validate:"required"`
	QuantityFilterMin                 float64 `gorm:"type:numeric" json:"quantity_filter_min" validate:"required"`
	QuantityFilterMax                 float64 `gorm:"type:numeric" json:"quantity_filter_max" validate:"required"`
	QuantityFilterStepSize            float64 `gorm:"type:numeric" json:"quantity_filter_step_size" validate:"required"`
	NotionalFilterMin                 float64 `gorm:"type:numeric" json:"notional_filter_min" validate:"required"`
	NotionalFilterAverageToMarket     int16   `gorm:"type:numeric" json:"notional_filter_average_to_market" validate:"required"`
	NotionalFilterAveragePriceMinutes int16   `gorm:"type:numeric" json:"notional_filter_average_price_minutes" validate:"required"`
	BaseAssetId                       int16   `gorm:"type:numeric" json:"base_asset_id" validate:"required"`
	QuoteAssetId                      int16   `gorm:"type:numeric" json:"quote_asset_id" validate:"required"`
	MakerCommission                   float64 `gorm:"type:numeric" json:"maker_commission" validate:"required"`
	TakerCommission                   float64 `gorm:"type:numeric" json:"taker_commission" validate:"required"`
	QuoteOrderQtyMarketAllowed        int     `gorm:"type:smallint" json:"quote_order_qty_market_allowed" validate:"required"`
	MarketOrderAllowed                int     `gorm:"type:smallint" json:"market_order_allowed" validate:"required"`
	IsMarginTradingAllowed            int     `gorm:"type:smallint" json:"is_margin_trading_allowed" validate:"required"`
	NotionalFilterApplyToMarket       int     `gorm:"type:numeric" json:"notional_filter_apply_to_market" validate:"required"`
	CreatedAt                         int64   `gorm:"type:bigint" json:"created_at"`
	UpdatedAt                         int64   `gorm:"type:bigint" json:"updated_at"`
	DeletedAt                         int64   `gorm:"type:bigint" json:"deleted_at"`
}

type AssetsSearch struct {
	ID                                *int64   `form:"id"`
	Name                              *string  `form:"name"`
	Enabled                           *int     `form:"enabled"`
	BaseAssetPrecision                *int16   `form:"base_asset_precision"`
	QuoteAssetPrecision               *int16   `form:"quote_asset_precision"`
	IcebergAllowed                    *int     `form:"iceberg_allowed"`
	OcoAllowed                        *int     `form:"oco_allowed"`
	IsSpotTradingAllowed              *int     `form:"is_spot_trading_allowed"`
	IsMarketTradingAllowed            *int     `form:"is_market_trading_allowed"`
	ExternalLiquidity                 *string  `form:"external_liquidity"`
	BuyTolerance                      *float64 `form:"buy_tolerance"`
	SellTolerance                     *float64 `form:"sell_tolerance"`
	PriceFilterMin                    *float64 `form:"price_filter_min"`
	PriceFilterMax                    *float64 `form:"price_filter_max"`
	PriceFilterStepSize               *float64 `form:"price_filter_step_size"`
	QuantityFilterMin                 *float64 `form:"quantity_filter_min"`
	QuantityFilterMax                 *float64 `form:"quantity_filter_max"`
	QuantityFilterStepSize            *float64 `form:"quantity_filter_step_size"`
	NotionalFilterMin                 *float64 `form:"notional_filter_min"`
	NotionalFilterAverageToMarket     *int16   `form:"notional_filter_average_to_market"`
	NotionalFilterAveragePriceMinutes *int16   `form:"notional_filter_average_price_minutes"`
	BasseAssetId                      *int16   `form:"base_asset_id"`
	QuoteAssetId                      *int16   `form:"quote_asset_id"`
	MakerCommission                   *float64 `form:"maker_commission"`
	TakerCommission                   *float64 `form:"taker_commission"`
	QuoteOrderQtyMarketAllowed        *int     `form:"quote_order_qty_market_allowed"`
	MarketOrderAllowed                *int     `form:"market_order_allowed"`
	IsMarginTradingAllowed            *int     `form:"is_margin_trading_allowed"`
	NotionalFilterApplyToMarket       *int     `form:"notional_filter_apply_to_market"`
	CreatedAt                         *int64   `form:"created_at"`
	UpdatedAt                         *int64   `form:"updated_at"`
	DeletedAt                         *int64   `form:"deleted_at"`
	dsl.DSLFields                     `gorm:"-" json:"-"`
}

func ValidateAssets(asset *Assets) error {
	return validateAssets.Struct(asset)
}
