package orders

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (z *ordersRepository) ExceptExchangeBotUser(db *gorm.DB) *gorm.DB {
	db.Where("user_id != ? ", 1)
	return db
}

func (z *ordersRepository) JoinWithTradeOrders(db *gorm.DB) *gorm.DB {
	query := fmt.Sprintf("LEFT JOIN %s t ON %s.client_order_id = t.maker_order_id", z.repoConfig.TradeOrdersTable, z.repoConfig.OrdersTable)
	return db.Joins(query)
}

func (z *ordersRepository) GroupByOrderID(db *gorm.DB) *gorm.DB {
	return db.Group(fmt.Sprintf("%s.id", z.repoConfig.OrdersTable))
}

func (z *ordersRepository) SelectFieldsWithCommission(fields []string, targetStruct interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		defaultFields, err := z.builders.SelectFields(fields, z.repoConfig.OrdersTable, targetStruct)
		if err != nil {
			z.logger.Error("select fields failed", zap.Error(err))
		}

		commissionSelect := fmt.Sprintf("COALESCE(SUM(t.taker_commission), %s.commission) AS calculated_commission", z.repoConfig.OrdersTable)
		commissionSelectTRY := fmt.Sprintf("COALESCE(SUM(t.taker_commission_try), %s.commission_try) AS calculated_commission_try", z.repoConfig.OrdersTable)
		commissionSelectUSDT := fmt.Sprintf("COALESCE(SUM(t.taker_commission_usdt), %s.commission_usdt) AS calculated_commission_usdt", z.repoConfig.OrdersTable)

		selectFields := []string{commissionSelect, commissionSelectTRY, commissionSelectUSDT}

		selectFields = append([]string{defaultFields}, selectFields...)
		return db.Select(strings.Join(selectFields, ", "))
	}
}
