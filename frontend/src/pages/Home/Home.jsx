import { Link } from 'react-router-dom';

export default function Home() {
  return (
    <div>
      {/* Hero Section */}
      <div className="bg-gradient-to-r from-blue-600 to-blue-800 rounded-2xl p-12 mb-8 text-white">
        <div className="max-w-2xl">
          <h1 className="text-4xl font-bold mb-4">欢迎来到 Go-Mall</h1>
          <p className="text-xl mb-6">基于 Go-Zero 微服务架构的现代电商平台</p>
          <Link
            to="/products"
            className="inline-block bg-white text-blue-600 px-6 py-3 rounded-lg font-medium hover:bg-gray-100 transition-colors"
          >
            浏览商品
          </Link>
        </div>
      </div>

      {/* Features */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="bg-white rounded-xl p-6 shadow-sm">
          <div className="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center mb-4">
            <svg className="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
          </div>
          <h3 className="text-lg font-semibold text-gray-900 mb-2">高性能</h3>
          <p className="text-gray-600">基于 Go-Zero 微服务架构，支持高并发场景</p>
        </div>

        <div className="bg-white rounded-xl p-6 shadow-sm">
          <div className="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center mb-4">
            <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
          </div>
          <h3 className="text-lg font-semibold text-gray-900 mb-2">安全可靠</h3>
          <p className="text-gray-600">JWT 认证，数据加密传输</p>
        </div>

        <div className="bg-white rounded-xl p-6 shadow-sm">
          <div className="w-12 h-12 bg-purple-100 rounded-lg flex items-center justify-center mb-4">
            <svg className="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" />
            </svg>
          </div>
          <h3 className="text-lg font-semibold text-gray-900 mb-2">分布式事务</h3>
          <p className="text-gray-600">集成 DTM，保证数据一致性</p>
        </div>
      </div>

      {/* Quick Links */}
      <div className="mt-8 grid grid-cols-2 md:grid-cols-4 gap-4">
        <Link
          to="/products"
          className="bg-white rounded-xl p-4 shadow-sm hover:shadow-md transition-shadow text-center"
        >
          <span className="text-2xl mb-2 block">🛍️</span>
          <span className="text-gray-700 font-medium">商品列表</span>
        </Link>
        <Link
          to="/flash"
          className="bg-white rounded-xl p-4 shadow-sm hover:shadow-md transition-shadow text-center"
        >
          <span className="text-2xl mb-2 block">⚡</span>
          <span className="text-gray-700 font-medium">秒杀活动</span>
        </Link>
        <Link
          to="/cart"
          className="bg-white rounded-xl p-4 shadow-sm hover:shadow-md transition-shadow text-center"
        >
          <span className="text-2xl mb-2 block">🛒</span>
          <span className="text-gray-700 font-medium">购物车</span>
        </Link>
        <Link
          to="/orders"
          className="bg-white rounded-xl p-4 shadow-sm hover:shadow-md transition-shadow text-center"
        >
          <span className="text-2xl mb-2 block">📦</span>
          <span className="text-gray-700 font-medium">我的订单</span>
        </Link>
      </div>
    </div>
  );
}