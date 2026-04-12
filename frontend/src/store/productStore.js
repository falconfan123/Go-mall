import { create } from 'zustand';
import { productApi, flashApi } from '../services/api';

export const useProductStore = create((set, get) => ({
  products: [],
  flashProducts: [],
  currentProduct: null,
  loading: false,
  error: null,

  // 获取商品列表
  fetchProducts: async (params = {}) => {
    set({ loading: true });
    try {
      // 确保传递正确的分页参数
      const queryParams = {
        page: 1,
        pageSize: 100,
        ...params
      };
      const response = await productApi.list(queryParams);
      // 后端返回: { statusCode: 0, products: [...] }
      // 兼容 response.data.products 或 response.products
      let products = [];
      if (response.data?.products) {
        products = response.data.products;
      } else if (response.products) {
        products = response.products;
      } else if (Array.isArray(response.data)) {
        products = response.data;
      }
      // 转换价格单位（分 → 元）
      products = products.map(p => ({
        ...p,
        price: (parseFloat(p.price || 0) / 100).toFixed(2),
        original_price: p.original_price ? (parseFloat(p.original_price) / 100).toFixed(2) : undefined,
      }));
      set({ products, loading: false });
    } catch (error) {
      set({ error: error.message, loading: false, products: [] });
    }
  },

  // 获取商品详情
  fetchProductDetail: async (id) => {
    set({ loading: true });
    try {
      const response = await productApi.detail(id);
      set({ currentProduct: response.data, loading: false });
    } catch (error) {
      set({ error: error.message, loading: false });
    }
  },

  // 搜索商品
  searchProducts: async (keyword) => {
    set({ loading: true });
    try {
      const response = await productApi.search(keyword);
      set({ products: response.data || [], loading: false });
    } catch (error) {
      set({ error: error.message, loading: false });
    }
  },

  // 获取秒杀商品列表
  fetchFlashProducts: async () => {
    set({ loading: true });
    try {
      const response = await flashApi.list();
      // 后端返回格式: { statusCode: 0, products: [...] }
      let products = [];
      if (response.data?.products) {
        products = response.data.products;
      } else if (response.products) {
        products = response.products;
      } else if (Array.isArray(response.data)) {
        products = response.data;
      }
      // 转换价格单位（分 → 元）
      products = products.map(p => ({
        ...p,
        price: (parseFloat(p.price || 0) / 100).toFixed(2),
        original_price: p.original_price ? (parseFloat(p.original_price) / 100).toFixed(2) : (parseFloat(p.price || 0) * 1.2 / 100).toFixed(2),
      }));
      set({ flashProducts: products, loading: false });
    } catch (error) {
      set({ error: error.message, loading: false, flashProducts: [] });
    }
  },

  // 获取秒杀商品详情
  fetchFlashProductDetail: async (id) => {
    set({ loading: true });
    try {
      const response = await flashApi.detail(id);
      set({ currentProduct: response.data, loading: false });
    } catch (error) {
      set({ error: error.message, loading: false });
    }
  },

  // 清除当前商品
  clearCurrentProduct: () => {
    set({ currentProduct: null });
  },
}));