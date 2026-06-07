import { useState } from 'react';
import { Link } from 'react-router-dom';
import { searchApi, normalizeSearchResponse, normalizeParseQueryResponse, getErrorMessage } from '../../services/api';

export default function Search() {
  const [query, setQuery] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [searchResult, setSearchResult] = useState(null);
  const [parsedQuery, setParsedQuery] = useState(null);

  const handleSearch = async (e) => {
    e.preventDefault();
    if (!query.trim()) {
      return;
    }

    setLoading(true);
    setError('');

    try {
      // 1. 先解析用户查询意图
      const parseResponse = await searchApi.parseQuery(query);
      const parseData = normalizeParseQueryResponse(parseResponse.data);
      setParsedQuery(parseData);

      // 2. 执行搜索
      const searchResponse = await searchApi.search({
        query: parseData.normalized_query || query,
        category: parseData.predicted_category,
        page: 1,
        page_size: 10,
      });
      const data = normalizeSearchResponse(searchResponse.data);
      setSearchResult(data);
    } catch (err) {
      console.error('Search error:', err);
      setError(getErrorMessage(err, '搜索失败，请稍后重试'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">智能搜索</h1>

      {/* 搜索表单 */}
      <form onSubmit={handleSearch} className="mb-6">
        <div className="flex gap-3">
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="试试: 帮我推荐几款适合夏天的裙子，便宜点的"
            className="flex-1 px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <button
            type="submit"
            disabled={loading || !query.trim()}
            className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
          >
            {loading ? '搜索中...' : '搜索'}
          </button>
        </div>
        <p className="text-sm text-gray-500 mt-2">
          支持自然语言搜索，如"便宜的学生外套"或"适合夏天的连衣裙"
        </p>
      </form>

      {/* 解析结果展示 */}
      {parsedQuery && (
        <div className="mb-6 p-4 bg-gray-50 rounded-lg">
          <h3 className="font-medium text-gray-700 mb-2">查询理解</h3>
          <div className="flex flex-wrap gap-2">
            {parsedQuery.predicted_category && (
              <span className="px-3 py-1 bg-blue-100 text-blue-700 rounded-full text-sm">
                分类: {parsedQuery.predicted_category}
              </span>
            )}
            {parsedQuery.brands?.map((brand) => (
              <span
                key={brand}
                className="px-3 py-1 bg-green-100 text-green-700 rounded-full text-sm"
              >
                品牌: {brand}
              </span>
            ))}
            {parsedQuery.modifiers?.map((mod) => (
              <span
                key={mod}
                className="px-3 py-1 bg-purple-100 text-purple-700 rounded-full text-sm"
              >
                {mod}
              </span>
            ))}
          </div>
        </div>
      )}

      {/* 错误提示 */}
      {error && (
        <div className="mb-6 p-4 bg-red-50 text-red-700 rounded-lg">
          {error}
        </div>
      )}

      {/* 搜索结果 */}
      {searchResult && (
        <div>
          <div className="flex items-center justify-between mb-4">
            <span className="text-gray-600">
              找到 {searchResult.total} 个相关商品
            </span>
          </div>

          {searchResult.results.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              <p className="text-lg mb-2">未找到相关商品</p>
              <p className="text-sm">试试用其他关键词搜索</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
              {searchResult.results.map((product) => (
                <Link
                  key={product.id}
                  to={`/products/${product.id}`}
                  className="bg-white rounded-xl p-4 shadow-sm hover:shadow-md transition-shadow"
                >
                  <div className="aspect-square bg-gray-100 rounded-lg mb-3 overflow-hidden">
                    {product.image_url ? (
                      <img
                        src={product.image_url}
                        alt={product.name}
                        className="w-full h-full object-cover"
                      />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center text-gray-400">
                        <svg className="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                        </svg>
                      </div>
                    )}
                  </div>
                  <h3 className="font-medium text-gray-900 mb-1 line-clamp-2">
                    {product.name}
                  </h3>
                  {product.description && (
                    <p className="text-sm text-gray-500 mb-2 line-clamp-1">
                      {product.description}
                    </p>
                  )}
                  <div className="flex items-center justify-between">
                    <span className="text-lg font-bold text-red-600">
                      ¥{product.price}
                    </span>
                    {product.score > 0 && (
                      <span className="text-sm text-gray-400">
                        匹配度: {(product.score * 100).toFixed(0)}%
                      </span>
                    )}
                  </div>
                  {product.category && (
                    <span className="inline-block mt-2 text-xs text-gray-500">
                      {product.category}
                    </span>
                  )}
                </Link>
              ))}
            </div>
          )}
        </div>
      )}

      {/* 初始状态 */}
      {!searchResult && !loading && !error && !query && (
        <div className="text-center py-12 text-gray-500">
          <div className="text-6xl mb-4">🔍</div>
          <p className="text-lg mb-2">输入搜索关键词开始智能搜索</p>
          <p className="text-sm">支持自然语言描述你想要的产品</p>
        </div>
      )}
    </div>
  );
}