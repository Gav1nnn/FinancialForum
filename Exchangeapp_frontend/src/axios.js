import axios from 'axios';
// 统一 API 客户端，默认走 /api 前缀（由 dev proxy 或网关转发）。
const instance = axios.create({
    baseURL: '/api',
});
// 请求拦截器：自动附加本地保存的 Bearer Token。
instance.interceptors.request.use(config => {
    const token = localStorage.getItem('token');
    if (token) {
        config.headers.Authorization = token;
    }
    return config;
});
export default instance;
