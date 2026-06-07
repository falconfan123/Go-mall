import axios from "axios";
import { useAuthStore } from "../store/authStore";

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || "";

const api = axios.create({
  baseURL: apiBaseUrl,
  timeout: 10000,
  headers: {
    "Content-Type": "application/json",
  },
});

const refreshClient = axios.create({
  baseURL: apiBaseUrl,
  timeout: 10000,
  headers: {
    "Content-Type": "application/json",
  },
});

function normalizeErrorMessage(response) {
  if (!response) {
    return "Network error";
  }

  const data = response.data;
  return (
    data?.status_msg ||
    data?.statusMsg ||
    data?.message ||
    data?.error ||
    `Request failed with status ${response.status}`
  );
}

api.interceptors.request.use(
  (config) => {
    const { longToken, shortToken, user } = useAuthStore.getState();

    if (longToken) {
      config.headers["Long-Token"] = longToken;
    }
    if (shortToken) {
      config.headers["Short-Token"] = shortToken;
    }
    if (user?.user_id) {
      config.headers["user_id"] = String(user.user_id);
      config.headers["X-User-Id"] = String(user.user_id);
    }

    return config;
  },
  (error) => Promise.reject(error),
);

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const { response } = error;

    if (response?.status === 401) {
      const statusCode = getStatusCode(response.data);
      const originalRequest = error.config || {};

      if (statusCode === 10001 && !originalRequest._retry) {
        originalRequest._retry = true;
        try {
          const refreshed = await refreshToken();
          if (refreshed) {
            return api(originalRequest);
          }
        } catch (refreshError) {
          console.error("Token refresh failed:", refreshError);
        }
      }

      useAuthStore.getState().logout();
      window.location.href = "/login";
    } else if (response) {
      console.error(
        "Request error:",
        normalizeErrorMessage(response),
        response.data,
      );
    } else {
      console.error("Network error");
    }

    return Promise.reject(error);
  },
);

export default api;

export function getErrorMessage(error, fallback = "请求失败") {
  return normalizeErrorMessage(error?.response) || fallback;
}

export function getStatusCode(data) {
  return data?.status_code ?? data?.statusCode ?? 0;
}

export function getStatusMsg(data, fallback = "") {
  return data?.status_msg ?? data?.statusMsg ?? data?.message ?? fallback;
}

export function getString(value, legacy = "") {
  return value ?? legacy ?? "";
}

export function getNumber(value, legacy = 0) {
  const candidate = value ?? legacy;
  const parsed = Number(candidate);
  return Number.isFinite(parsed) ? parsed : 0;
}

export function getBoolean(value, legacy = false) {
  return value ?? legacy ?? false;
}

const ORDER_STATUS_MAP = {
  ORDER_STATUS_UNSPECIFIED: 0,
  ORDER_STATUS_CREATED: 1,
  ORDER_STATUS_PENDING_PAYMENT: 2,
  ORDER_STATUS_PAID: 3,
  ORDER_STATUS_COMPLETED: 4,
  ORDER_STATUS_CANCELLED: 5,
  ORDER_STATUS_CLOSED: 6,
  ORDER_STATUS_REFUND: 7,
};

const PAYMENT_STATUS_MAP = {
  PAYMENT_STATUS_UNSPECIFIED: 0,
  PAYMENT_STATUS_NOT_PAID: 1,
  PAYMENT_STATUS_PAYING: 2,
  PAYMENT_STATUS_PAID: 3,
  PAYMENT_STATUS_EXPIRED: 4,
  PAYMENT_STATUS_REFUND: 5,
};

const PAYMENT_RECORD_STATUS_MAP = {
  PAYMENT_STATUS_UNSPECIFIED: 0,
  PAYMENT_STATUS_UNPAID: 1,
  PAYMENT_STATUS_PAID: 2,
  PAYMENT_STATUS_FAILED: 3,
  PAYMENT_STATUS_FULLY_REFUNDED: 4,
  PAYMENT_STATUS_EXPIRED: 5,
};

export const PAYMENT_EXIST_STATUS_CODE = 10001;

function getEnumNumber(value, legacy, mapping, fallback = 0) {
  const candidate = value ?? legacy;
  if (typeof candidate === "number") {
    return Number.isFinite(candidate) ? candidate : fallback;
  }
  if (typeof candidate === "string") {
    const trimmed = candidate.trim();
    if (!trimmed) {
      return fallback;
    }
    const parsed = Number(trimmed);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
    if (Object.prototype.hasOwnProperty.call(mapping, trimmed)) {
      return mapping[trimmed];
    }
  }
  return fallback;
}

export function getShortToken(data) {
  return data?.short_token ?? data?.shortToken ?? "";
}

export function getLongToken(data) {
  return data?.long_token ?? data?.longToken ?? "";
}

