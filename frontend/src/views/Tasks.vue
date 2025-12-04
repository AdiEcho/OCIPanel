<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex justify-between items-center">
      <div class="flex items-center gap-4">
        <h1 class="text-3xl font-bold">任务列表</h1>
        <span v-if="selectedTaskIds.length > 0" class="text-sm text-slate-400">
          已选择 {{ selectedTaskIds.length }} 项
        </span>
      </div>
      <div class="flex gap-2">
        <button 
          v-if="selectedTaskIds.length > 0" 
          class="btn btn-danger" 
          @click="batchDeleteTasks"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
          批量删除
        </button>
        <select v-model="statusFilter" class="input w-32" @change="loadTasks(1)">
          <option value="">全部状态</option>
          <option value="running">运行中</option>
          <option value="stopped">已停止</option>
          <option value="completed">已完成</option>
          <option value="error">失败</option>
        </select>
        <button class="btn btn-secondary" @click="loadTasks(currentPage)">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          刷新
        </button>
      </div>
    </div>

    <!-- 任务列表卡片 -->
    <div class="card">
      <div class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-slate-700/50">
            <tr>
              <th class="px-4 py-4 text-center w-12">
                <input 
                  type="checkbox" 
                  class="checkbox" 
                  :checked="isAllSelected" 
                  :indeterminate="isIndeterminate"
                  @change="toggleSelectAll"
                />
              </th>
              <th class="px-4 py-4 text-left text-xs font-semibold text-slate-300 uppercase">配置名</th>
              <th class="px-4 py-4 text-left text-xs font-semibold text-slate-300 uppercase">区域</th>
              <th class="px-4 py-4 text-left text-xs font-semibold text-slate-300 uppercase">配置</th>
              <th class="px-4 py-4 text-left text-xs font-semibold text-slate-300 uppercase">间隔</th>
              <th class="px-4 py-4 text-left text-xs font-semibold text-slate-300 uppercase">状态</th>
              <th class="px-4 py-4 text-left text-xs font-semibold text-slate-300 uppercase">执行次数</th>
              <th class="px-4 py-4 text-left text-xs font-semibold text-slate-300 uppercase">最后执行</th>
              <th class="px-4 py-4 text-left text-xs font-semibold text-slate-300 uppercase">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-700">
            <tr v-if="loading">
              <td colspan="9" class="px-6 py-8 text-center text-slate-400">
                <svg class="animate-spin h-8 w-8 mx-auto" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
              </td>
            </tr>
            <tr v-else-if="!tasks.length">
              <td colspan="9" class="px-6 py-8 text-center text-slate-400">暂无任务</td>
            </tr>
            <tr v-for="task in tasks" v-else :key="task.id" class="hover:bg-slate-700/30">
              <td class="px-4 py-4 text-center">
                <input 
                  type="checkbox" 
                  class="checkbox" 
                  :checked="selectedTaskIds.includes(task.id)"
                  @change="toggleSelectTask(task.id)"
                />
              </td>
              <td class="px-4 py-4 font-medium">{{ task.username || '-' }}</td>
              <td class="px-4 py-4">
                <span class="px-2 py-1 text-xs font-semibold rounded-full bg-blue-500/20 text-blue-300">
                  {{ task.ociRegion }}
                </span>
              </td>
              <td class="px-4 py-4 text-sm">
                <div>{{ task.architecture }} / {{ task.operationSystem }}</div>
                <div class="text-slate-400 text-xs">{{ task.ocpus }}核 / {{ task.memory }}GB / {{ task.disk }}GB</div>
              </td>
              <td class="px-4 py-4">{{ task.interval }}秒</td>
              <td class="px-4 py-4">
                <span 
                  class="px-2 py-1 text-xs font-semibold rounded-full"
                  :class="{
                    'bg-green-500/20 text-green-300': task.status === 'running',
                    'bg-yellow-500/20 text-yellow-300': task.status === 'stopped',
                    'bg-blue-500/20 text-blue-300': task.status === 'completed',
                    'bg-red-500/20 text-red-300': task.status === 'error'
                  }"
                >
                  {{ statusMap[task.status] || task.status }}
                </span>
              </td>
              <td class="px-4 py-4">
                <span class="text-slate-300">{{ task.executeCount }}</span>
                <span class="text-green-400 ml-1">({{ task.successCount }} 成功)</span>
              </td>
              <td class="px-4 py-4 text-sm">
                <div v-if="task.lastExecuteTime">{{ task.lastExecuteTime }}</div>
                <div class="text-slate-400 text-xs truncate max-w-[150px]" :title="task.lastMessage">
                  {{ task.lastMessage || '-' }}
                </div>
              </td>
              <td class="px-4 py-4">
                <div class="flex gap-2">
                  <button 
                    v-if="task.status === 'stopped'" 
                    class="btn btn-success text-xs py-1 px-2" 
                    title="启动"
                    :disabled="actionLoading[task.id]"
                    @click="startTask(task.id)"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                    </svg>
                  </button>
                  <button 
                    v-if="task.status === 'running'" 
                    class="btn btn-warning text-xs py-1 px-2" 
                    title="停止"
                    :disabled="actionLoading[task.id]"
                    @click="stopTask(task.id)"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 10a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z" />
                    </svg>
                  </button>
                  <button 
                    class="btn btn-primary text-xs py-1 px-2" 
                    title="查看日志"
                    @click="viewLogs(task)"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                  </button>
                  <button 
                    class="btn btn-danger text-xs py-1 px-2" 
                    title="删除"
                    :disabled="actionLoading[task.id]"
                    @click="deleteTask(task.id)"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 分页 -->
      <div v-if="totalPages > 1" class="p-6 border-t border-slate-700 flex justify-center gap-2">
        <button :disabled="currentPage === 1" class="btn btn-secondary" @click="loadTasks(currentPage - 1)">
          上一页
        </button>
        <span class="flex items-center px-4 text-slate-300">第 {{ currentPage }} / {{ totalPages }} 页</span>
        <button :disabled="currentPage === totalPages" class="btn btn-secondary" @click="loadTasks(currentPage + 1)">
          下一页
        </button>
      </div>
    </div>

    <!-- 日志弹窗 -->
    <div
      v-if="showLogsModal"
      class="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50 p-4"
    >
      <div class="card max-w-4xl w-full max-h-[80vh] flex flex-col">
        <div class="p-6 border-b border-slate-700 flex justify-between items-center flex-shrink-0">
          <div>
            <h3 class="text-xl font-bold">任务执行日志</h3>
            <p class="text-sm text-slate-400 mt-1">{{ currentTask?.username }} - {{ currentTask?.ociRegion }}</p>
          </div>
          <div class="flex gap-2">
            <button class="btn btn-secondary text-sm" @click="clearLogs">清空日志</button>
            <button class="text-slate-400 hover:text-white" @click="closeLogsModal">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        <div class="flex-1 overflow-y-auto p-6">
          <div v-if="logsLoading" class="text-center py-8">
            <svg class="animate-spin h-8 w-8 mx-auto text-blue-500" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
          </div>
          <div v-else-if="!logs.length" class="text-center py-8 text-slate-400">
            暂无日志记录
          </div>
          <div v-else class="space-y-2">
            <div 
              v-for="log in logs" 
              :key="log.id" 
              class="p-3 rounded-lg bg-slate-700/30 border-l-4"
              :class="{
                'border-green-500': log.status === 'success',
                'border-red-500': log.status === 'error',
                'border-blue-500': log.status === 'info'
              }"
            >
              <div class="flex justify-between items-start">
                <span 
                  class="px-2 py-0.5 text-xs rounded"
                  :class="{
                    'bg-green-500/20 text-green-300': log.status === 'success',
                    'bg-red-500/20 text-red-300': log.status === 'error',
                    'bg-blue-500/20 text-blue-300': log.status === 'info'
                  }"
                >
                  {{ log.status === 'success' ? '成功' : log.status === 'error' ? '失败' : '信息' }}
                </span>
                <span class="text-xs text-slate-400">{{ formatTime(log.executeTime) }}</span>
              </div>
              <p class="mt-2 text-sm text-slate-300 break-all">{{ log.message }}</p>
            </div>
          </div>
        </div>

        <div v-if="logsTotalPages > 1" class="p-4 border-t border-slate-700 flex justify-center gap-2 flex-shrink-0">
          <button :disabled="logsPage === 1" class="btn btn-secondary btn-sm" @click="loadLogs(logsPage - 1)">
            上一页
          </button>
          <span class="flex items-center px-3 text-sm text-slate-300">{{ logsPage }} / {{ logsTotalPages }}</span>
          <button :disabled="logsPage === logsTotalPages" class="btn btn-secondary btn-sm" @click="loadLogs(logsPage + 1)">
            下一页
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import api from '../utils/api'
import { toast } from '../utils/toast'

