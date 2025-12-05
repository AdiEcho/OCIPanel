<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  Info,
  Database,
  RefreshCw,
  Github,
  FileText,
  Server,
  Code,
  Loader2,
  FolderOpen,
  Activity,
  Send,
  Play,
  Square,
  TestTube
} from 'lucide-vue-next'
import api from '@/lib/api'
import { toast } from '@/composables/useToast'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'

const loading = ref(false)
const refreshing = ref(false)

const config = ref({
  keyDirPath: '',
  logLevel: ''
})

const cacheConfig = ref({
  cacheEnabled: false,
  cacheInterval: 30
})

const telegramConfig = ref({
  botToken: '',
  chatId: '',
  enabled: false,
  running: false
})

const telegramLoading = ref(false)
const testingConnection = ref(false)
const sendingTestMessage = ref(false)

const loadConfig = async () => {
  loading.value = true
  try {
    const response = await api.post('/sys/getSysCfg', {})
    if (response.data) {
      config.value = response.data
      cacheConfig.value.cacheEnabled = response.data.cacheEnabled || false
      cacheConfig.value.cacheInterval = response.data.cacheInterval || 30
    }
  } catch {
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
  } catch {
    toast.error('更新缓存配置失败')
  }
}

const refreshCache = async () => {
  refreshing.value = true
  try {
    await api.post('/sys/refreshCache', {})
    toast.success('缓存刷新任务已启动')
  } catch {
    toast.error('刷新缓存失败')
  } finally {
    refreshing.value = false
  }
}

const loadTelegramConfig = async () => {
  telegramLoading.value = true
  try {
    const response = await api.post('/telegram/getConfig', {})
    if (response.data) {
      telegramConfig.value = response.data
    }
  } catch {
    toast.error('加载 Telegram 配置失败')
  } finally {
    telegramLoading.value = false
  }
}

const updateTelegramConfig = async () => {
  try {
    await api.post('/telegram/updateConfig', {
      botToken: telegramConfig.value.botToken,
      chatId: telegramConfig.value.chatId,
      enabled: telegramConfig.value.enabled
    })
    toast.success('Telegram 配置已更新')
    await loadTelegramConfig()
  } catch {
    toast.error('更新 Telegram 配置失败')
  }
}

const testTelegramConnection = async () => {
  testingConnection.value = true
  try {
    await api.post('/telegram/testConnection', {})
    toast.success('连接测试成功')
  } catch {
    toast.error('连接测试失败，请检查 Bot Token')
  } finally {
    testingConnection.value = false
  }
}

const sendTelegramTestMessage = async () => {
  sendingTestMessage.value = true
  try {
    await api.post('/telegram/sendTestMessage', {})
    toast.success('测试消息发送成功')
  } catch {
    toast.error('发送测试消息失败')
  } finally {
    sendingTestMessage.value = false
  }
}

const startTelegramBot = async () => {
  try {
    await api.post('/telegram/startBot', {})
    toast.success('Telegram Bot 已启动')
    await loadTelegramConfig()
  } catch {
    toast.error('启动 Bot 失败')
  }
}

const stopTelegramBot = async () => {
  try {
    await api.post('/telegram/stopBot', {})
    toast.success('Telegram Bot 已停止')
    await loadTelegramConfig()
  } catch {
    toast.error('停止 Bot 失败')
  }
}

onMounted(() => {
  loadConfig()
  loadTelegramConfig()
})

