import { Link, useNavigate } from 'react-router-dom';
import { useAuthStore } from '../../store/authStore';
import { useCartStore } from '../../store/cartStore';

export function Navbar() {
  const { user, logout } = useAuthStore();
  const { getItemCount } = useCartStore();
  const navigate = useNavigate();

  const cartCount = getItemCount();

  const handleLogout = () => {
    logout();
    navigate('/');
  };

  return (
    <nav className="bg-white shadow-sm border-b">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          {/* Logo */}
          <Link to="/" className="flex items-center">
            <span className="text-xl font-bold text-blue-600">Go-Mall</span>
          </Link>

          {/* 导航链接 */}
          <div className="flex items-center space-x-4">
            <Link
              to="/"
              className="text-gray-600 hover:text-blue-600 px-3 py-2 rounded-md text-sm font-medium"
            >
              首页
            </Link>
            <Link
              to="/products"
              className="text-gray-600 hover:text-blue-600 px-3 py-2 rounded-md text-sm font-medium"
            >
              商品
            </Link>
            <Link
              to="/flash"
              className="text-gray-600 hover:text-blue-600 px-3 py-2 rounded-md text-sm font-medium"
            >
              秒杀活动
            </Link>
            <Link
              to="/cart"
              className="text-gray-600 hover:text-blue-600 px-3 py-2 rounded-md text-sm font-medium relative"
            >
              购物车
              {cartCount > 0 && (
                <span className="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center">
                  {cartCount}
                </span>
              )}
            </Link>
            {user && (
              <Link
                to="/orders"
                className="text-gray-600 hover:text-blue-600 px-3 py-2 rounded-md text-sm font-medium"
              >
                订单
              </Link>
            )}
          </div>

          {/* 认证状态 */}
          <div className="flex items-center space-x-4">
            {user ? (
              <div className="flex items-center space-x-4">
                <span className="text-sm text-gray-700">{user.username || user.user_id}</span>
                <button
                  onClick={handleLogout}
                  className="text-gray-600 hover:text-gray-900 text-sm"
                >
                  退出
                </button>
              </div>
            ) : (
              <div className="flex items-center space-x-2">
                <Link
                  to="/login"
                  className="text-gray-600 hover:text-blue-600 px-3 py-2 text-sm"
                >
                  登录
                </Link>
                <Link
                  to="/register"
                  className="bg-blue-600 text-white px-4 py-2 rounded-md text-sm hover:bg-blue-700"
                >
                  注册
                </Link>
              </div>
            )}
          </div>
        </div>
      </div>
    </nav>
  );
}