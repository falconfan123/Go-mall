import { create } from 'zustand';
import { cartApi } from '../services/api';

export const useCartStore = create((set, get) => ({
  items: [],
  loading: false,
  error: null,

  // 获取购物车列表
  fetchCart: async () => {
    set({ loading: true });
    try {
      const response = await cartApi.list();
      set({ items: response.data || [], loading: false });
    } catch (error) {
      set({ error: error.message, loading: false });
    }
  },

  // 添加商品到购物车
  addItem: async (productId, quantity = 1) => {
    try {
      await cartApi.add({ product_id: productId, quantity });
      await get().fetchCart();
      return true;
    } catch (error) {
      set({ error: error.message });
      return false;
    }
  },

  // 更新商品数量
  updateQuantity: async (itemId, quantity) => {
    try {
      await cartApi.update({ id: itemId, quantity });
      await get().fetchCart();
      return true;
    } catch (error) {
      set({ error: error.message });
      return false;
    }
  },

  // 移除商品
  removeItem: async (itemId) => {
    try {
      await cartApi.remove(itemId);
      await get().fetchCart();
      return true;
    } catch (error) {
      set({ error: error.message });
      return false;
    }
  },

  // 清空购物车
  clearCart: async () => {
    try {
      await cartApi.clear();
      set({ items: [] });
      return true;
    } catch (error) {
      set({ error: error.message });
      return false;
    }
  },

  // 计算总价
  getTotalPrice: () => {
    return get().items.reduce((total, item) => {
      return total + (item.price || 0) * (item.quantity || 0);
    }, 0);
  },

  // 获取商品数量
  getItemCount: () => {
    return get().items.reduce((count, item) => count + (item.quantity || 0), 0);
  },
}));