const systemInfo = [
  { label: '应用名称', value: 'OCI Panel', icon: Server },
  { label: '版本号', value: 'v1.0.0', icon: Info },
  { label: '后端框架', value: 'Gin (Go)', icon: Code },
  { label: '前端框架', value: 'Vue 3 + Vite + Tailwind CSS', icon: Code },
  { label: '数据库', value: 'SQLite', icon: Database }
]
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div v-motion :initial="{ opacity: 0, y: -20 }" :enter="{ opacity: 1, y: 0 }">
      <h1 class="text-3xl font-display font-bold">系统设置</h1>
      <p class="text-muted-foreground mt-1">管理系统配置和查看系统信息</p>
    </div>

    <div class="grid gap-6">
      <!-- System Info Card -->
      <Card
        v-motion
        :initial="{ opacity: 0, y: 20 }"
        :enter="{ opacity: 1, y: 0, transition: { delay: 100 } }"
        class="border-border/50"
      >
        <CardHeader class="border-b border-border/50">
          <CardTitle class="flex items-center gap-2">
            <Info class="w-5 h-5 text-primary" />
            系统信息
          </CardTitle>
        </CardHeader>
        <CardContent class="p-0">
          <div class="divide-y divide-border/50">
            <div
              v-for="(item, index) in systemInfo"
              :key="item.label"
              v-motion
              :initial="{ opacity: 0, x: -20 }"
              :enter="{ opacity: 1, x: 0, transition: { delay: 150 + index * 50 } }"
              class="flex items-center justify-between px-6 py-4"
            >
              <div class="flex items-center gap-3">
                <div class="w-8 h-8 rounded bg-secondary flex items-center justify-center">
                  <component :is="item.icon" class="w-4 h-4 text-muted-foreground" />
                </div>
                <span class="text-muted-foreground">{{ item.label }}</span>
              </div>
              <span class="font-medium">{{ item.value }}</span>
            </div>
            <div class="flex items-center justify-between px-6 py-4">
              <div class="flex items-center gap-3">
                <div class="w-8 h-8 rounded bg-success/10 flex items-center justify-center">
                  <Activity class="w-4 h-4 text-success" />
                </div>
                <span class="text-muted-foreground">运行状态</span>
              </div>
              <Badge variant="success">正常运行</Badge>
            </div>
          </div>
        </CardContent>
      </Card>

      <!-- Cache Settings Card -->
      <Card
        v-motion
        :initial="{ opacity: 0, y: 20 }"
        :enter="{ opacity: 1, y: 0, transition: { delay: 200 } }"
        class="border-border/50"
      >
        <CardHeader class="border-b border-border/50">
          <CardTitle class="flex items-center gap-2">
            <Database class="w-5 h-5 text-primary" />
            缓存设置
          </CardTitle>
        </CardHeader>
        <CardContent class="p-0 divide-y divide-border/50">
          <div class="flex items-center justify-between px-6 py-4">
            <div>
              <p class="font-medium">启用数据缓存</p>
              <p class="text-sm text-muted-foreground mt-1">
                启用后将定时缓存配置的实例数据到数据库，减少对OCI API的请求
              </p>
            </div>
            <Switch v-model="cacheConfig.cacheEnabled" @update:model-value="updateCacheConfig" />
          </div>
          <div class="flex items-center justify-between px-6 py-4">
            <div>
              <p class="font-medium">缓存刷新间隔</p>
              <p class="text-sm text-muted-foreground mt-1">定时任务检查并更新缓存的间隔时间（分钟）</p>
            </div>
            <div class="flex items-center gap-2">
              <Input
                v-model.number="cacheConfig.cacheInterval"
                type="number"
                min="5"
                max="1440"
                class="w-20 text-center"
                :disabled="!cacheConfig.cacheEnabled"
                @change="updateCacheConfig"
              />
              <span class="text-muted-foreground">分钟</span>
            </div>
          </div>
          <div class="flex items-center justify-between px-6 py-4">
            <div>
              <p class="font-medium">手动刷新缓存</p>
              <p class="text-sm text-muted-foreground mt-1">立即更新所有配置的缓存数据</p>
            </div>
            <Button :disabled="!cacheConfig.cacheEnabled || refreshing" @click="refreshCache">
              <RefreshCw v-if="!refreshing" class="w-4 h-4" />
              <Loader2 v-else class="w-4 h-4 animate-spin" />
              {{ refreshing ? '刷新中...' : '立即刷新' }}
            </Button>
          </div>
        </CardContent>
      </Card>

      <!-- Telegram Settings Card -->
      <Card
        v-motion
        :initial="{ opacity: 0, y: 20 }"
        :enter="{ opacity: 1, y: 0, transition: { delay: 300 } }"
        class="border-border/50"
      >
        <CardHeader class="border-b border-border/50">
          <CardTitle class="flex items-center gap-2">
            <Send class="w-5 h-5 text-primary" />
            Telegram 通知
          </CardTitle>
        </CardHeader>
        <CardContent class="p-0 divide-y divide-border/50">
          <div v-if="telegramLoading" class="text-center py-8">
            <Loader2 class="w-8 h-8 mx-auto animate-spin text-primary" />
          </div>
          <template v-else>
            <div class="flex items-center justify-between px-6 py-4">
              <div>
                <p class="font-medium">启用 Telegram 通知</p>
                <p class="text-sm text-muted-foreground mt-1">接收任务执行结果和系统通知</p>
              </div>
              <Switch v-model="telegramConfig.enabled" @update:model-value="updateTelegramConfig" />
            </div>
            <div class="px-6 py-4 space-y-4">
              <div>
                <label class="text-sm font-medium mb-2 block">Bot Token</label>
                <Input
                  v-model="telegramConfig.botToken"
                  placeholder="输入 Telegram Bot Token"
                  class="font-mono text-sm"
                  @blur="updateTelegramConfig"
                />
                <p class="text-xs text-muted-foreground mt-1">从 @BotFather 获取</p>
              </div>
              <div>
                <label class="text-sm font-medium mb-2 block">Chat ID</label>
                <Input
                  v-model="telegramConfig.chatId"
                  placeholder="输入您的 Telegram Chat ID"
                  class="font-mono"
                  @blur="updateTelegramConfig"
                />
                <p class="text-xs text-muted-foreground mt-1">从 @userinfobot 获取</p>
              </div>
            </div>
            <div class="flex items-center justify-between px-6 py-4">
              <div>
                <p class="font-medium">Bot 运行状态</p>
                <p class="text-sm text-muted-foreground mt-1">
                  Bot 启动后可通过 /start 命令与 Bot 交互
                </p>
              </div>
              <div class="flex items-center gap-2">
                <Badge :variant="telegramConfig.running ? 'success' : 'secondary'">
                  {{ telegramConfig.running ? '运行中' : '已停止' }}
                </Badge>
                <Button
                  v-if="telegramConfig.running"
                  variant="outline"
                  size="sm"
                  @click="stopTelegramBot"
                >
                  <Square class="w-4 h-4" />
                  停止
                </Button>
                <Button
                  v-else
                  variant="outline"
                  size="sm"
                  :disabled="!telegramConfig.enabled || !telegramConfig.botToken || !telegramConfig.chatId"
                  @click="startTelegramBot"
                >
                  <Play class="w-4 h-4" />
                  启动
                </Button>
              </div>
            </div>
            <div class="flex items-center justify-between px-6 py-4">
              <div>
                <p class="font-medium">测试功能</p>
                <p class="text-sm text-muted-foreground mt-1">测试连接或发送测试消息</p>
              </div>
              <div class="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  :disabled="!telegramConfig.botToken || testingConnection"
                  @click="testTelegramConnection"
                >
                  <Loader2 v-if="testingConnection" class="w-4 h-4 animate-spin" />
                  <TestTube v-else class="w-4 h-4" />
                  测试连接
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  :disabled="!telegramConfig.enabled || !telegramConfig.botToken || !telegramConfig.chatId || sendingTestMessage"
                  @click="sendTelegramTestMessage"
                >
                  <Loader2 v-if="sendingTestMessage" class="w-4 h-4 animate-spin" />
                  <Send v-else class="w-4 h-4" />
                  发送测试
                </Button>
              </div>
            </div>
          </template>
        </CardContent>
      </Card>

      <!-- System Configuration Card -->
      <Card
        v-motion
        :initial="{ opacity: 0, y: 20 }"
        :enter="{ opacity: 1, y: 0, transition: { delay: 400 } }"
        class="border-border/50"
      >
        <CardHeader class="border-b border-border/50">
          <CardTitle class="flex items-center gap-2">
            <FolderOpen class="w-5 h-5 text-primary" />
            系统配置
          </CardTitle>
        </CardHeader>
        <CardContent class="p-0 divide-y divide-border/50">
          <div v-if="loading" class="text-center py-8">
            <Loader2 class="w-8 h-8 mx-auto animate-spin text-primary" />
          </div>
          <template v-else>
            <div class="flex items-center justify-between px-6 py-4">
              <span class="text-muted-foreground">密钥目录</span>
              <span class="font-mono text-sm">{{ config.keyDirPath || 'N/A' }}</span>
            </div>
            <div class="flex items-center justify-between px-6 py-4">
              <span class="text-muted-foreground">日志级别</span>
              <span class="font-medium">{{ config.logLevel || 'N/A' }}</span>
            </div>
          </template>
        </CardContent>
      </Card>

      <!-- About Card -->
      <Card
        v-motion
        :initial="{ opacity: 0, y: 20 }"
        :enter="{ opacity: 1, y: 0, transition: { delay: 500 } }"
        class="border-border/50"
      >
        <CardHeader class="border-b border-border/50">
          <CardTitle class="flex items-center gap-2">
            <Info class="w-5 h-5 text-primary" />
            关于
          </CardTitle>
        </CardHeader>
        <CardContent class="py-6">
          <p class="text-muted-foreground leading-relaxed mb-6">
            OCI Panel 是一个基于 Go + Vue 3 开发的 Oracle Cloud Infrastructure 管理面板，
            提供实例管理、网络配置、任务调度等功能，帮助用户更便捷地管理 OCI 资源。
          </p>
          <div class="flex gap-3">
            <Button variant="outline" as="a" href="https://github.com" target="_blank">
              <Github class="w-4 h-4" />
              GitHub
            </Button>
            <Button variant="outline" as="a" href="https://docs.oracle.com/iaas" target="_blank">
              <FileText class="w-4 h-4" />
              OCI文档
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
