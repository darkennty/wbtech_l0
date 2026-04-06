package repository

import (
	"WBTech_L0/internal/model"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type OrderPostgres struct {
	db *sqlx.DB
}

func NewOrderPostgres(db *sqlx.DB) *OrderPostgres {
	return &OrderPostgres{db: db}
}

func (r *OrderPostgres) Insert(order model.Order) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	query := fmt.Sprintf("INSERT INTO %s (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11);", orderTable)
	_, err = tx.Exec(query, order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature, order.CustomerID, order.DeliveryService, order.Shardkey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return rollbackErr
		}
		return err
	}

	query = fmt.Sprintf("INSERT INTO %s (order_uid, name, phone, zip, city, address, region, email) VALUES ($1,$2,$3,$4,$5,$6,$7,$8);", deliveryTable)
	_, err = tx.Exec(query, order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return rollbackErr
		}
		return err
	}

	query = fmt.Sprintf("INSERT INTO %s (order_uid, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11);", paymentTable)
	_, err = tx.Exec(query, order.OrderUID, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency, order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return rollbackErr
		}
		return err
	}

	for _, v := range order.Items {
		query = fmt.Sprintf("INSERT INTO %s (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12);", itemTable)
		_, err = tx.Exec(query, order.OrderUID, v.ChrtID, v.TrackNumber, v.Price, v.Rid, v.Name, v.Sale, v.Size, v.TotalPrice, v.NmID, v.Brand, v.Status)
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				return err
			}
			return err
		}
	}

	return tx.Commit()
}

func (r *OrderPostgres) GetOrderByID(orderUUID uuid.UUID) (model.Order, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return model.Order{}, err
	}

	var order model.Order

	query := fmt.Sprintf(`
		SELECT 
			o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature, 
			o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
			d.name as delivery_name, d.phone as delivery_phone, d.zip as delivery_zip, 
			d.city as delivery_city, d.address as delivery_address, d.region as delivery_region, d.email as delivery_email,
			p.transaction, p.request_id, p.currency, p.provider, p.amount, 
			p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee
		FROM %s o
		LEFT JOIN %s d ON o.order_uid = d.order_uid
		LEFT JOIN %s p ON o.order_uid = p.order_uid
		WHERE o.order_uid = $1;
	`, orderTable, deliveryTable, paymentTable)

	var result model.DBResult

	if err = r.db.Get(&result, query, orderUUID); err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return model.Order{}, rollbackErr
		}
		return model.Order{}, err
	}

	order.OrderUID = result.OrderUID
	order.TrackNumber = result.TrackNumber
	order.Entry = result.Entry
	order.Locale = result.Locale
	order.InternalSignature = result.InternalSignature
	order.CustomerID = result.CustomerID
	order.DeliveryService = result.DeliveryService
	order.Shardkey = result.Shardkey
	order.SmID = result.SmID
	order.DateCreated = result.DateCreated
	order.OofShard = result.OofShard

	order.Delivery = model.Delivery{
		Name:    result.DeliveryName,
		Phone:   result.DeliveryPhone,
		Zip:     result.DeliveryZip,
		City:    result.DeliveryCity,
		Address: result.DeliveryAddress,
		Region:  result.DeliveryRegion,
		Email:   model.Email(result.DeliveryEmail),
	}

	order.Payment = model.Payment{
		Transaction:  result.PaymentTransaction,
		RequestID:    result.PaymentRequestID,
		Currency:     result.PaymentCurrency,
		Provider:     result.PaymentProvider,
		Amount:       result.PaymentAmount,
		PaymentDt:    result.PaymentPaymentDt,
		Bank:         result.PaymentBank,
		DeliveryCost: result.PaymentDeliveryCost,
		GoodsTotal:   result.PaymentGoodsTotal,
		CustomFee:    result.PaymentCustomFee,
	}

	itemQuery := fmt.Sprintf(`
		SELECT i.chrt_id, i.track_number, i.price, i.rid, i.name, i.sale, 
		       i.size, i.total_price, i.nm_id, i.brand, i.status 
		FROM %s i 
		WHERE i.order_uid = $1;
	`, itemTable)

	if err = r.db.Select(&order.Items, itemQuery, orderUUID); err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return model.Order{}, rollbackErr
		}
		return model.Order{}, err
	}

	if err = tx.Commit(); err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return model.Order{}, rollbackErr
		}
		return model.Order{}, err
	}

	return order, nil
}

func (r *OrderPostgres) GetOrdersForCache(capacity int) ([]model.Order, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	var orders []model.Order

	query := fmt.Sprintf("SELECT * FROM %s LIMIT $1;", orderTable)
	if err = r.db.Select(&orders, query, capacity); err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return nil, rollbackErr
		}
		return nil, err
	}

	for i := 0; i < len(orders); i++ {
		order := &orders[i]

		query = fmt.Sprintf("SELECT d.name, d.phone, d.zip, d.city, d.address, d.region, d.email FROM %s d WHERE d.order_uid = $1;", deliveryTable)
		if err = r.db.Get(&(order.Delivery), query, order.OrderUID); err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				return nil, rollbackErr
			}
			return nil, err
		}

		query = fmt.Sprintf("SELECT p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee FROM %s p WHERE p.order_uid = $1;", paymentTable)
		if err = r.db.Get(&(order.Payment), query, order.OrderUID); err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				return nil, rollbackErr
			}
			return nil, err
		}

		query = fmt.Sprintf("SELECT i.chrt_id, i.track_number, i.price, i.rid, i.name, i.sale, i.size, i.total_price, i.nm_id, i.brand, i.status FROM %s i WHERE i.order_uid = $1;", itemTable)
		if err = r.db.Select(&(order.Items), query, order.OrderUID); err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				return nil, rollbackErr
			}
			return nil, err
		}
	}

	return orders, tx.Commit()
}