export function normalizeUser(payload) {
  if (!payload) {
    return null;
  }

  return {
    user_id: getNumber(payload.user_id, payload.userId),
    username: getString(
      payload.username,
      payload.user_name ?? payload.userName,
    ),
    email: getString(payload.email),
    avatar_url: getString(payload.avatar_url, payload.avatarUrl),
  };
}

export function normalizeProduct(payload) {
  if (!payload) {
    return null;
  }

  const priceInCents = getNumber(payload.price);
  const originalPriceInCents = getNumber(
    payload.original_price,
    payload.originalPrice,
  );
  const image = getString(
    payload.thumbnail_url,
    payload.picture ?? payload.image_url ?? payload.image,
  );

  return {
    id: getNumber(payload.id, payload.product_id),
    product_id: getNumber(payload.product_id, payload.id),
    name: getString(payload.name),
    description: getString(payload.description),
    image,
    image_url: image,
    price: Number((priceInCents / 100).toFixed(2)),
    original_price:
      originalPriceInCents > 0
        ? Number((originalPriceInCents / 100).toFixed(2))
        : null,
    stock: getNumber(payload.stock),
    sold: getNumber(payload.sold, payload.sold_count),
  };
}

export function normalizeCartItem(payload) {
  if (!payload) {
    return null;
  }

  const price = getNumber(
    payload.product_price,
    payload.productPrice ?? payload.price,
  );
  const productId = getNumber(
    payload.product_id,
    payload.productId ?? payload.id,
  );
  const image = getString(
    payload.product_image,
    payload.productImage ?? payload.image_url ?? payload.image,
  );

  return {
    id: productId,
    product_id: productId,
    name: getString(payload.product_name, payload.productName ?? payload.name),
    image,
    image_url: image,
    price,
    quantity: getNumber(payload.quantity),
    checked: getBoolean(payload.checked, true),
  };
}

export function normalizeOrderItem(payload) {
  if (!payload) {
    return null;
  }

  const unitPriceInCents = getNumber(
    payload.unit_price,
    payload.unitPrice ?? payload.price,
  );
  return {
    product_id: getNumber(payload.product_id, payload.productId),
    name: getString(payload.product_name, payload.productName ?? payload.name),
    description: getString(
      payload.product_desc,
      payload.productDesc ?? payload.description,
    ),
    quantity: getNumber(payload.quantity),
    price: Number((unitPriceInCents / 100).toFixed(2)),
  };
}

export function normalizeOrder(payload) {
  if (!payload) {
    return null;
  }

  const payableAmount = getNumber(
    payload.payable_amount,
    payload.payableAmount,
  );
  const orderStatus = getEnumNumber(
    payload.order_status,
    payload.orderStatus ?? payload.status,
    ORDER_STATUS_MAP,
  );
  const paymentStatus = getEnumNumber(
    payload.payment_status,
    payload.paymentStatus,
    PAYMENT_STATUS_MAP,
  );

  return {
    order_id: getString(payload.order_id, payload.orderId),
    pre_order_id: getString(payload.pre_order_id, payload.preOrderId),
    user_id: getNumber(payload.user_id, payload.userId),
    order_status: orderStatus,
    payment_status: paymentStatus,
    total_price: Number((payableAmount / 100).toFixed(2)),
    payable_amount: payableAmount,
    created_at: getString(payload.created_at, payload.createdAt),
    items: Array.isArray(payload.items)
      ? payload.items.map(normalizeOrderItem).filter(Boolean)
      : [],
  };
}

export function normalizePayment(payload) {
  if (!payload) {
    return null;
  }

  const status = getEnumNumber(
    payload.status,
    payload.payment_status ?? payload.paymentStatus,
    PAYMENT_RECORD_STATUS_MAP,
  );

  return {
    payment_id: getString(payload.payment_id, payload.paymentId),
    order_id: getString(payload.order_id, payload.orderId),
    pre_order_id: getString(payload.pre_order_id, payload.preOrderId),
    pay_url: getString(payload.pay_url, payload.payUrl),
    status,
  };
}

export function normalizePaymentStatus(payload) {
  if (!payload) {
    return null;
  }

  return {
    payment: normalizePayment(payload.payment),
    order_status: getNumber(payload.order_status, payload.orderStatus),
    order_payment_status: getNumber(
      payload.order_payment_status,
      payload.orderPaymentStatus,
    ),
    status_code: getStatusCode(payload),
    status_msg: getStatusMsg(payload, ""),
  };
}

export function normalizeSeckillStatus(payload) {
  if (!payload) {
    return {
      is_purchased: false,
      order_id: "",
    };
  }

  return {
    is_purchased: getBoolean(payload.is_purchased, payload.isPurchased),
    order_id: getString(payload.order_id, payload.orderId),
  };
}

