/**
 * Token工具函数
 */

/**
 * 保存token到本地存储并设置过期时间
 * @param {string} token - 从API获取的token
 */
export const saveToken = (token) => {
  try {
    // 保存token
    localStorage.setItem('token', token);
    
    // 计算24小时后的过期时间
    const expiryTime = new Date();
    expiryTime.setHours(expiryTime.getHours() + 24);
    
    // 保存过期时间
    localStorage.setItem('tokenExpiry', expiryTime.getTime().toString());
  } catch (error) {
    console.error('保存token失败:', error);
  }
};

/**
 * 检查token是否过期
 * @returns {boolean} - 如果token已过期或不存在则返回true，否则返回false
 */
export const isTokenExpired = () => {
  try {
    // 获取token和过期时间
    const token = localStorage.getItem('token');
    const expiryTime = localStorage.getItem('tokenExpiry');
    
    // 如果token或过期时间不存在，则视为已过期
    if (!token || !expiryTime) return true;
    
    // 检查是否已过期
    const now = new Date().getTime();
    return now > parseInt(expiryTime, 10);
  } catch (error) {
    console.error('检查token过期失败:', error);
    return true; // 如果出错，视为已过期
  }
};

/**
 * 清除token和相关数据
 */
export const clearToken = () => {
  localStorage.removeItem('token');
  localStorage.removeItem('tokenExpiry');
  localStorage.removeItem('user');
};