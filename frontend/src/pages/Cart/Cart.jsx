import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useCartStore } from '../../store/cartStore';
import { useAuthStore } from '../../store/authStore';
import { Button } from '../../components/common/Button';
import { Spinner } from '../../components/common/Spinner';
import { toast } from '../../components/common/Toast';

// Emoji 占位符生成器
const EMOJIS = ['📱', '💻', '⌚', '🎧', '📷', '🎮', '📺', '🏠', '👕', '👟', '👜', '💄', '🧴', '☕', '🍳', '🧹', '💡', '🔌', '🔋', '⌨️'];
const COLORS = ['f3f4f6', 'fee2e2', 'fef3c7', 'dcfce7', 'dbeafe', 'e0e7ff', 'fce7f3', 'f3e8ff'];

const getPlaceholder = (id) => {
  const index = id % EMOJIS.length;
  const colorIndex = id % COLORS.length;
  const emoji = EMOJIS[index];
  const bgColor = COLORS[colorIndex];
  return `data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='80' height='80' viewBox='0 0 80 80'%3E%3Crect fill='%23${bgColor}' width='80' height='80'/%3E%3Ctext fill='%236b7280' font-family='sans-serif' font-size='24' x='50%25' y='50%25' text-anchor='middle' dy='.3em'%3E${encodeURIComponent(emoji)}%3C/text%3E%3C/svg%3E`;
};

export default function Cart() {
  const { user } = useAuthStore();
  const navigate = useNavigate();
  const { items, loading, fetchCart, updateQuantity, removeItem, clearCart, getTotalPrice } = useCartStore();
  const [removingId, setRemovingId] = useState(null);

  useEffect(() => {
    if (user) {
      fetchCart();
    }
  }, [user, fetchCart]);

  const handleQuantityChange = async (itemId, newQuantity) => {
    if (newQuantity < 1) return;
    await updateQuantity(itemId, newQuantity);
  };

  const handleRemove = async (itemId) => {
    setRemovingId(itemId);
    const success = await removeItem(itemId);
    setRemovingId(null);
    if (success) {
      toast.success('移除成功');
    }
  };

  const handleCheckout = () => {
    navigate('/payment');
  };

  if (!user) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500 mb-4">请先登录后查看购物车</p>
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

  if (items.length === 0) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500 mb-4">购物车是空的</p>
        <Button onClick={() => navigate('/products')}>去购物</Button>
      </div>
    );
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 mb-6">购物车</h2>

      <div className="bg-white rounded-xl shadow-sm overflow-hidden">
        {/* 购物车项 */}
        <div className="divide-y divide-gray-200">
          {items.map((item) => (
            <div key={item.product_id || item.id} className="p-4 flex items-center gap-4">
              <img
                src={item.image || item.image_url || getPlaceholder(item.product_id || item.id)}
                alt={item.name}
                className="w-20 h-20 object-cover rounded-lg"
              />
              <div className="flex-1">
                <h3 className="font-medium text-gray-900">{item.name}</h3>
                <p className="text-sm text-gray-500">¥{item.price}</p>
              </div>
              <div className="flex items-center gap-2">
                <button
                  className="px-2 py-1 border rounded hover:bg-gray-50"
                  onClick={() => handleQuantityChange(item.product_id || item.id, item.quantity - 1)}
                >
                  -
                </button>
                <span className="w-8 text-center">{item.quantity}</span>
                <button
                  className="px-2 py-1 border rounded hover:bg-gray-50"
                  onClick={() => handleQuantityChange(item.product_id || item.id, item.quantity + 1)}
                >
                  +
                </button>
              </div>
              <div className="text-right">
                <p className="font-medium text-gray-900">
                  ¥{(item.price * item.quantity).toFixed(2)}
                </p>
                <button
                  className="text-sm text-red-600 hover:text-red-700"
                  onClick={() => handleRemove(item.product_id || item.id)}
                  disabled={removingId === (item.product_id || item.id)}
                >
                  {removingId === (item.product_id || item.id) ? '移除中...' : '移除'}
                </button>
              </div>
            </div>
          ))}
        </div>

        {/* 底部结算 */}
        <div className="p-4 border-t border-gray-200">
          <div className="flex items-center justify-between">
            <div>
              <span className="text-gray-600">总计: </span>
              <span className="text-2xl font-bold text-blue-600">¥{getTotalPrice().toFixed(2)}</span>
            </div>
            <div className="flex gap-4">
              <Button variant="secondary" onClick={clearCart}>
                清空购物车
              </Button>
              <Button onClick={handleCheckout}>
                结算
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
