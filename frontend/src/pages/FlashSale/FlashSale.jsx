import { useEffect, useState } from "react";
import { create } from "zustand";
import { useProductStore } from "../../store/productStore";
import {
  getBoolean,
  getNumber,
  getStatusCode,
  getStatusMsg,
  getString,
  normalizeSeckillStatus,
  seckillApi,
  userApi,
} from "../../services/api";
import { useAuthStore } from "../../store/authStore";
import { Button } from "../../components/common/Button";
import { Spinner } from "../../components/common/Spinner";
import { toast } from "../../components/common/Toast";

// 秒杀系统状态管理
const useSeckillStore = create((set, get) => ({
  timeOffset: 0,
  timeLastSync: 0,
  isSyncing: false,
  pathKeys: {},

  // 获取服务器时间
  getServerTime: () => Date.now() + get().timeOffset,

  // 同步服务器时间
  syncServerTime: async () => {
    if (get().isSyncing) return;
    set({ isSyncing: true });
    try {
      const response = await seckillApi.systemTime();
      const serverTime = getNumber(response.data?.now);
      const now = Date.now();
      set({ timeOffset: serverTime - now, timeLastSync: now });
    } catch (error) {
      console.error("Time sync error:", error);
    } finally {
      set({ isSyncing: false });
    }
  },

  // 获取秒杀 Token
  getSeckillToken: async (productId) => {
    const cached = get().pathKeys[productId];
    if (cached && cached.expiresAt > get().getServerTime()) {
      return cached.pathKey;
    }
    try {
      const response = await seckillApi.activityToken(productId);
      const pathKey = getString(
        response.data?.pathKey,
        response.data?.path_key,
      );
      if (pathKey) {
        set((state) => ({
          pathKeys: {
            ...state.pathKeys,
            [productId]: {
              pathKey,
              expiresAt: get().getServerTime() + 5 * 60 * 1000,
            },
          },
        }));
        return pathKey;
      }
    } catch (error) {
      console.error("Get token error:", error);
    }
    return null;
  },

  // 提交秒杀订单
  submitSeckill: async (productId, pathKey, userId, addressId) => {
    try {
      const response = await seckillApi.submit({
        product_id: productId,
        path_key: pathKey,
        user_id: userId,
        address_id: addressId,
      });
      return response.data;
    } catch (error) {
      console.error("Seckill error:", error);
      return { status_code: 1, message: "秒杀请求失败" };
    }
  },

  getSeckillStatus: async (productId) => {
    try {
      const response = await seckillApi.activityStatus(productId);
      return normalizeSeckillStatus(response.data);
    } catch (error) {
      console.error("Get seckill status error:", error);
      return normalizeSeckillStatus(null);
    }
  },

  // 计算按钮状态
  calculateButtonState: (productId, activityStartTime) => {
    const now = get().getServerTime();
    const diff = activityStartTime - now;

    if (diff > 10000) {
      return { status: "waiting", text: "即将开始", class: "bg-gray-400" };
    }
    if (diff > 0) {
      return {
        status: "critical",
        text: `${Math.ceil(diff / 1000)}秒`,
        class: "bg-red-600 animate-pulse",
      };
    }
    return {
      status: "trigger",
      text: "立即抢购",
      class: "bg-red-600 hover:bg-red-700",
    };
  },
}));

