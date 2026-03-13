// API 配置（使用相对路径，通过代理服务器转发）
const API_BASE = {
    user: '/douyin/user',
    product: '/douyin/product',
    cart: '/douyin/cart',
    order: '/douyin/order',
    coupon: '/douyin/coupon',
    checkout: '/douyin/checkout',
    payment: '/douyin/payment',
    flash: '/douyin/flash',
    // 秒杀系统 API
    systemTime: '/api/v1/system/time',
    activityToken: '/api/v1/activity/token',
    seckill: '/api/v1/order/seckill'
};

// 秒杀系统状态管理
const seckillState = {
    timeOffset: 0,           // 服务器时间偏移量
    timeLastSync: 0,         // 上次同步时间
    isSyncing: false,        // 是否正在同步
    pathKeys: {},            // 缓存的 path_key
    buttonStates: {},        // 按钮状态缓存
    lastRequestTime: 0,      // 上次请求时间（防抖）
    isRequesting: false      // 是否正在请求
};

// 状态管理
let state = {
    user: null,
    longToken: null,    // 长令牌，用于身份验证
    shortToken: null,   // 短令牌，用于接口访问
    shortTokenExpire: null, // 短令牌过期时间
    cart: [],
    products: [],
    orders: []
};

// 设备ID管理
const DEVICE_ID_KEY = 'device_id';
let deviceId = null;

// 获取或生成设备ID
function getDeviceId() {
    if (deviceId) return deviceId;

    let storedDeviceId = localStorage.getItem(DEVICE_ID_KEY);
    if (storedDeviceId) {
        deviceId = storedDeviceId;
        return deviceId;
    }

    // 生成UUID
    deviceId = 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        const r = Math.random() * 16 | 0;
        const v = c === 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
    });
    localStorage.setItem(DEVICE_ID_KEY, deviceId);
    return deviceId;
}

// ==================== 秒杀系统函数 ====================

// 获取服务器时间（带偏移量校正）
function getServerTime() {
    return Date.now() + seckillState.timeOffset;
}

// 初始化时钟同步
async function initTimeSync() {
    await syncServerTime();
}

// 同步服务器时间
async function syncServerTime() {
    if (seckillState.isSyncing) return;

    seckillState.isSyncing = true;
    const clientTimeBefore = Date.now();

    try {
        const response = await fetch(API_BASE.systemTime, {
            method: 'GET',
            headers: { 'Content-Type': 'application/json' }
        });

        if (!response.ok) throw new Error('Time sync failed');

        const clientTimeAfter = Date.now();
        const roundTrip = clientTimeAfter - clientTimeBefore;

        const data = await response.json();
        const serverTime = data.time; // 服务器返回的 Unix 毫秒时间

        // 计算偏移量：服务器时间 + 单程延迟 - 客户端当前时间
        const estimatedServerTime = serverTime + (roundTrip / 2);
        seckillState.timeOffset = estimatedServerTime - clientTimeAfter;
        seckillState.timeLastSync = Date.now();

        console.log('[TimeSync] Offset:', seckillState.timeOffset, 'ms');
    } catch (error) {
        console.error('[TimeSync] Error:', error);
    } finally {
        seckillState.isSyncing = false;
    }
}

// 获取秒杀 Token（path_key）
async function getSeckillToken(productId) {
    // 检查缓存
    const cached = seckillState.pathKeys[productId];
    if (cached && cached.expiresAt > getServerTime()) {
        return cached.pathKey;
    }

    try {
        const response = await fetch(API_BASE.activityToken, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                'X-User-Id': state.user?.id?.toString() || getDeviceId()
            }
        });

        if (!response.ok) throw new Error('Token request failed');

        const data = await response.json();
        const pathKey = data.path_key;

        // 缓存 token（假设有效期 5 分钟）
        seckillState.pathKeys[productId] = {
            pathKey: pathKey,
            expiresAt: getServerTime() + 5 * 60 * 1000
        };

        return pathKey;
    } catch (error) {
        console.error('[getSeckillToken] Error:', error);
        return null;
    }
}

// 提交秒杀请求
async function submitSeckillOrder(productId, pathKey) {
    try {
        const response = await fetch(API_BASE.seckill, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-User-Id': state.user?.id?.toString() || getDeviceId()
            },
            body: JSON.stringify({
                product_id: productId,
                path_key: pathKey,
                user_id: state.user?.id || 0
            })
        });

        const data = await response.json();
        return data;
    } catch (error) {
        console.error('[submitSeckillOrder] Error:', error);
        return { status_code: 1, message: '秒杀请求失败' };
    }
}

// 计算按钮状态
function calculateButtonState(productId, activityStartTime) {
    const now = getServerTime();
    const diff = activityStartTime - now;

    // 活动未开始
    if (diff > 10 * 1000) { // 超过10秒
        return { status: 'waiting', text: '即将开始', class: 'btn-disabled' };
    }

    // 即将开始 (10秒内)
    if (diff > 0) {
        return { status: 'critical', text: Math.ceil(diff / 1000) + '秒', class: 'btn-critical' };
    }

    // 可开始秒杀
    return { status: 'trigger', text: '立即抢购', class: 'btn-trigger' };
}

// 更新按钮状态（每秒调用）
function updateButtonStates() {
    const buttons = document.querySelectorAll('.flash-btn[data-product-id]');
    if (buttons.length === 0) return;

    // 假设活动开始时间（可以根据后端配置）
    const activityStartTime = 1739/* TODO: 从后端获取 */;

    buttons.forEach(btn => {
        const productId = parseInt(btn.dataset.productId);
        const state = calculateButtonState(productId, activityStartTime);

        // 更新按钮状态
        if (state.status === 'waiting') {
            btn.disabled = true;
            btn.className = 'flash-btn btn-disabled';
            btn.textContent = state.text;
        } else if (state.status === 'critical') {
            btn.disabled = false;
            btn.className = 'flash-btn btn-critical';
            btn.textContent = state.text;
        } else if (state.status === 'trigger') {
            btn.disabled = false;
            btn.className = 'flash-btn btn-trigger';
            btn.textContent = state.text;
        }
    });
}

