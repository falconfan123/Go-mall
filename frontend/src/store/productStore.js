import { create } from "zustand";
import { getErrorMessage, normalizeProduct, productApi } from "../services/api";
import { useAuthStore } from "./authStore";

export const useProductStore = create((set, get) => ({
  products: [],
  flashProducts: [],
  currentProduct: null,
  loading: false,
  error: null,

  fetchProducts: async (params = {}) => {
    set({ loading: true, error: null });
    try {
      const response = await productApi.list({
        cursor: 0,
        limit: 100,
        ...params,
      });
      const data = response.data || {};
      const products = Array.isArray(data.products)
        ? data.products.map(normalizeProduct).filter(Boolean)
        : [];
      set({
        products,
        flashProducts: products,
        loading: false,
      });
    } catch (error) {
      set({
        error: getErrorMessage(error, "获取商品失败"),
        loading: false,
        products: [],
        flashProducts: [],
      });
    }
  },

  fetchProductDetail: async (id) => {
    set({ loading: true, error: null });
    try {
      const userId = useAuthStore.getState().user?.user_id;
      const response = await productApi.detail(id, userId);
      const product = normalizeProduct(response.data?.product || response.data);
      set({ currentProduct: product, loading: false });
    } catch (error) {
      set({
        error: getErrorMessage(error, "获取商品详情失败"),
        loading: false,
        currentProduct: null,
      });
    }
  },

  searchProducts: async (keyword) => {
    const sourceProducts = get().products;
    if (!keyword.trim()) {
      return get().fetchProducts();
    }

    const normalizedKeyword = keyword.trim().toLowerCase();
    const filtered = sourceProducts.filter((product) => {
      return (
        product.name.toLowerCase().includes(normalizedKeyword) ||
        product.description.toLowerCase().includes(normalizedKeyword)
      );
    });

    set({ products: filtered, loading: false });
  },

  fetchFlashProducts: async () => {
    await get().fetchProducts();
    set({ flashProducts: get().products });
  },

  fetchFlashProductDetail: async (id) => {
    await get().fetchProductDetail(id);
  },

  clearCurrentProduct: () => {
    set({ currentProduct: null });
  },
}));
