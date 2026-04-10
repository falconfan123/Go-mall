import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useCartStore } from '../../store/cartStore';
import { orderApi, paymentApi } from '../../services/api';
import { useAuthStore } from '../../store/authStore';
import { Button } from '../../components/common/Button';
import { Spinner } from '../../components/common/Spinner';
import { toast } from '../../components/common/Toast';

export default function Payment() {
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const { items, getTotalPrice, clearCart, fetchCart } = useCartStore();
  const [loading, setLoading] = useState(false);
  const [paymentMethod, setPaymentMethod] = useState('alipay');
  const [orderId, setOrderId] = useState(null);

  useEffect(() => {
    if (!user) {
      navigate('/login');
    }
  }, [user, navigate]);

  useEffect(() => {
    fetchCart();
  }, [fetchCart]);

  const handlePayment = async () => {
    if (items.length === 0) {
      toast.error('购物车为空');
      return;
    }

    setLoading(true);

    try {
      // 创建订单
      const orderResponse = await orderApi.create({
        items: items.map((item) => ({
          product_id: item.product_id || item.id,
          quantity: item.quantity,
        })),
      });

      const newOrderId = orderResponse.data.order_id || orderResponse.data.id;
      setOrderId(newOrderId);

      // 创建支付
      const paymentResponse = await paymentApi.create({
        order_id: newOrderId,
        amount: getTotalPrice(),
        payment_method: paymentMethod,
      });

      if (paymentResponse.data.status === 'success' || paymentResponse.data.status_code === 0) {
        toast.success('支付成功');
        await clearCart();
        navigate('/orders');
      } else {
        toast.error(paymentResponse.data.message || '支付失败');
      }
    } catch (error) {
      console.error('Payment error:', error);
      toast.error('支付失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  if (!user) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500 mb-4">请先登录</p>
        <Button onClick={() => navigate('/login')}>去登录</Button>
      </div>
    );
  }

  if (items.length === 0) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500 mb-4">购物车是空的</p>
        <Button onClick={() => navigate('/products')}>去购物</Button>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto">
      <h2 className="text-2xl font-bold text-gray-900 mb-6">订单支付</h2>

      <div className="bg-white rounded-xl shadow-sm p-6 mb-6">
        <h3 className="text-lg font-semibold mb-4">订单信息</h3>
        <div className="space-y-3 mb-6">
          {items.map((item) => (
            <div key={item.id} className="flex justify-between">
              <span className="text-gray-600">{item.name} x {item.quantity}</span>
              <span className="font-medium">¥{(item.price * item.quantity).toFixed(2)}</span>
            </div>
          ))}
        </div>
        <div className="border-t pt-4 flex justify-between">
          <span className="text-lg font-semibold">总计</span>
          <span className="text-2xl font-bold text-blue-600">¥{getTotalPrice().toFixed(2)}</span>
        </div>
      </div>

      <div className="bg-white rounded-xl shadow-sm p-6 mb-6">
        <h3 className="text-lg font-semibold mb-4">支付方式</h3>
        <div className="space-y-3">
          <label className="flex items-center p-3 border rounded-lg cursor-pointer hover:bg-gray-50">
            <input
              type="radio"
              name="paymentMethod"
              value="alipay"
              checked={paymentMethod === 'alipay'}
              onChange={(e) => setPaymentMethod(e.target.value)}
              className="mr-3"
            />
            <span>支付宝</span>
          </label>
          <label className="flex items-center p-3 border rounded-lg cursor-pointer hover:bg-gray-50">
            <input
              type="radio"
              name="paymentMethod"
              value="wechat"
              checked={paymentMethod === 'wechat'}
              onChange={(e) => setPaymentMethod(e.target.value)}
              className="mr-3"
            />
            <span>微信支付</span>
          </label>
        </div>
      </div>

      <Button
        className="w-full"
        size="large"
        loading={loading}
        onClick={handlePayment}
      >
        确认支付
      </Button>
    </div>
  );
}