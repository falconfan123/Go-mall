package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/falconfan123/Go-mall/dal/model/order"
	"github.com/falconfan123/Go-mall/services/order/internal/domain/entity"
	"github.com/falconfan123/Go-mall/services/order/internal/domain/repository"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	ErrOrderNotFound = errors.New("order not found")
)

// OrderRepositoryImpl 订单仓储实现
type OrderRepositoryImpl struct {
	conn                sqlx.SqlConn
	ordersModel         order.OrdersModel
	orderItemsModel     order.OrderItemsModel
	orderAddressesModel order.OrderAddressesModel
}

// NewOrderRepositoryImpl 创建订单仓储实现
func NewOrderRepositoryImpl(conn sqlx.SqlConn) repository.OrderRepository {
	return &OrderRepositoryImpl{
		conn:                conn,
		ordersModel:         order.NewOrdersModel(conn),
		orderItemsModel:     order.NewOrderItemsModel(conn),
		orderAddressesModel: order.NewOrderAddressesModel(conn),
	}
}

// GetByID 根据订单ID查询
func (r *OrderRepositoryImpl) GetByID(ctx context.Context, orderID string) (*entity.Order, error) {
	// 查询订单主表
	o, err := r.ordersModel.FindOne(ctx, orderID)
	if err != nil {
		if err == order.ErrNotFound {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	order := r.toDomainOrder(o)

	// 查询订单项
	items, err := r.orderItemsModel.QueryOrderItemsByOrderID(ctx, orderID)
	if err == nil && len(items) > 0 {
		order.Items = r.toDomainItems(items)
	}

	// 查询地址快照
	addr, err := r.orderAddressesModel.GetOrderAddressByOrderID(ctx, orderID)
	if err == nil && addr != nil {
		order.Address = r.toDomainAddress(addr)
	}

	return order, nil
}

// GetByPreOrderID 根据预订单ID查询
func (r *OrderRepositoryImpl) GetByPreOrderID(ctx context.Context, preOrderID string) (*entity.Order, error) {
	o, err := r.ordersModel.FindOneByPreOrderId(ctx, preOrderID)
	if err != nil {
		if err == order.ErrNotFound {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	order := r.toDomainOrder(o)

	// 查询订单项
	items, err := r.orderItemsModel.QueryOrderItemsByOrderID(ctx, o.OrderId)
	if err == nil && len(items) > 0 {
		order.Items = r.toDomainItems(items)
	}

	// 查询地址快照
	addr, err := r.orderAddressesModel.GetOrderAddressByOrderID(ctx, o.OrderId)
	if err == nil && addr != nil {
		order.Address = r.toDomainAddress(addr)
	}

	return order, nil
}

// GetByUserID 根据用户ID查询
func (r *OrderRepositoryImpl) GetByUserID(ctx context.Context, userID int64) ([]*entity.Order, error) {
	// 使用分页查询获取用户的订单
	orders, err := r.ordersModel.GetOrdersByUserID(ctx, int32(userID), 1, 1000)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.Order, 0, len(orders))
	for _, o := range orders {
		result = append(result, r.toDomainOrder(o))
	}
	return result, nil
}

// Save 保存订单
func (r *OrderRepositoryImpl) Save(ctx context.Context, order *entity.Order) error {
	// 使用 TransactCtx 处理事务
	return r.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 重新创建带 session 的 model
		ordersModel := r.ordersModel.WithSession(session)
		orderItemsModel := r.orderItemsModel.WithSession(session)
		orderAddressesModel := r.orderAddressesModel.WithSession(session)

		// 插入订单主表
		orderDO := r.toDOOrder(order)
		_, err := ordersModel.Insert(ctx, orderDO)
		if err != nil {
			return err
		}

		// 插入订单项
		if len(order.Items) > 0 {
			itemsDO := r.toDOItems(order.OrderID, order.Items)
			err = orderItemsModel.BulkInsert(session, itemsDO)
			if err != nil {
				return err
			}
		}

		// 插入地址快照
		if order.Address != nil {
			addressDO := r.toDOAddress(order.OrderID, order.Address)
			_, err = orderAddressesModel.Insert(ctx, addressDO)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// Update 更新订单
func (r *OrderRepositoryImpl) Update(ctx context.Context, order *entity.Order) error {
	orderDO := r.toDOOrder(order)
	return r.ordersModel.Update(ctx, orderDO)
}

// Delete 删除订单
func (r *OrderRepositoryImpl) Delete(ctx context.Context, orderID string) error {
	return r.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		ordersModel := r.ordersModel.WithSession(session)
		orderItemsModel := r.orderItemsModel.WithSession(session)
		orderAddressesModel := r.orderAddressesModel.WithSession(session)

		err := ordersModel.DeleteOrderByOrderID(ctx, session, orderID)
		if err != nil {
			return err
		}

		// 删除订单项
		err = orderItemsModel.DeleteOrderItemByOrderID(ctx, session, orderID)
		if err != nil {
			return err
		}

		// 删除地址快照
		err = orderAddressesModel.DeleteOrderAddressByOrderID(ctx, session, orderID)
		if err != nil {
			return err
		}

		return nil
	})
}

// ListByUserID 查询用户的订单列表
func (r *OrderRepositoryImpl) ListByUserID(ctx context.Context, userID int64, status *entity.OrderStatus, page, pageSize int) ([]*entity.Order, int64, error) {
	// 简化实现：直接查询所有订单
	orders, err := r.ordersModel.GetOrdersByUserID(ctx, int32(userID), int32(page), int32(pageSize))
	if err != nil {
		return nil, 0, err
	}

	result := make([]*entity.Order, 0, len(orders))
	for _, o := range orders {
		order := r.toDomainOrder(o)
		// 如果有状态过滤
		if status != nil && order.OrderStatus != *status {
			continue
		}
		// 查询订单项
		items, err := r.orderItemsModel.QueryOrderItemsByOrderID(ctx, o.OrderId)
		if err == nil && len(items) > 0 {
			order.Items = r.toDomainItems(items)
		}
		// 查询地址快照
		address, err := r.orderAddressesModel.GetOrderAddressByOrderID(ctx, o.OrderId)
		if err == nil && address != nil {
			order.Address = r.toDomainAddress(address)
		}
		result = append(result, order)
	}

	return result, int64(len(result)), nil
}

// ListByStatus 根据状态查询订单列表
func (r *OrderRepositoryImpl) ListByStatus(ctx context.Context, status entity.OrderStatus, page, pageSize int) ([]*entity.Order, int64, error) {
	// 简化实现
	return nil, 0, errors.New("not implemented")
}

// FindExpired 查找已过期的订单
func (r *OrderRepositoryImpl) FindExpired(ctx context.Context, limit int) ([]*entity.Order, error) {
	// 简化实现
	return nil, errors.New("not implemented")
}

// toDomainOrder 转换数据库模型到领域模型
func (r *OrderRepositoryImpl) toDomainOrder(o *order.Orders) *entity.Order {
	var paidAt *time.Time
	if o.PaidAt.Valid {
		t := time.Unix(o.PaidAt.Int64, 0)
		paidAt = &t
	}

	return &entity.Order{
		OrderID:        o.OrderId,
		PreOrderID:     o.PreOrderId,
		UserID:         int64(o.UserId),
		CouponID:       o.CouponId,
		PaymentMethod:  int(o.PaymentMethod.Int64),
		TransactionID:  o.TransactionId.String,
		PaidAt:         paidAt,
		OriginalAmount: o.OriginalAmount,
		DiscountAmount: o.DiscountAmount,
		PayableAmount:  o.PayableAmount,
		PaidAmount:     o.PaidAmount.Int64,
		OrderStatus:    entity.OrderStatus(o.OrderStatus),
		PaymentStatus:  entity.PaymentStatus(o.PaymentStatus),
		Reason:         o.Reason.String,
		ExpireTime:     time.Unix(o.ExpireTime, 0),
		CreatedAt:      o.CreatedAt,
		UpdatedAt:      o.UpdatedAt,
	}
}

// toDomainItems 转换订单项
func (r *OrderRepositoryImpl) toDomainItems(items []*order.OrderItems) []*entity.OrderItem {
	result := make([]*entity.OrderItem, 0, len(items))
	for _, item := range items {
		result = append(result, &entity.OrderItem{
			OrderID:     item.OrderId,
			ProductID:   int64(item.ProductId),
			Quantity:    int(item.Quantity),
			Price:       item.Price,
			ProductName: item.ProductName,
			ProductDesc: item.ProductDesc,
			CreatedAt:   item.CreatedAt,
		})
	}
	return result
}

// toDomainAddress 转换地址快照
func (r *OrderRepositoryImpl) toDomainAddress(addr *order.OrderAddresses) *entity.OrderAddress {
	return &entity.OrderAddress{
		OrderID:         addr.OrderId,
		AddressID:       int64(addr.AddressId),
		RecipientName:   addr.RecipientName,
		PhoneNumber:     addr.PhoneNumber.String,
		Province:        addr.Province.String,
		City:            addr.City,
		DetailedAddress: addr.DetailedAddress,
		CreatedAt:       addr.CreatedAt,
		UpdatedAt:       addr.UpdatedAt,
	}
}

// toDOOrder 转换领域模型到数据库模型
func (r *OrderRepositoryImpl) toDOOrder(o *entity.Order) *order.Orders {
	var paymentMethod sql.NullInt64
	if o.PaymentMethod > 0 {
		paymentMethod = sql.NullInt64{Int64: int64(o.PaymentMethod), Valid: true}
	}

	var transactionID sql.NullString
	if o.TransactionID != "" {
		transactionID = sql.NullString{String: o.TransactionID, Valid: true}
	}

	var paidAt sql.NullInt64
	if o.PaidAt != nil {
		paidAt = sql.NullInt64{Int64: o.PaidAt.Unix(), Valid: true}
	}

	var paidAmount sql.NullInt64
	if o.PaidAmount > 0 {
		paidAmount = sql.NullInt64{Int64: o.PaidAmount, Valid: true}
	}

	var reason sql.NullString
	if o.Reason != "" {
		reason = sql.NullString{String: o.Reason, Valid: true}
	}

	return &order.Orders{
		OrderId:        o.OrderID,
		PreOrderId:     o.PreOrderID,
		UserId:         uint64(o.UserID),
		CouponId:       o.CouponID,
		PaymentMethod:  paymentMethod,
		TransactionId:  transactionID,
		PaidAt:         paidAt,
		OriginalAmount: o.OriginalAmount,
		DiscountAmount: o.DiscountAmount,
		PayableAmount:  o.PayableAmount,
		PaidAmount:     paidAmount,
		OrderStatus:    int64(o.OrderStatus),
		PaymentStatus:  int64(o.PaymentStatus),
		Reason:         reason,
		ExpireTime:     o.ExpireTime.Unix(),
		CreatedAt:      o.CreatedAt,
		UpdatedAt:      o.UpdatedAt,
	}
}

// toDOItems 转换订单项到数据库模型
func (r *OrderRepositoryImpl) toDOItems(orderID string, items []*entity.OrderItem) []*order.OrderItems {
	result := make([]*order.OrderItems, 0, len(items))
	for _, item := range items {
		result = append(result, &order.OrderItems{
			OrderId:     orderID,
			ProductId:   uint64(item.ProductID),
			Quantity:    uint64(item.Quantity),
			Price:       item.Price,
			ProductName: item.ProductName,
			ProductDesc: item.ProductDesc,
			CreatedAt:   item.CreatedAt,
		})
	}
	return result
}

// toDOAddress 转换地址到数据库模型
func (r *OrderRepositoryImpl) toDOAddress(orderID string, addr *entity.OrderAddress) *order.OrderAddresses {
	var phoneNumber sql.NullString
	if addr.PhoneNumber != "" {
		phoneNumber = sql.NullString{String: addr.PhoneNumber, Valid: true}
	}

	var province sql.NullString
	if addr.Province != "" {
		province = sql.NullString{String: addr.Province, Valid: true}
	}

	return &order.OrderAddresses{
		OrderId:         orderID,
		AddressId:       uint64(addr.AddressID),
		RecipientName:   addr.RecipientName,
		PhoneNumber:     phoneNumber,
		Province:        province,
		City:            addr.City,
		DetailedAddress: addr.DetailedAddress,
		CreatedAt:       addr.CreatedAt,
		UpdatedAt:       addr.UpdatedAt,
	}
}
