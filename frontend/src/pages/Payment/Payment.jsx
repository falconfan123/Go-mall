import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useCartStore } from "../../store/cartStore";
import { useAuthStore } from "../../store/authStore";
import {
  checkoutApi,
  getStatusCode,
  getStatusMsg,
  normalizeOrder,
  normalizePayment,
  orderApi,
  PAYMENT_EXIST_STATUS_CODE,
  paymentApi,
  userApi,
} from "../../services/api";
import { Button } from "../../components/common/Button";
import { Input } from "../../components/common/Input";
import { toast } from "../../components/common/Toast";

const emptyAddress = {
  recipient_name: "",
  phone_number: "",
  province: "",
  city: "",
  detailed_address: "",
};

export default function Payment() {
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const { items, getTotalPrice, clearCart, fetchCart } = useCartStore();
  const [loading, setLoading] = useState(false);
  const [address, setAddress] = useState(emptyAddress);

  useEffect(() => {
    if (!user) {
      navigate("/login");
      return;
    }
    fetchCart();
  }, [user, navigate, fetchCart]);

  const handleAddressChange = (field, value) => {
    setAddress((prev) => ({ ...prev, [field]: value }));
  };

  const ensureAddress = async () => {
    const values = Object.values(address).map((value) => value.trim());
    const hasAnyValue = values.some(Boolean);
    if (!hasAnyValue) {
      toast.error("请先填写收货地址");
      return null;
    }
    if (values.some((value) => !value)) {
      toast.error("请完整填写收货地址");
      return null;
    }

    const response = await userApi.addAddress({
      ...address,
      user_id: user.user_id,
      is_default: true,
      ip: "127.0.0.1",
    });
    const data = response.data || {};
    if (getStatusCode(data) !== 0) {
      throw new Error(getStatusMsg(data, "创建地址失败"));
    }
    return data.data?.addressId ?? data.data?.address_id ?? null;
  };

  const handlePayment = async () => {
    if (items.length === 0) {
      toast.error("购物车为空");
      return;
    }

    setLoading(true);

    try {
      const addressId = await ensureAddress();
      if (!addressId) {
        setLoading(false);
        return;
      }

      const checkoutResponse = await checkoutApi.prepare({
        user_id: user.user_id,
        address_id: addressId,
        coupon_id: "",
        order_items: items.map((item) => ({
          product_id: item.product_id,
          quantity: item.quantity,
        })),
      });
      const checkoutData = checkoutResponse.data || {};
      if (getStatusCode(checkoutData) !== 0) {
        throw new Error(getStatusMsg(checkoutData, "预结算失败"));
      }

      const paymentMethodCode = 3;
      const preOrderId = checkoutData.preOrderId ?? checkoutData.pre_order_id;

      const orderResponse = await orderApi.create({
        pre_order_id: preOrderId,
        user_id: user.user_id,
        coupon_id: "",
        address_id: addressId,
        payment_method: paymentMethodCode,
      });
      const orderData = orderResponse.data || {};
      if (getStatusCode(orderData) !== 0) {
        throw new Error(getStatusMsg(orderData, "创建订单失败"));
      }

      const normalizedOrder = normalizeOrder(orderData.order ?? orderData);
      const newOrderId = normalizedOrder?.order_id;
      if (!newOrderId) {
        throw new Error("订单号缺失");
      }

      const paymentResponse = await paymentApi.create({
        user_id: user.user_id,
        order_id: newOrderId,
        payment_method: paymentMethodCode,
      });
      const paymentData = paymentResponse.data || {};
      const paymentStatusCode = getStatusCode(paymentData);
      if (
        paymentStatusCode !== 0 &&
        paymentStatusCode !== PAYMENT_EXIST_STATUS_CODE
      ) {
        throw new Error(getStatusMsg(paymentData, "创建支付失败"));
      }
      const createdPayment = normalizePayment(paymentData.payment ?? paymentData);

      await clearCart();

      if (!createdPayment?.pay_url) {
        throw new Error("Stripe 支付链接缺失");
      }

      toast.info("正在跳转到 Stripe 支付页面...");
      window.location.href = createdPayment.pay_url;
    } catch (error) {
      console.error("Payment error:", error);
      toast.error(error.message || "支付失败，请重试");
    } finally {
      setLoading(false);
    }
  };

  if (!user) {
    return null;
  }

  if (items.length === 0) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500 mb-4">购物车是空的</p>
        <Button onClick={() => navigate("/products")}>去购物</Button>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto">
      <h2 className="text-2xl font-bold text-gray-900 mb-6">订单支付</h2>

      <div className="bg-white rounded-xl shadow-sm p-6 mb-6">
        <h3 className="text-lg font-semibold mb-4">收货地址</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Input
            label="收件人"
            value={address.recipient_name}
            onChange={(e) =>
              handleAddressChange("recipient_name", e.target.value)
            }
          />
          <Input
            label="手机号"
            value={address.phone_number}
            onChange={(e) =>
              handleAddressChange("phone_number", e.target.value)
            }
          />
          <Input
            label="省份"
            value={address.province}
            onChange={(e) => handleAddressChange("province", e.target.value)}
          />
          <Input
            label="城市"
            value={address.city}
            onChange={(e) => handleAddressChange("city", e.target.value)}
          />
          <div className="md:col-span-2">
            <Input
              label="详细地址"
              value={address.detailed_address}
              onChange={(e) =>
                handleAddressChange("detailed_address", e.target.value)
              }
            />
          </div>
        </div>
      </div>

      <div className="bg-white rounded-xl shadow-sm p-6 mb-6">
        <h3 className="text-lg font-semibold mb-4">订单信息</h3>
        <div className="space-y-3 mb-6">
          {items.map((item) => (
            <div
              key={item.product_id || item.id}
              className="flex justify-between"
            >
              <span className="text-gray-600">
                {item.name} x {item.quantity}
              </span>
              <span className="font-medium">
                ¥{(item.price * item.quantity).toFixed(2)}
              </span>
            </div>
          ))}
        </div>
        <div className="border-t pt-4 flex justify-between">
          <span className="text-lg font-semibold">总计</span>
          <span className="text-2xl font-bold text-blue-600">
            ¥{getTotalPrice().toFixed(2)}
          </span>
        </div>
      </div>

      <div className="bg-white rounded-xl shadow-sm p-6 mb-6">
        <h3 className="text-lg font-semibold mb-4">支付方式</h3>
        <div className="rounded-lg border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-700">
          当前开发环境固定使用 Stripe Checkout。
        </div>
      </div>

      <Button
        className="w-full"
        size="large"
        loading={loading}
        onClick={handlePayment}
      >
        创建订单并发起支付
      </Button>
    </div>
  );
}
