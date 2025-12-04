<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMotion } from '@vueuse/motion'
import { Cloud, Lock, User, ArrowRight, Loader2 } from 'lucide-vue-next'
import { useAuthStore } from '@/stores/auth'
import { toast } from '@/composables/useToast'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent } from '@/components/ui/card'

const router = useRouter()
const authStore = useAuthStore()

const form = ref({
  account: '',
  password: ''
})

const loading = ref(false)
const error = ref('')

const cardRef = ref<HTMLElement>()

useMotion(cardRef, {
  initial: { opacity: 0, y: 50, scale: 0.95 },
  enter: {
    opacity: 1,
    y: 0,
    scale: 1,
    transition: {
      duration: 600,
      ease: [0.16, 1, 0.3, 1]
    }
  }
})

const handleLogin = async () => {
  error.value = ''
  loading.value = true

  try {
    await authStore.login(form.value.account, form.value.password)
    toast.success('登录成功')
    router.push('/')
  } catch (err: any) {
    error.value = err.message || '登录失败，请检查账号密码'
    toast.error(error.value)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-4 relative overflow-hidden">
    <!-- Animated Background -->
    <div class="absolute inset-0 mesh-gradient" />
    <div class="absolute inset-0 grid-pattern opacity-30" />
    <div class="absolute inset-0 noise-overlay" />

    <!-- Floating Elements -->
    <div class="absolute top-1/4 left-1/4 w-64 h-64 bg-primary/10 rounded-full blur-3xl animate-pulse" />
    <div class="absolute bottom-1/4 right-1/4 w-96 h-96 bg-cyan-500/5 rounded-full blur-3xl animate-pulse delay-1000" />

    <!-- Login Card -->
    <Card ref="cardRef" class="relative z-10 w-full max-w-md glass border-border/50">
      <CardContent class="p-8">
        <!-- Logo -->
        <div class="text-center mb-8">
          <div
            v-motion
            :initial="{ scale: 0, rotate: -180 }"
            :enter="{ scale: 1, rotate: 0, transition: { delay: 200, duration: 500 } }"
            class="inline-flex items-center justify-center w-16 h-16 bg-gradient-to-br from-primary to-cyan-400 rounded-2xl mb-4 shadow-lg glow-primary"
          >
            <Cloud class="w-8 h-8 text-primary-foreground" />
          </div>
          <h1
            v-motion
            :initial="{ opacity: 0, y: 20 }"
            :enter="{ opacity: 1, y: 0, transition: { delay: 300 } }"
            class="text-3xl font-display font-bold text-gradient"
          >
            OCI Panel
          </h1>
          <p
            v-motion
            :initial="{ opacity: 0 }"
            :enter="{ opacity: 1, transition: { delay: 400 } }"
            class="text-muted-foreground mt-2"
          >
            Oracle Cloud Infrastructure 管理面板
          </p>
        </div>

        <!-- Form -->
        <form class="space-y-6" @submit.prevent="handleLogin">
          <div v-motion :initial="{ opacity: 0, x: -20 }" :enter="{ opacity: 1, x: 0, transition: { delay: 500 } }">
            <label class="block text-sm font-medium mb-2">
              <User class="w-4 h-4 inline mr-2 text-muted-foreground" />
              账号
            </label>
            <Input
              v-model="form.account"
              type="text"
              placeholder="请输入账号"
              required
              autocomplete="username"
              class="h-11 bg-secondary/50 border-border/50 focus:border-primary"
            />
          </div>

          <div v-motion :initial="{ opacity: 0, x: -20 }" :enter="{ opacity: 1, x: 0, transition: { delay: 600 } }">
            <label class="block text-sm font-medium mb-2">
              <Lock class="w-4 h-4 inline mr-2 text-muted-foreground" />
              密码
            </label>
            <Input
              v-model="form.password"
              type="password"
              placeholder="请输入密码"
              required
              autocomplete="current-password"
              class="h-11 bg-secondary/50 border-border/50 focus:border-primary"
            />
          </div>

          <div v-motion :initial="{ opacity: 0, y: 20 }" :enter="{ opacity: 1, y: 0, transition: { delay: 700 } }">
            <Button
              type="submit"
              :disabled="loading"
              :loading="loading"
              class="w-full h-11 text-base font-medium group"
            >
              <template v-if="!loading">
                登录
                <ArrowRight class="w-4 h-4 ml-2 transition-transform group-hover:translate-x-1" />
              </template>
              <template v-else>
                <Loader2 class="w-4 h-4 mr-2 animate-spin" />
                登录中...
              </template>
            </Button>
          </div>
        </form>

        <!-- Error Message -->
        <Transition
          enter-active-class="transition-all duration-300"
          leave-active-class="transition-all duration-300"
          enter-from-class="opacity-0 -translate-y-2"
          leave-to-class="opacity-0 -translate-y-2"
        >
          <div
            v-if="error"
            class="mt-4 p-3 bg-destructive/10 border border-destructive/30 rounded-lg text-destructive text-sm"
          >
            {{ error }}
          </div>
        </Transition>

        <!-- Footer -->
        <p
          v-motion
          :initial="{ opacity: 0 }"
          :enter="{ opacity: 1, transition: { delay: 800 } }"
          class="text-center text-xs text-muted-foreground mt-6"
        >
          安全登录 · 数据加密传输
        </p>
      </CardContent>
    </Card>
  </div>
</template>
