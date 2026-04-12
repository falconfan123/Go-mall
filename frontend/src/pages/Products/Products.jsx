import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useProductStore } from '../../store/productStore';
import { useCartStore } from '../../store/cartStore';
import { Button } from '../../components/common/Button';
import { Input } from '../../components/common/Input';
import { Spinner } from '../../components/common/Spinner';
import { toast } from '../../components/common/Toast';

export default function Products() {
  const { products, loading, fetchProducts, searchProducts } = useProductStore();
  const { addItem } = useCartStore();
  const [keyword, setKeyword] = useState('');
  const [addingId, setAddingId] = useState(null);

  useEffect(() => {
    fetchProducts();
  }, [fetchProducts]);

  const handleSearch = (e) => {
    e.preventDefault();
    if (keyword.trim()) {
      searchProducts(keyword);
    } else {
      fetchProducts();
    }
  };

  const handleAddToCart = async (productId) => {
    setAddingId(productId);
    const success = await addItem(productId, 1);
    setAddingId(null);
    if (success) {
      toast.success('添加成功');
    }
  };

  return (
    <div>
      {/* 搜索栏 */}
      <form onSubmit={handleSearch} className="mb-6">
        <div className="flex gap-4">
          <div className="flex-1">
            <Input
              placeholder="搜索商品..."
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
            />
          </div>
          <Button type="submit">搜索</Button>
          <Button
            type="button"
            variant="secondary"
            onClick={() => {
              setKeyword('');
              fetchProducts();
            }}
          >
            重置
          </Button>
        </div>
      </form>

      {/* 商品列表 */}
      {loading ? (
        <div className="flex justify-center py-12">
          <Spinner size="large" />
        </div>
      ) : products.length === 0 ? (
        <div className="text-center py-12 text-gray-500">
          暂无商品
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {products.map((product) => (
            <div
              key={product.id || product.product_id}
              className="bg-white rounded-xl shadow-sm overflow-hidden hover:shadow-md transition-shadow"
            >
              <Link to={`/products/${product.id || product.product_id}`}>
                <div className="aspect-w-16 aspect-h-9 bg-gray-200">
                  <img
                    src={product.image || product.image_url || '/placeholder.png'}
                    alt={product.name}
                    className="object-cover w-full h-48"
                    onError={(e) => {
                      e.target.src = 'https://via.placeholder.com/300x200?text=No+Image';
                    }}
                  />
                </div>
              </Link>
              <div className="p-4">
                <Link to={`/products/${product.id || product.product_id}`}>
                  <h3 className="font-semibold text-gray-900 mb-2 hover:text-blue-600">
                    {product.name}
                  </h3>
                </Link>
                <p className="text-sm text-gray-500 mb-3 line-clamp-2">
                  {product.description}
                </p>
                <div className="flex items-center justify-between">
                  <span className="text-xl font-bold text-blue-600">
                    ¥{product.price}
                  </span>
                  <Button
                    size="small"
                    loading={addingId === product.id || addingId === product.product_id}
                    onClick={() => handleAddToCart(product.id || product.product_id)}
                  >
                    加入购物车
                  </Button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}