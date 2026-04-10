import { create } from 'zustand';
import { userApi } from '../services/api';

// 设备ID管理
const getDeviceId = () => {
  const DEVICE_ID_KEY = 'device_id';
  let deviceId = localStorage.getItem(DEVICE_ID_KEY);
  if (!deviceId) {
    deviceId = 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
      const r = Math.random() * 16 | 0;
      const v = c === 'x' ? r : (r & 0x3 | 0x8);
      return v.toString(16);
    });
    localStorage.setItem(DEVICE_ID_KEY, deviceId);
  }
  return deviceId;
};

export const useAuthStore = create((set, get) => ({
  user: null,
  longToken: localStorage.getItem('longToken') || null,
  shortToken: localStorage.getItem('shortToken') || null,
  loading: false,
  error: null,

  // 登录
  login: async (credentials) => {
    set({ loading: true, error: null });
    try {
      // 添加 device_id
      const data = { ...credentials, device_id: getDeviceId() };
      const response = await userApi.login(data);
      const { long_token, short_token, user } = response.data;
      localStorage.setItem('longToken', long_token);
      localStorage.setItem('shortToken', short_token);
      set({
        user,
        longToken: long_token,
        shortToken: short_token,
        loading: false,
      });
      return true;
    } catch (error) {
      set({ error: error.response?.data?.message || '登录失败', loading: false });
      return false;
    }
  },

  // 注册
  register: async (data) => {
    set({ loading: true, error: null });
    try {
      const response = await userApi.register(data);
      return response.data;
    } catch (error) {
      set({ error: error.response?.data?.message || '注册失败', loading: false });
      return false;
    }
  },

  // 获取用户信息
  fetchUserInfo: async () => {
    try {
      const response = await userApi.info();
      set({ user: response.data });
    } catch (error) {
      console.error('Failed to fetch user info:', error);
    }
  },

  // 设置短令牌
  setShortToken: (token) => {
    localStorage.setItem('shortToken', token);
    set({ shortToken: token });
  },

  // 登出
  logout: () => {
    localStorage.removeItem('longToken');
    localStorage.removeItem('shortToken');
    set({ user: null, longToken: null, shortToken: null });
  },

  // 检查登录状态
  checkAuth: () => {
    const { longToken } = get();
    if (longToken) {
      get().fetchUserInfo();
    }
  },
}));