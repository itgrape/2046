import { useState, useEffect } from "react";
import {
  Typography,
  Grid,
  Box,
  Card,
  CardContent,
  Chip,
  Alert,
  CircularProgress,
  LinearProgress,
  Divider,
} from "@mui/material";
import { useTheme } from "@mui/material/styles";
import DnsIcon from "@mui/icons-material/Dns";
import MemoryIcon from "@mui/icons-material/Memory";
import apiService from "../services/api";

// 根据节点状态返回不同的颜色
const getNodeStateColor = (state) => {
  if (state.includes("ALLOCATED")) return "error";
  if (state.includes("MIXED")) return "warning";
  if (state.includes("IDLE")) return "success";
  return "default";
};

// 资源使用率进度条组件
const ResourceProgress = ({ title, used, total }) => {
  const theme = useTheme();
  const percentage = total > 0 ? (used / total) * 100 : 0;
  let progressColor = theme.palette.success.main;
  if (percentage > 85) {
    progressColor = theme.palette.error.main;
  } else if (percentage > 60) {
    progressColor = theme.palette.warning.main;
  }

  return (
    <Box sx={{ width: "100%", my: 1.5 }}>
      <Box sx={{ display: "flex", justifyContent: "space-between", mb: 0.5 }}>
        <Typography variant="body2" color="text.secondary">
          {title}
        </Typography>
        <Typography
          variant="body2"
          color="text.secondary"
        >{`${used} / ${total}`}</Typography>
      </Box>
      <LinearProgress
        variant="determinate"
        value={percentage}
        sx={{
          height: 8,
          borderRadius: 4,
          "& .MuiLinearProgress-bar": {
            backgroundColor: progressColor,
          },
        }}
      />
    </Box>
  );
};

function Dashboard() {
  const [clusterStatus, setClusterStatus] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchClusterStatus = async () => {
      try {
        const data = await apiService.getClusterStatus();

        // --- 数据过滤逻辑 ---
        const filteredPartitions = data.partitions.filter(
          (p) => p.name !== "login"
        );
        const filteredNodes = data.nodes.filter(
          (node) =>
            !(node.partitions.length === 1 && node.partitions[0] === "login")
        );

        setClusterStatus({
          ...data,
          partitions: filteredPartitions,
          nodes: filteredNodes,
        });
      } catch (err) {
        setError("获取集群状态失败。请检查API端点和网络连接。");
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    fetchClusterStatus();
    const intervalId = setInterval(fetchClusterStatus, 30000);
    return () => clearInterval(intervalId);
  }, []);

  if (loading) {
    return (
      <Box
        sx={{
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          height: "80vh",
        }}
      >
        <CircularProgress />
        <Typography sx={{ ml: 2 }}>加载中...</Typography>
      </Box>
    );
  }

  if (error) {
    return <Alert severity="error">{error}</Alert>;
  }

  if (!clusterStatus) {
    return <Alert severity="info">没有可用的集群数据。</Alert>;
  }

  const { partitions, nodes } = clusterStatus;

  return (
    <Box sx={{ flexGrow: 1, p: 1 }}>
      <Typography variant="h4" sx={{ mb: 3, fontWeight: 500 }}>
        集群总览
      </Typography>

      {/* 分区信息 */}
      <Box sx={{ mb: 4 }}>
        <Typography
          variant="h6"
          sx={{ mb: 2, display: "flex", alignItems: "center" }}
        >
          <DnsIcon sx={{ mr: 1, color: "text.secondary" }} /> 分区信息
        </Typography>
        <Box sx={{ display: "flex", flexWrap: "wrap", gap: 2 }}>
          {partitions.map((partition) => (
            <Card
              key={partition.name}
              variant="outlined"
              sx={{ minWidth: 200 }}
            >
              <CardContent>
                <Typography
                  sx={{ fontWeight: "bold", textTransform: "uppercase" }}
                >
                  {partition.name}
                </Typography>
                <Divider sx={{ my: 1 }} />
                <Box sx={{ display: "flex", flexWrap: "wrap", gap: 0.5 }}>
                  {partition.nodes.map((node) => (
                    <Chip key={node} label={node} size="small" />
                  ))}
                </Box>
              </CardContent>
            </Card>
          ))}
        </Box>
      </Box>

      {/* 节点状态信息 */}
      <Box>
        <Typography
          variant="h6"
          sx={{ mb: 2, display: "flex", alignItems: "center" }}
        >
          <MemoryIcon sx={{ mr: 1, color: "text.secondary" }} /> 节点状态
        </Typography>
        <Grid container spacing={3}>
          {nodes.map((node) => (
            <Grid item xs={12} sm={6} md={4} key={node.name}>
              <Card variant="outlined" sx={{ height: "100%" }}>
                <CardContent>
                  <Box
                    sx={{
                      display: "flex",
                      justifyContent: "space-between",
                      alignItems: "center",
                      mb: 1,
                    }}
                  >
                    <Typography variant="h6" component="div" fontWeight="500">
                      {node.name}
                    </Typography>
                    &nbsp;&nbsp;
                    <Chip
                      label={node.state.join(", ")}
                      color={getNodeStateColor(node.state)}
                      size="small"
                    />
                  </Box>
                  <Typography
                    variant="body2"
                    color="text.secondary"
                    sx={{ mb: 2 }}
                  >
                    分区: {node.partitions.join(", ")}
                  </Typography>

                  <Divider sx={{ my: 1 }} />

                  <ResourceProgress
                    title="CPU"
                    used={node.allocated_cpus}
                    total={node.total_cpus}
                  />

                  {node.gpus && node.gpus.length > 0 ? (
                    node.gpus.map((gpu) => (
                      <ResourceProgress
                        key={gpu.type}
                        title={`GPU (${gpu.type})`}
                        used={gpu.allocated}
                        total={gpu.total}
                      />
                    ))
                  ) : (
                    <Typography
                      variant="body2"
                      sx={{
                        mt: 2,
                        color: "text.secondary",
                        textAlign: "center",
                      }}
                    >
                      无 GPU 资源
                    </Typography>
                  )}
                </CardContent>
              </Card>
            </Grid>
          ))}
        </Grid>
      </Box>
    </Box>
  );
}

export default Dashboard;
