<script setup lang="ts">
import { ref, onUnmounted, nextTick } from 'vue'
import { Wifi, WifiOff, Trash2, Terminal } from 'lucide-vue-next'
import { toast } from '@/composables/useToast'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

interface LogEntry {
  message: string
  type: 'info' | 'error' | 'warning' | 'success'
}

const logs = ref<LogEntry[]>([])
const isConnected = ref(false)
const ws = ref<WebSocket | null>(null)
const logConsole = ref<HTMLElement>()

const addLog = (message: string, type: LogEntry['type'] = 'info') => {
  const timestamp = new Date().toLocaleTimeString('zh-CN')
  logs.value.push({
    message: `[${timestamp}] ${message}`,
    type
  })

  nextTick(() => {
    if (logConsole.value) {
      logConsole.value.scrollTop = logConsole.value.scrollHeight
    }
  })
}

const connectWebSocket = () => {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/ws/logs`

  try {
    ws.value = new WebSocket(wsUrl)

    ws.value.onopen = () => {
      isConnected.value = true
      addLog('WebSocket 连接成功', 'success')
      toast.success('日志连接成功')
    }

    ws.value.onmessage = event => {
      addLog(event.data, 'info')
    }

    ws.value.onerror = () => {
      addLog('WebSocket 连接错误', 'error')
      toast.error('WebSocket连接错误')
    }

    ws.value.onclose = () => {
      isConnected.value = false
      addLog('WebSocket 连接已断开', 'warning')
    }
  } catch {
    toast.error('无法建立WebSocket连接')
  }
}

const disconnectWebSocket = () => {
  if (ws.value) {
    ws.value.close()
    ws.value = null
  }
}

const toggleConnection = () => {
  if (isConnected.value) {
    disconnectWebSocket()
    toast.info('已断开连接')
  } else {
    connectWebSocket()
  }
}

const clearLogs = () => {
  logs.value = []
  toast.info('日志已清空')
}

const getLogColor = (type: LogEntry['type']) => {
  switch (type) {
    case 'success':
      return 'text-success'
    case 'error':
      return 'text-destructive'
    case 'warning':
      return 'text-warning'
    default:
      return 'text-primary'
  }
}

onUnmounted(() => {
  disconnectWebSocket()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div
      v-motion
      :initial="{ opacity: 0, y: -20 }"
      :enter="{ opacity: 1, y: 0 }"
      class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4"
    >
      <div class="flex items-center gap-3">
        <h1 class="text-3xl font-display font-bold">实时日志</h1>
        <Badge :variant="isConnected ? 'success' : 'secondary'" class="gap-1">
          <span class="relative flex h-2 w-2">
            <span
              v-if="isConnected"
              class="animate-ping absolute inline-flex h-full w-full rounded-full bg-success opacity-75"
            />
            <span
              class="relative inline-flex rounded-full h-2 w-2"
              :class="isConnected ? 'bg-success' : 'bg-muted-foreground'"
            />
          </span>
          {{ isConnected ? '已连接' : '未连接' }}
        </Badge>
      </div>
      <div class="flex gap-2">
        <Button variant="outline" @click="clearLogs">
          <Trash2 class="w-4 h-4" />
          清空日志
        </Button>
        <Button :variant="isConnected ? 'destructive' : 'default'" @click="toggleConnection">
          <WifiOff v-if="isConnected" class="w-4 h-4" />
          <Wifi v-else class="w-4 h-4" />
          {{ isConnected ? '断开连接' : '连接' }}
        </Button>
      </div>
    </div>

    <!-- Console Card -->
    <Card
      v-motion
      :initial="{ opacity: 0, y: 20 }"
      :enter="{ opacity: 1, y: 0, transition: { delay: 100 } }"
      class="border-border/50"
    >
      <CardHeader class="border-b border-border/50 py-3">
        <CardTitle class="flex items-center gap-2 text-base">
          <Terminal class="w-4 h-4 text-primary" />
          控制台输出
        </CardTitle>
      </CardHeader>
      <CardContent class="p-0">
        <div ref="logConsole" class="bg-background rounded-b-lg p-4 h-[600px] overflow-y-auto font-mono text-sm">
          <div
            v-for="(log, index) in logs"
            :key="index"
            v-motion
            :initial="{ opacity: 0, x: -10 }"
            :enter="{ opacity: 1, x: 0 }"
            class="mb-1 leading-relaxed"
            :class="getLogColor(log.type)"
          >
            {{ log.message }}
          </div>
          <div v-if="!logs.length" class="text-muted-foreground text-center py-8">
            <Terminal class="w-12 h-12 mx-auto mb-4 opacity-50" />
            <p>{{ isConnected ? '等待日志输出...' : '未连接，请点击"连接"按钮' }}</p>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
