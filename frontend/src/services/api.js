import axios from 'axios';
import { useAuthStore } from '../store/authStore';

// 创建 axios 实例
const api = axios.create({
  baseURL: '/douyin',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器 - 自动携带令牌
api.interceptors.request.use(
  (config) => {
    const { longToken, shortToken } = useAuthStore.getState();
    if (longToken) {
      config.headers['Long-Token'] = longToken;
    }
    if (shortToken) {
      config.headers['Short-Token'] = shortToken;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器 - 统一错误处理和令牌刷新
api.interceptors.response.use(
  (response) => {
    // 自动刷新短令牌
    const newToken = response.headers['short-token-refresh'];
    if (newToken) {
      useAuthStore.getState().setShortToken(newToken);
    }
    return response;
  },
  (error) => {
    const { response } = error;

    // 统一错误处理
    if (response) {
      switch (response.status) {
        case 401:
          // 未授权，清除登录状态
          useAuthStore.getState().logout();
          window.location.href = '/login';
          break;
        case 403:
          // 禁止访问
          console.error('Access forbidden');
          break;
        case 500:
          // 服务器错误
          console.error('Server error');
          break;
        default:
          console.error('Request error:', response.data);
      }
    } else {
      console.error('Network error');
    }

    return Promise.reject(error);
  }
);

export default api;

// 用户 API
export const userApi = {
  login: (data) => api.post('/user/login', data),
  register: (data) => api.post('/user/register', data),
  info: () => api.get('/user/info'),
};

// 商品 API
export const productApi = {
  list: (params) => api.get('/product/list', { params }),
  detail: (id) => api.get(`/product/detail?id=${id}`),
  search: (keyword) => api.get('/product/search', { params: { keyword } }),
};

// 购物车 API
export const cartApi = {
  list: () => api.get('/cart/list'),
  add: (data) => api.post('/cart/add', data),
  update: (data) => api.put('/cart/update', data),
  remove: (id) => api.delete(`/cart/remove?id=${id}`),
  clear: () => api.delete('/cart/clear'),
};

// 订单 API
export const orderApi = {
  list: () => api.get('/order/list'),
  detail: (id) => api.get(`/order/detail?id=${id}`),
  create: (data) => api.post('/order/create', data),
  cancel: (id) => api.post(`/order/cancel?id=${id}`),
};

// 支付 API
export const paymentApi = {
  create: (data) => api.post('/payment/create', data),
  status: (id) => api.get(`/payment/status?id=${id}`),
};

// 秒杀 API
export const flashApi = {
  list: () => api.get('/flash/list'),
  detail: (id) => api.get(`/flash/detail?id=${id}`),
  status: (activityId) => api.get(`/flash/status?activity_id=${activityId}`),
};

// 秒杀系统 API
export const seckillApi = {
  systemTime: () => api.get('/api/v1/system/time'),
  activityToken: (activityId) => api.get(`/api/v1/activity/token?activity_id=${activityId}`),
  activityStatus: (activityId) => api.get(`/api/v1/activity/status?activity_id=${activityId}`),
  submit: (data) => api.post('/api/v1/order/seckill', data),
};