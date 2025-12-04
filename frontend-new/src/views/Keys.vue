<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Plus, Search, Trash2, Edit, Copy, ChevronLeft, ChevronRight, Loader2, Key } from 'lucide-vue-next'
import api from '@/lib/api'
import { toast } from '@/composables/useToast'
import { copyToClipboard, truncateId } from '@/lib/utils'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import { Select } from '@/components/ui/select'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table'
import { Dialog, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog'

interface KeyItem {
  id: number
  name: string
  keyType: string
  publicKey: string
  configName?: string
  createTime: string
}

const keys = ref<KeyItem[]>([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = 10
const totalPages = ref(0)
const searchText = ref('')
const filterType = ref('')

const showAddModal = ref(false)
const editingKey = ref<KeyItem | null>(null)
const submitting = ref(false)

const form = ref({
  name: '',
  publicKey: ''
})

const loadKeys = async (page = 1) => {
  loading.value = true
  try {
    const response = await api.post('/key/list', {
      page,
      pageSize,
      name: searchText.value,
      keyType: filterType.value
    })
    keys.value = response.data.list || []
    currentPage.value = response.data.page
    totalPages.value = Math.ceil(response.data.total / pageSize)
  } catch (error: any) {
    toast.error(error.message || '加载失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => loadKeys(1)

const closeModal = () => {
  showAddModal.value = false
  editingKey.value = null
  form.value = { name: '', publicKey: '' }
}

const submitForm = async () => {
  submitting.value = true
  try {
    if (editingKey.value) {
      await api.post('/key/update', {
        id: editingKey.value.id,
        name: form.value.name,
        publicKey: form.value.publicKey
      })
      toast.success('更新成功')
    } else {
      await api.post('/key/create', {
        name: form.value.name,
        publicKey: form.value.publicKey
      })
      toast.success('添加成功')
    }
    closeModal()
    await loadKeys(currentPage.value)
  } catch (error: any) {
    toast.error(error.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

const editKey = (key: KeyItem) => {
  editingKey.value = key
  form.value = {
    name: key.name,
    publicKey: key.publicKey
  }
  showAddModal.value = true
}

const deleteKey = async (id: number) => {
  if (!confirm('确定要删除此密钥吗？')) return
  try {
    await api.post('/key/delete', { ids: [id] })
    toast.success('删除成功')
    await loadKeys(currentPage.value)
  } catch (error: any) {
    toast.error(error.message || '删除失败')
  }
}

const handleCopy = async (text: string) => {
  try {
    await copyToClipboard(text)
    toast.success('已复制到剪贴板')
  } catch {
    toast.error('复制失败')
  }
}

onMounted(() => {
  loadKeys()
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
      <h1 class="text-3xl font-display font-bold">密钥管理</h1>
      <Button @click="showAddModal = true">
        <Plus class="w-4 h-4" />
        添加公钥
      </Button>
    </div>

    <!-- Main Card -->
    <Card
      v-motion
      :initial="{ opacity: 0, y: 20 }"
      :enter="{ opacity: 1, y: 0, transition: { delay: 100 } }"
      class="border-border/50"
    >
      <!-- Search -->
      <CardHeader class="border-b border-border/50 pb-4">
        <div class="flex gap-4">
          <div class="relative max-w-md flex-1">
            <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <Input v-model="searchText" placeholder="搜索密钥名..." class="pl-10" @input="handleSearch" />
          </div>
          <Select v-model="filterType" class="w-40" @change="loadKeys(1)">
            <option value="">全部类型</option>
            <option value="standalone">独立上传</option>
            <option value="config">配置关联</option>
          </Select>
        </div>
      </CardHeader>

      <!-- Table -->
      <CardContent class="p-0">
        <Table>
          <TableHeader>
            <TableRow class="hover:bg-transparent">
              <TableHead>名称</TableHead>
              <TableHead>类型</TableHead>
              <TableHead>公钥</TableHead>
              <TableHead>关联配置</TableHead>
              <TableHead>创建时间</TableHead>
              <TableHead class="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="loading">
              <TableCell colspan="6" class="h-32 text-center">
                <Loader2 class="w-8 h-8 mx-auto animate-spin text-primary" />
              </TableCell>
            </TableRow>

            <TableRow v-else-if="!keys.length">
              <TableCell colspan="6" class="h-32 text-center text-muted-foreground">暂无密钥</TableCell>
            </TableRow>

            <TableRow
              v-for="(key, index) in keys"
              v-else
              :key="key.id"
              v-motion
              :initial="{ opacity: 0, x: -20 }"
              :enter="{ opacity: 1, x: 0, transition: { delay: 50 * index } }"
              class="group"
            >
              <TableCell class="font-medium">
                <div class="flex items-center gap-2">
                  <div class="w-8 h-8 rounded bg-primary/10 flex items-center justify-center">
                    <Key class="w-4 h-4 text-primary" />
                  </div>
                  {{ key.name }}
                </div>
              </TableCell>
              <TableCell>
                <Badge :variant="key.keyType === 'standalone' ? 'info' : 'secondary'">
                  {{ key.keyType === 'standalone' ? '独立上传' : '配置关联' }}
                </Badge>
              </TableCell>
              <TableCell>
                <div class="flex items-center gap-2">
                  <span class="font-mono text-xs text-muted-foreground">
                    {{ truncateId(key.publicKey, 40) }}
                  </span>
                  <Button
                    variant="ghost"
                    size="icon"
                    class="h-6 w-6"
                    title="复制公钥"
                    @click="handleCopy(key.publicKey)"
                  >
                    <Copy class="w-3 h-3" />
                  </Button>
                </div>
              </TableCell>
              <TableCell class="text-muted-foreground">{{ key.configName || '-' }}</TableCell>
              <TableCell class="text-muted-foreground text-sm">{{ key.createTime }}</TableCell>
              <TableCell class="text-right">
                <div class="flex justify-end gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  <Button
                    v-if="key.keyType === 'standalone'"
                    variant="ghost"
                    size="icon"
                    title="编辑"
                    @click="editKey(key)"
                  >
                    <Edit class="w-4 h-4" />
                  </Button>
                  <Button
                    v-if="key.keyType === 'standalone'"
                    variant="ghost"
                    size="icon"
                    title="删除"
                    class="text-destructive hover:text-destructive"
                    @click="deleteKey(key.id)"
                  >
                    <Trash2 class="w-4 h-4" />
                  </Button>
                </div>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>

        <!-- Pagination -->
        <div v-if="totalPages > 1" class="flex items-center justify-center gap-2 p-4 border-t border-border/50">
          <Button variant="outline" size="sm" :disabled="currentPage === 1" @click="loadKeys(currentPage - 1)">
            <ChevronLeft class="w-4 h-4" />
            上一页
          </Button>
          <span class="px-4 text-sm text-muted-foreground">第 {{ currentPage }} / {{ totalPages }} 页</span>
          <Button variant="outline" size="sm" :disabled="currentPage === totalPages" @click="loadKeys(currentPage + 1)">
            下一页
            <ChevronRight class="w-4 h-4" />
          </Button>
        </div>
      </CardContent>
    </Card>

    <!-- Add/Edit Modal -->
    <Dialog v-model:open="showAddModal">
      <DialogHeader class="mb-4">
        <DialogTitle>{{ editingKey ? '编辑公钥' : '添加公钥' }}</DialogTitle>
        <DialogDescription>
          {{ editingKey ? '修改现有SSH公钥信息' : '添加新的SSH公钥用于实例创建' }}
        </DialogDescription>
      </DialogHeader>

      <form class="space-y-4" @submit.prevent="submitForm">
        <div>
          <label class="block text-sm font-medium mb-2">密钥名称</label>
          <Input v-model="form.name" placeholder="例: 我的SSH公钥" required />
        </div>

        <div>
          <label class="block text-sm font-medium mb-2">公钥内容</label>
          <Textarea
            v-model="form.publicKey"
            :rows="6"
            class="font-mono text-sm"
            placeholder="ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC..."
            required
          />
          <p class="text-xs text-muted-foreground mt-2">粘贴您的 SSH 公钥内容（通常位于 ~/.ssh/id_rsa.pub）</p>
        </div>

        <DialogFooter class="mt-6">
          <Button type="button" variant="outline" @click="closeModal">取消</Button>
          <Button type="submit" :loading="submitting">
            {{ submitting ? '提交中...' : '提交' }}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  </div>
</template>
