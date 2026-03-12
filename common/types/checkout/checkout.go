package checkout

// CheckoutItem 预结算商品项
type CheckoutItem struct {
	Id          int64  // 主键ID
	PreOrderId  string // 预订单ID
	ProductId   uint64 // 商品ID
	ProductName string // 商品名称
	ProductDesc string // 商品描述
	Price       int64  // 单价（分）
	Quantity    uint64 // 数量
	TotalPrice  int64  // 总价（分）
}
