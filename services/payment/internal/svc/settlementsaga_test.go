package svc

import (
	"context"
	"encoding/json"
	"testing"

	inventorypb "github.com/falconfan123/Go-mall/services/inventory/pb"
	"github.com/falconfan123/Go-mall/services/payment/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
)

// gid 拼接断言：空后缀 / :r2 / :r3（design 决策 5 前置声明）。
func TestGidSuffixConcatenation(t *testing.T) {
	s := NewSettlementSaga(config.DtmConfig{Server: "localhost:36790"}, logx.WithContext(context.Background()))
	if got := s.gid("o-1", ""); got != "settle:o-1" {
		t.Fatalf("gid empty suffix = %q", got)
	}
	if got := s.gid("o-1", ":r2"); got != "settle:o-1:r2" {
		t.Fatalf("gid r2 = %q", got)
	}
	if got := s.gid("o-1", ":r3"); got != "settle:o-1:r3" {
		t.Fatalf("gid r3 = %q", got)
	}
}

// outbox payload 序列化 roundtrip：encoding/json 写读两端一致（design 决策 3）。
func TestSettlementInputJSONRoundtrip(t *testing.T) {
	in := &SettlementInput{
		OrderID: "o-1", UserID: 7, PreOrderID: "pre-1",
		TransactionID: "tx-1", PaidAmount: 9900, PaidAt: 123,
		Items:    []*inventorypb.InventoryReq_Items{{ProductId: 42, Quantity: 2}},
		CouponID: "c-1", DiscountAmount: 900, OriginAmount: 10000, GidSuffix: ":r2",
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out SettlementInput
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.OrderID != in.OrderID || out.GidSuffix != ":r2" || out.PaidAmount != 9900 ||
		len(out.Items) != 1 || out.Items[0].ProductId != 42 || out.CouponID != "c-1" {
		t.Fatalf("roundtrip mismatch: %+v", out)
	}
}