export function normalizeSearchResult(payload) {
  if (!payload) {
    return null;
  }

  const priceInCents = getNumber(payload.price);
  return {
    id: getNumber(payload.id),
    name: getString(payload.name),
    description: getString(payload.description),
    image_url: getString(payload.image_url, payload.imageUrl),
    category: getString(payload.category),
    brand: getString(payload.brand),
    price: priceInCents > 0 ? Number((priceInCents / 100).toFixed(2)) : 0,
    score: getNumber(payload.score),
  };
}

export function normalizeSearchResponse(payload) {
  if (!payload) {
    return {
      results: [],
      total: 0,
      page: 1,
      page_size: 10,
      total_pages: 0,
    };
  }

  return {
    results: Array.isArray(payload.results)
      ? payload.results.map(normalizeSearchResult).filter(Boolean)
      : [],
    total: getNumber(payload.total),
    page: getNumber(payload.page, 1),
    page_size: getNumber(payload.page_size, 10),
    total_pages: getNumber(payload.total_pages),
  };
}

export function normalizeParseQueryResponse(payload) {
  if (!payload) {
    return {
      original_query: "",
      normalized_query: "",
      predicted_category: "",
      brands: [],
      product_words: [],
      modifiers: [],
    };
  }

  return {
    original_query: getString(payload.original_query, payload.originalQuery),
    normalized_query: getString(payload.normalized_query, payload.normalizedQuery),
    predicted_category: getString(payload.predicted_category, payload.predictedCategory),
    brands: Array.isArray(payload.brands) ? payload.brands : [],
    product_words: Array.isArray(payload.product_words, payload.productWords)
      ? payload.product_words
      : [],
    modifiers: Array.isArray(payload.modifiers) ? payload.modifiers : [],
  };
}

export const userApi = {
  login: (data) => api.post("/api/v1/users/login", data),
  register: (data) => api.post("/api/v1/users/register", data),
  logout: (data) => api.post("/api/v1/users/logout", data),
  info: (userId) =>
    api.get("/api/v1/users/me", {
      params: userId ? { user_id: userId } : undefined,
    }),
  addAddress: (data) => api.post("/api/v1/users/addresses", data),
  getAddress: (addressId, userId) =>
    api.get("/api/v1/users/address", {
      params: {
        address_id: addressId,
        user_id: userId,
      },
    }),
};

export const productApi = {
  list: (params) => api.get("/api/v1/products", { params }),
  detail: (id, userId) =>
    api.get("/api/v1/products/detail", {
      params: {
        id,
        ...(userId ? { user_id: userId } : {}),
      },
    }),
};

export const cartApi = {
  list: (userId) =>
    api.get("/api/v1/cart/items", { params: { user_id: userId, id: userId } }),
  add: (data) => api.post("/api/v1/cart/items", data),
  sub: (data) => api.post("/api/v1/cart/items/sub", data),
  remove: (data) => api.post("/api/v1/cart/items/delete", data),
};

export const checkoutApi = {
  prepare: (data) => api.post("/api/v1/checkout/prepare", data),
};

export const orderApi = {
  list: (params) => api.get("/api/v1/orders", { params }),
  detail: (params) => api.get("/api/v1/orders/detail", { params }),
  create: (data) => api.post("/api/v1/orders", data),
  cancel: (data) => api.post("/api/v1/orders/cancel", data),
};

export const paymentApi = {
  list: (params) => api.get("/api/v1/payments", { params }),
  create: (data) => api.post("/api/v1/payments", data),
  status: (params) => api.get("/api/v1/payments/status", { params }),
};

export const seckillApi = {
  systemTime: () => api.get("/api/v1/system/time"),
  activityToken: (activityId) =>
    api.get("/api/v1/activity/token", { params: { activity_id: activityId } }),
  activityStatus: (activityId) =>
    api.get("/api/v1/activity/status", { params: { activity_id: activityId } }),
  submit: (data) => api.post("/api/v1/orders/seckill", data),
};

export const searchApi = {
  search: (params) =>
    api.get("/api/v1/search", {
      params: {
        query: params.query,
        category: params.category || "",
        page: params.page || 1,
        page_size: params.page_size || 10,
        sort_by: params.sort_by || "score",
        sort_order: params.sort_order || "desc",
      },
    }),
  parseQuery: (query) =>
    api.post("/api/v1/search/parse", { query }),
};

export async function refreshToken() {
  const { longToken, shortToken, setShortToken } = useAuthStore.getState();
  if (!longToken) {
    return false;
  }

  const response = await refreshClient.post(
    "/api/v1/auth/refresh",
    {
      long_token: longToken,
      short_token: shortToken || "",
      client_ip: "127.0.0.1",
    },
    {
      headers: {
        "Long-Token": longToken,
        "Short-Token": shortToken || "",
      },
    },
  );

  const data = response.data || {};
  if (getStatusCode(data) !== 0) {
    return false;
  }

  const nextShortToken = getShortToken(data);
  if (!nextShortToken) {
    return false;
  }

  setShortToken(nextShortToken);
  return true;
}
