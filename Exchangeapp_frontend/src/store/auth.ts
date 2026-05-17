import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import axios from '../axios';

// decodeUsername 从 JWT payload 中解析用户名，用于前端显示和鉴权判断。
const decodeUsername = (token: string | null): string | null => {
  if (!token) {
    return null;
  }

  const parts = token.replace('Bearer ', '').split('.');
  if (parts.length !== 3) {
    return null;
  }

  try {
    const normalized = parts[1].replace(/-/g, '+').replace(/_/g, '/');
    const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '=');
    const payload = JSON.parse(atob(padded));
    return typeof payload.username === 'string' ? payload.username : null;
  } catch {
    return null;
  }
};

// useAuthStore 维护登录态、用户名及登录/注册/登出动作。
export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'));
  const username = computed(() => decodeUsername(token.value));

  // 只要存在 token 就视为已登录（接口权限由后端最终校验）。
  const isAuthenticated = computed(() => !!token.value);

  // login 调用后端登录接口并持久化 token。
  const login = async (username: string, password: string) => {
    try {
      const response = await axios.post('/auth/login', { username, password });
      token.value = response.data.token;
      localStorage.setItem('token', token.value || '');
    } catch (error) {
      throw new Error(`Login failed! ${error}`);
    }
  };

  // register 调用后端注册接口并持久化 token。
  const register = async (username: string, password: string) => {
    try {
      const response = await axios.post('/auth/register', { username, password });
      token.value = response.data.token;
      localStorage.setItem('token', token.value || '');
    } catch (error) {
      throw new Error(`Register failed! ${error}`);
    }
  };

  // logout 清理本地登录状态。
  const logout = () => {
    token.value = null;
    localStorage.removeItem('token');
  };

  return {
    token,
    username,
    isAuthenticated,
    login,
    register,
    logout
  };
});
