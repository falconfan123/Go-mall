import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  getStatusCode,
  getStatusMsg,
  normalizeOrder,
  orderApi,
  PAYMENT_EXIST_STATUS_CODE,
  paymentApi,
  normalizePayment,
} from "../../services/api";
import { useAuthStore } from "../../store/authStore";
import { Button } from "../../components/common/Button";
import { Spinner } from "../../components/common/Spinner";
import { toast } from "../../components/common/Toast";

export default function Orders() {
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(false);
  const [cancellingId, setCancellingId] = useState(null);
  const [payingId, setPayingId] = useState(null);

  useEffect(() => {
    if (!user) {
      navigate("/login");
    } else {
      fetchOrders();
    }
  }, [user, navigate]);

  const fetchOrders = async () => {
    setLoading(true);
    try {
      const response = await orderApi.list({ user_id: user.user_id });
      const data = response.data || {};
      if (getStatusCode(data) !== 0) {
        throw new Error(getStatusMsg(data, "获取订单失败"));
      }
      const orderList = Array.isArray(data.orders)
        ? data.orders.map(normalizeOrder).filter(Boolean)
        : [];
      setOrders(orderList);
    } catch (error) {
      console.error("Fetch orders error:", error);
      toast.error("获取订单失败");
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = async (orderId) => {
    setCancellingId(orderId);
    try {
      const response = await orderApi.cancel({
        order_id: orderId,
        user_id: user.user_id,
        cancel_reason: "user_cancelled",
        initiative: true,
      });
      const data = response.data || {};
      if (getStatusCode(data) !== 0) {
        throw new Error(getStatusMsg(data, "取消失败"));
      }
      toast.success("取消成功");
      fetchOrders();
    } catch (error) {
      console.error("Cancel order error:", error);
      toast.error("取消失败");
    } finally {
      setCancellingId(null);
    }
  };

  const handlePay = async (orderId) => {
    setPayingId(orderId);
    try {
      const response = await paymentApi.create({
        user_id: user.user_id,
        order_id: orderId,
        payment_method: 3,
      });
      const data = response.data || {};
      const statusCode = getStatusCode(data);
      if (statusCode !== 0 && statusCode !== PAYMENT_EXIST_STATUS_CODE) {
        throw new Error(getStatusMsg(data, "创建支付失败"));
      }

      const payment = normalizePayment(data.payment ?? data);
      if (!payment?.pay_url) {
        throw new Error("Stripe 支付链接缺失");
      }

      toast.info("正在跳转到 Stripe 支付页面...");
      window.location.href = payment.pay_url;
    } catch (error) {
      console.error("Create Stripe payment error:", error);
      toast.error(error.message || "创建支付失败");
    } finally {
      setPayingId(null);
    }
  };

  const getStatusText = (status) => {
    const statusMap = {
      1: "已创建",
      2: "待支付",
      3: "已支付",
      4: "已完成",
      5: "已取消",
      6: "已关闭",
      7: "已退款",
    };
    return statusMap[status] || status;
  };

  const getStatusClass = (status) => {
    const classMap = {
      1: "bg-gray-100 text-gray-800",
      2: "bg-yellow-100 text-yellow-800",
      3: "bg-blue-100 text-blue-800",
      4: "bg-green-100 text-green-800",
      5: "bg-gray-100 text-gray-800",
      6: "bg-gray-100 text-gray-800",
      7: "bg-red-100 text-red-800",
    };
    return classMap[status] || "bg-gray-100";
  };

  if (!user) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500 mb-4">请先登录</p>
        <Button onClick={() => navigate("/login")}>去登录</Button>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="flex justify-center py-12">
        <Spinner size="large" />
      </div>
    );
  }

  if (orders.length === 0) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500 mb-4">暂无订单</p>
        <Button onClick={() => navigate("/products")}>去购物</Button>
      </div>
    );
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 mb-6">我的订单</h2>

      <div className="space-y-4">
        {orders.map((order) => (
          <div
            key={order.id || order.order_id}
            className="bg-white rounded-xl shadow-sm p-6"
          >
            <div className="flex items-center justify-between mb-4">
              <div>
                <span className="text-sm text-gray-500">订单号: </span>
                <span className="font-medium">{order.order_id}</span>
              </div>
              <span
                className={`px-3 py-1 rounded-full text-sm ${getStatusClass(order.order_status)}`}
              >
                {getStatusText(order.order_status)}
              </span>
            </div>

            <div className="border-t border-b py-4 mb-4">
              {order.items?.map((item, index) => (
                <div key={index} className="flex justify-between mb-2">
                  <span className="text-gray-600">
                    {item.name} x {item.quantity}
                  </span>
                  <span className="font-medium">
                    ¥{(item.price * item.quantity).toFixed(2)}
                  </span>
                </div>
              ))}
            </div>

            <div className="flex items-center justify-between">
              <div>
                <span className="text-gray-600">总计: </span>
                <span className="text-xl font-bold text-blue-600">
                  ¥{order.total_price.toFixed(2)}
                </span>
              </div>
              <div className="flex gap-2">
                {order.order_status === 1 || order.order_status === 2 ? (
                  <>
                    <Button
                      size="small"
                      loading={payingId === order.order_id}
                      onClick={() => handlePay(order.order_id)}
                    >
                      去支付
                    </Button>
                    <Button
                      variant="outline"
                      size="small"
                      loading={cancellingId === order.order_id}
                      onClick={() => handleCancel(order.order_id)}
                    >
                      取消订单
                    </Button>
                  </>
                ) : null}
                {order.order_status === 3 && (
                  <Button size="small">查看物流</Button>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
