<script setup lang="ts">
import { ref, computed, onMounted, reactive, watch } from 'vue'
import { useMotion } from '@vueuse/motion'
import { RouterLink } from 'vue-router'
import {
  Plus,
  Search,
  Trash2,
  Edit,
  Eye,
  Server,
  RefreshCw,
  Upload,
  X,
  ChevronLeft,
  ChevronRight,
  Loader2,
  Play,
  Square,
  RotateCcw,
  Globe,
  Settings,
  Terminal,
  HardDrive,
  Network,
  Shield,
  Check,
  Copy,
  MoreHorizontal,
  BarChart3,
  Users,
  KeyRound,
  LockKeyhole,
  ShieldOff,
  Mail
} from 'lucide-vue-next'
import api from '@/lib/api'
import { toast } from '@/composables/useToast'
import { formatFileSize } from '@/lib/utils'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import { Switch } from '@/components/ui/switch'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table'
import { Dialog, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog'
import { Dropdown, DropdownItem } from '@/components/ui/dropdown'

import EditInstanceModal from './configs/components/EditInstanceModal.vue'
import CloudShellModal from './configs/components/CloudShellModal.vue'
import VolumeEditModal from './configs/components/VolumeEditModal.vue'
import SecurityListModal from './configs/components/SecurityListModal.vue'

interface Config {
  id: number
  username: string
  tenantName?: string
  tenantCreateTime?: string
  ociRegion: string
  ociUserId?: string
  ociFingerprint?: string
  ociTenantId?: string
  ociKeyPath?: string
  instanceCount?: number
  runningInstances?: number
}

interface Instance {
  id: string
  displayName: string
  state: string
  shape: string
  ocpus: number
  memory: number
  bootVolumeSize?: number
  bootVolumeVpu?: number
  region: string
  publicIps?: string[]
  ipv6?: string
  imageName?: string
}

interface SSHKey {
  id: number
  name: string
}
interface Image {
  id: string
  operatingSystem: string
  operatingSystemVersion: string
}

// 配置列表状态
const configs = ref<Config[]>([])
const loading = ref(false)
const searchText = ref('')
const currentPage = ref(1)
const pageSize = 10
const totalPages = ref(0)
const selectedConfigIds = ref<number[]>([])

// 弹窗状态
const showAddModal = ref(false)
const showCreateInstanceModal = ref(false)
const showBatchCreateModal = ref(false)
const showConfigDetailsSidebar = ref(false)
const editingConfig = ref<Config | null>(null)
const selectedConfigForInstance = ref<Config | null>(null)
const configDetails = ref<any>(null)

// 子组件弹窗状态
const showEditInstanceModal = ref(false)
const showCloudShellModal = ref(false)
const showVolumeEditModal = ref(false)
const showSecurityListModal = ref(false)
const selectedInstance = ref<Instance | null>(null)
const selectedVolume = ref<any>(null)
const selectedVcn = ref<any>(null)
const cloudShellInstanceId = ref('')

// 加载状态
const submitting = ref(false)
const submittingInstance = ref(false)
const loadingDetails = ref(false)
const loadingTab = ref(false)
const loadingImages = ref(false)
const loadingTraffic = ref(false)

// 密码过期时间编辑
const editingPasswordExpiry = ref(false)
const passwordExpiryInput = ref(0)
const updatingPasswordExpiry = ref(false)

// 用户编辑
const showEditUserModal = ref(false)
const editingUser = ref<any>(null)
const userForm = ref({ email: '', dbUserName: '', description: '' })

// 文件上传
const isDragging = ref(false)
const uploadedFile = ref<File | null>(null)
const fileInput = ref<HTMLInputElement>()

// 标签页状态
const activeTab = ref('basic')
const tabInstances = ref<Instance[]>([])
const tabVolumes = ref<any[]>([])
const tabVCNs = ref<any[]>([])
const tabTenant = ref<any>(null)
const tabTraffic = ref<{ time: string[]; inbound: string[]; outbound: string[] }>({
  time: [],
  inbound: [],
  outbound: []
})
const instanceActionLoading = reactive<Record<string, boolean>>({})

// 流量查询
const trafficCondition = ref<{ instances: { value: string; label: string }[] }>({ instances: [] })
const trafficVnics = ref<{ value: string; label: string }[]>([])
const trafficForm = ref({ instanceId: '', vnicId: '', startTime: '', endTime: '' })

// 表单
const form = ref({ username: '', configContent: '' })
const instanceForm = ref({
  ociRegion: '',
  ocpus: 1,
  memory: 6,
  disk: 50,
  bootVolumeVpu: 10,
  architecture: 'ARM',
  operationSystem: 'Ubuntu',
  imageId: '',
  sshKeyId: '',
  interval: 60,
  isTaskMode: true
})
const sshKeys = ref<SSHKey[]>([])
const availableImages = ref<Image[]>([])
const filteredImages = ref<Image[]>([])

// 计算属性
const isAllSelected = computed(
  () => configs.value.length > 0 && selectedConfigIds.value.length === configs.value.length
)
const isIndeterminate = computed(
  () => selectedConfigIds.value.length > 0 && selectedConfigIds.value.length < configs.value.length
)

// 工具函数
const parseConfigContent = (content: string) => {
  const config: Record<string, string> = {}
  content.split('\n').forEach(line => {
    const [key, value] = line.split('=').map(s => s.trim())
    if (key && value) config[key.toLowerCase()] = value
  })
  return {
    ociTenantId: config['tenancy'] || '',
    ociUserId: config['user'] || '',
    ociFingerprint: config['fingerprint'] || '',
    ociRegion: config['region'] || ''
  }
}

const copyToClipboard = (text: string) => {
  navigator.clipboard
    .writeText(text)
    .then(() => toast.success('已复制到剪贴板'))
    .catch(() => toast.error('复制失败'))
}

const formatDateTime = (date: Date) => {
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

// API 操作
const loadConfigs = async (page = 1) => {
  loading.value = true
  try {
    const response = await api.post('/oci/userPage', { page, pageSize, username: searchText.value })
    configs.value = response.data.list || []
    currentPage.value = response.data.page
    totalPages.value = Math.ceil(response.data.total / pageSize)
  } catch (error: any) {
    toast.error(error.message || '加载失败')
  } finally {
    loading.value = false
  }
}

const loadSSHKeys = async () => {
  try {
    const response = await api.get('/key/standalone')
    sshKeys.value = response.data || []
  } catch {
    sshKeys.value = []
  }
}

const loadImages = async (configId: number, region: string, architecture: string) => {
  if (!configId || !region || !architecture) return
  loadingImages.value = true
  try {
    const response = await api.post('/oci/images', { configId, region, architecture })
    availableImages.value = response.data || []
    filterImagesByOS()
  } catch {
    availableImages.value = []
    filteredImages.value = []
  } finally {
    loadingImages.value = false
  }
}

const filterImagesByOS = () => {
  const os = instanceForm.value.operationSystem.toLowerCase()
  filteredImages.value = availableImages.value.filter(img =>
    img.operatingSystem.toLowerCase().includes(os === 'oracle linux' ? 'oracle' : os)
  )
  instanceForm.value.imageId = filteredImages.value.length > 0 ? filteredImages.value[0].id : ''
}

// 配置操作
const handleSearch = () => {
  selectedConfigIds.value = []
  loadConfigs(1)
}
const toggleSelectConfig = (id: number) => {
  const index = selectedConfigIds.value.indexOf(id)
  if (index === -1) selectedConfigIds.value.push(id)
  else selectedConfigIds.value.splice(index, 1)
}
const toggleSelectAll = () => {
  if (isAllSelected.value) selectedConfigIds.value = []
  else selectedConfigIds.value = configs.value.map(c => c.id)
}
const closeModal = () => {
  showAddModal.value = false
  editingConfig.value = null
  form.value = { username: '', configContent: '' }
  uploadedFile.value = null
}
const handleFileDrop = (e: DragEvent) => {
  isDragging.value = false
  if (e.dataTransfer?.files?.length) uploadedFile.value = e.dataTransfer.files[0]
}
const handleFileSelect = (e: Event) => {
  const target = e.target as HTMLInputElement
  if (target.files?.length) uploadedFile.value = target.files[0]
}
const clearFile = () => {
  uploadedFile.value = null
  if (fileInput.value) fileInput.value.value = ''
}

const submitForm = async () => {
  if (!uploadedFile.value && !editingConfig.value) {
    toast.error('请选择密钥文件')
    return
  }
  submitting.value = true
  try {
    let keyPath = ''
    if (uploadedFile.value) {
      const formData = new FormData()
      formData.append('file', uploadedFile.value)
      const uploadResponse = await api.post('/oci/uploadKey', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
      keyPath = uploadResponse.data
    }
    const parsedConfig = parseConfigContent(form.value.configContent)
    if (editingConfig.value) {
      await api.post('/oci/updateCfgName', {
        id: editingConfig.value.id,
        username: form.value.username,
        ociKeyPath: uploadedFile.value ? keyPath : undefined
      })
      toast.success('配置已更新')
    } else {
      await api.post('/oci/addCfg', {
        username: form.value.username,
        tenantName: form.value.username,
        ...parsedConfig,
        ociKeyPath: keyPath
      })
      toast.success('配置已添加')
    }
    closeModal()
    await loadConfigs(currentPage.value)
  } catch (error: any) {
    toast.error(error.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

const editConfig = (config: Config) => {
  editingConfig.value = config
  form.value = {
    username: config.username,
    configContent: `user=${config.ociUserId || ''}\nfingerprint=${config.ociFingerprint || ''}\ntenancy=${config.ociTenantId || ''}\nregion=${config.ociRegion || ''}`
  }
  showAddModal.value = true
}

const deleteConfig = async (id: number) => {
  if (!confirm('确定要删除此配置吗？')) return
  try {
    await api.post('/oci/removeCfg', { ids: [id] })
    toast.success('配置已删除')
    await loadConfigs(currentPage.value)
  } catch (error: any) {
    toast.error(error.message || '删除失败')
  }
}

const batchDeleteConfigs = async () => {
  if (!confirm(`确定要删除选中的 ${selectedConfigIds.value.length} 个配置吗？`)) return
  try {
    await api.post('/oci/removeCfg', { ids: selectedConfigIds.value })
    toast.success(`成功删除 ${selectedConfigIds.value.length} 个配置`)
    selectedConfigIds.value = []
    await loadConfigs(currentPage.value)
  } catch (error: any) {
    toast.error(error.message || '批量删除失败')
  }
}

// 创建实例
const createInstance = async (config: Config) => {
  selectedConfigForInstance.value = config
  instanceForm.value.ociRegion = config.ociRegion
  await loadSSHKeys()
  await loadImages(config.id, config.ociRegion, instanceForm.value.architecture)
  showCreateInstanceModal.value = true
}
const closeInstanceModal = () => {
  showCreateInstanceModal.value = false
  selectedConfigForInstance.value = null
  availableImages.value = []
  filteredImages.value = []
}
const onArchitectureChange = () => {
  if (selectedConfigForInstance.value)
    loadImages(selectedConfigForInstance.value.id, instanceForm.value.ociRegion, instanceForm.value.architecture)
}
const onOperationSystemChange = () => {
  filterImagesByOS()
}

const submitInstanceTask = async () => {
  if (!instanceForm.value.sshKeyId) {
    toast.warning('请选择SSH公钥')
    return
  }
  submittingInstance.value = true
  try {
    await api.post('/task/create', {
      userId: selectedConfigForInstance.value?.id,
      ociRegion: instanceForm.value.ociRegion,
      ocpus: instanceForm.value.ocpus,
      memory: instanceForm.value.memory,
      disk: instanceForm.value.disk,
      bootVolumeVpu: instanceForm.value.bootVolumeVpu,
      architecture: instanceForm.value.architecture,
      operationSystem: instanceForm.value.operationSystem,
      imageId: instanceForm.value.imageId || undefined,
      sshKeyId: instanceForm.value.sshKeyId,
      interval: instanceForm.value.interval || 60,
      executeOnce: !instanceForm.value.isTaskMode
    })
    toast.success(instanceForm.value.isTaskMode ? '任务已创建，可在任务列表查看' : '实例创建请求已提交')
    closeInstanceModal()
  } catch (error: any) {
    toast.error(error.message || '创建失败')
  } finally {
    submittingInstance.value = false
  }
}

// 批量创建实例
const batchCreateInstance = async () => {
  if (selectedConfigIds.value.length === 0) {
    toast.warning('请先选择配置')
    return
  }
  await loadSSHKeys()
  showBatchCreateModal.value = true
}
const closeBatchCreateModal = () => {
  showBatchCreateModal.value = false
}

const submitBatchInstanceTask = async () => {
  if (!instanceForm.value.sshKeyId) {
    toast.warning('请选择SSH公钥')
    return
  }
  submittingInstance.value = true
  try {
    for (const configId of selectedConfigIds.value) {
      const config = configs.value.find(c => c.id === configId)
      await api.post('/task/create', {
        userId: configId,
        ociRegion: config?.ociRegion || instanceForm.value.ociRegion,
        ocpus: instanceForm.value.ocpus,
        memory: instanceForm.value.memory,
        disk: instanceForm.value.disk,
        architecture: instanceForm.value.architecture,
        operationSystem: instanceForm.value.operationSystem,
        sshKeyId: instanceForm.value.sshKeyId,
        interval: instanceForm.value.interval || 60
      })
    }
    toast.success(`已为 ${selectedConfigIds.value.length} 个配置创建定时任务`)
    closeBatchCreateModal()
    selectedConfigIds.value = []
  } catch (error: any) {
    toast.error(error.message || '批量创建失败')
  } finally {
    submittingInstance.value = false
  }
}

// 配置详情
const viewConfigDetails = async (config: Config) => {
  activeTab.value = 'basic'
  tabInstances.value = []
  tabVolumes.value = []
  tabVCNs.value = []
  tabTenant.value = null
  tabTraffic.value = { time: [], inbound: [], outbound: [] }
  trafficCondition.value = { instances: [] }
  trafficVnics.value = []
  trafficForm.value = { instanceId: '', vnicId: '', startTime: '', endTime: '' }
  showConfigDetailsSidebar.value = true
  loadingDetails.value = true
  try {
    const response = await api.post('/oci/details', { configId: config.id })
    configDetails.value = response.data
    await loadTenant()
  } catch (error: any) {
    toast.error(error.message || '加载配置详情失败')
    closeConfigDetailsSidebar()
  } finally {
    loadingDetails.value = false
  }
}

const closeConfigDetailsSidebar = () => {
  showConfigDetailsSidebar.value = false
  configDetails.value = null
  tabInstances.value = []
  tabVolumes.value = []
  tabVCNs.value = []
  tabTenant.value = null
  tabTraffic.value = { time: [], inbound: [], outbound: [] }
}

// 标签页数据加载
const loadTenant = async (clearCache = false) => {
  if (!configDetails.value) return
  loadingTab.value = true
  try {
    const response = await api.post('/oci/tenant/info', { configId: configDetails.value.userId, clearCache })
    tabTenant.value = response.data
  } catch (error: any) {
    toast.error(error.message || '加载租户详情失败')
    tabTenant.value = null
  } finally {
    loadingTab.value = false
  }
}

const loadInstances = async (clearCache = false) => {
  if (!configDetails.value) return
  loadingTab.value = true
  try {
    const response = await api.post('/oci/details/instances', { configId: configDetails.value.userId, clearCache })
    tabInstances.value = response.data || []
  } catch (error: any) {
    toast.error(error.message || '加载实例列表失败')
    tabInstances.value = []
  } finally {
    loadingTab.value = false
  }
}

const loadVolumes = async (clearCache = false) => {
  if (!configDetails.value) return
  loadingTab.value = true
  try {
    const response = await api.post('/oci/details/volumes', { configId: configDetails.value.userId, clearCache })
    tabVolumes.value = response.data || []
  } catch (error: any) {
    toast.error(error.message || '加载存储卷列表失败')
    tabVolumes.value = []
  } finally {
    loadingTab.value = false
  }
}

const loadVCNs = async (clearCache = false) => {
  if (!configDetails.value) return
  loadingTab.value = true
  try {
    const response = await api.post('/oci/details/vcns', { configId: configDetails.value.userId, clearCache })
    tabVCNs.value = response.data || []
  } catch (error: any) {
    toast.error(error.message || '加载VCN列表失败')
    tabVCNs.value = []
  } finally {
    loadingTab.value = false
  }
}

const loadTrafficCondition = async () => {
  if (!configDetails.value) return
  try {
    const response = await api.get('/oci/traffic/condition', { params: { configId: configDetails.value.userId } })
    trafficCondition.value = response.data || { instances: [] }
    const now = new Date()
    const oneHourAgo = new Date(now.getTime() - 60 * 60 * 1000)
    trafficForm.value.endTime = formatDateTime(now)
    trafficForm.value.startTime = formatDateTime(oneHourAgo)
  } catch (error) {
    console.error('加载流量条件失败:', error)
  }
}

const loadInstanceVnics = async () => {
  if (!configDetails.value || !trafficForm.value.instanceId) return
  try {
    const response = await api.get('/oci/traffic/vnics', {
      params: { configId: configDetails.value.userId, instanceId: trafficForm.value.instanceId }
    })
    trafficVnics.value = response.data || []
    trafficForm.value.vnicId = ''
  } catch {
    trafficVnics.value = []
  }
}

const loadTrafficData = async () => {
  if (!configDetails.value || !trafficForm.value.instanceId || !trafficForm.value.vnicId) {
    toast.error('请选择实例和VNIC')
    return
  }
  loadingTraffic.value = true
  try {
    const response = await api.post('/oci/traffic/data', {
      configId: configDetails.value.userId,
      instanceId: trafficForm.value.instanceId,
      vnicId: trafficForm.value.vnicId,
      startTime: trafficForm.value.startTime,
      endTime: trafficForm.value.endTime
    })
    tabTraffic.value = response.data || { time: [], inbound: [], outbound: [] }
  } catch (error: any) {
    toast.error(error.message || '加载流量数据失败')
    tabTraffic.value = { time: [], inbound: [], outbound: [] }
  } finally {
    loadingTraffic.value = false
  }
}

const refreshCurrentTab = async () => {
  if (activeTab.value === 'basic') await loadTenant(true)
  else if (activeTab.value === 'instances') await loadInstances(true)
  else if (activeTab.value === 'volumes') await loadVolumes(true)
  else if (activeTab.value === 'vcns') await loadVCNs(true)
  else if (activeTab.value === 'traffic') await loadTrafficCondition()
}

watch(activeTab, newTab => {
  if (newTab === 'basic' && !tabTenant.value) loadTenant()
  else if (newTab === 'instances' && tabInstances.value.length === 0) loadInstances()
  else if (newTab === 'volumes' && tabVolumes.value.length === 0) loadVolumes()
  else if (newTab === 'vcns' && tabVCNs.value.length === 0) loadVCNs()
  else if (newTab === 'traffic' && trafficCondition.value.instances.length === 0) loadTrafficCondition()
})

watch(
  () => trafficForm.value.instanceId,
  newVal => {
    if (newVal) loadInstanceVnics()
    else {
      trafficVnics.value = []
      trafficForm.value.vnicId = ''
    }
  }
)

// 密码过期时间编辑
const startEditPasswordExpiry = () => {
  passwordExpiryInput.value = tabTenant.value?.passwordExpiresAfter || 0
  editingPasswordExpiry.value = true
}
const cancelEditPasswordExpiry = () => {
  editingPasswordExpiry.value = false
}
const savePasswordExpiry = async () => {
  updatingPasswordExpiry.value = true
  try {
    await api.post('/oci/tenant/updatePwdEx', {
      cfgId: configDetails.value?.userId,
      passwordExpiresAfter: passwordExpiryInput.value
    })
    toast.success('密码过期时间更新成功')
    if (tabTenant.value) tabTenant.value.passwordExpiresAfter = passwordExpiryInput.value
    editingPasswordExpiry.value = false
  } catch (error: any) {
    toast.error(error.message || '更新失败')
  } finally {
    updatingPasswordExpiry.value = false
  }
}

// 用户管理
const editUser = (user: any) => {
  editingUser.value = user
  userForm.value = { email: user.email || '', dbUserName: user.name || '', description: '' }
  showEditUserModal.value = true
}
const closeEditUserModal = () => {
  showEditUserModal.value = false
  editingUser.value = null
  userForm.value = { email: '', dbUserName: '', description: '' }
}
const saveUserInfo = async () => {
  if (!editingUser.value) return
  try {
    await api.post('/oci/tenant/updateUserInfo', {
      ociCfgId: configDetails.value?.userId,
      userId: editingUser.value.id,
      email: userForm.value.email,
      dbUserName: userForm.value.dbUserName,
      description: userForm.value.description
    })
    toast.success('用户信息更新成功')
    closeEditUserModal()
    await loadTenant(true)
  } catch (error: any) {
    toast.error(error.message || '更新失败')
  }
}
const resetUserPassword = async (user: any) => {
  if (!confirm(`确定要重置用户 ${user.name} 的密码吗？`)) return
  try {
    await api.post('/oci/tenant/resetPassword', { ociCfgId: configDetails.value?.userId, userId: user.id })
    toast.success('密码重置成功')
  } catch (error: any) {
    toast.error(error.message || '重置密码失败')
  }
}
const clearUserMfa = async (user: any) => {
  if (!confirm(`确定要清除用户 ${user.name} 的 MFA 设备吗？`)) return
  try {
    await api.post('/oci/tenant/deleteMfaDevice', { ociCfgId: configDetails.value?.userId, userId: user.id })
    toast.success('MFA 设备清除成功')
    await loadTenant(true)
  } catch (error: any) {
    toast.error(error.message || 'MFA 清除失败')
  }
}
const clearUserApiKeys = async (user: any) => {
  if (!confirm(`确定要清除用户 ${user.name} 的所有 API 密钥吗？`)) return
  try {
    await api.post('/oci/tenant/deleteApiKey', { ociCfgId: configDetails.value?.userId, userId: user.id })
    toast.success('API 密钥清除成功')
  } catch (error: any) {
    toast.error(error.message || 'API 密钥清除失败')
  }
}
const deleteUser = async (user: any) => {
  if (!confirm(`确定要删除用户 ${user.name} 吗？此操作不可恢复！`)) return
  try {
    await api.post('/oci/tenant/deleteUser', { ociCfgId: configDetails.value?.userId, userId: user.id })
    toast.success('用户删除成功')
    await loadTenant(true)
  } catch (error: any) {
    toast.error(error.message || '删除用户失败')
  }
}

// 实例操作
const controlInstance = async (instanceId: string, action: string) => {
  const actionMap: Record<string, { endpoint: string; message: string }> = {
    START: { endpoint: '/instance/start', message: '启动' },
    STOP: { endpoint: '/instance/stop', message: '停止' },
    SOFTRESET: { endpoint: '/instance/reboot', message: '重启' }
  }
  instanceActionLoading[instanceId] = true
  try {
    await api.post(actionMap[action].endpoint, { userId: configDetails.value?.userId, instanceId })
    toast.success(`${actionMap[action].message}操作已提交`)
    setTimeout(() => loadInstances(true), 3000)
  } catch (error: any) {
    toast.error(error.message || '操作失败')
  } finally {
    delete instanceActionLoading[instanceId]
  }
}

const terminateInstance = async (instanceId: string) => {
  if (!confirm('确定要删除此实例吗？此操作不可恢复！')) return
  instanceActionLoading[instanceId] = true
  try {
    await api.post('/instance/terminate', { userId: configDetails.value?.userId, instanceId })
    toast.success('删除操作已提交')
    setTimeout(() => loadInstances(true), 3000)
  } catch (error: any) {
    toast.error(error.message || '删除失败')
  } finally {
    delete instanceActionLoading[instanceId]
  }
}

const changeIP = async (instanceId: string) => {
  if (!confirm('确定要更改此实例的公网IP吗？')) return
  instanceActionLoading[instanceId] = true
  try {
    const response = await api.post('/instance/changeIP', { userId: configDetails.value?.userId, instanceId })
    toast.success(response.data?.newIP ? `IP更改成功，新IP: ${response.data.newIP}` : 'IP更换请求已提交')
    setTimeout(() => loadInstances(true), 2000)
  } catch (error: any) {
    toast.error(error.message || '更换IP失败')
  } finally {
    delete instanceActionLoading[instanceId]
  }
}

// 打开子组件弹窗
const openEditInstance = (instance: Instance) => {
  selectedInstance.value = instance
  showEditInstanceModal.value = true
}
const openCloudShell = (instanceId: string) => {
  cloudShellInstanceId.value = instanceId
  showCloudShellModal.value = true
}
const openVolumeEdit = (volume: any) => {
  selectedVolume.value = volume
  showVolumeEditModal.value = true
}
const openSecurityList = (vcn: any) => {
  selectedVcn.value = vcn
  showSecurityListModal.value = true
}

onMounted(() => {
  loadConfigs()
})

const headerRef = ref<HTMLElement>()
useMotion(headerRef, { initial: { opacity: 0, y: -20 }, enter: { opacity: 1, y: 0 } })
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div ref="headerRef" class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div class="flex items-center gap-4">
        <h1 class="text-3xl font-display font-bold">配置管理</h1>
        <Badge v-if="selectedConfigIds.length > 0" variant="secondary">已选择 {{ selectedConfigIds.length }} 项</Badge>
      </div>
      <div class="flex gap-2">
        <Button v-if="selectedConfigIds.length > 0" variant="success" @click="batchCreateInstance">
          <Plus class="w-4 h-4" />
          批量创建实例
        </Button>
        <Button v-if="selectedConfigIds.length > 0" variant="destructive" @click="batchDeleteConfigs">
          <Trash2 class="w-4 h-4" />
          批量删除
        </Button>
        <Button @click="showAddModal = true">
          <Plus class="w-4 h-4" />
          添加配置
        </Button>
      </div>
    </div>

    <!-- Main Card -->
    <Card
      v-motion
      :initial="{ opacity: 0, y: 20 }"
      :enter="{ opacity: 1, y: 0, transition: { delay: 100 } }"
      class="border-border/50"
    >
      <CardHeader class="border-b border-border/50 pb-4">
        <div class="relative max-w-md">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input v-model="searchText" placeholder="搜索配置名..." class="pl-10" @input="handleSearch" />
        </div>
      </CardHeader>
      <CardContent class="p-0">
        <Table>
          <TableHeader>
            <TableRow class="hover:bg-transparent">
              <TableHead class="w-12">
                <Checkbox
                  :model-value="isAllSelected"
                  :indeterminate="isIndeterminate"
                  @update:model-value="toggleSelectAll"
                />
              </TableHead>
              <TableHead>配置名</TableHead>
              <TableHead>租户名称</TableHead>
              <TableHead class="hidden lg:table-cell">创建时间</TableHead>
              <TableHead>区域</TableHead>
              <TableHead>实例</TableHead>
              <TableHead class="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="loading">
              <TableCell colspan="7" class="h-32 text-center">
                <Loader2 class="w-8 h-8 mx-auto animate-spin text-primary" />
              </TableCell>
            </TableRow>
            <TableRow v-else-if="!configs.length">
              <TableCell colspan="7" class="h-32 text-center text-muted-foreground">暂无配置</TableCell>
            </TableRow>
            <TableRow
              v-for="(config, index) in configs"
              v-else
              :key="config.id"
              v-motion
              :initial="{ opacity: 0, x: -20 }"
              :enter="{ opacity: 1, x: 0, transition: { delay: 50 * index } }"
            >
              <TableCell>
                <Checkbox
                  :model-value="selectedConfigIds.includes(config.id)"
                  @update:model-value="toggleSelectConfig(config.id)"
                />
              </TableCell>
              <TableCell class="font-medium">{{ config.username }}</TableCell>
              <TableCell class="text-muted-foreground">{{ config.tenantName || '-' }}</TableCell>
              <TableCell class="text-sm text-muted-foreground hidden lg:table-cell">
                {{ config.tenantCreateTime || '-' }}
              </TableCell>
              <TableCell>
                <Badge variant="info">{{ config.ociRegion }}</Badge>
              </TableCell>
              <TableCell>
                <span>{{ config.instanceCount || 0 }}</span>
                <span v-if="config.runningInstances" class="text-success text-xs ml-1">
                  ({{ config.runningInstances }}运行)
                </span>
              </TableCell>
              <TableCell class="text-right">
                <div class="flex justify-end items-center gap-1">
                  <Button size="sm" variant="success" @click="createInstance(config)">
                    <Plus class="w-3.5 h-3.5" />
                    <span class="hidden sm:inline ml-1">创建实例</span>
                  </Button>
                  <Button size="sm" variant="outline" @click="viewConfigDetails(config)">
                    <Eye class="w-3.5 h-3.5" />
                    <span class="hidden sm:inline ml-1">详情</span>
                  </Button>
                  <Dropdown align="right">
                    <template #trigger>
                      <Button size="sm" variant="ghost"><MoreHorizontal class="w-4 h-4" /></Button>
                    </template>
                    <DropdownItem @click="editConfig(config)">
                      <Edit class="w-4 h-4" />
                      编辑配置
                    </DropdownItem>
                    <DropdownItem destructive @click="deleteConfig(config.id)">
                      <Trash2 class="w-4 h-4" />
                      删除配置
                    </DropdownItem>
                  </Dropdown>
                </div>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
        <div v-if="totalPages > 1" class="flex items-center justify-center gap-2 p-4 border-t border-border/50">
          <Button variant="outline" size="sm" :disabled="currentPage === 1" @click="loadConfigs(currentPage - 1)">
            <ChevronLeft class="w-4 h-4" />
            上一页
          </Button>
          <span class="px-4 text-sm text-muted-foreground">第 {{ currentPage }} / {{ totalPages }} 页</span>
          <Button
            variant="outline"
            size="sm"
            :disabled="currentPage === totalPages"
            @click="loadConfigs(currentPage + 1)"
          >
            下一页
            <ChevronRight class="w-4 h-4" />
          </Button>
        </div>
      </CardContent>
    </Card>

    <!-- Add/Edit Config Modal -->
    <Dialog v-model:open="showAddModal">
      <DialogHeader class="mb-4">
        <DialogTitle>{{ editingConfig ? '编辑配置' : '添加OCI配置' }}</DialogTitle>
        <DialogDescription>{{ editingConfig ? '修改现有配置信息' : '添加新的Oracle Cloud配置' }}</DialogDescription>
      </DialogHeader>
      <form class="space-y-4" @submit.prevent="submitForm">
        <div>
          <label class="block text-sm font-medium mb-2">配置名称</label>
          <Input v-model="form.username" placeholder="例: 我的OCI配置" required />
        </div>
        <div>
          <label class="block text-sm font-medium mb-2">配置内容</label>
          <Textarea
            v-model="form.configContent"
            :rows="6"
            class="font-mono text-sm"
            placeholder="user=ocid1.user.oc1..xxx&#10;fingerprint=xx:xx:xx&#10;tenancy=ocid1.tenancy.oc1..xxx&#10;region=ap-singapore-1"
            required
          />
          <p class="text-xs text-muted-foreground mt-2">格式：user、fingerprint、tenancy、region（每行一个）</p>
        </div>
        <div>
          <label class="block text-sm font-medium mb-2">密钥文件</label>
          <div
            :class="[
              'border-2 border-dashed rounded-lg p-6 transition-colors cursor-pointer text-center',
              isDragging ? 'border-primary bg-primary/5' : 'border-border hover:border-primary/50'
            ]"
            @drop.prevent="handleFileDrop"
            @dragover.prevent="isDragging = true"
            @dragleave.prevent="isDragging = false"
            @click="fileInput?.click()"
          >
            <input ref="fileInput" type="file" accept=".pem,.key" class="hidden" @change="handleFileSelect" />
            <div v-if="!uploadedFile">
              <Upload class="mx-auto h-10 w-10 text-muted-foreground mb-2" />
              <p class="text-sm text-muted-foreground">点击或拖拽文件</p>
              <p class="text-xs text-muted-foreground mt-1">支持 .pem 或 .key</p>
            </div>
            <div v-else class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-lg bg-success/10 flex items-center justify-center">
                  <Check class="w-5 h-5 text-success" />
                </div>
                <div class="text-left">
                  <p class="text-sm font-medium">{{ uploadedFile.name }}</p>
                  <p class="text-xs text-muted-foreground">{{ formatFileSize(uploadedFile.size) }}</p>
                </div>
              </div>
              <Button type="button" variant="ghost" size="icon" @click.stop="clearFile"><X class="w-4 h-4" /></Button>
            </div>
          </div>
        </div>
        <DialogFooter class="mt-6">
          <Button type="button" variant="outline" @click="closeModal">取消</Button>
          <Button type="submit" :disabled="submitting">
            <Loader2 v-if="submitting" class="w-4 h-4 animate-spin" />
            {{ submitting ? '提交中...' : '提交' }}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>

    <!-- Create Instance Modal -->
    <Dialog v-model:open="showCreateInstanceModal">
      <DialogHeader class="mb-4">
        <DialogTitle>创建实例任务</DialogTitle>
        <DialogDescription>
          为配置
          <span class="text-primary font-medium">{{ selectedConfigForInstance?.username }}</span>
          创建实例
        </DialogDescription>
      </DialogHeader>
      <form class="space-y-4" @submit.prevent="submitInstanceTask">
        <div>
          <label class="block text-sm font-medium mb-2">区域</label>
          <Input v-model="instanceForm.ociRegion" placeholder="例: ap-singapore-1" required />
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium mb-2">CPU核心数</label>
            <Input v-model.number="instanceForm.ocpus" type="number" step="0.1" min="0.1" required />
          </div>
          <div>
            <label class="block text-sm font-medium mb-2">内存(GB)</label>
            <Input v-model.number="instanceForm.memory" type="number" step="0.1" min="0.1" required />
          </div>
        </div>
        <div class="grid grid-cols-3 gap-4">
          <div>
            <label class="block text-sm font-medium mb-2">磁盘(GB)</label>
            <Input v-model.number="instanceForm.disk" type="number" min="50" required />
          </div>
          <div>
            <label class="block text-sm font-medium mb-2">VPU/GB</label>
            <select
              v-model.number="instanceForm.bootVolumeVpu"
              class="w-full h-10 px-3 rounded-md border border-input bg-background text-sm"
            >
              <option v-for="vpu in [10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 110, 120]" :key="vpu" :value="vpu">
                {{ vpu }}
              </option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium mb-2">架构</label>
            <select
              v-model="instanceForm.architecture"
              class="w-full h-10 px-3 rounded-md border border-input bg-background text-sm"
              @change="onArchitectureChange"
            >
              <option value="ARM">ARM</option>
              <option value="AMD">AMD</option>
            </select>
          </div>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium mb-2">操作系统</label>
            <select
              v-model="instanceForm.operationSystem"
              class="w-full h-10 px-3 rounded-md border border-input bg-background text-sm"
              @change="onOperationSystemChange"
            >
              <option value="Ubuntu">Ubuntu</option>
              <option value="CentOS">CentOS</option>
              <option value="Oracle Linux">Oracle Linux</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium mb-2">系统版本</label>
            <select
              v-model="instanceForm.imageId"
              class="w-full h-10 px-3 rounded-md border border-input bg-background text-sm"
              :disabled="loadingImages"
            >
              <option value="">
                {{ loadingImages ? '加载中...' : filteredImages.length === 0 ? '无可用镜像' : '自动选择最新' }}
              </option>
              <option v-for="img in filteredImages" :key="img.id" :value="img.id">
                {{ img.operatingSystem }} {{ img.operatingSystemVersion }}
              </option>
            </select>
          </div>
        </div>
        <div>
          <label class="block text-sm font-medium mb-2">SSH公钥</label>
          <select
            v-model="instanceForm.sshKeyId"
            class="w-full h-10 px-3 rounded-md border border-input bg-background text-sm"
            required
          >
            <option value="">请选择SSH公钥</option>
            <option v-for="key in sshKeys" :key="key.id" :value="key.id">{{ key.name }}</option>
          </select>
          <p class="text-xs text-muted-foreground mt-2">
            请先在
            <RouterLink to="/keys" class="text-primary hover:underline">密钥管理</RouterLink>
            中添加SSH公钥
          </p>
        </div>
        <div class="flex items-center gap-3 py-2">
          <Switch v-model="instanceForm.isTaskMode" />
          <span class="text-sm font-medium">抢占实例任务</span>
          <span class="text-xs text-muted-foreground">（持续尝试直到成功）</span>
        </div>
        <div v-if="instanceForm.isTaskMode">
          <label class="block text-sm font-medium mb-2">执行间隔（秒）</label>
          <Input v-model.number="instanceForm.interval" type="number" min="10" placeholder="60" />
        </div>
        <DialogFooter class="mt-6">
          <Button type="button" variant="outline" @click="closeInstanceModal">取消</Button>
          <Button type="submit" :disabled="submittingInstance">
            <Loader2 v-if="submittingInstance" class="w-4 h-4 animate-spin" />
            {{ submittingInstance ? '创建中...' : instanceForm.isTaskMode ? '创建任务' : '创建实例' }}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>

    <!-- Batch Create Instance Modal -->
    <Dialog v-model:open="showBatchCreateModal">
      <DialogHeader class="mb-4">
        <DialogTitle>批量创建实例任务</DialogTitle>
        <DialogDescription>
          将为
          <span class="text-primary font-medium">{{ selectedConfigIds.length }}</span>
          个配置批量创建实例
        </DialogDescription>
      </DialogHeader>
      <form class="space-y-4" @submit.prevent="submitBatchInstanceTask">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium mb-2">CPU核心数</label>
            <Input v-model.number="instanceForm.ocpus" type="number" step="0.1" min="0.1" required />
          </div>
          <div>
            <label class="block text-sm font-medium mb-2">内存(GB)</label>
            <Input v-model.number="instanceForm.memory" type="number" step="0.1" min="0.1" required />
          </div>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium mb-2">磁盘(GB)</label>
            <Input v-model.number="instanceForm.disk" type="number" min="50" required />
          </div>
          <div>
            <label class="block text-sm font-medium mb-2">架构</label>
            <select
              v-model="instanceForm.architecture"
              class="w-full h-10 px-3 rounded-md border border-input bg-background text-sm"
            >
              <option value="ARM">ARM</option>
              <option value="AMD">AMD</option>
            </select>
          </div>
        </div>
        <div>
          <label class="block text-sm font-medium mb-2">操作系统</label>
          <select
            v-model="instanceForm.operationSystem"
            class="w-full h-10 px-3 rounded-md border border-input bg-background text-sm"
          >
            <option value="Ubuntu">Ubuntu</option>
            <option value="CentOS">CentOS</option>
            <option value="Oracle Linux">Oracle Linux</option>
          </select>
        </div>
        <div>
          <label class="block text-sm font-medium mb-2">SSH公钥</label>
          <select
            v-model="instanceForm.sshKeyId"
            class="w-full h-10 px-3 rounded-md border border-input bg-background text-sm"
            required
          >
            <option value="">请选择SSH公钥</option>
            <option v-for="key in sshKeys" :key="key.id" :value="key.id">{{ key.name }}</option>
          </select>
        </div>
        <div>
          <label class="block text-sm font-medium mb-2">执行间隔（秒）</label>
          <Input v-model.number="instanceForm.interval" type="number" min="10" placeholder="60" />
        </div>
        <DialogFooter class="mt-6">
          <Button type="button" variant="outline" @click="closeBatchCreateModal">取消</Button>
          <Button type="submit" :disabled="submittingInstance">
            <Loader2 v-if="submittingInstance" class="w-4 h-4 animate-spin" />
            {{ submittingInstance ? '创建中...' : '批量创建' }}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>

    <!-- Config Details Sidebar -->
    <Teleport to="body">
      <Transition name="fade">
        <div
          v-if="showConfigDetailsSidebar"
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm"
          @click.self="closeConfigDetailsSidebar"
        >
          <div
            class="bg-card rounded-xl shadow-2xl w-full max-w-6xl max-h-[90vh] overflow-hidden flex flex-col border border-border"
            v-motion
            :initial="{ opacity: 0, scale: 0.95 }"
            :enter="{ opacity: 1, scale: 1 }"
          >
            <div class="flex items-center justify-between p-6 border-b border-border bg-card/80 backdrop-blur">
              <div>
                <h2 class="text-xl font-bold">配置详情</h2>
                <p class="text-sm text-muted-foreground mt-1">{{ configDetails?.username }}</p>
              </div>
              <Button variant="ghost" size="icon" @click="closeConfigDetailsSidebar"><X class="w-5 h-5" /></Button>
            </div>
            <div class="flex-1 overflow-y-auto p-6">
              <div v-if="loadingDetails" class="flex items-center justify-center py-12">
                <Loader2 class="w-10 h-10 animate-spin text-primary" />
              </div>
              <div v-else-if="configDetails" class="space-y-6">
                <!-- Tabs -->
                <div class="flex items-center justify-between flex-wrap gap-4">
                  <div class="flex gap-1 p-1 bg-muted/50 rounded-lg flex-wrap">
                    <button
                      v-for="tab in [
                        { key: 'basic', label: '基本信息', icon: Settings },
                        { key: 'instances', label: '实例列表', icon: Server },
                        { key: 'volumes', label: '引导卷', icon: HardDrive },
                        { key: 'vcns', label: 'VCN网络', icon: Network },
                        { key: 'traffic', label: '流量统计', icon: BarChart3 }
                      ]"
                      :key="tab.key"
                      :class="[
                        'flex items-center gap-2 px-4 py-2 rounded-md text-sm font-medium transition-all',
                        activeTab === tab.key
                          ? 'bg-primary text-primary-foreground shadow-sm'
                          : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                      ]"
                      @click="activeTab = tab.key"
                    >
                      <component :is="tab.icon" class="w-4 h-4" />
                      {{ tab.label }}
                    </button>
                  </div>
                  <Button variant="outline" size="sm" :disabled="loadingTab" @click="refreshCurrentTab">
                    <RefreshCw :class="['w-4 h-4', loadingTab && 'animate-spin']" />
                    刷新
                  </Button>
                </div>

                <!-- Basic Info Tab -->
                <div v-show="activeTab === 'basic'" class="space-y-6">
                  <Card v-if="loadingTab" class="p-8">
                    <div class="flex items-center justify-center">
                      <Loader2 class="w-8 h-8 animate-spin text-primary" />
                    </div>
                  </Card>
                  <Card v-else class="p-6">
                    <h3 class="text-lg font-semibold mb-4 flex items-center gap-2">
                      <Settings class="w-5 h-5 text-primary" />
                      配置与租户信息
                    </h3>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-6 text-sm">
                      <div class="space-y-1">
                        <label class="text-muted-foreground text-xs uppercase tracking-wide">配置名称</label>
                        <p class="font-medium text-lg">{{ configDetails.username }}</p>
                      </div>
                      <div class="space-y-1">
                        <label class="text-muted-foreground text-xs uppercase tracking-wide">当前区域</label>
                        <p>
                          <Badge variant="info" class="text-sm">{{ configDetails.region }}</Badge>
                        </p>
                      </div>
                      <div v-if="tabTenant" class="space-y-1 md:col-span-2">
                        <label class="text-muted-foreground text-xs uppercase tracking-wide">租户名称</label>
                        <p class="font-medium text-lg">{{ tabTenant.name }}</p>
                      </div>
                      <div v-if="tabTenant" class="space-y-1 md:col-span-2">
                        <label class="text-muted-foreground text-xs uppercase tracking-wide">租户ID</label>
                        <div class="flex items-center gap-2 bg-muted/50 p-3 rounded-lg">
                          <code class="text-xs font-mono flex-1 break-all">{{ tabTenant.id }}</code>
                          <Button
                            variant="ghost"
                            size="icon"
                            class="h-8 w-8 shrink-0"
                            @click="copyToClipboard(tabTenant.id)"
                          >
                            <Copy class="w-4 h-4" />
                          </Button>
                        </div>
                      </div>
                      <div v-if="tabTenant" class="space-y-1">
                        <label class="text-muted-foreground text-xs uppercase tracking-wide">主区域</label>
                        <p class="font-medium">{{ tabTenant.homeRegionKey }}</p>
                      </div>
                      <div v-if="tabTenant?.createTime" class="space-y-1">
                        <label class="text-muted-foreground text-xs uppercase tracking-wide">账户创建时间</label>
                        <p>{{ tabTenant.createTime }}</p>
                      </div>
                      <div class="space-y-1 md:col-span-2">
                        <label class="text-muted-foreground text-xs uppercase tracking-wide">用户ID</label>
                        <div class="bg-muted/50 p-3 rounded-lg">
                          <code class="text-xs font-mono break-all">{{ configDetails.userId }}</code>
                        </div>
                      </div>
                      <div class="space-y-1">
                        <label class="text-muted-foreground text-xs uppercase tracking-wide">指纹</label>
                        <div class="bg-muted/50 p-3 rounded-lg">
                          <code class="text-xs font-mono break-all">{{ configDetails.fingerprint }}</code>
                        </div>
                      </div>
                      <div class="space-y-1">
                        <label class="text-muted-foreground text-xs uppercase tracking-wide">密钥文件</label>
                        <p class="text-sm">{{ configDetails.keyPath }}</p>
                      </div>
                    </div>
                    <!-- 密码过期时间设置 -->
                    <div v-if="tabTenant" class="mt-6 pt-6 border-t border-border">
                      <label class="text-muted-foreground text-xs uppercase tracking-wide block mb-2">
                        密码过期时间
                      </label>
                      <div class="flex items-center gap-2">
                        <Input
                          v-if="editingPasswordExpiry"
                          v-model.number="passwordExpiryInput"
                          type="number"
                          min="0"
                          class="w-32"
                          placeholder="0"
                        />
                        <span v-else class="font-medium">
                          {{
                            tabTenant.passwordExpiresAfter === 0 ? '永不过期' : tabTenant.passwordExpiresAfter + ' 天'
                          }}
                        </span>
                        <Button
                          v-if="!editingPasswordExpiry"
                          variant="ghost"
                          size="icon"
                          class="h-8 w-8"
                          @click="startEditPasswordExpiry"
                        >
                          <Edit class="w-4 h-4" />
                        </Button>
                        <div v-else class="flex gap-1">
                          <Button
                            variant="ghost"
                            size="icon"
                            class="h-8 w-8 text-success"
                            :disabled="updatingPasswordExpiry"
                            @click="savePasswordExpiry"
                          >
                            <Check class="w-4 h-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            class="h-8 w-8 text-destructive"
                            :disabled="updatingPasswordExpiry"
                            @click="cancelEditPasswordExpiry"
                          >
                            <X class="w-4 h-4" />
                          </Button>
                        </div>
                      </div>
                      <p class="text-xs text-muted-foreground mt-1">设置为 0 表示永不过期</p>
                    </div>
                    <div v-if="tabTenant?.regions?.length" class="mt-6 pt-6 border-t border-border">
                      <label class="text-muted-foreground text-xs uppercase tracking-wide block mb-3">
                        订阅区域 ({{ tabTenant.regions.length }})
                      </label>
                      <div class="flex flex-wrap gap-2">
                        <Badge v-for="region in tabTenant.regions" :key="region" variant="secondary">
                          {{ region }}
                        </Badge>
                      </div>
                    </div>
                  </Card>

                  <!-- 用户列表卡片 -->
                  <Card v-if="tabTenant?.userList?.length" class="p-6">
                    <h3 class="text-lg font-semibold mb-4 flex items-center gap-2">
                      <Users class="w-5 h-5 text-primary" />
                      用户列表 ({{ tabTenant.userList.length }})
                    </h3>
                    <div class="space-y-4">
                      <div
                        v-for="user in tabTenant.userList"
                        :key="user.id"
                        class="border border-border rounded-lg p-4 hover:border-primary/50 transition-colors"
                      >
                        <div class="flex justify-between items-start mb-3">
                          <div class="flex-1">
                            <h4 class="font-semibold text-lg">{{ user.name }}</h4>
                            <p v-if="user.email" class="text-sm text-muted-foreground mt-1 flex items-center gap-1">
                              <Mail class="w-3.5 h-3.5" />
                              {{ user.email }}
                            </p>
                          </div>
                          <div class="flex gap-2 items-center">
                            <Badge v-if="user.isMfaActivated" variant="success" class="text-xs">MFA</Badge>
                            <Badge v-if="user.emailVerified" variant="info" class="text-xs">已验证</Badge>
                            <Badge :variant="user.state === 'ACTIVE' ? 'success' : 'destructive'" class="text-xs">
                              {{ user.state }}
                            </Badge>
                          </div>
                        </div>
                        <div class="text-xs text-muted-foreground mb-4">
                          创建时间: {{ user.createTime }}
                          <span v-if="user.lastSuccessfulLoginTime" class="ml-4">
                            最近登录: {{ user.lastSuccessfulLoginTime }}
                          </span>
                        </div>
                        <div class="flex flex-wrap gap-2">
                          <Button size="sm" variant="outline" @click="editUser(user)">
                            <Edit class="w-3.5 h-3.5" />
                            编辑
                          </Button>
                          <Button size="sm" variant="warning" @click="resetUserPassword(user)">
                            <KeyRound class="w-3.5 h-3.5" />
                            重置密码
                          </Button>
                          <Button v-if="user.isMfaActivated" size="sm" variant="outline" @click="clearUserMfa(user)">
                            <LockKeyhole class="w-3.5 h-3.5" />
                            清除MFA
                          </Button>
                          <Button size="sm" variant="outline" @click="clearUserApiKeys(user)">
                            <ShieldOff class="w-3.5 h-3.5" />
                            清除API
                          </Button>
                          <Button size="sm" variant="destructive" @click="deleteUser(user)">
                            <Trash2 class="w-3.5 h-3.5" />
                            删除
                          </Button>
                        </div>
                      </div>
                    </div>
                  </Card>
                </div>

                <!-- Instances Tab -->
                <div v-show="activeTab === 'instances'">
                  <div v-if="loadingTab" class="flex items-center justify-center py-12">
                    <Loader2 class="w-10 h-10 animate-spin text-primary" />
                  </div>
                  <div v-else-if="!tabInstances.length" class="text-center py-16">
                    <div class="w-20 h-20 mx-auto mb-4 rounded-full bg-muted/50 flex items-center justify-center">
                      <Server class="w-10 h-10 text-muted-foreground" />
                    </div>
                    <p class="text-muted-foreground text-lg">暂无实例</p>
                  </div>
                  <div v-else class="grid grid-cols-1 xl:grid-cols-2 gap-4">
                    <Card
                      v-for="instance in tabInstances"
                      :key="instance.id"
                      class="p-5 hover:border-primary/50 transition-colors"
                    >
                      <div class="flex justify-between items-start mb-4">
                        <div class="flex-1 min-w-0 pr-4">
                          <h4 class="font-semibold text-lg truncate">{{ instance.displayName }}</h4>
                          <p class="text-xs text-muted-foreground font-mono truncate mt-1">{{ instance.id }}</p>
                        </div>
                        <Badge
                          :variant="
                            instance.state === 'RUNNING'
                              ? 'success'
                              : instance.state === 'STOPPED'
                                ? 'destructive'
                                : 'warning'
                          "
                          class="shrink-0"
                        >
                          {{ instance.state }}
                        </Badge>
                      </div>
                      <div class="grid grid-cols-2 gap-3 text-sm mb-5">
                        <div class="bg-muted/30 rounded-lg p-3">
                          <span class="text-muted-foreground text-xs block mb-1">规格</span>
                          <span class="font-medium">{{ instance.shape }}</span>
                        </div>
                        <div class="bg-muted/30 rounded-lg p-3">
                          <span class="text-muted-foreground text-xs block mb-1">CPU / 内存</span>
                          <span class="font-medium">{{ instance.ocpus }}核 / {{ instance.memory }}GB</span>
                        </div>
                        <div class="bg-muted/30 rounded-lg p-3">
                          <span class="text-muted-foreground text-xs block mb-1">引导卷</span>
                          <span class="font-medium">{{ instance.bootVolumeSize || '-' }} GB</span>
                        </div>
                        <div class="bg-muted/30 rounded-lg p-3">
                          <span class="text-muted-foreground text-xs block mb-1">区域</span>
                          <span class="font-medium">{{ instance.region }}</span>
                        </div>
                        <div class="bg-muted/30 rounded-lg p-3 col-span-2">
                          <span class="text-muted-foreground text-xs block mb-1">公网IP</span>
                          <span class="font-mono text-sm font-medium">
                            {{ instance.publicIps?.join(', ') || '无' }}
                          </span>
                        </div>
                        <div
                          v-if="instance.ipv6"
                          class="bg-primary/5 border border-primary/20 rounded-lg p-3 col-span-2"
                        >
                          <span class="text-primary text-xs block mb-1">IPv6</span>
                          <span class="font-mono text-sm text-primary">{{ instance.ipv6 }}</span>
                        </div>
                      </div>
                      <div class="flex flex-wrap gap-2">
                        <Button
                          size="sm"
                          variant="success"
                          :disabled="instance.state === 'RUNNING' || instanceActionLoading[instance.id]"
                          @click="controlInstance(instance.id, 'START')"
                        >
                          <Play class="w-3.5 h-3.5" />
                          启动
                        </Button>
                        <Button
                          size="sm"
                          variant="warning"
                          :disabled="instance.state !== 'RUNNING' || instanceActionLoading[instance.id]"
                          @click="controlInstance(instance.id, 'STOP')"
                        >
                          <Square class="w-3.5 h-3.5" />
                          停止
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          :disabled="instance.state !== 'RUNNING' || instanceActionLoading[instance.id]"
                          @click="controlInstance(instance.id, 'SOFTRESET')"
                        >
                          <RotateCcw class="w-3.5 h-3.5" />
                          重启
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          :disabled="instanceActionLoading[instance.id]"
                          @click="changeIP(instance.id)"
                        >
                          <Globe class="w-3.5 h-3.5" />
                          更改IP
                        </Button>
                        <Button size="sm" variant="outline" @click="openEditInstance(instance)">
                          <Settings class="w-3.5 h-3.5" />
                          编辑配置
                        </Button>
                        <Button size="sm" variant="outline" @click="openCloudShell(instance.id)">
                          <Terminal class="w-3.5 h-3.5" />
                          Cloud Shell
                        </Button>
                        <Button
                          size="sm"
                          variant="destructive"
                          :disabled="instanceActionLoading[instance.id]"
                          @click="terminateInstance(instance.id)"
                        >
                          <Trash2 class="w-3.5 h-3.5" />
                          删除
                        </Button>
                      </div>
                    </Card>
                  </div>
                </div>

                <!-- Volumes Tab -->
                <div v-show="activeTab === 'volumes'">
                  <div v-if="loadingTab" class="flex items-center justify-center py-12">
                    <Loader2 class="w-10 h-10 animate-spin text-primary" />
                  </div>
                  <div v-else-if="!tabVolumes.length" class="text-center py-16">
                    <div class="w-20 h-20 mx-auto mb-4 rounded-full bg-muted/50 flex items-center justify-center">
                      <HardDrive class="w-10 h-10 text-muted-foreground" />
                    </div>
                    <p class="text-muted-foreground text-lg">暂无引导卷</p>
                  </div>
                  <div v-else class="space-y-4">
                    <Card v-for="volume in tabVolumes" :key="volume.id" class="p-5">
                      <div class="flex justify-between items-start mb-4">
                        <div class="flex items-center gap-4">
                          <div class="w-12 h-12 rounded-lg bg-primary/10 flex items-center justify-center">
                            <HardDrive class="w-6 h-6 text-primary" />
                          </div>
                          <div>
                            <h4 class="font-semibold text-lg">{{ volume.displayName }}</h4>
                            <p class="text-xs text-muted-foreground font-mono">{{ volume.id?.substring(0, 40) }}...</p>
                          </div>
                        </div>
                        <div class="flex items-center gap-2">
                          <Badge :variant="volume.attached ? 'success' : 'warning'">
                            {{ volume.attached ? '已附加' : '未附加' }}
                          </Badge>
                          <Badge :variant="volume.state === 'AVAILABLE' ? 'success' : 'warning'">
                            {{ volume.state }}
                          </Badge>
                          <Button size="sm" variant="outline" @click="openVolumeEdit(volume)">
                            <Edit class="w-4 h-4" />
                            编辑
                          </Button>
                        </div>
                      </div>
                      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
                        <div class="bg-muted/30 rounded-lg p-4">
                          <span class="text-muted-foreground text-xs block mb-1">磁盘大小</span>
                          <span class="font-semibold text-lg">{{ volume.sizeInGBs }} GB</span>
                        </div>
                        <div class="bg-muted/30 rounded-lg p-4">
                          <span class="text-muted-foreground text-xs block mb-1">性能 (VPU/GB)</span>
                          <span class="font-semibold text-lg">{{ volume.vpusPerGB || 10 }}</span>
                        </div>
                        <div v-if="volume.instanceName" class="bg-muted/30 rounded-lg p-4">
                          <span class="text-muted-foreground text-xs block mb-1">附加实例</span>
                          <span class="text-primary font-semibold">{{ volume.instanceName }}</span>
                        </div>
                        <div class="bg-muted/30 rounded-lg p-4">
                          <span class="text-muted-foreground text-xs block mb-1">可用域</span>
                          <span class="text-sm">{{ volume.availabilityDomain?.split(':').pop() || '-' }}</span>
                        </div>
                      </div>
                    </Card>
                  </div>
                </div>

                <!-- VCNs Tab -->
                <div v-show="activeTab === 'vcns'">
                  <div v-if="loadingTab" class="flex items-center justify-center py-12">
                    <Loader2 class="w-10 h-10 animate-spin text-primary" />
                  </div>
                  <div v-else-if="!tabVCNs.length" class="text-center py-16">
                    <div class="w-20 h-20 mx-auto mb-4 rounded-full bg-muted/50 flex items-center justify-center">
                      <Network class="w-10 h-10 text-muted-foreground" />
                    </div>
                    <p class="text-muted-foreground text-lg">暂无VCN</p>
                  </div>
                  <div v-else class="space-y-4">
                    <Card v-for="vcn in tabVCNs" :key="vcn.id" class="p-5">
                      <div class="flex justify-between items-start mb-4">
                        <div class="flex items-center gap-4">
                          <div class="w-12 h-12 rounded-lg bg-primary/10 flex items-center justify-center">
                            <Network class="w-6 h-6 text-primary" />
                          </div>
                          <div>
                            <h4 class="font-semibold text-lg">{{ vcn.displayName }}</h4>
                            <p class="text-sm text-muted-foreground font-mono">CIDR: {{ vcn.cidrBlock }}</p>
                          </div>
                        </div>
                        <div class="flex items-center gap-2">
                          <Badge variant="success">{{ vcn.state }}</Badge>
                          <Button size="sm" variant="outline" @click="openSecurityList(vcn)">
                            <Shield class="w-4 h-4" />
                            安全列表
                          </Button>
                        </div>
                      </div>
                      <div v-if="vcn.createTime" class="text-sm text-muted-foreground mb-4">
                        创建时间: {{ vcn.createTime }}
                      </div>
                      <div v-if="vcn.subnets?.length">
                        <h5 class="text-sm font-medium mb-3">子网 ({{ vcn.subnets.length }}个)</h5>
                        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                          <div v-for="subnet in vcn.subnets" :key="subnet.id" class="bg-muted/30 rounded-lg p-4">
                            <div class="flex justify-between items-center mb-2">
                              <span class="font-medium">{{ subnet.displayName }}</span>
                              <Badge :variant="subnet.isPublic ? 'success' : 'warning'" class="text-xs">
                                {{ subnet.isPublic ? '公有' : '私有' }}
                              </Badge>
                            </div>
                            <p class="text-sm text-muted-foreground font-mono">{{ subnet.cidrBlock }}</p>
                          </div>
                        </div>
                      </div>
                    </Card>
                  </div>
                </div>

                <!-- Traffic Tab -->
                <div v-show="activeTab === 'traffic'" class="space-y-4">
                  <Card class="p-4">
                    <h4 class="font-semibold mb-4 flex items-center gap-2">
                      <BarChart3 class="w-5 h-5 text-primary" />
                      查询条件
                    </h4>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <div>
                        <label class="block text-sm text-muted-foreground mb-1">选择实例</label>
                        <select
                          v-model="trafficForm.instanceId"
                          class="w-full h-10 px-3 rounded-md border border-input bg-background text-sm"
                        >
                          <option value="">请选择实例</option>
                          <option v-for="inst in trafficCondition.instances" :key="inst.value" :value="inst.value">
                            {{ inst.label }}
                          </option>
                        </select>
                      </div>
                      <div>
                        <label class="block text-sm text-muted-foreground mb-1">选择VNIC</label>
                        <select
                          v-model="trafficForm.vnicId"
                          class="w-full h-10 px-3 rounded-md border border-input bg-background text-sm"
                          :disabled="!trafficVnics.length"
                        >
                          <option value="">请选择VNIC</option>
                          <option v-for="vnic in trafficVnics" :key="vnic.value" :value="vnic.value">
                            {{ vnic.label }}
                          </option>
                        </select>
                      </div>
                      <div>
                        <label class="block text-sm text-muted-foreground mb-1">开始时间</label>
                        <Input v-model="trafficForm.startTime" placeholder="YYYY-MM-DD HH:mm:ss" />
                      </div>
                      <div>
                        <label class="block text-sm text-muted-foreground mb-1">结束时间</label>
                        <Input v-model="trafficForm.endTime" placeholder="YYYY-MM-DD HH:mm:ss" />
                      </div>
                    </div>
                    <Button class="mt-4" :disabled="loadingTraffic" @click="loadTrafficData">
                      <Loader2 v-if="loadingTraffic" class="w-4 h-4 animate-spin" />
                      {{ loadingTraffic ? '查询中...' : '查询流量' }}
                    </Button>
                  </Card>

                  <div v-if="loadingTraffic" class="flex items-center justify-center py-12">
                    <Loader2 class="w-10 h-10 animate-spin text-primary" />
                  </div>
                  <Card v-else-if="tabTraffic.time?.length" class="p-4">
                    <h4 class="font-semibold mb-4">流量数据 (单位: MB)</h4>
                    <div class="overflow-x-auto">
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>时间</TableHead>
                            <TableHead class="text-success">入站 (MB)</TableHead>
                            <TableHead class="text-primary">出站 (MB)</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          <TableRow v-for="(time, index) in tabTraffic.time" :key="index">
                            <TableCell class="text-muted-foreground">{{ time }}</TableCell>
                            <TableCell class="text-success">{{ tabTraffic.inbound[index] || '0' }}</TableCell>
                            <TableCell class="text-primary">{{ tabTraffic.outbound[index] || '0' }}</TableCell>
                          </TableRow>
                        </TableBody>
                      </Table>
                    </div>
                  </Card>
                  <div v-else class="text-center py-16">
                    <div class="w-20 h-20 mx-auto mb-4 rounded-full bg-muted/50 flex items-center justify-center">
                      <BarChart3 class="w-10 h-10 text-muted-foreground" />
                    </div>
                    <p class="text-muted-foreground">请选择实例和VNIC后查询流量数据</p>
                  </div>
                </div>
              </div>
            </div>
            <div class="p-6 border-t border-border bg-card/80 backdrop-blur">
              <div class="flex justify-end">
                <Button variant="outline" @click="closeConfigDetailsSidebar">关闭</Button>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- 编辑用户弹窗 -->
    <Dialog v-model:open="showEditUserModal">
      <DialogHeader class="mb-4">
        <DialogTitle>编辑用户信息</DialogTitle>
        <DialogDescription>
          修改用户
          <span class="text-primary font-medium">{{ editingUser?.name }}</span>
          的信息
        </DialogDescription>
      </DialogHeader>
      <form class="space-y-4" @submit.prevent="saveUserInfo">
        <div>
          <label class="block text-sm font-medium mb-2">用户名</label>
          <Input v-model="userForm.dbUserName" placeholder="输入用户名" required />
        </div>
        <div>
          <label class="block text-sm font-medium mb-2">邮箱</label>
          <Input v-model="userForm.email" type="email" placeholder="输入邮箱地址" required />
        </div>
        <div>
          <label class="block text-sm font-medium mb-2">描述（可选）</label>
          <Textarea v-model="userForm.description" :rows="3" placeholder="输入用户描述" />
        </div>
        <DialogFooter class="mt-6">
          <Button type="button" variant="outline" @click="closeEditUserModal">取消</Button>
          <Button type="submit">保存</Button>
        </DialogFooter>
      </form>
    </Dialog>

    <!-- 子组件弹窗 -->
    <EditInstanceModal
      v-model:open="showEditInstanceModal"
      :instance="selectedInstance"
      :user-id="configDetails?.userId || ''"
      @refresh="loadInstances(true)"
    />
    <CloudShellModal
      v-model:open="showCloudShellModal"
      :instance-id="cloudShellInstanceId"
      :user-id="configDetails?.userId || ''"
    />
    <VolumeEditModal
      v-model:open="showVolumeEditModal"
      :volume="selectedVolume"
      :user-id="configDetails?.userId || ''"
      @refresh="loadVolumes(true)"
    />
    <SecurityListModal
      v-model:open="showSecurityListModal"
      :vcn="selectedVcn"
      :user-id="configDetails?.userId || ''"
      @refresh="loadVCNs(true)"
    />
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