// 防抖请求
async function debouncedSeckillRequest(productId) {
    const now = Date.now();

    // 500ms 防抖
    if (now - seckillState.lastRequestTime < 500) {
        showToast('请求过于频繁，请稍后再试', 'error');
        return;
    }

    // 检查是否正在请求
    if (seckillState.isRequesting) {
        showToast('秒杀处理中，请稍候', 'warning');
        return;
    }

    seckillState.isRequesting = true;
    seckillState.lastRequestTime = now;

    // 添加随机延迟 (0-200ms)，打散请求
    await new Promise(resolve => setTimeout(resolve, Math.random() * 200));

    try {
        // 1. 获取 Token
        const pathKey = await getSeckillToken(productId);
        if (!pathKey) {
            showToast('获取秒杀资格失败', 'error');
            return;
        }

        // 2. 提交秒杀请求
        const result = await submitSeckillOrder(productId, pathKey);

        // 3. 处理结果
        if (result.status_code === 0) {
            showToast('秒杀成功！订单号: ' + result.order_id, 'success');
            // 跳转支付页面
            setTimeout(() => {
                showPaymentPage({
                    id: result.order_id,
                    order_no: result.order_id,
                    total: 0.01
                });
            }, 1500);
        } else {
            showToast(result.message || '秒杀失败', 'error');
        }
    } catch (error) {
        console.error('[Seckill] Error:', error);
        showToast('系统错误，请稍后重试', 'error');
    } finally {
        seckillState.isRequesting = false;
    }
}

// 初始化
document.addEventListener('DOMContentLoaded', () => {
    loadFromStorage();
    updateNav();
    showPage('home');
    loadProducts();
    // 启动时钟同步
    initTimeSync();
    // 启动定时同步（每30秒）
    setInterval(syncServerTime, 30000);
    // 启动按钮状态更新（每秒检查）
    setInterval(updateButtonStates, 1000);
});

// 从 localStorage 加载状态
function loadFromStorage() {
    const longToken = localStorage.getItem('long_token');
    const shortToken = localStorage.getItem('short_token');
    const shortTokenExpire = localStorage.getItem('short_token_expire');
    const user = localStorage.getItem('user');
    if (longToken && user) {
        state.longToken = longToken;
        state.shortToken = shortToken;
        state.shortTokenExpire = shortTokenExpire ? parseInt(shortTokenExpire) : null;
        state.user = JSON.parse(user);
    }
}

// 保存到 localStorage
function saveToStorage() {
    if (state.longToken) {
        localStorage.setItem('long_token', state.longToken);
        if (state.shortToken) {
            localStorage.setItem('short_token', state.shortToken);
        }
        if (state.shortTokenExpire) {
            localStorage.setItem('short_token_expire', state.shortTokenExpire.toString());
        }
        localStorage.setItem('user', JSON.stringify(state.user));
    } else {
        localStorage.removeItem('long_token');
        localStorage.removeItem('short_token');
        localStorage.removeItem('short_token_expire');
        localStorage.removeItem('user');
    }
}

// 更新导航栏
function updateNav() {
    const navAuth = document.getElementById('navAuth');
    const navUser = document.getElementById('navUser');
    const userName = document.getElementById('userName');
    const ordersNav = document.getElementById('ordersNav');

    if (state.user && state.longToken) {
        navAuth.style.display = 'none';
        navUser.style.display = 'flex';
        ordersNav.style.display = 'inline';
        userName.textContent = state.user.user_name || state.user.email || '用户';
    } else {
        navAuth.style.display = 'flex';
        navUser.style.display = 'none';
        ordersNav.style.display = 'none';
    }

    updateCartCount();
}

// 更新购物车数量
function updateCartCount() {
    const cartCount = document.getElementById('cartCount');
    const count = state.cart.reduce((sum, item) => sum + item.quantity, 0);
    cartCount.textContent = count;
}

// 显示页面
function showPage(pageName) {
    console.log('showPage function called with pageName:', pageName);
    // 隐藏所有页面
    document.querySelectorAll('.page').forEach(page => {
        page.style.display = 'none';
    });

    // 显示目标页面
    const targetPage = document.getElementById(`page-${pageName}`);
    if (targetPage) {
        targetPage.style.display = 'block';
    }

    // 根据页面加载数据
    if (pageName === 'products') {
        loadProducts();
    } else if (pageName === 'cart') {
        loadCart();
    } else if (pageName === 'orders') {
        loadOrders();
    } else if (pageName === 'flash') {
        loadFlashProducts();
        startCountdown();
    }
}

// 秒杀倒计时
let countdownInterval;

