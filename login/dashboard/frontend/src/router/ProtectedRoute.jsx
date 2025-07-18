import { Navigate, Outlet } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { CircularProgress, Box } from '@mui/material';

// 受保护的路由组件
function ProtectedRoute() {
  const { isAuthenticated, loading } = useAuth();

  // 如果正在加载认证状态，显示加载指示器
  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <CircularProgress />
      </Box>
    );
  }

  // 如果用户未登录，重定向到登录页面
  if (!isAuthenticated()) {
    return <Navigate to="/login" />;
  }

  // 如果用户已登录，渲染子路由
  return <Outlet />;
}

export default ProtectedRoute;