export default function FlashSale() {
  const { flashProducts, loading, fetchFlashProducts } = useProductStore();
  const { user } = useAuthStore();
  const [countdown, setCountdown] = useState({
    hours: 0,
    minutes: 0,
    seconds: 0,
  });
  const [buttonStates, setButtonStates] = useState({});
  const {
    syncServerTime,
    getSeckillToken,
    submitSeckill,
    calculateButtonState,
    getServerTime,
    getSeckillStatus,
  } = useSeckillStore();
  const [submittingId, setSubmittingId] = useState(null);
  const [seckillAddressId, setSeckillAddressId] = useState(null);

  useEffect(() => {
    fetchFlashProducts();
    syncServerTime();

    // 每10秒同步一次时间
    const timeInterval = setInterval(syncServerTime, 10000);

    // 每秒更新倒计时和按钮状态
    const buttonInterval = setInterval(() => {
      const newStates = {};
      flashProducts.forEach((product) => {
        const startTime =
          product.start_time || product.startTime || Date.now() - 1000;
        newStates[product.id] = calculateButtonState(product.id, startTime);
      });
      setButtonStates(newStates);
    }, 1000);

    return () => {
      clearInterval(timeInterval);
      clearInterval(buttonInterval);
    };
  }, [flashProducts, fetchFlashProducts, syncServerTime, calculateButtonState]);

  // 计算活动倒计时
  useEffect(() => {
    const timer = setInterval(() => {
      const now = getServerTime();
      // 假设活动在2小时后结束
      const endTime = now + 2 * 60 * 60 * 1000;
      const diff = endTime - now;

      if (diff > 0) {
        setCountdown({
          hours: Math.floor(diff / (1000 * 60 * 60)),
          minutes: Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60)),
          seconds: Math.floor((diff % (1000 * 60)) / 1000),
        });
      }
    }, 1000);

    return () => clearInterval(timer);
  }, [getServerTime]);

  const ensureSeckillAddress = async () => {
    if (!user?.user_id) {
      return null;
    }
    if (seckillAddressId) {
      return seckillAddressId;
    }

    const response = await userApi.addAddress({
      recipient_name: user.username || `user-${user.user_id}`,
      phone_number: "13800000000",
      province: "Shanghai",
      city: "Shanghai",
      detailed_address: "flash sale address",
      is_default: true,
      user_id: user.user_id,
      ip: "127.0.0.1",
    });
    const data = response.data || {};
    if (getStatusCode(data) !== 0) {
      throw new Error(getStatusMsg(data, "创建秒杀地址失败"));
    }

    const addressId = getNumber(data.data?.addressId, data.data?.address_id);
    if (!addressId) {
      throw new Error("秒杀地址缺失");
    }

    setSeckillAddressId(addressId);
    return addressId;
  };

  const handleSeckill = async (product) => {
    const productId = product.id;

    if (!user) {
      toast.warning("请先登录");
      return;
    }

    setSubmittingId(productId);

    try {
      const addressId = await ensureSeckillAddress();
      if (!addressId) {
        throw new Error("缺少秒杀地址");
      }

      const pathKey = await getSeckillToken(productId);
      if (!pathKey) {
        throw new Error("获取秒杀资格失败");
      }

      const result = await submitSeckill(
        productId,
        pathKey,
        user.user_id,
        addressId,
      );

      if (
        getStatusCode(result) === 0 ||
        getString(result.status) === "success"
      ) {
        toast.success("秒杀成功！");
        const status = await getSeckillStatus(productId);
        if (getBoolean(status.is_purchased)) {
          setButtonStates((prev) => ({
            ...prev,
            [productId]: {
              status: "done",
              text: "已抢购",
              class: "bg-gray-400",
            },
          }));
        }
      } else {
        throw new Error(getStatusMsg(result, "秒杀失败"));
      }
    } catch (error) {
      console.error("Seckill action error:", error);
      toast.error(error.message || "秒杀失败");
    } finally {
      setSubmittingId(null);
    }
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold text-gray-900">秒杀活动</h2>
        <div className="bg-red-600 text-white px-4 py-2 rounded-lg">
          距离活动结束: {String(countdown.hours).padStart(2, "0")}:
          {String(countdown.minutes).padStart(2, "0")}:
          {String(countdown.seconds).padStart(2, "0")}
        </div>
      </div>

      {loading ? (
        <div className="flex justify-center py-12">
          <Spinner size="large" />
        </div>
      ) : flashProducts.length === 0 ? (
        <div className="text-center py-12 text-gray-500">暂无秒杀活动</div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {flashProducts.map((product) => {
            const productId = product.id;
            const buttonState = buttonStates[productId] || {
              status: "waiting",
              text: "即将开始",
              class: "bg-gray-400",
            };

            return (
              <div
                key={productId}
                className="bg-white rounded-xl shadow-sm overflow-hidden border-2 border-red-500"
              >
                <div className="relative">
                  <img
                    src={
                      product.image ||
                      "https://via.placeholder.com/300x200?text=Seckill"
                    }
                    alt={product.name}
                    className="w-full h-48 object-cover"
                    onError={(e) => {
                      e.target.src =
                        "https://via.placeholder.com/300x200?text=Seckill";
                    }}
                  />
                  <div className="absolute top-2 right-2 bg-red-600 text-white px-2 py-1 rounded text-sm font-bold">
                    秒杀
                  </div>
                </div>
                <div className="p-4">
                  <h3 className="font-semibold text-gray-900 mb-2">
                    {product.name}
                  </h3>
                  <div className="flex items-center gap-2 mb-4">
                    <span className="text-2xl font-bold text-red-600">
                      ¥{product.price}
                    </span>
                    <span className="text-sm text-gray-400 line-through">
                      ¥{product.original_price || product.originalPrice}
                    </span>
                  </div>
                  <div className="mb-4">
                    <div className="flex justify-between text-sm mb-1">
                      <span className="text-gray-600">已抢</span>
                      <span className="text-red-600">{product.sold || 0}%</span>
                    </div>
                    <div className="h-2 bg-gray-200 rounded-full overflow-hidden">
                      <div
                        className="h-full bg-red-500"
                        style={{ width: `${product.sold || 0}%` }}
                      />
                    </div>
                  </div>
                  <Button
                    className={`w-full text-white ${buttonState.class}`}
                    loading={submittingId === productId}
                    onClick={() => handleSeckill(product)}
                  >
                    {buttonState.text}
                  </Button>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