const tasks = ref([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = 10
const totalPages = ref(0)
const statusFilter = ref('')
const actionLoading = reactive({})
const selectedTaskIds = ref([])

const statusMap = {
  running: '运行中',
  stopped: '已停止',
  completed: '已完成',
  error: '失败'
}

// 选择相关
const isAllSelected = computed(() => tasks.value.length > 0 && selectedTaskIds.value.length === tasks.value.length)
const isIndeterminate = computed(() => selectedTaskIds.value.length > 0 && selectedTaskIds.value.length < tasks.value.length)

const toggleSelectTask = (taskId) => {
  const index = selectedTaskIds.value.indexOf(taskId)
  if (index === -1) {
    selectedTaskIds.value.push(taskId)
  } else {
    selectedTaskIds.value.splice(index, 1)
  }
}

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedTaskIds.value = []
  } else {
    selectedTaskIds.value = tasks.value.map(t => t.id)
  }
}

const batchDeleteTasks = async () => {
  if (!confirm(`确定要删除选中的 ${selectedTaskIds.value.length} 个任务吗？相关日志也将被删除。`)) return
  try {
    await api.post('/task/batchDelete', { taskIds: selectedTaskIds.value })
    toast.success(`成功删除 ${selectedTaskIds.value.length} 个任务`)
    selectedTaskIds.value = []
    await loadTasks(currentPage.value)
  } catch (error) {
    toast.error(error.message || '批量删除失败')
  }
}

