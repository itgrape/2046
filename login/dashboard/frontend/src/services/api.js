import axios from "axios";

// 创建axios实例
const api = axios.create({
  baseURL: "/api",
  timeout: 3000,
  headers: {
    "Content-Type": "application/json",
  },
});

// 请求拦截器 - 添加token到请求头
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem("token");
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器 - 处理常见错误
api.interceptors.response.use(
  (response) => {
    return response.data;
  },
  (error) => {
    // 处理401错误 - 未授权/token过期
    if (error.response && error.response.status === 401) {
      if (error.config.url.endsWith("/login")) {
        return Promise.reject(error);
      }

      // 清除本地存储的token和用户信息
      localStorage.removeItem("token");
      localStorage.removeItem("user");
      // 重定向到登录页
      window.location.href = "/login";
    }
    return Promise.reject(error);
  }
);

// API服务
const apiService = {
  // 登录
  login: async (username, password) => {
    try {
      const response = await api.post("/login", { username, password });
      return response;
    } catch (error) {
      console.error("登录失败:", error);
      throw error;
    }
  },

  // 获取集群状态
  getClusterStatus: async () => {
    try {
      const response = await api.get("/v1/cluster/status");
      return response;
    } catch (error) {
      console.error("获取集群状态失败:", error);
      throw error;
    }
  },

  getJobs: async (params) => {
    // 接受查询参数
    try {
      const response = await api.get("/v1/jobs", { params });
      return response;
    } catch (error) {
      console.error("获取作业列表失败:", error);
      throw error;
    }
  },

  // 删除作业 (新增)
  deleteJob: async (jobId) => {
    try {
      const response = await api.delete(`/v1/job/${jobId}`);
      return response;
    } catch (error) {
      console.error(`删除作业 ${jobId} 失败:`, error);
      throw error;
    }
  },
};

export default apiService;
