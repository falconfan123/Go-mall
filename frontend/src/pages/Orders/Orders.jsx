import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { orderApi } from '../../services/api';
import { useAuthStore } from '../../store/authStore';
import { Button } from '../../components/common/Button';
import { Spinner } from '../../components/common/Spinner';
import { toast } from '../../components/common/Toast';

export default function Orders() {
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(false);
  const [cancellingId, setCancellingId] = useState(null);

  useEffect(() => {
    if (!user) {
      navigate('/login');
    } else {
      fetchOrders();
    }
  }, [user, navigate]);

  const fetchOrders = async () => {
    setLoading(true);
    try {
      const response = await orderApi.list();
      setOrders(response.data || []);
    } catch (error) {
      console.error('Fetch orders error:', error);
      toast.error('获取订单失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = async (orderId) => {
    setCancellingId(orderId);
    try {
      await orderApi.cancel(orderId);
      toast.success('取消成功');
      fetchOrders();
    } catch (error) {
      console.error('Cancel order error:', error);
      toast.error('取消失败');
    } finally {
      setCancellingId(null);
    }
  };

  const getStatusText = (status) => {
    const statusMap = {
      pending: '待支付',
      paid: '已支付',
      shipped: '已发货',
      completed: '已完成',
      cancelled: '已取消',
    };
    return statusMap[status] || status;
  };

  const getStatusClass = (status) => {
    const classMap = {
      pending: 'bg-yellow-100 text-yellow-800',
      paid: 'bg-blue-100 text-blue-800',
      shipped: 'bg-purple-100 text-purple-800',
      completed: 'bg-green-100 text-green-800',
      cancelled: 'bg-gray-100 text-gray-800',
    };
    return classMap[status] || 'bg-gray-100';
  };

  if (!user) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500 mb-4">请先登录</p>
        <Button onClick={() => navigate('/login')}>去登录</Button>
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
        <Button onClick={() => navigate('/products')}>去购物</Button>
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
                <span className="font-medium">{order.order_id || order.id}</span>
              </div>
              <span className={`px-3 py-1 rounded-full text-sm ${getStatusClass(order.status)}`}>
                {getStatusText(order.status)}
              </span>
            </div>

            <div className="border-t border-b py-4 mb-4">
              {order.items?.map((item, index) => (
                <div key={index} className="flex justify-between mb-2">
                  <span className="text-gray-600">{item.name} x {item.quantity}</span>
                  <span className="font-medium">¥{(item.price * item.quantity).toFixed(2)}</span>
                </div>
              ))}
            </div>

            <div className="flex items-center justify-between">
              <div>
                <span className="text-gray-600">总计: </span>
                <span className="text-xl font-bold text-blue-600">¥{order.total_price || order.total || 0}</span>
              </div>
              <div className="flex gap-2">
                {order.status === 'pending' && (
                  <Button
                    variant="outline"
                    size="small"
                    loading={cancellingId === (order.id || order.order_id)}
                    onClick={() => handleCancel(order.id || order.order_id)}
                  >
                    取消订单
                  </Button>
                )}
                {order.status === 'paid' && (
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