const showLogsModal = ref(false)
const currentTask = ref(null)
const logs = ref([])
const logsLoading = ref(false)
const logsPage = ref(1)
const logsPageSize = 20
const logsTotalPages = ref(0)

const loadTasks = async (page = 1) => {
  loading.value = true
  try {
    const response = await api.post('/task/list', { 
      page, 
      pageSize,
      status: statusFilter.value 
    })
    tasks.value = response.data.list || []
    currentPage.value = response.data.page
    totalPages.value = Math.ceil(response.data.total / pageSize)
  } catch (error) {
    toast.error(error.message || '加载失败')
  } finally {
    loading.value = false
  }
}

const startTask = async (taskId) => {
  actionLoading[taskId] = true
  try {
    await api.post('/task/start', { taskId })
    toast.success('任务已启动')
    await loadTasks(currentPage.value)
  } catch (error) {
    toast.error(error.message || '启动失败')
  } finally {
    delete actionLoading[taskId]
  }
}

const stopTask = async (taskId) => {
  actionLoading[taskId] = true
  try {
    await api.post('/task/stop', { taskId })
    toast.success('任务已停止')
    await loadTasks(currentPage.value)
  } catch (error) {
    toast.error(error.message || '停止失败')
  } finally {
    delete actionLoading[taskId]
  }
}

const deleteTask = async (taskId) => {
  if (!confirm('确定要删除此任务吗？相关日志也将被删除。')) return
  actionLoading[taskId] = true
  try {
    await api.post('/task/delete', { taskId })
    toast.success('任务已删除')
    await loadTasks(currentPage.value)
  } catch (error) {
    toast.error(error.message || '删除失败')
  } finally {
    delete actionLoading[taskId]
  }
}

const viewLogs = async (task) => {
  currentTask.value = task
  showLogsModal.value = true
  logsPage.value = 1
  await loadLogs(1)
}

const loadLogs = async (page = 1) => {
  logsLoading.value = true
  try {
    const response = await api.post('/task/logs', {
      taskId: currentTask.value.id,
      page,
      pageSize: logsPageSize
    })
    logs.value = response.data.list || []
    logsPage.value = response.data.page
    logsTotalPages.value = Math.ceil(response.data.total / logsPageSize)
  } catch (error) {
    toast.error(error.message || '加载日志失败')
  } finally {
    logsLoading.value = false
  }
}

const clearLogs = async () => {
  if (!confirm('确定要清空此任务的所有日志吗？')) return
  try {
    await api.post('/task/clearLogs', { taskId: currentTask.value.id })
    toast.success('日志已清空')
    logs.value = []
    logsTotalPages.value = 0
  } catch (error) {
    toast.error(error.message || '清空日志失败')
  }
}

const closeLogsModal = () => {
  showLogsModal.value = false
  currentTask.value = null
  logs.value = []
}

const formatTime = (time) => {
  if (!time) return '-'
  return time.replace('T', ' ').substring(0, 19)
}

onMounted(() => {
  loadTasks()
})
</script>
