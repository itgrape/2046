--[[
  Slurm job_submit.lua script
  功能：
  1. 禁止用户在提交作业时使用 --nodelist 参数。
  2. 禁止用户在提交作业时使用 --exclude 参数。
--]]

-- init() 函数在插件加载时被调用。
function init()
    slurm.log_info("job_submit/lua: the custom script is loaded")
    return slurm.SUCCESS
end

-- slurm_job_submit() 函数在新作业提交时被调用 (sbatch, salloc)。
function slurm_job_submit(job_desc, part_list, submit_uid)
    -- 检查用户是否指定了 --nodelist
    -- job_desc.nodes 对应 --nodelist 参数
    if job_desc.nodes ~= nil then
        slurm.log_user("错误：不允许使用 --nodelist 参数指定节点。请移除该参数后重试。")
        return slurm.ERROR
    end

    -- 检查用户是否指定了 --exclude
    -- job_desc.exc_nodes 对应 --exclude 参数
    if job_desc.exc_nodes ~= nil then
        slurm.log_user("错误：不允许使用 --exclude 参数排除节点。请移除该参数后重试。")
        return slurm.ERROR
    end

    -- 如果所有检查都通过，则允许作业提交
    return slurm.SUCCESS
end

-- slurm_job_modify() 函数在作业被修改时被调用 (scontrol update)。
-- 这是 slurm_job_submit/lua 插件必需的函数。
-- 即使我们在这里不做任何检查，也必须定义它并返回成功。
function slurm_job_modify(job_desc, job_rec, part_list, modify_uid)
    -- 在本场景中，我们不在修改作业时应用相同的限制，
    return slurm.SUCCESS
end

-- slurm_job_p_requeue() 函数在作业被重新排队时调用。
function slurm_job_p_requeue(job_desc, part_list, requeue_uid)
    return slurm.SUCCESS
end

-- fini() 函数在插件卸载时调用。
function fini()
    slurm.log_info("job_submit/lua: the custom script is unloaded")
    return slurm.SUCCESS
end
