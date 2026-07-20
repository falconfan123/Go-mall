import { useEffect, useRef, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import {
  getStatusCode,
  getStatusMsg,
  normalizePaymentStatus,
  paymentApi,
} from "../../services/api";
import { Button } from "../../components/common/Button";
import { Spinner } from "../../components/common/Spinner";

const POLL_INTERVAL_MS = 1500;

const ORDER_STATUS_TEXT = {
  1: "已创建",
  2: "待支付",
  3: "已支付",
  4: "已完成",
  5: "已取消",
  6: "已关闭",
  7: "已退款",
};

const PAYMENT_STATUS_TEXT = {
  1: "待支付",
  2: "支付中",
  3: "已支付",
  4: "已过期",
  5: "已退款",
};

function getOrderStatusText(status) {
  return ORDER_STATUS_TEXT[status] || "处理中";
}

function getPaymentStatusText(status) {
  return PAYMENT_STATUS_TEXT[status] || "处理中";
}

function isPaidState(order) {
  return order?.order_status === 3 && order?.payment_status === 3;
}

function isTerminalFailureState(order) {
  return (
    order?.order_status === 5 ||
    order?.order_status === 6 ||
    order?.order_status === 7 ||
    order?.payment_status === 4 ||
    order?.payment_status === 5
  );
}

function getTerminalMessage(order) {
  if (isPaidState(order)) {
    return "支付成功，订单状态已更新";
  }
  if (order?.order_status === 7 || order?.payment_status === 5) {
    return "该订单已退款";
  }
  if (order?.order_status === 5) {
    return "订单已取消";
  }
  if (order?.order_status === 6) {
    return "订单已关闭";
  }
  if (order?.payment_status === 4) {
    return "支付已过期，订单未完成";
  }
  return "订单状态已更新";
}

function formatCheckedAt(timestamp) {
  if (!timestamp) {
    return "";
  }
  return new Date(timestamp).toLocaleTimeString("zh-CN", {
    hour12: false,
  });
}

export default function PaymentSuccess() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [paymentState, setPaymentState] = useState(null);
  const [initialLoading, setInitialLoading] = useState(true);
  const [isPolling, setIsPolling] = useState(false);
  const [message, setMessage] = useState("支付已完成，正在同步订单状态...");
  const [lastCheckedAt, setLastCheckedAt] = useState(null);

  const orderId = searchParams.get("order_id") || "";
  const paymentId = searchParams.get("payment_id") || "";
  const pollTimerRef = useRef(null);
  const requestInFlightRef = useRef(false);
  const refreshStatusRef = useRef(null);

  // Determine if payment is in processing state
  const isProcessingState = paymentState &&
    !isPaidState(paymentState) &&
    !isTerminalFailureState(paymentState);

  useEffect(() => {
    if (!paymentId) {
      setInitialLoading(false);
      setIsPolling(false);
      setMessage("缺少支付单号，无法确认支付结果");
      return;
    }

    let cancelled = false;

    const clearPollTimer = () => {
      if (pollTimerRef.current) {
        window.clearTimeout(pollTimerRef.current);
        pollTimerRef.current = null;
      }
    };

    const scheduleNextPoll = () => {
      clearPollTimer();
      if (cancelled) {
        return;
      }
      pollTimerRef.current = window.setTimeout(() => {
        refreshStatusRef.current?.();
      }, POLL_INTERVAL_MS);
    };

    const refreshStatus = async () => {
      if (cancelled || requestInFlightRef.current) {
        return;
      }

      clearPollTimer();
      requestInFlightRef.current = true;
      setIsPolling(true);

      let shouldContinuePolling = true;

      try {
        const paymentResp = await paymentApi.status({
          payment_id: paymentId,
          order_id: orderId,
        });
        const paymentData = paymentResp.data || {};
        if (getStatusCode(paymentData) !== 0) {
          throw new Error(getStatusMsg(paymentData, "获取支付状态失败"));
        }
        if (cancelled) {
          return;
        }

        const nextState = normalizePaymentStatus(paymentData);
        setPaymentState(nextState);
        setLastCheckedAt(Date.now());

        const nextOrder = nextState
          ? {
              order_status: nextState.order_status,
              payment_status: nextState.order_payment_status,
            }
          : null;

        if (isPaidState(nextOrder) || isTerminalFailureState(nextOrder)) {
          shouldContinuePolling = false;
          setMessage(getTerminalMessage(nextOrder));
        } else {
          setMessage("支付已完成，正在同步订单状态...");
        }
      } catch (error) {
        if (!cancelled) {
          setLastCheckedAt(Date.now());
          setMessage(
            error?.message === "缺少支付单号，无法确认支付结果"
              ? error.message
              : "订单状态同步异常，正在重试...",
          );
        }
      } finally {
        requestInFlightRef.current = false;
        if (cancelled) {
          return;
        }

        setInitialLoading(false);
        if (shouldContinuePolling) {
          setIsPolling(true);
          scheduleNextPoll();
        } else {
          setIsPolling(false);
        }
      }
    };

    refreshStatusRef.current = refreshStatus;
    refreshStatus();

    return () => {
      cancelled = true;
      requestInFlightRef.current = false;
      refreshStatusRef.current = null;
      clearPollTimer();
    };
  }, [orderId, paymentId]);

  // Handle view order button click
  const handleViewOrder = () => {
    if (orderId) {
      navigate(`/orders?highlight=${orderId}`);
    } else {
      navigate("/orders");
    }
  };

  return (
    <div className="max-w-2xl mx-auto">
      <div className="bg-white rounded-xl shadow-sm p-8">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">
            支付结果
          </h2>

          {initialLoading ? (
            <div className="flex flex-col items-center gap-4 py-8">
              <Spinner size="large" />
              <p className="text-gray-600">{message}</p>
            </div>
          ) : (
            <div className="space-y-4">
              <p className="text-lg font-medium text-gray-900">{message}</p>
              <div className="flex items-center justify-center gap-2 text-sm text-gray-500">
                {isPolling ? (
                  <>
                    <Spinner size="small" />
                    <span>实时同步中</span>
                  </>
                ) : lastCheckedAt ? (
                  <span>最近刷新于 {formatCheckedAt(lastCheckedAt)}</span>
                ) : (
                  <span>等待首次同步结果</span>
                )}
              </div>
              {isProcessingState && (
                <p className="text-sm text-yellow-600">
                  支付已完成，订单同步中。您可以先查看订单页面了解最新状态。
                </p>
              )}
            </div>
          )}
        </div>

        <div className="mt-6 rounded-lg bg-gray-50 p-4 text-left">
          <p className="text-sm text-gray-500">订单号</p>
          <p className="font-medium text-gray-900 break-all">
            {orderId || "未提供"}
          </p>

          <p className="mt-3 text-sm text-gray-500">支付单号</p>
          <p className="font-medium text-gray-900 break-all">
            {paymentId || "未提供"}
          </p>

          <p className="mt-3 text-sm text-gray-500">订单状态</p>
          <p className="font-medium text-gray-900">
            {paymentState
              ? getOrderStatusText(paymentState.order_status)
              : "同步中"}
          </p>

          <p className="mt-3 text-sm text-gray-500">支付状态</p>
          <p className="font-medium text-gray-900">
            {paymentState
              ? getPaymentStatusText(paymentState.order_payment_status)
              : "同步中"}
          </p>
        </div>

        <div className="mt-6 flex justify-center gap-3">
          <Button
            variant="primary"
            onClick={handleViewOrder}
          >
            查看订单
          </Button>
          <Button
            variant="secondary"
            onClick={() => refreshStatusRef.current?.()}
            disabled={!paymentId || initialLoading}
          >
            刷新状态
          </Button>
          <Button variant="outline" onClick={() => navigate("/products")}>
            继续购物
          </Button>
        </div>
      </div>
    </div>
  );
}
