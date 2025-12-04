import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '@/lib/api'

export interface User {
  username: string
  account: string
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const user = ref<User | null>(JSON.parse(localStorage.getItem('user') || 'null'))

  const isAuthenticated = computed(() => !!token.value)

  async function login(account: string, password: string): Promise<boolean> {
    const response = await api.post('/sys/login', { account, password })
    token.value = response.data.token
    user.value = {
      username: response.data.username,
      account: account
    }

    localStorage.setItem('token', token.value)
    localStorage.setItem('user', JSON.stringify(user.value))

    return true
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  function checkAuth() {
    if (!token.value) {
      logout()
    }
  }

  return {
    token,
    user,
    isAuthenticated,
    login,
    logout,
    checkAuth
  }
})
