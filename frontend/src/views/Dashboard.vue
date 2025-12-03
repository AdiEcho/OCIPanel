<template>
  <div class="space-y-6">
    <h1 class="text-3xl font-bold">系统概览</h1>

    <!-- Stats Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div class="card p-6 hover:scale-105 transition-transform duration-200">
        <div class="flex items-center gap-4">
          <div class="w-14 h-14 bg-gradient-to-br from-blue-500 to-blue-600 rounded-xl flex items-center justify-center">
            <svg class="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
          </div>
          <div>
            <p class="text-slate-400 text-sm">配置总数</p>
            <p class="text-3xl font-bold">{{ stats.totalConfigs }}</p>
          </div>
        </div>
      </div>

      <div class="card p-6 hover:scale-105 transition-transform duration-200">
        <div class="flex items-center gap-4">
          <div class="w-14 h-14 bg-gradient-to-br from-yellow-500 to-orange-600 rounded-xl flex items-center justify-center">
            <svg class="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
            </svg>
          </div>
          <div>
            <p class="text-slate-400 text-sm">任务总数</p>
            <p class="text-3xl font-bold">{{ stats.totalTasks }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Quick Actions -->
    <div class="card p-6">
      <h2 class="text-xl font-bold mb-4">快速操作</h2>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <router-link to="/configs" class="btn btn-primary justify-center">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          添加配置
        </router-link>
        <router-link to="/instances" class="btn btn-success justify-center">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
          </svg>
          管理实例
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../utils/api'
import { toast } from '../utils/toast'

const stats = ref({
  totalConfigs: 0,
  totalTasks: 0
})

const loadStats = async () => {
  try {
    const response = await api.post('/sys/getGlance', {})
    if (response.data) {
      stats.value = response.data
    }
  } catch (error) {
    console.error('加载统计数据失败:', error)
    toast.error('加载统计数据失败')
  }
}

onMounted(() => {
  loadStats()
})
</script>
