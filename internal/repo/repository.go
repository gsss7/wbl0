package repo

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"wbl0/internal/models"
)

type Repository struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (r *Repository) InsertOrUpdateOrder(ctx context.Context, o *models.Order) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	_, err = tx.Exec(ctx, `
        INSERT INTO orders(order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
        ON CONFLICT (order_uid) DO UPDATE SET
          track_number=EXCLUDED.track_number,
          entry=EXCLUDED.entry,
          locale=EXCLUDED.locale,
          internal_signature=EXCLUDED.internal_signature,
          customer_id=EXCLUDED.customer_id,
          delivery_service=EXCLUDED.delivery_service,
          shardkey=EXCLUDED.shardkey,
          sm_id=EXCLUDED.sm_id,
          date_created=EXCLUDED.date_created,
          oof_shard=EXCLUDED.oof_shard
    `,
		o.OrderUID, o.TrackNumber, o.Entry, o.Locale, o.InternalSignature, o.CustomerID, o.DeliveryService, o.ShardKey, o.SmID, o.DateCreated, o.OofShard)
	if err != nil {
		return err
	}

	d := o.Delivery
	_, err = tx.Exec(ctx, `
        INSERT INTO deliveries(order_uid, name, phone, zip, city, address, region, email)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
        ON CONFLICT (order_uid) DO UPDATE SET
          name=EXCLUDED.name, phone=EXCLUDED.phone, zip=EXCLUDED.zip, city=EXCLUDED.city,
          address=EXCLUDED.address, region=EXCLUDED.region, email=EXCLUDED.email
    `, o.OrderUID, d.Name, d.Phone, d.Zip, d.City, d.Address, d.Region, d.Email)
	if err != nil {
		return err
	}

	p := o.Payment
	_, err = tx.Exec(ctx, `
        INSERT INTO payments(order_uid, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
        ON CONFLICT (order_uid) DO UPDATE SET
          transaction=EXCLUDED.transaction, request_id=EXCLUDED.request_id, currency=EXCLUDED.currency, provider=EXCLUDED.provider,
          amount=EXCLUDED.amount, payment_dt=EXCLUDED.payment_dt, bank=EXCLUDED.bank, delivery_cost=EXCLUDED.delivery_cost,
          goods_total=EXCLUDED.goods_total, custom_fee=EXCLUDED.custom_fee
    `, o.OrderUID, p.Transaction, p.RequestID, p.Currency, p.Provider, p.Amount, p.PaymentDT, p.Bank, p.DeliveryCost, p.GoodsTotal, p.CustomFee)
	if err != nil {
		return err
	}

	batch := &pgx.Batch{}
	for _, it := range o.Items {
		batch.Queue(`
            INSERT INTO items(order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
            VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
            ON CONFLICT (order_uid, rid) DO UPDATE SET
              chrt_id=EXCLUDED.chrt_id, track_number=EXCLUDED.track_number, price=EXCLUDED.price,
              name=EXCLUDED.name, sale=EXCLUDED.sale, size=EXCLUDED.size, total_price=EXCLUDED.total_price,
              nm_id=EXCLUDED.nm_id, brand=EXCLUDED.brand, status=EXCLUDED.status
        `,
			o.OrderUID, it.ChrtID, it.TrackNumber, it.Price, it.RID, it.Name, it.Sale, it.Size, it.TotalPrice, it.NmID, it.Brand, it.Status)
	}
	br := tx.SendBatch(ctx, batch)
	if err = br.Close(); err != nil {
		return err
	}

	err = tx.Commit(ctx)
	return err
}

func (r *Repository) GetOrder(ctx context.Context, id string) (*models.Order, error) {
	var o models.Order
	err := r.pool.QueryRow(ctx, `SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders WHERE order_uid=$1`, id).
		Scan(&o.OrderUID, &o.TrackNumber, &o.Entry, &o.Locale, &o.InternalSignature, &o.CustomerID, &o.DeliveryService, &o.ShardKey, &o.SmID, &o.DateCreated, &o.OofShard)
	if err != nil {
		return nil, err
	}

	err = r.pool.QueryRow(ctx, `SELECT name, phone, zip, city, address, region, email FROM deliveries WHERE order_uid=$1`, id).
		Scan(&o.Delivery.Name, &o.Delivery.Phone, &o.Delivery.Zip, &o.Delivery.City, &o.Delivery.Address, &o.Delivery.Region, &o.Delivery.Email)
	if err != nil {
		return nil, err
	}

	err = r.pool.QueryRow(ctx, `SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee FROM payments WHERE order_uid=$1`, id).
		Scan(&o.Payment.Transaction, &o.Payment.RequestID, &o.Payment.Currency, &o.Payment.Provider, &o.Payment.Amount, &o.Payment.PaymentDT, &o.Payment.Bank, &o.Payment.DeliveryCost, &o.Payment.GoodsTotal, &o.Payment.CustomFee)
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items WHERE order_uid=$1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var it models.Item
		if err := rows.Scan(&it.ChrtID, &it.TrackNumber, &it.Price, &it.RID, &it.Name, &it.Sale, &it.Size, &it.TotalPrice, &it.NmID, &it.Brand, &it.Status); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, it)
	}
	return &o, nil
}

func (r *Repository) LoadRecentOrders(ctx context.Context, n int) ([]*models.Order, error) {
	if n <= 0 {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx, `SELECT order_uid FROM orders ORDER BY date_created DESC LIMIT $1`, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]string, 0, n)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}

	out := make([]*models.Order, 0, len(ids))
	for _, id := range ids {
		o, err := r.GetOrder(ctx, id)
		if err == nil {
			out = append(out, o)
		}
	}
	return out, nil
}
