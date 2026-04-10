import { useEffect, useState } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useProductStore } from '../../store/productStore';
import { useCartStore } from '../../store/cartStore';
import { Button } from '../../components/common/Button';
import { Spinner } from '../../components/common/Spinner';
import { toast } from '../../components/common/Toast';

export default function ProductDetail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { currentProduct, loading, fetchProductDetail, clearCurrentProduct } = useProductStore();
  const { addItem } = useCartStore();
  const [quantity, setQuantity] = useState(1);
  const [adding, setAdding] = useState(false);

  useEffect(() => {
    fetchProductDetail(id);
    return () => clearCurrentProduct();
  }, [id, fetchProductDetail, clearCurrentProduct]);

  const handleAddToCart = async () => {
    setAdding(true);
    const success = await addItem(id, quantity);
    setAdding(false);
    if (success) {
      toast.success('添加成功');
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center py-12">
        <Spinner size="large" />
      </div>
    );
  }

  if (!currentProduct) {
    return (
      <div className="text-center py-12 text-gray-500">
        商品不存在
      </div>
    );
  }

  return (
    <div>
      <Link to="/products" className="text-blue-600 hover:text-blue-700 mb-4 inline-block">
        ← 返回商品列表
      </Link>

      <div className="bg-white rounded-xl shadow-sm p-8">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
          {/* 商品图片 */}
          <div className="aspect-w-1 aspect-h-1 bg-gray-200 rounded-lg overflow-hidden">
            <img
              src={currentProduct.image || currentProduct.image_url || 'https://via.placeholder.com/400x400?text=No+Image'}
              alt={currentProduct.name}
              className="object-cover w-full h-full"
              onError={(e) => {
                e.target.src = 'https://via.placeholder.com/400x400?text=No+Image';
              }}
            />
          </div>

          {/* 商品信息 */}
          <div>
            <h1 className="text-2xl font-bold text-gray-900 mb-4">
              {currentProduct.name}
            </h1>
            <p className="text-3xl font-bold text-blue-600 mb-6">
              ¥{currentProduct.price}
            </p>
            <p className="text-gray-600 mb-6">
              {currentProduct.description}
            </p>

            <div className="mb-6">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                数量
              </label>
              <div className="flex items-center gap-4">
                <button
                  className="px-3 py-1 border rounded-md hover:bg-gray-50"
                  onClick={() => setQuantity(Math.max(1, quantity - 1))}
                >
                  -
                </button>
                <span className="text-lg">{quantity}</span>
                <button
                  className="px-3 py-1 border rounded-md hover:bg-gray-50"
                  onClick={() => setQuantity(quantity + 1)}
                >
                  +
                </button>
              </div>
            </div>

            <Button
              size="large"
              loading={adding}
              onClick={handleAddToCart}
            >
              加入购物车
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}