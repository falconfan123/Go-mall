import { create } from "zustand";
import {
  getErrorMessage,
  getLongToken,
  getShortToken,
  getStatusCode,
  getStatusMsg,
  normalizeUser,
  userApi,
} from "../services/api";

const DEVICE_ID_KEY = "device_id";

function getDeviceId() {
  let deviceId = localStorage.getItem(DEVICE_ID_KEY);
  if (!deviceId) {
    deviceId = "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (c) => {
      const r = (Math.random() * 16) | 0;
      const v = c === "x" ? r : (r & 0x3) | 0x8;
      return v.toString(16);
    });
    localStorage.setItem(DEVICE_ID_KEY, deviceId);
  }
  return deviceId;
}

export const useAuthStore = create((set, get) => ({
  user: (() => {
    const raw = localStorage.getItem("user");
    if (!raw) {
      return null;
    }
    try {
      return JSON.parse(raw);
    } catch {
      return null;
    }
  })(),
  longToken: localStorage.getItem("longToken") || null,
  shortToken: localStorage.getItem("shortToken") || null,
  loading: false,
  error: null,

  login: async (credentials) => {
    set({ loading: true, error: null });
    try {
      const response = await userApi.login({
        ...credentials,
        email: credentials.email || "",
        username: credentials.username || "",
        ip: "127.0.0.1",
        device_id: getDeviceId(),
      });

      const data = response.data || {};
      if (getStatusCode(data) !== 0) {
        set({ error: getStatusMsg(data, "登录失败"), loading: false });
        return false;
      }

      const shortToken = getShortToken(data);
      const longToken = getLongToken(data);

      localStorage.setItem("shortToken", shortToken);
      localStorage.setItem("longToken", longToken);

      const user = normalizeUser({
        user_id: data.user_id ?? data.userId,
        username: data.user_name ?? data.userName ?? credentials.username,
        email: credentials.email,
      });
      localStorage.setItem("user", JSON.stringify(user));

      set({
        user,
        longToken,
        shortToken,
        loading: false,
        error: null,
      });

      await get().fetchUserInfo();
      return true;
    } catch (error) {
      set({ error: getErrorMessage(error, "登录失败"), loading: false });
      return false;
    }
  },

  register: async (payload) => {
    set({ loading: true, error: null });
    try {
      const response = await userApi.register({
        username: payload.username,
        password: payload.password,
        confirm_password: payload.confirm_password || payload.password,
        email: payload.email || "",
        ip: "127.0.0.1",
        device_id: getDeviceId(),
      });
      const data = response.data || {};
      if (getStatusCode(data) !== 0) {
        set({ error: getStatusMsg(data, "注册失败"), loading: false });
        return false;
      }
      set({ loading: false, error: null });
      return true;
    } catch (error) {
      set({ error: getErrorMessage(error, "注册失败"), loading: false });
      return false;
    }
  },

  fetchUserInfo: async () => {
    const current = get().user;
    if (!current?.user_id) {
      return null;
    }

    try {
      const response = await userApi.info(current.user_id);
      const data = response.data || {};
      if (getStatusCode(data) !== 0) {
        return null;
      }

      const user = normalizeUser(data);
      localStorage.setItem("user", JSON.stringify(user));
      set({ user });
      return user;
    } catch (error) {
      console.error("Failed to fetch user info:", error);
      return null;
    }
  },

  setShortToken: (token) => {
    localStorage.setItem("shortToken", token);
    set({ shortToken: token });
  },

  logout: async () => {
    const { user, longToken } = get();
    if (longToken) {
      try {
        await userApi.logout({
          user_id: user?.user_id || 0,
          long_token: longToken,
          ip: "127.0.0.1",
        });
      } catch (error) {
        console.error("Failed to notify logout:", error);
      }
    }

    localStorage.removeItem("longToken");
    localStorage.removeItem("shortToken");
    localStorage.removeItem("user");
    set({
      user: null,
      longToken: null,
      shortToken: null,
      error: null,
      loading: false,
    });
  },

  checkAuth: async () => {
    const { longToken, shortToken, user } = get();
    if (!longToken && !shortToken) {
      return;
    }
    if (user?.user_id) {
      await get().fetchUserInfo();
    }
  },
}));