function startCountdown() {
    // 设置倒计时结束时间（2小时后）
    const endTime = Date.now() + 2 * 60 * 60 * 1000;

    if (countdownInterval) {
        clearInterval(countdownInterval);
    }

    countdownInterval = setInterval(() => {
        const remaining = endTime - Date.now();
        if (remaining <= 0) {
            clearInterval(countdownInterval);
            document.getElementById('countdownTime').textContent = '00:00:00';
            return;
        }

        const hours = Math.floor(remaining / (1000 * 60 * 60));
        const minutes = Math.floor((remaining % (1000 * 60 * 60)) / (1000 * 60));
        const seconds = Math.floor((remaining % (1000 * 60)) / 1000);

        const timeStr = `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
        document.getElementById('countdownTime').textContent = timeStr;
    }, 1000);
}

// 加载秒杀商品
async function loadFlashProducts() {
    const flashProductList = document.getElementById('flashProductList');

    // 先显示一些演示数据
    const demoFlashProducts = [
        { id: 1, name: 'iPhone 15 Pro', description: '最新款苹果手机，搭载 A17 芯片', price: 7999, flashPrice: 5999, stock: 50, picture: '📱' },
        { id: 2, name: 'MacBook Pro 14"', description: '高性能笔记本电脑，M3 Pro 芯片', price: 14999, flashPrice: 11999, stock: 30, picture: '💻' },
        { id: 3, name: 'Sony PS5', description: '次世代游戏主机', price: 3899, flashPrice: 2999, stock: 20, picture: '🎮' },
        { id: 4, name: 'AirPods Pro 2', description: '最新款降噪耳机', price: 1899, flashPrice: 1499, stock: 100, picture: '🎧' }
    ];

    renderFlashProducts(demoFlashProducts);
}

// 渲染秒杀商品列表
function renderFlashProducts(products) {
    const flashProductList = document.getElementById('flashProductList');
    flashProductList.innerHTML = products.map(product => `
        <div class="flash-product-card">
            <div class="flash-product-image">
                <span class="flash-badge">秒杀</span>
                ${product.picture || '📦'}
            </div>
            <div class="flash-product-info">
                <h3>${product.name}</h3>
                <p class="description">${product.description || ''}</p>
                <div class="flash-price-container">
                    <span class="flash-original-price">¥${product.price}</span>
                    <span class="flash-price">¥${product.flashPrice}</span>
                </div>
                <div class="flash-stock">
                    库存: <span class="stock-number">${product.stock}</span> 件
                </div>
                <div class="flash-action">
                    <button class="flash-btn" data-product-id="${product.id}" onclick="handleFlashBuy(${product.id})" ${product.stock <= 0 ? 'disabled' : ''}>
                        ${product.stock > 0 ? '即将开始' : '已售罄'}
                    </button>
                </div>
            </div>
        </div>
    `).join('');
}

// 处理秒杀购买
async function handleFlashBuy(productId) {
    if (!state.longToken) {
        showToast('请先登录', 'error');
        showPage('login');
        return;
    }

    // 检查时钟同步状态
    if (Math.abs(seckillState.timeOffset) > 5000) {
        showToast('时钟未同步，请稍后重试', 'warning');
        await syncServerTime();
        return;
    }

    // 获取商品信息
    const products = [
        { id: 1, name: 'iPhone 15 Pro', price: 7999, flashPrice: 5999, stock: 50 },
        { id: 2, name: 'MacBook Pro 14"', price: 14999, flashPrice: 11999, stock: 30 },
        { id: 3, name: 'Sony PS5', price: 3899, flashPrice: 2999, stock: 20 },
        { id: 4, name: 'AirPods Pro 2', price: 1899, flashPrice: 1499, stock: 100 }
    ];

    const product = products.find(p => p.id === productId);
    if (!product) {
        showToast('商品不存在', 'error');
        return;
    }

    // 使用新的秒杀流程（防抖 + Token + 随机延迟）
    await debouncedSeckillRequest(productId);
}

// 显示支付页面
function showPaymentPage(order) {
    const paymentContainer = document.getElementById('paymentContainer');

    // 渲染商品信息（处理单个商品或多个商品的情况）
    let productDetails = '';
    if (order.product) {
        // 单个商品（秒杀订单）
        productDetails = `
            <div class="detail-row">
                <span class="detail-label">商品</span>
                <span class="detail-value">${order.product.name}</span>
            </div>
            <div class="detail-row">
                <span class="detail-label">单价</span>
                <span class="detail-value">¥${order.product.flashPrice || order.product.price}</span>
            </div>
            <div class="detail-row">
                <span class="detail-label">数量</span>
                <span class="detail-value">${order.quantity}</span>
            </div>
        `;
    } else if (order.items && order.items.length > 0) {
        // 多个商品（购物车订单）
        productDetails = order.items.map(item => `
            <div class="detail-row">
                <span class="detail-label">${item.product?.name || '商品'}</span>
                <span class="detail-value">${item.quantity} × ¥${item.product?.price || 0}</span>
            </div>
        `).join('');
    }

    paymentContainer.innerHTML = `
        <div class="payment-info">
            <h3>订单信息</h3>
            <div class="payment-details">
                <div class="detail-row">
                    <span class="detail-label">订单号</span>
                    <span class="detail-value">${order.order_no}</span>
                </div>
                ${productDetails}
            </div>
        </div>

        <div class="payment-methods">
            <h3>支付方式</h3>
            <div class="payment-method-list">
                <div class="payment-method active">
                    <input type="radio" name="paymentMethod" value="alipay" checked>
                    <span>支付宝</span>
                </div>
                <div class="payment-method">
                    <input type="radio" name="paymentMethod" value="wechat">
                    <span>微信支付</span>
                </div>
                <div class="payment-method">
                    <input type="radio" name="paymentMethod" value="credit">
                    <span>信用卡</span>
                </div>
            </div>
        </div>

        <div class="payment-summary">
            <div class="total-row">
                <span>应付总额</span>
                <span>¥${order.total}</span>
            </div>
        </div>

        <div class="payment-actions">
            <button class="btn btn-secondary" onclick="showPage('orders')">取消支付</button>
            <button class="btn btn-primary" onclick="handlePayment('${order.order_no}')">立即支付</button>
        </div>
    `;

    showPage('payment');
}

// 处理支付
async function handlePayment(orderNo) {
    try {
        // 模拟支付过程
        showToast('正在处理支付...', 'info');
        await new Promise(resolve => setTimeout(resolve, 1500));

        // 调用后端支付接口
        await apiRequest(`${API_BASE.payment}/create`, {
            method: 'POST',
            body: JSON.stringify({
                order_id: orderNo,
                payment_method: 1 // 1-WECHAT_PAY 2-ALIPAY
            })
        });

        // 更新订单状态
        const order = state.orders.find(o => o.order_no === orderNo);
        if (order) {
            order.status = 'paid';
        }

        showToast('支付成功！', 'success');

        // 跳转订单页面
        setTimeout(() => {
            showPage('orders');
        }, 1500);

    } catch (error) {
        showToast(error.message || '支付失败，请重试', 'error');
    }
}

// 显示消息提示
function showToast(message, type = 'info') {
    const toast = document.getElementById('toast');
    toast.textContent = message;
    toast.className = `toast show ${type}`;

    setTimeout(() => {
        toast.className = 'toast';
    }, 30000);
}

// Mock 数据配置
const MOCK_ENABLED = false; // 关闭mock模式
const MOCK_DELAY = 500; // mock请求延迟毫秒

// Mock 响应数据
const mockResponses = {
    // 用户注册
    'POST:/douyin/user/register': (body) => {
        // 生成mock的长短令牌
        const deviceId = body.device_id || 'unknown_device';
        const timestamp = Date.now();
        return {
            code: 0,
            msg: '注册成功',
            data: {
                // 长令牌：有效期7天，用于身份验证和刷新短令牌
                long_token: 'mock_long_token_' + deviceId + '_' + timestamp,
                // 短令牌：有效期2小时，用于接口访问
                short_token: 'mock_short_token_' + deviceId + '_' + timestamp,
                // 短令牌过期时间（时间戳）
                short_token_expire: timestamp + 2 * 60 * 60 * 1000,
                user_id: 1,
                email: body.email
            }
        };
    },
    // 用户登录
    'POST:/douyin/user/login': (body) => {
        // 生成mock的长短令牌
        const deviceId = body.device_id || 'unknown_device';
        const timestamp = Date.now();
        return {
            code: 0,
            msg: '登录成功',
            data: {
                // 长令牌：有效期7天，用于身份验证和刷新短令牌
                long_token: 'mock_long_token_' + deviceId + '_' + timestamp,
                // 短令牌：有效期2小时，用于接口访问
                short_token: 'mock_short_token_' + deviceId + '_' + timestamp,
                // 短令牌过期时间（时间戳）
                short_token_expire: timestamp + 2 * 60 * 60 * 1000,
                user_id: 1,
                email: body.email
            }
        };
    },
    // 获取用户信息
    'GET:/douyin/user/info': () => {
        return {
            code: 0,
            msg: 'success',
            data: {
                user_id: 1,
                email: 'test@example.com',
                user_name: '测试用户',
                avatar: 'https://example.com/avatar.jpg',
                phone: '13800138000'
            }
        };
    },
    // 更新用户信息
    'PUT:/douyin/user/update': (body) => {
        return {
            code: 0,
            msg: '更新成功',
            data: {
                user_id: 1,
                email: 'test@example.com',
                user_name: body.user_name || '测试用户',
                avatar: body.avatar || 'https://example.com/avatar.jpg'
            }
        };
    },
    // 添加用户地址
    'POST:/douyin/user/address': (body) => {
        return {
            code: 0,
            msg: '添加成功',
            data: {
                id: 1,
                ...body,
                user_id: 1
            }
        };
    },
    // 获取地址列表
    'GET:/douyin/user/address/list': () => {
        return {
            code: 0,
            msg: 'success',
            data: [
                {
                    id: 1,
                    recipient_name: '张三',
                    phone_number: '13800138000',
                    province: '北京市',
                    city: '北京市',
                    detailed_address: '朝阳区某某街道123号',
                    is_default: true
                }
            ]
        };
    },
    // 用户登出
    'POST:/douyin/user/logout': (body) => {
        return {
            code: 0,
            msg: '登出成功',
            data: {
                logout_at: Date.now(),
                device_id: body.device_id
            }
        };
    },
    // 刷新短令牌
    'POST:/douyin/user/refresh': (body) => {
        const deviceId = body.device_id || 'unknown_device';
        const timestamp = Date.now();
        return {
            code: 0,
            msg: '令牌刷新成功',
            data: {
                // 新的短令牌
                short_token: 'mock_short_token_refreshed_' + deviceId + '_' + timestamp,
                // 短令牌过期时间（2小时后）
                short_token_expire: timestamp + 2 * 60 * 60 * 1000
            }
        };
    },
    // 获取商品列表
    'GET:/douyin/product/list': () => {
        return {
            code: 0,
            msg: 'success',
            products: [
                { id: 1, name: 'iPhone 15 Pro', description: '最新款苹果手机，搭载 A17 芯片', price: 7999, stock: 100, picture: '📱', thumbnailUrl: '📱' },
                { id: 2, name: 'MacBook Pro 14"', description: '高性能笔记本电脑，M3 Pro 芯片', price: 14999, stock: 50, picture: '💻', thumbnailUrl: '💻' },
                { id: 3, name: 'Sony PS5', description: '次世代游戏主机', price: 3899, stock: 200, picture: '🎮', thumbnailUrl: '🎮' },
                { id: 4, name: 'AirPods Pro 2', description: '最新款降噪耳机', price: 1899, stock: 150, picture: '🎧', thumbnailUrl: '🎧' },
                { id: 5, name: 'iPad Pro', description: '12.9英寸平板电脑，M2芯片', price: 8999, stock: 75, picture: '📱', thumbnailUrl: '📱' },
                { id: 6, name: 'Apple Watch', description: '智能手表，健康监测', price: 2999, stock: 120, picture: '⌚', thumbnailUrl: '⌚' }
            ]
        };
    },
    // 获取商品详情
    'GET:/douyin/product/': (_, params) => {
        const id = parseInt(params.get('id'));
        const products = [
            { id: 1, name: 'iPhone 15 Pro', description: '最新款苹果手机，搭载 A17 芯片', price: 7999, stock: 100, picture: '📱' },
            { id: 2, name: 'MacBook Pro 14"', description: '高性能笔记本电脑，M3 Pro 芯片', price: 14999, stock: 50, picture: '💻' },
            { id: 3, name: 'Sony PS5', description: '次世代游戏主机', price: 3899, stock: 200, picture: '🎮' }
        ];
        return {
            code: 0,
            msg: 'success',
            data: products.find(p => p.id === id) || products[0]
        };
    },
    // 添加商品到购物车
    'POST:/douyin/cart/add': (body) => {
        // 检查商品是否已在购物车中
        const existingItem = state.cart.find(item => item.product_id === body.product_id);
        if (existingItem) {
            existingItem.quantity += 1;
        } else {
            state.cart.push({
                id: Date.now(),
                product_id: body.product_id,
                quantity: 1
            });
        }
        return {
            code: 0,
            msg: '添加成功',
            data: {
                id: Date.now(),
                product_id: body.product_id,
                quantity: existingItem ? existingItem.quantity : 1
            }
        };
    },
    // 获取购物车列表
    'GET:/douyin/cart/list': () => {
        // 从本地状态获取购物车数据
        return {
            code: 0,
            msg: 'success',
            data: state.cart.map(item => ({
                id: item.id,
                product_id: item.product_id,
                quantity: item.quantity
            }))
        };
    },
    // 减少购物车商品数量
    'POST:/douyin/cart/sub': (body) => {
        const existingItem = state.cart.find(item => item.product_id === body.product_id);
        if (existingItem) {
            existingItem.quantity -= body.quantity || 1;
            if (existingItem.quantity <= 0) {
                state.cart = state.cart.filter(item => item.product_id !== body.product_id);
            }
        }
        return {
            code: 0,
            msg: '更新成功',
            data: {
                product_id: body.product_id,
                quantity: existingItem ? existingItem.quantity : 0
            }
        };
    },
    // 删除购物车商品
    'POST:/douyin/cart/delete': (body) => {
        state.cart = state.cart.filter(item => item.product_id !== body.product_id);
        return {
            code: 0,
            msg: '删除成功',
            data: {
                success: true
            }
        };
    },
    // 获取优惠券列表
    'GET:/douyin/coupon/list': () => {
        return {
            code: 0,
            msg: 'success',
            data: [
                { id: 'coupon_001', name: '满100减20优惠券', discount: 20, min_amount: 100, expire_time: '2026-12-31' },
                { id: 'coupon_002', name: '满200减50优惠券', discount: 50, min_amount: 200, expire_time: '2026-12-31' }
            ]
        };
    },
    // 领取优惠券
    'POST:/douyin/coupon/claim': (body) => {
        return {
            code: 0,
            msg: '领取成功',
            data: {
                coupon_id: body.coupon_id,
                receive_time: Date.now()
            }
        };
    },
    // 获取我的优惠券列表
    'GET:/douyin/coupon/my/list': () => {
        return {
            code: 0,
            msg: 'success',
            data: [
                { id: 'coupon_001', name: '满100减20优惠券', discount: 20, min_amount: 100, expire_time: '2026-12-31', used: false }
            ]
        };
    },
    // 预结算
    'POST:/douyin/checkout/prepare': (body) => {
        const total = body.order_items.reduce((sum, item) => {
            const product = state.products.find(p => p.id === item.product_id);
            return sum + (product ? product.price * item.quantity : 0);
        }, 0);
        return {
            code: 0,
            msg: 'success',
            data: {
                pre_order_id: 'pre_' + Date.now(),
                total_amount: total,
                discount_amount: body.coupon_id ? 20 : 0,
                payable_amount: total - (body.coupon_id ? 20 : 0)
            }
        };
    },
    // 获取结算详情
    'GET:/douyin/checkout/detail': (_, params) => {
        return {
            code: 0,
            msg: 'success',
            data: {
                pre_order_id: params.get('pre_order_id'),
                order_items: state.cart,
                total_amount: 7999,
                discount_amount: 0,
                payable_amount: 7999,
                address: {
                    id: 1,
                    recipient_name: '张三',
                    phone_number: '13800138000',
                    province: '北京市',
                    city: '北京市',
                    detailed_address: '朝阳区某某街道123号'
                }
            }
        };
    },
    // 创建订单
    'POST:/douyin/order/create': (body) => {
        return {
            code: 0,
            msg: '订单创建成功',
            data: {
                order_id: 'order_' + Date.now(),
                pre_order_id: body.pre_order_id,
                order_status: 0,
                created_at: new Date().toISOString()
            }
        };
    },
    // 获取订单列表
    'GET:/douyin/order/list': () => {
        return {
            code: 0,
            msg: 'success',
            data: {
                orders: state.orders.map(order => ({
                    order_id: order.id,
                    order_no: order.order_no,
                    items: order.items.map(item => ({
                        product_id: item.product_id,
                        product_name: item.product?.name || '商品',
                        unit_price: item.product?.price || 0,
                        quantity: item.quantity
                    })),
                    payable_amount: order.total,
                    order_status: order.status === 'pending' ? 0 : 1,
                    created_at: order.created_at
                }))
            }
        };
    },
    // 获取订单详情
    'GET:/douyin/order/detail': (_, params) => {
        const order = state.orders.find(o => o.id === params.get('order_id'));
        return {
            code: 0,
            msg: 'success',
            data: order ? {
                order_id: order.id,
                order_no: order.order_no,
                items: order.items.map(item => ({
                    product_id: item.product_id,
                    product_name: item.product?.name || '商品',
                    unit_price: item.product?.price || 0,
                    quantity: item.quantity
                })),
                payable_amount: order.total,
                order_status: order.status === 'pending' ? 0 : 1,
                created_at: order.created_at,
                address: {
                    recipient_name: '张三',
                    phone_number: '13800138000',
                    province: '北京市',
                    city: '北京市',
                    detailed_address: '朝阳区某某街道123号'
                }
            } : null
        };
    },
    // 取消订单
    'POST:/douyin/order/cancel': (body) => {
        return {
            code: 0,
            msg: '订单取消成功',
            data: {
                order_id: body.order_id,
                order_status: 4
            }
        };
    },
    // 创建支付订单
    'POST:/douyin/payment/create': (body) => {
        return {
            code: 0,
            msg: '支付订单创建成功',
            data: {
                payment_id: 'pay_' + Date.now(),
                order_id: body.order_id,
                amount: 7999,
                payment_url: 'https://example.com/pay/' + Date.now()
            }
        };
    },
    // 获取支付列表
    'GET:/douyin/payment/list': () => {
        return {
            code: 0,
            msg: 'success',
            data: [
                {
                    id: 'pay_123',
                    order_id: 'order_123',
                    amount: 7999,
                    status: 'success',
                    created_at: new Date().toISOString()
                }
            ]
        };
    },
    // 获取秒杀商品列表
    'GET:/douyin/flash/products': () => {
        return {
            code: 0,
            msg: 'success',
            data: [
                { id: 1, name: 'iPhone 15 Pro', description: '最新款苹果手机，搭载 A17 芯片', price: 7999, flashPrice: 5999, stock: 50, picture: '📱' },
                { id: 2, name: 'MacBook Pro 14"', description: '高性能笔记本电脑，M3 Pro 芯片', price: 14999, flashPrice: 11999, stock: 30, picture: '💻' }
            ]
        };
    },
    // 秒杀商品
    'POST:/douyin/flash/buy': (body) => {
        // 模拟10%的秒杀失败概率
        if (Math.random() < 0.1) {
            return {
                code: 1001,
                msg: '秒杀失败，商品已售罄'
            };
        }
        return {
            code: 0,
            msg: '秒杀成功',
            data: {
                orderId: 'order_flash_' + Date.now(),
                orderNo: 'FLASH' + Date.now(),
                total: 599900 // 单位分
            }
        };
    }
};

// API 请求工具
async function apiRequest(url, options = {}) {
    // Mock模式下拦截请求
    if (MOCK_ENABLED) {
        await new Promise(resolve => setTimeout(resolve, MOCK_DELAY));

        const method = options.method || 'GET';
        const urlObj = new URL(url, window.location.origin);
        const path = urlObj.pathname;
        const params = urlObj.searchParams;

        // 匹配mock响应
        const mockKey = `${method}:${path}`;
        let mockHandler = mockResponses[mockKey];

        // 处理带参数的路径，如 /douyin/product/
        if (!mockHandler) {
            for (const key of Object.keys(mockResponses)) {
                const [m, p] = key.split(':');
                if (m === method && path.startsWith(p)) {
                    mockHandler = mockResponses[key];
                    break;
                }
            }
        }

        if (mockHandler) {
            let body = null;
            if (options.body) {
                body = JSON.parse(options.body);
            }
            const mockResponse = mockHandler(body, params);
            console.log(`[Mock] ${method} ${url}`, mockResponse);
            return mockResponse;
        }

        throw new Error(`Mock 响应未找到: ${method} ${url}`);
    }

    const headers = {
        'Content-Type': 'application/json',
        ...options.headers
    };

    // 携带长短令牌（除登录、注册接口外）
    const isAuthEndpoint = url.includes('/login') || url.includes('/register');
    if (!isAuthEndpoint && state.longToken) {
        headers['Long-Token'] = state.longToken;
    }
    if (!isAuthEndpoint && state.shortToken) {
        headers['Short-Token'] = state.shortToken;
    }

    try {
        const response = await fetch(url, {
            ...options,
            headers,
            mode: 'cors',
            credentials: 'same-origin'
        });

        // Handle 404 (Gateway Not Found) or 500
        if (!response.ok) {
             throw new Error(`HTTP Error: ${response.status} ${response.statusText}`);
        }

        const data = await response.json();

        // 检查响应头中的 Short-Token-Refresh 字段，如果存在说明短令牌已更新
        const newShortToken = response.headers.get('Short-Token-Refresh');
        if (newShortToken) {
            console.log('Short token refreshed automatically');
            state.shortToken = newShortToken;
            // 从响应中获取新的过期时间
            const newExpire = response.headers.get('Short-Token-Expire');
            if (newExpire) {
                state.shortTokenExpire = parseInt(newExpire);
            }
            saveToStorage();
        }

        // 兼容 Gateway (statusCode) 和 API (code)
        const code = data.code !== undefined ? data.code : data.statusCode;
        const msg = data.msg || data.statusMsg || '请求失败';

        // 短令牌过期，自动重试（服务端返回了新的短令牌）
        if (code === 10004) {
            console.log('Token renewed, retrying request...');
            // 重试原请求
            return apiRequest(url, options);
        }

        if (code !== 0 && code !== undefined) {
            // 令牌非法或Session不存在，清除状态并跳转登录页
            if (code === 10001 || code === 10003) {
                logout();
            }
            throw new Error(msg);
        }

        return data;
    } catch (error) {
        console.error('API Error:', error);
        // 处理CORS错误
        if (error.message.includes('Failed to fetch') || error.message.includes('CORS')) {
            console.error('CORS 错误提示：请确保后端网关已配置允许跨域访问，或使用代理访问');
            throw new Error('网络请求失败，请检查后端服务是否正常运行或联系管理员配置跨域');
        }
        throw error;
    }
}

// 用户注册
async function handleRegister(event) {
    event.preventDefault();

    const username = document.getElementById('registerEmail').value;
    const email = username; // 使用用户名作为邮箱（兼容后端）
    const password = document.getElementById('registerPassword').value;
    const confirmPassword = document.getElementById('registerConfirmPassword').value;

    if (password !== confirmPassword) {
        showToast('两次密码输入不一致', 'error');
        return;
    }

    try {
        const data = await apiRequest(`${API_BASE.user}/register`, {
            method: 'POST',
            body: JSON.stringify({ username, email, password, confirm_password: confirmPassword, device_id: getDeviceId() })
        });

        // 用户信息在 handleLogin 中已处理
        saveToStorage();
        updateNav();
        showToast('注册成功！', 'success');
        showPage('home');
    } catch (error) {
        showToast(error.message || '注册失败', 'error');
    }
}

// 用户登录
async function handleLogin(event) {
    event.preventDefault();

    const username = document.getElementById('loginEmail').value;
    const password = document.getElementById('loginPassword').value;

    // 获取设备ID
    const deviceId = getDeviceId();

    try {
        const data = await apiRequest(`${API_BASE.user}/login`, {
            method: 'POST',
            body: JSON.stringify({ username, password, device_id: deviceId })
        });

        // 兼容 Gateway 和 API，提取长短令牌
        // 长令牌 (long_token): 有效期较长，用于身份验证和刷新短令牌
        // 短令牌 (short_token): 有效期较短，用于接口访问
        state.longToken = data.longToken || data.long_token ||
            data.data?.longToken || data.data?.long_token ||
            data.accessToken || data.access_token || data.data?.accessToken || data.data?.access_token;
        state.shortToken = data.shortToken || data.short_token ||
            data.data?.shortToken || data.data?.short_token ||
            data.refreshToken || data.refresh_token || data.data?.refreshToken || data.data?.refresh_token;
        // 从响应中获取短令牌过期时间
        state.shortTokenExpire = data.shortTokenExpire || data.short_token_expire ||
            data.data?.shortTokenExpire || data.data?.short_token_expire ||
            Date.now() + 2 * 60 * 60 * 1000; // 默认2小时

        state.user = { username, user_id: data.userId || data.user_id || data.data?.userId || data.data?.user_id };
        saveToStorage();
        updateNav();
        showToast('登录成功！', 'success');
        showPage('home');
    } catch (error) {
        showToast(error.message || '登录失败', 'error');
    }
}

// 用户登出
async function logout() {
    // 调用登出接口（如果已登录）
    if (state.longToken) {
        try {
            await apiRequest(`${API_BASE.user}/logout`, {
                method: 'POST',
                body: JSON.stringify({ device_id: getDeviceId() })
            });
        } catch (error) {
            // 忽略登出接口错误
            console.log('Logout API error (ignored):', error.message);
        }
    }

    state.user = null;
    state.longToken = null;
    state.shortToken = null;
    state.shortTokenExpire = null;
    state.cart = [];
    saveToStorage();
    updateNav();
    showToast('已退出登录', 'info');
    showPage('home');
}

// 加载商品列表
async function loadProducts() {
    const productList = document.getElementById('productList');

    try {
        const data = await apiRequest(`${API_BASE.product}/list?page=1&pageSize=100`, {
             method: 'GET'
        });
        console.log('Product API Response:', data);
        
        if (data.products && data.products.length > 0) {
             state.products = data.products.map(p => ({
                 id: p.id,
                 name: p.name || '商品', // Fallback if name is empty
                 description: p.description,
                 price: parseFloat(p.price),
                 stock: parseInt(p.stock),
                 picture: p.thumbnailUrl || p.picture || '📦'
             }));
             renderProducts(state.products);
             return;
        }
        
        productList.innerHTML = '<div class="empty-state">暂无商品</div>';
    } catch (error) {
        console.error("Failed to load products:", error);
        showToast(`加载商品失败: ${error.message}`, 'error');
        productList.innerHTML = '<div class="error-state">加载失败，请刷新重试</div>';
    }
}

// 渲染商品列表
function renderProducts(products) {
    const productList = document.getElementById('productList');
    productList.innerHTML = products.map(product => `
        <div class="product-card" onclick="showProductDetail(${product.id})">
            <div class="product-image">${product.picture || '📦'}</div>
            <div class="product-info">
                <h3>${product.name}</h3>
                <p class="description">${product.description || ''}</p>
                <p class="price">¥${product.price}</p>
                <div class="meta">
                    <span>库存: ${product.stock || 0}</span>
                    <button class="btn btn-primary btn-small" onclick="event.stopPropagation(); addToCart(${product.id})">加入购物车</button>
                </div>
            </div>
        </div>
    `).join('');
}

// 筛选商品
function filterProducts() {
    const searchTerm = document.getElementById('searchInput').value.toLowerCase();
    const filtered = state.products.filter(p =>
        p.name.toLowerCase().includes(searchTerm) ||
        (p.description && p.description.toLowerCase().includes(searchTerm))
    );
    renderProducts(filtered);
}

// 显示商品详情
function showProductDetail(productId) {
    const product = state.products.find(p => p.id === productId);
    if (!product) return;

    const productDetail = document.getElementById('productDetail');
    productDetail.innerHTML = `
        <div class="product-detail-image">${product.picture || '📦'}</div>
        <div class="product-detail-info">
            <h2>${product.name}</h2>
            <p class="price">¥${product.price}</p>
            <p class="description">${product.description || '暂无描述'}</p>
            <div class="stock">
                <strong>库存:</strong> ${product.stock || 0} 件
            </div>
            <div class="product-actions">
                <button class="btn btn-primary btn-large" onclick="addToCart(${product.id})">加入购物车</button>
                <button class="btn btn-secondary btn-large" onclick="showPage('products')">返回列表</button>
            </div>
        </div>
    `;
    showPage('product-detail');
}

// 添加到购物车
async function addToCart(productId) {
    if (!state.longToken) {
        showToast('请先登录', 'error');
        showPage('login');
        return;
    }

    try {
        await apiRequest(`${API_BASE.cart}/add`, {
            method: 'POST',
            body: JSON.stringify({
                product_id: productId,
                quantity: 1
            })
        });

        updateCartCount();
        showToast('已添加到购物车', 'success');
        // Refresh cart if we are on cart page, but usually we are on product list
        if (document.getElementById('page-cart').style.display === 'block') {
            loadCart();
        } else {
            // Optimistically update count or just fetch count?
            // For now, let's just reload cart to be safe and update count
            loadCart();
        }
    } catch (error) {
        showToast(error.message || '添加失败', 'error');
    }
}

// 加载购物车
async function loadCart() {
    const cartList = document.getElementById('cartList');
    const cartSummary = document.getElementById('cartSummary');

    if (!state.longToken) {
         state.cart = [];
         renderCart();
         return;
    }

    try {
        const data = await apiRequest(`${API_BASE.cart}/list`, {
            method: 'GET'
        });

        if (data.data) {
            state.cart = data.data.map(item => {
                const product = state.products.find(p => p.id === item.product_id);
                return {
                    id: item.id,
                    product_id: item.product_id,
                    quantity: item.quantity,
                    product: product || { name: '加载中...', price: 0, picture: '📦' }
                };
            });
        } else {
            state.cart = [];
        }
    } catch (error) {
        console.error('加载购物车失败:', error);
        // Don't clear cart if error, maybe just show toast
        showToast('加载购物车失败', 'error');
    }

    renderCart();
}

function renderCart() {
    const cartList = document.getElementById('cartList');
    const cartSummary = document.getElementById('cartSummary');

    if (state.cart.length === 0) {
        cartList.innerHTML = '<div class="cart-empty">购物车是空的，快去选购商品吧！</div>';
        cartSummary.style.display = 'none';
        return;
    }

    cartList.innerHTML = state.cart.map(item => `
        <div class="cart-item">
            <div class="cart-item-image">${item.product?.picture || '📦'}</div>
            <div class="cart-item-info">
                <h3>${item.product?.name || '商品'}</h3>
                <p class="price">¥${item.product?.price || 0}</p>
            </div>
            <div class="cart-item-quantity">
                <button class="quantity-btn" onclick="updateCartItem(${item.product_id}, -1)">-</button>
                <span class="quantity-display">${item.quantity}</span>
                <button class="quantity-btn" onclick="updateCartItem(${item.product_id}, 1)">+</button>
            </div>
            <button class="btn btn-danger btn-small" onclick="removeFromCart(${item.product_id})">删除</button>
        </div>
    `).join('');

    const total = state.cart.reduce((sum, item) => sum + (item.product?.price || 0) * item.quantity, 0);
    document.getElementById('cartTotal').textContent = total;
    cartSummary.style.display = 'block';
    
    updateCartCount();
}

// 更新购物车商品数量
async function updateCartItem(productId, delta) {
    try {
        if (delta > 0) {
            await apiRequest(`${API_BASE.cart}/add`, {
                method: 'POST',
                body: JSON.stringify({
                    product_id: productId,
                    quantity: 1
                })
            });
        } else {
            // Check if quantity is 1, if so, maybe ask to delete? 
            // Current backend SubCartItem returns error if quantity <= 1.
            // So we should check local state first.
            const item = state.cart.find(i => i.product_id === productId);
            if (item && item.quantity <= 1) {
                if (confirm('确定要删除该商品吗？')) {
                    await removeFromCart(productId);
                }
                return;
            }

            await apiRequest(`${API_BASE.cart}/sub`, {
                method: 'POST',
                body: JSON.stringify({
                    product_id: productId,
                    quantity: 1
                })
            });
        }
        loadCart();
    } catch (error) {
        showToast(error.message || '更新失败', 'error');
    }
}

// 从购物车删除
async function removeFromCart(productId) {
    try {
        await apiRequest(`${API_BASE.cart}/delete`, {
            method: 'POST',
            body: JSON.stringify({
                product_id: productId
            })
        });
        showToast('已从购物车删除', 'success');
        loadCart();
    } catch (error) {
        showToast(error.message || '删除失败', 'error');
    }
}

// 结算
async function checkout() {
    if (!state.longToken) {
        showToast('请先登录', 'error');
        showPage('login');
        return;
    }

    if (state.cart.length === 0) {
        showToast('购物车是空的', 'error');
        return;
    }

    try {
        // 调用后端结算 API
        const orderItems = state.cart.map(item => ({
            product_id: item.product_id,
            quantity: item.quantity
        }));

        const data = await apiRequest(`${API_BASE.checkout}/prepare`, {
            method: 'POST',
            body: JSON.stringify({
                coupon_id: '',
                order_items: orderItems,
                address_id: 1 // 暂时使用默认地址
            })
        });

        // 调用创建订单 API
        const createData = await apiRequest(`${API_BASE.order}/create`, {
            method: 'POST',
            body: JSON.stringify({
                pre_order_id: data.data.pre_order_id,
                coupon_id: '',
                address_id: 1,
                payment_method: 1 // 1-WECHAT_PAY 2-ALIPAY
            })
        });

        // 创建订单对象
        const order = {
            id: createData.data.order_id,
            order_no: createData.data.order_id,
            items: [...state.cart],
            total: state.cart.reduce((sum, item) => sum + (item.product?.price || 0) * item.quantity, 0),
            status: 'pending',
            created_at: new Date().toISOString()
        };

        state.orders.unshift(order);
        state.cart = [];
        updateCartCount();
        showToast('订单创建成功！正在跳转支付页面...', 'success');

        // 跳转支付页面
        setTimeout(() => {
            showPaymentPage(order);
        }, 1500);
    } catch (error) {
        showToast(error.message || '结算失败，请重试', 'error');
        console.error('Checkout Error:', error);
    }
}

// MinIO Upload Function
async function uploadFile(file) {
    if (!state.longToken) {
        showToast('Please login first', 'error');
        return;
    }

    try {
        // 1. Get Presigned URL
        const presignRes = await apiRequest(`${API_BASE.product}/upload`, {
            method: 'POST',
            body: JSON.stringify({
                filename: file.name,
                contentType: file.type
            })
        });

        if (presignRes.statusCode !== 0) {
            throw new Error(presignRes.statusMsg || 'Get upload url failed');
        }

        const { uploadUrl, formData, key } = presignRes;

        // 2. Construct FormData
        const data = new FormData();
        for (const k in formData) {
            data.append(k, formData[k]);
        }
        data.append('file', file);

        // 3. Upload to MinIO
        const uploadRes = await fetch(uploadUrl, {
            method: 'POST',
            body: data
        });

        if (!uploadRes.ok) {
            throw new Error('Upload to MinIO failed');
        }

        showToast('Upload success!', 'success');
        return key; // Return the key (path) for saving to DB
    } catch (error) {
        console.error('Upload error:', error);
        showToast(error.message, 'error');
        throw error;
    }
}

// 加载订单
async function loadOrders() {
    console.log('loadOrders function called');
    const orderList = document.getElementById('orderList');

    // 先尝试从后端获取真实的订单数据
    if (state.longToken) {
        try {
            const data = await apiRequest(`${API_BASE.order}/list`, {
                method: 'GET'
            });

            if (data.data && data.data.orders && data.data.orders.length > 0) {
                state.orders = data.data.orders.map(order => ({
                    id: order.order_id,
                    order_no: order.order_id,
                    items: order.items || [],
                    total: parseFloat(order.payable_amount), // 返回的是元
                    status: order.order_status,
                    created_at: order.created_at
                }));
            }
        } catch (error) {
            console.error('加载订单失败:', error);
        }
    }

    if (state.orders.length === 0) {
        orderList.innerHTML = '<div class="order-empty">暂无订单</div>';
        return;
    }

    // 输出 item.product_id 的值
    state.orders.forEach(order => {
        order.items.forEach(item => {
            console.log('item.product_id:', item.product_id, typeof item.product_id);
        });
    });

    orderList.innerHTML = state.orders.map(order => `
        <div class="order-card">
            <div class="order-header">
                <span class="order-id">订单号: ${order.order_no}</span>
                <span class="order-status ${order.status}">${getStatusText(order.status)}</span>
            </div>
            <div class="order-items">
                ${order.items.map(item => {
                    // 优先使用后端返回的商品名称，如果没有则使用硬编码映射（兼容旧数据）
                    let productName = item.product_name || '商品';
                    
                    // 兼容旧数据的硬编码映射
                    if (!item.product_name) {
                        if (item.product_id === 1) {
                            productName = 'iPhone 15 Pro';
                        } else if (item.product_id === 2) {
                            productName = 'MacBook Pro 14"';
                        } else if (item.product_id === 3) {
                            productName = 'Sony PS5';
                        } else if (item.product_id === 4) {
                            productName = 'AirPods Pro 2';
                        }
                    }

                    return `
                        <div class="order-item">
                            <span>${productName} x ${item.quantity}</span>
                            <span>¥${parseFloat(item.unit_price) * item.quantity}</span>
                        </div>
                    `;
                }).join('')}
            </div>
            <div class="order-footer">
                <span>下单时间: ${new Date(order.created_at).toLocaleString()}</span>
                <strong>总计: ¥${order.total.toFixed(2)}</strong>
            </div>
        </div>
    `).join('');
}

// 获取状态文本
function getStatusText(status) {
    // 后端返回的是数字状态，对应 OrderStatus 枚举
    const statusMap = {
        0: '待支付',
        1: '已支付',
        2: '已发货',
        3: '已完成',
        4: '已取消',
        5: '已退款',
        6: '已关闭',
        'pending': '待支付',
        'paid': '已支付',
        'shipped': '已发货',
        'completed': '已完成',
        'cancelled': '已取消',
        'refund': '已退款',
        'closed': '已关闭'
    };
    return statusMap[status] || status;
}

// 模拟商品数据（用于演示）
const mockProducts = [
    { id: 1, name: 'iPhone 15', description: '最新款苹果手机，搭载 A17 芯片', price: 7999, stock: 100, picture: '📱' },
    { id: 2, name: 'MacBook Pro', description: '高性能笔记本电脑，M3 Pro 芯片', price: 14999, stock: 50, picture: '💻' },
    { id: 3, name: 'Nike Air Max', description: '舒适的运动鞋子，气垫设计', price: 899, stock: 200, picture: '👟' },
    { id: 4, name: 'Sony WH-1000XM5', description: '顶级降噪耳机', price: 2699, stock: 75, picture: '🎧' },
    { id: 5, name: 'iPad Pro', description: '12.9英寸平板电脑，M2芯片', price: 8999, stock: 40, picture: '📱' },
    { id: 6, name: 'Apple Watch', description: '智能手表，健康监测', price: 2999, stock: 60, picture: '⌚' }
];

// 初始化时设置演示商品
state.products = mockProducts;
