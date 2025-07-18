import {
  Box,
  Typography,
  Paper,
  Divider,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  Chip,
} from "@mui/material";
import DashboardIcon from "@mui/icons-material/Dashboard";
import ListAltIcon from "@mui/icons-material/ListAlt";
import CodeIcon from "@mui/icons-material/Code";
import HelpOutlineIcon from "@mui/icons-material/HelpOutline";
import FilterAltIcon from "@mui/icons-material/FilterAlt";
import DeleteIcon from "@mui/icons-material/Delete";
import MouseIcon from "@mui/icons-material/Mouse";
import LinkIcon from "@mui/icons-material/Link";

function Tutorial() {
  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Slurm 仪表盘使用教程
      </Typography>

      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography paragraph>
          本系统旨在为您提供一个直观、易用的界面来监控集群状态、管理您的计算任务以及快速生成
          Slurm 作业脚本。本教程将引导您了解系统的主要功能和使用方法。
        </Typography>
      </Paper>

      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography
          variant="h5"
          gutterBottom
          sx={{ display: "flex", alignItems: "center" }}
        >
          <DashboardIcon sx={{ mr: 1 }} />
          仪表盘概览
        </Typography>
        <Divider sx={{ my: 2 }} />
        <Typography paragraph>
          登录后您看到的第一个页面就是仪表盘，它为您提供了集群的实时状态快照。
        </Typography>
        <List>
          <ListItem>
            <ListItemText
              primary="分区信息"
              secondary="这里以卡片形式展示了集群中所有可用的计算分区（如 cpu, gpu）以及每个分区包含的节点列表。"
            />
          </ListItem>
          <ListItem>
            <ListItemText
              primary="节点状态"
              secondary="下方则详细列出了每个计算节点的状态卡片。您可以看到节点名称、当前状态（如 IDLE 空闲, ALLOCATED 已分配, MIXED 混合）、CPU和GPU的资源使用率等核心信息。"
            />
          </ListItem>
        </List>
      </Paper>

      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography
          variant="h5"
          gutterBottom
          sx={{ display: "flex", alignItems: "center" }}
        >
          <ListAltIcon sx={{ mr: 1 }} />
          作业管理
        </Typography>
        <Divider sx={{ my: 2 }} />
        <Typography paragraph>
          “作业管理”页面是您与自己提交的任务进行交互的核心区域。
        </Typography>
        <List>
          <ListItem>
            <ListItemIcon>
              <FilterAltIcon />
            </ListItemIcon>
            <ListItemText
              primary="筛选作业"
              secondary="页面顶部提供了一个筛选表单，您可以按“用户名”或“作业状态”（如 RUNNING, COMPLETED, FAILED）快速找到您关心的任务。"
            />
          </ListItem>
          <ListItem>
            <ListItemIcon>
              <MouseIcon />
            </ListItemIcon>
            <ListItemText
              primary="查看资源详情"
              secondary="直接点击表格中的任意一行，系统会弹出一个窗口，详细展示该任务请求和实际分配到的计算资源（TRES 和 GRES 详情）。"
            />
          </ListItem>
          <ListItem>
            <ListItemIcon>
              <DeleteIcon color="error" />
            </ListItemIcon>
            <ListItemText
              primary="取消作业"
              secondary="在您自己提交的作业（用户名与您当前登录用户匹配）最右侧，会有一个红色的删除图标。点击它并确认后，即可向 Slurm 系统请求取消该作业。"
            />
          </ListItem>
        </List>
      </Paper>

      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography
          variant="h5"
          gutterBottom
          sx={{ display: "flex", alignItems: "center" }}
        >
          <CodeIcon sx={{ mr: 1 }} />
          脚本生成器
        </Typography>
        <Divider sx={{ my: 2 }} />
        <Typography paragraph>
          为了简化 Slurm
          脚本的编写，“脚本生成器”提供了一个表单化的方式来创建任务脚本。
        </Typography>
        <List>
          <ListItem>
            <ListItemText
              primary="选择任务类型"
              secondary={
                <>
                  <Chip label="调试任务 (salloc)" size="small" sx={{ mr: 1 }} />
                  用于快速申请资源进行交互式调试。表单较为简单，填写后会生成一行
                  `salloc` 命令。
                </>
              }
            />
          </ListItem>
          <ListItem>
            <ListItemText
              primary=""
              secondary={
                <>
                  <Chip label="训练任务 (sbatch)" size="small" sx={{ mr: 1 }} />
                  用于提交非交互式的后台批处理任务。表单内容更丰富，并按功能进行了分组：
                  <List dense>
                    <ListItem>
                      <strong>任务配置：</strong>设置任务名称、分区等基本信息。
                    </ListItem>
                    <ListItem>
                      <strong>资源配置：</strong>
                      精细化设置节点、CPU、GPU等计算资源。
                    </ListItem>
                    <ListItem>
                      <strong>日志与工作目录：</strong>
                      指定日志输出路径和工作目录。
                    </ListItem>
                    <ListItem>
                      <strong>邮件通知：</strong>
                      设置在任务开始、结束或失败时接收邮件提醒。
                    </ListItem>
                    <ListItem>
                      <strong>任务脚本：</strong>
                      在页面最下方的大输入框中，填写您实际要执行的命令（例如
                      `srun python train.py`）。
                    </ListItem>
                  </List>
                </>
              }
            />
          </ListItem>
          <ListItem>
            <ListItemText
              primary="使用方式"
              secondary={
                <div>
                  点击右上角的“复制”按钮，即可方便地将脚本内容保存到剪贴板。
                  <br></br>
                  对于调试任务，直接在终端中粘贴执行即可。<br></br>
                  对于批处理任务，您可以将生成的脚本内容保存为一个文件（如
                  `my_job.sh`），然后在终端中使用 `sbatch my_job.sh` 提交任务。
                </div>
              }
            />
          </ListItem>
        </List>
      </Paper>

      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography
          variant="h5"
          gutterBottom
          sx={{ display: "flex", alignItems: "center" }}
        >
          <LinkIcon sx={{ mr: 1 }} />
          连接到正在运行的作业
        </Typography>
        <Divider sx={{ my: 2 }} />
        <Typography paragraph>
          在您提交作业的那个目录下，系统会自动生成名为 connect-jobid .log 和
          info-jobid.log 的文件。
          <br />
          <br />
          connect-jobid.log
          文件包含了连接到正在运行作业的方法以及命令，您可以参考该文件来连接到作业。
          <br />
          info-jobid.log 文件则包含了作业因资源使用不足而被强制断掉的具体原因。
        </Typography>
      </Paper>

      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography
          variant="h5"
          gutterBottom
          sx={{ display: "flex", alignItems: "center" }}
        >
          <HelpOutlineIcon sx={{ mr: 1 }} />
          获取帮助
        </Typography>
        <Divider sx={{ my: 2 }} />
        <Typography paragraph>
          如果您在使用过程中遇到任何问题或有功能建议，请联系系统管理员。
        </Typography>
      </Paper>
    </Box>
  );
}

export default Tutorial;
