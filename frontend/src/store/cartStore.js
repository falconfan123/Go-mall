import { create } from "zustand";
import { cartApi, getErrorMessage, normalizeCartItem } from "../services/api";
import { useAuthStore } from "./authStore";

async function repeatCall(times, fn) {
  for (let i = 0; i < times; i += 1) {
    await fn();
  }
}

export const useCartStore = create((set, get) => ({
  items: [],
  loading: false,
  error: null,

  fetchCart: async () => {
    const userId = useAuthStore.getState().user?.user_id;
    if (!userId) {
      set({ items: [], loading: false });
      return [];
    }

    set({ loading: true, error: null });
    try {
      const response = await cartApi.list(userId);
      const data = response.data || {};
      const items = Array.isArray(data.data)
        ? data.data.map(normalizeCartItem).filter(Boolean)
        : [];
      set({ items, loading: false });
      return items;
    } catch (error) {
      set({
        error: getErrorMessage(error, "获取购物车失败"),
        loading: false,
        items: [],
      });
      return [];
    }
  },

  addItem: async (product, quantity = 1) => {
    const userId = useAuthStore.getState().user?.user_id;
    if (!userId) {
      set({ error: "请先登录" });
      return false;
    }

    try {
      await repeatCall(quantity, () =>
        cartApi.add({
          user_id: userId,
          product_id: product.product_id ?? product.id ?? product,
          product_name: product.name ?? "",
          product_image: product.image_url ?? product.image ?? "",
          productPrice: product.price ?? 0,
          quantity: 0,
          checked: true,
        }),
      );
      await get().fetchCart();
      return true;
    } catch (error) {
      set({ error: getErrorMessage(error, "加入购物车失败") });
      return false;
    }
  },

  updateQuantity: async (productId, quantity) => {
    const userId = useAuthStore.getState().user?.user_id;
    const currentItem = get().items.find(
      (item) => item.product_id === productId || item.id === productId,
    );
    if (!userId || !currentItem) {
      return false;
    }

    try {
      const delta = quantity - currentItem.quantity;
      if (delta > 0) {
        await repeatCall(delta, () =>
          cartApi.add({
            user_id: userId,
            product_id: currentItem.product_id,
            product_name: currentItem.name,
            product_image: currentItem.image_url || currentItem.image,
            productPrice: currentItem.price,
            quantity: 0,
            checked: currentItem.checked,
          }),
        );
      } else if (delta < 0) {
        await repeatCall(Math.abs(delta), () =>
          cartApi.sub({
            user_id: userId,
            product_id: currentItem.product_id,
          }),
        );
      }

      await get().fetchCart();
      return true;
    } catch (error) {
      set({ error: getErrorMessage(error, "更新数量失败") });
      return false;
    }
  },

  removeItem: async (productId) => {
    const userId = useAuthStore.getState().user?.user_id;
    if (!userId) {
      return false;
    }

    try {
      await cartApi.remove({
        user_id: userId,
        product_id: productId,
      });
      await get().fetchCart();
      return true;
    } catch (error) {
      set({ error: getErrorMessage(error, "移除商品失败") });
      return false;
    }
  },

  clearCart: async () => {
    try {
      const items = [...get().items];
      for (const item of items) {
        await get().removeItem(item.product_id);
      }
      set({ items: [] });
      return true;
    } catch (error) {
      set({ error: getErrorMessage(error, "清空购物车失败") });
      return false;
    }
  },

  getTotalPrice: () => {
    return get().items.reduce(
      (total, item) => total + (Number(item.price) || 0) * (item.quantity || 0),
      0,
    );
  },

  getItemCount: () => {
    return get().items.reduce((count, item) => count + (item.quantity || 0), 0);
  },
}));
