package delay

import (
	"context"
	"errors"
	"testing"

	"github.com/falconfan123/Go-mall/common/consts/code"
	ordertypes "github.com/falconfan123/Go-mall/common/types/order"
	order2 "github.com/falconfan123/Go-mall/dal/model/order"
	"github.com/falconfan123/Go-mall/services/inventory/inventoryclient"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc"
)

// ---------- fakes ----------

type fakeTxConn struct {
	sqlx.SqlConn
}

func (f *fakeTxConn) TransactCtx(ctx context.Context, fn func(ctx context.Context, session sqlx.Session) error) error {
	return fn(ctx, nil)
}

type fakeOrdersModel struct {
	order2.OrdersModel
	order     *order2.Orders
	orderEr   error
	statusSet int64
	updated   int
}

func (f *fakeOrdersModel) WithSession(sqlx.Session) order2.OrdersModel { return f }

func (f *fakeOrdersModel) GetOrderByOrderIDAndUserIDWithLock(context.Context, string, int32) (*order2.Orders, error) {
	return f.order, f.orderEr
}

func (f *fakeOrdersModel) UpdateOrderStatusByOrderIDAndUserID(_ context.Context, _ string, _ int32,
	orderStatus ordertypes.OrderStatus, _ ordertypes.PaymentStatus) error {
	f.updated++
	f.statusSet = int64(orderStatus)
	return nil
}

type fakeOrderItemsModel struct {
	order2.OrderItemsModel
	itemEr error
}

func (f *fakeOrderItemsModel) QueryOrderItemsByOrderID(context.Context, string) ([]*order2.OrderItems, error) {
	if f.itemEr != nil {
		return nil, f.itemEr
	}
	return []*order2.OrderItems{{ProductId: 3, Quantity: 1}}, nil
}

type fakeInventory struct {
	inventoryclient.Inventory
	called    int
	transport bool
}

func (f *fakeInventory) ReturnPreInventory(context.Context, *inventoryclient.InventoryReq, ...grpc.CallOption) (*inventoryclient.InventoryResp, error) {
	f.called++
	if f.transport {
		return nil, errInventoryDown
	}
	return &inventoryclient.InventoryResp{StatusCode: code.Success, StatusMsg: code.SuccessMsg}, nil
}

var errInventoryDown = errors.New("inventory rpc down")

func newCreatedOrder() *order2.Orders {
	return &order2.Orders{
		OrderId:     "o-200",
		UserId:      9,
		PreOrderId:  "pre-2",
		OrderStatus: int64(ordertypes.OrderStatusCreated),
	}
}

func newDelayHandlerFor(order *order2.Orders, orderEr error, inv *fakeInventory) (*delayHandler, *fakeOrdersModel) {
	om := &fakeOrdersModel{order: order, orderEr: orderEr}
	h := &delayHandler{
		Model:           &fakeTxConn{},
		OrderModel:      om,
		OrderItemsModel: &fakeOrderItemsModel{},
		InventoryRpc:    inv,
	}
	return h, om
}

// ---------- 3.1 业务语义 ----------

func initLog(t *testing.T) {
	t.Helper()
	logx.SetUp(logx.LogConf{Stat: false})
}

// 创建态订单：关单 + 释放预扣库存
func TestHandleClosesCreatedOrderAndReturnsInventory(t *testing.T) {
	initLog(t)
	inv := &fakeInventory{}
	h, om := newDelayHandlerFor(newCreatedOrder(), nil, inv)

	if err := h.Handle(context.Background(), &OrderReq{OrderId: "o-200", UserID: 9}); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if om.updated != 1 || om.statusSet != int64(ordertypes.OrderStatusClosed) || inv.called != 1 {
		t.Fatalf("want close+return, got updated=%d status=%d inv=%d", om.updated, om.statusSet, inv.called)
	}
}

// 非创建态（已支付）：跳过关单与释放库存
func TestHandleSkipsNonCreatedOrder(t *testing.T) {
	initLog(t)
	o := newCreatedOrder()
	o.OrderStatus = int64(ordertypes.OrderStatusPaid)
	inv := &fakeInventory{}
	h, om := newDelayHandlerFor(o, nil, inv)

	if err := h.Handle(context.Background(), &OrderReq{OrderId: "o-200", UserID: 9}); err != nil {
		t.Fatalf("paid order should be skipped without error, got %v", err)
	}
	if om.updated != 0 || inv.called != 0 {
		t.Fatalf("non-created order must not trigger side effects, got updated=%d inv=%d", om.updated, inv.called)
	}
}

// 3.1 fall-through 回归：查单失败 → 返回 err，不释放库存，不 panic
func TestHandleGetOrderErrorStopsInventory(t *testing.T) {
	initLog(t)
	inv := &fakeInventory{}
	h, _ := newDelayHandlerFor(nil, sqlx.ErrNotFound, inv)

	if err := h.Handle(context.Background(), &OrderReq{OrderId: "o-404", UserID: 9}); err == nil {
		t.Fatal("expected error")
	}
	if inv.called != 0 {
		t.Fatalf("inventory must not run after order query failure")
	}
}

// 释放预扣库存失败 → 返回 err（交由骨架有界重试）
func TestHandleReturnInventoryError(t *testing.T) {
	initLog(t)
	inv := &fakeInventory{transport: true}
	h, _ := newDelayHandlerFor(newCreatedOrder(), nil, inv)

	if err := h.Handle(context.Background(), &OrderReq{OrderId: "o-200", UserID: 9}); err == nil {
		t.Fatal("expected error")
	}
}
