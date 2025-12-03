<template>
  <div class="space-y-6">
    <h1 class="text-3xl font-bold">系统设置</h1>

    <div class="card">
      <div class="p-6 border-b border-slate-700">
        <h2 class="text-xl font-semibold">系统信息</h2>
      </div>
      <div class="p-6 space-y-4">
        <div class="flex justify-between items-center py-3 border-b border-slate-700">
          <span class="text-slate-400">应用名称</span>
          <span class="font-semibold">OCI Panel</span>
        </div>
        <div class="flex justify-between items-center py-3 border-b border-slate-700">
          <span class="text-slate-400">版本号</span>
          <span class="font-semibold">v1.0.0</span>
        </div>
        <div class="flex justify-between items-center py-3 border-b border-slate-700">
          <span class="text-slate-400">运行状态</span>
          <span class="badge badge-success">正常运行</span>
        </div>
        <div class="flex justify-between items-center py-3 border-b border-slate-700">
          <span class="text-slate-400">后端框架</span>
          <span class="font-semibold">Gin (Go)</span>
        </div>
        <div class="flex justify-between items-center py-3 border-b border-slate-700">
          <span class="text-slate-400">前端框架</span>
          <span class="font-semibold">Vue 3 + Vite + Tailwind CSS</span>
        </div>
        <div class="flex justify-between items-center py-3">
          <span class="text-slate-400">数据库</span>
          <span class="font-semibold">SQLite</span>
        </div>
      </div>
    </div>

    <div class="card">
      <div class="p-6 border-b border-slate-700">
        <h2 class="text-xl font-semibold">缓存设置</h2>
      </div>
      <div class="p-6 space-y-4">
        <div class="flex justify-between items-center py-3 border-b border-slate-700">
          <div>
            <span class="text-slate-300">启用数据缓存</span>
            <p class="text-sm text-slate-500 mt-1">启用后将定时缓存配置的实例数据到数据库，减少对OCI API的请求</p>
          </div>
          <label class="relative inline-flex items-center cursor-pointer">
            <input type="checkbox" v-model="cacheConfig.cacheEnabled" class="sr-only peer" @change="updateCacheConfig">
            <div class="w-11 h-6 bg-slate-600 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-500"></div>
          </label>
        </div>
        <div class="flex justify-between items-center py-3 border-b border-slate-700">
          <div>
            <span class="text-slate-300">缓存刷新间隔</span>
            <p class="text-sm text-slate-500 mt-1">定时任务检查并更新缓存的间隔时间（分钟）</p>
          </div>
          <div class="flex items-center gap-2">
            <input 
              type="number" 
              v-model.number="cacheConfig.cacheInterval" 
              min="5" 
              max="1440"
              class="w-20 px-3 py-2 bg-slate-700 border border-slate-600 rounded-lg text-white text-center"
              :disabled="!cacheConfig.cacheEnabled"
              @change="updateCacheConfig"
            >
            <span class="text-slate-400">分钟</span>
          </div>
        </div>
        <div class="flex justify-between items-center py-3">
          <div>
            <span class="text-slate-300">手动刷新缓存</span>
            <p class="text-sm text-slate-500 mt-1">立即更新所有配置的缓存数据</p>
          </div>
          <button 
            @click="refreshCache" 
            :disabled="!cacheConfig.cacheEnabled || refreshing"
            class="btn btn-primary"
          >
            <svg v-if="refreshing" class="animate-spin h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            {{ refreshing ? '刷新中...' : '立即刷新' }}
          </button>
        </div>
      </div>
    </div>

    <div class="card">
      <div class="p-6 border-b border-slate-700">
        <h2 class="text-xl font-semibold">系统配置</h2>
      </div>
      <div class="p-6 space-y-4">
        <div v-if="loading" class="text-center py-8">
          <svg class="animate-spin h-8 w-8 mx-auto text-blue-500" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
        </div>
        <template v-else>
          <div class="flex justify-between items-center py-3 border-b border-slate-700">
            <span class="text-slate-400">密钥目录</span>
            <span class="font-mono text-sm">{{ config.keyDirPath || 'N/A' }}</span>
          </div>
          <div class="flex justify-between items-center py-3 border-b border-slate-700">
            <span class="text-slate-400">日志级别</span>
            <span class="font-semibold">{{ config.logLevel || 'N/A' }}</span>
          </div>
          <div class="flex justify-between items-center py-3">
            <span class="text-slate-400">AI功能</span>
            <span :class="config.aiEnabled ? 'badge-success' : 'badge-danger'" class="badge">
              {{ config.aiEnabled ? '已启用' : '未启用' }}
            </span>
          </div>
        </template>
      </div>
    </div>

    <div class="card">
      <div class="p-6 border-b border-slate-700">
        <h2 class="text-xl font-semibold">关于</h2>
      </div>
      <div class="p-6">
        <p class="text-slate-300 leading-relaxed">
          OCI Panel 是一个基于 Go + Vue 3 开发的 Oracle Cloud Infrastructure 管理面板，
          提供实例管理、网络配置、任务调度等功能，帮助用户更便捷地管理 OCI 资源。
        </p>
        <div class="mt-6 flex gap-4">
          <a 
            href="https://github.com" 
            target="_blank" 
            class="btn btn-secondary"
          >
            <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
              <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.840 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.430.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
            </svg>
            GitHub
          </a>
          <a 
            href="https://docs.oracle.com/iaas" 
            target="_blank" 
            class="btn btn-secondary"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            OCI文档
          </a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../utils/api'
import { toast } from '../utils/toast'

const loading = ref(false)
const refreshing = ref(false)
const config = ref({
  keyDirPath: '',
  logLevel: '',
  aiEnabled: false
})
const cacheConfig = ref({
  cacheEnabled: false,
  cacheInterval: 30
})

const loadConfig = async () => {
  loading.value = true
  try {
    const response = await api.post('/sys/getSysCfg', {})
    if (response.data) {
      config.value = response.data
      cacheConfig.value.cacheEnabled = response.data.cacheEnabled || false
      cacheConfig.value.cacheInterval = response.data.cacheInterval || 30
    }
  } catch (error) {
    console.error('加载配置失败:', error)
    toast.error('加载系统配置失败')
  } finally {
    loading.value = false
  }
}

const updateCacheConfig = async () => {
  try {
    await api.post('/sys/updateCacheCfg', {
      cacheEnabled: cacheConfig.value.cacheEnabled,
      cacheInterval: cacheConfig.value.cacheInterval
    })
    toast.success('缓存配置已更新')
  } catch (error) {
    console.error('更新缓存配置失败:', error)
    toast.error('更新缓存配置失败')
  }
}

const refreshCache = async () => {
  refreshing.value = true
  try {
    await api.post('/sys/refreshCache', {})
    toast.success('缓存刷新任务已启动')
  } catch (error) {
    console.error('刷新缓存失败:', error)
    toast.error('刷新缓存失败')
  } finally {
    refreshing.value = false
  }
}

onMounted(() => {
  loadConfig()
})
</script>
