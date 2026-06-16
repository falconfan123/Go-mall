import { useNavigate, useSearchParams } from "react-router-dom";
import { Button } from "../../components/common/Button";

export default function PaymentCancel() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const orderId = searchParams.get("order_id") || "";

  return (
    <div className="max-w-2xl mx-auto">
      <div className="bg-white rounded-xl shadow-sm p-8 text-center">
        <h2 className="text-2xl font-bold text-gray-900 mb-4">支付已取消</h2>
        <p className="text-gray-600">
          Stripe 支付已取消，订单仍保持待支付状态。
        </p>
        {orderId ? (
          <p className="mt-4 text-sm text-gray-500 break-all">
            订单号: {orderId}
          </p>
        ) : null}
        <div className="mt-6 flex justify-center gap-3">
          <Button variant="secondary" onClick={() => navigate("/products")}>
            继续逛逛
          </Button>
          <Button onClick={() => navigate("/orders")}>返回订单页</Button>
        </div>
      </div>
    </div>
  );
}
