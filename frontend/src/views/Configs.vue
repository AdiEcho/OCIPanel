<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex justify-between items-center">
      <div class="flex items-center gap-4">
        <h1 class="text-3xl font-bold">配置管理</h1>
        <span v-if="selectedConfigIds.length > 0" class="text-sm text-slate-400">
          已选择 {{ selectedConfigIds.length }} 项
        </span>
      </div>
      <div class="flex gap-2">
        <button 
          v-if="selectedConfigIds.length > 0" 
          class="btn btn-success" 
          @click="batchCreateInstance"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          批量创建实例
        </button>
        <button 
          v-if="selectedConfigIds.length > 0" 
          class="btn btn-danger" 
          @click="batchDeleteConfigs"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
          批量删除
        </button>
        <button class="btn btn-primary" @click="showAddModal = true">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          添加配置
        </button>
      </div>
    </div>

    <!-- 配置列表卡片 -->
    <div class="card">
      <div class="p-6 border-b border-slate-700">
        <input
          v-model="searchText"
          type="text"
          placeholder="搜索配置名..."
          class="input max-w-md"
          @input="handleSearch"
        />
      </div>

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
              <th class="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase">配置名</th>
              <th class="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase">租户名称</th>
              <th class="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase">租户创建时间</th>
              <th class="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase">区域</th>
              <th class="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase">实例数</th>
              <th class="px-6 py-4 text-left text-xs font-semibold text-slate-300 uppercase">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-700">
            <tr v-if="loading">
              <td colspan="7" class="px-6 py-8 text-center text-slate-400">
                <svg class="animate-spin h-8 w-8 mx-auto" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path
                    class="opacity-75"
                    fill="currentColor"
                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                  ></path>
                </svg>
              </td>
            </tr>
            <tr v-else-if="!configs.length">
              <td colspan="7" class="px-6 py-8 text-center text-slate-400">暂无配置</td>
            </tr>
            <tr v-for="config in configs" v-else :key="config.id" class="hover:bg-slate-700/30">
              <td class="px-4 py-4 text-center">
                <input 
                  type="checkbox" 
                  class="checkbox" 
                  :checked="selectedConfigIds.includes(config.id)"
                  @change="toggleSelectConfig(config.id)"
                />
              </td>
              <td class="px-6 py-4 font-medium">{{ config.username }}</td>
              <td class="px-6 py-4">{{ config.tenantName || '-' }}</td>
              <td class="px-6 py-4 text-sm text-slate-400">{{ config.tenantCreateTime || '-' }}</td>
              <td class="px-6 py-4">
                <span class="px-2 py-1 text-xs font-semibold rounded-full bg-blue-500/20 text-blue-300">
                  {{ config.ociRegion }}
                </span>
              </td>
              <td class="px-6 py-4">
                <div class="flex items-center gap-2">
                  <span class="text-slate-300">{{ config.instanceCount || 0 }}</span>
                  <span v-if="config.runningInstances > 0" class="text-xs text-green-400">
                    ({{ config.runningInstances }} 运行中)
                  </span>
                </div>
              </td>
              <td class="px-6 py-4">
                <div class="flex gap-2">
                  <button class="btn btn-success btn-sm" title="创建实例" @click="createInstance(config)">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                    </svg>
                  </button>
                  <button class="btn btn-primary btn-sm" title="配置详情" @click="viewConfigDetails(config)">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                      />
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
                      />
                    </svg>
                  </button>
                  <button class="btn btn-secondary btn-sm" title="编辑配置" @click="editConfig(config)">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                      />
                    </svg>
                  </button>
                  <button class="btn btn-danger btn-sm" title="删除配置" @click="deleteConfig(config.id)">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                      />
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
        <button :disabled="currentPage === 1" class="btn btn-secondary" @click="loadConfigs(currentPage - 1)">
          上一页
        </button>
        <span class="flex items-center px-4 text-slate-300">第 {{ currentPage }} / {{ totalPages }} 页</span>
        <button :disabled="currentPage === totalPages" class="btn btn-secondary" @click="loadConfigs(currentPage + 1)">
          下一页
        </button>
      </div>
    </div>

    <!-- 添加/编辑配置弹窗 -->
    <div
      v-if="showAddModal"
      class="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50 p-4"
    >
      <div class="card max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <div class="p-6 border-b border-slate-700 flex justify-between items-center">
          <h3 class="text-xl font-bold">{{ editingConfig ? '编辑配置' : '添加OCI配置' }}</h3>
          <button class="text-slate-400 hover:text-white" @click="closeModal">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <form class="p-6 space-y-4" @submit.prevent="submitForm">
          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">配置名称</label>
            <input v-model="form.username" type="text" class="input" placeholder="例: 我的OCI配置" required />
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">配置内容</label>
            <textarea
              v-model="form.configContent"
              class="input min-h-[200px] font-mono text-sm"
              placeholder="user=ocid1.user.oc1..aaaaaaaaxxx&#10;fingerprint=c6:1b:9f:cd:01:9d:7a:xxx&#10;tenancy=ocid1.tenancy.oc1..aaaaaaaaxxx&#10;region=sa-saopaulo-1"
              required
            ></textarea>
            <p class="text-xs text-slate-400 mt-2">
              请按照格式输入：user、fingerprint、tenancy、region（每行一个配置项）
            </p>
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">密钥文件</label>
            <div
              :class="[
                'border-2 border-dashed rounded-lg p-6 transition-colors cursor-pointer',
                isDragging ? 'border-blue-500 bg-blue-500/10' : 'border-slate-600 hover:border-slate-500'
              ]"
              @drop.prevent="handleFileDrop"
              @dragover.prevent="isDragging = true"
              @dragleave.prevent="isDragging = false"
              @click="$refs.fileInput.click()"
            >
              <input ref="fileInput" type="file" accept=".pem,.key" class="hidden" @change="handleFileSelect" />
              <div v-if="!uploadedFile" class="text-center">
                <svg class="mx-auto h-12 w-12 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
                  />
                </svg>
                <p class="mt-2 text-sm text-slate-300">点击选择文件或拖拽文件到此处</p>
                <p class="mt-1 text-xs text-slate-400">支持 .pem 或 .key 格式</p>
              </div>
              <div v-else class="flex items-center justify-between">
                <div class="flex items-center">
                  <svg class="h-8 w-8 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  </svg>
                  <div class="ml-3">
                    <p class="text-sm font-medium text-slate-200">{{ uploadedFile.name }}</p>
                    <p class="text-xs text-slate-400">{{ formatFileSize(uploadedFile.size) }}</p>
                  </div>
                </div>
                <button type="button" class="text-red-400 hover:text-red-300" @click.stop="clearFile">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>
            <p class="text-xs text-slate-400 mt-2">文件将自动上传到服务器的 keys 目录</p>
          </div>

          <div class="flex gap-3 pt-4">
            <button type="button" class="btn btn-secondary flex-1" @click="closeModal">取消</button>
            <button type="submit" class="btn btn-primary flex-1" :disabled="submitting">
              {{ submitting ? '提交中...' : '提交' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- 配置详情弹窗 -->
    <div
      v-if="showConfigDetailsSidebar"
      class="fixed inset-0 bg-black/70 backdrop-blur-sm z-50 flex items-center justify-center p-4"
      @click.self="closeConfigDetailsSidebar"
    >
      <div class="bg-slate-800 rounded-xl shadow-2xl w-full max-w-5xl max-h-[90vh] overflow-hidden flex flex-col">
        <div class="bg-slate-800 border-b border-slate-700 p-6 flex justify-between items-center flex-shrink-0">
          <div>
            <h3 class="text-xl font-bold">配置详情</h3>
            <p class="text-sm text-slate-400 mt-1">{{ configDetails?.username }}</p>
          </div>
          <button class="text-slate-400 hover:text-white" @click="closeConfigDetailsSidebar">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="overflow-y-auto flex-1">
          <div v-if="loadingDetails" class="p-12 text-center">
            <svg class="animate-spin h-12 w-12 mx-auto text-blue-500" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
          </div>

          <div v-else-if="configDetails" class="p-6">
            <!-- 标签导航 -->
            <div class="flex gap-2 mb-6 border-b border-slate-700">
              <button
                v-for="tab in detailTabs"
                :key="tab.key"
                :class="[
                  'px-4 py-3 text-sm font-medium transition-all border-b-2',
                  activeTab === tab.key
                    ? 'border-blue-500 text-blue-400'
                    : 'border-transparent text-slate-400 hover:text-slate-300'
                ]"
                @click="activeTab = tab.key"
              >
                {{ tab.label }}
              </button>
              <div class="ml-auto flex items-center gap-2">
                <button class="btn btn-secondary text-sm" :disabled="loadingTab" @click="refreshCurrentTab">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                    />
                  </svg>
                  刷新
                </button>
              </div>
            </div>

            <!-- 基本信息标签 -->
            <div v-show="activeTab === 'basic'" class="space-y-4">
              <div v-if="loadingTab" class="card p-6">
                <div class="text-center py-8">
                  <svg class="animate-spin h-8 w-8 mx-auto text-blue-500" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path
                      class="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    ></path>
                  </svg>
                  <p class="text-slate-400 mt-2">加载信息中...</p>
                </div>
              </div>
              <div v-else class="card p-6">
                <h4 class="text-lg font-semibold mb-4 flex items-center gap-2">
                  <svg class="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
                    />
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                    />
                  </svg>
                  配置与租户信息
                </h4>
                <div class="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <label class="text-slate-400">配置名称</label>
                    <p class="text-white font-medium">{{ configDetails.username }}</p>
                  </div>
                  <div>
                    <label class="text-slate-400">当前区域</label>
                    <p class="text-white">
                      <span class="px-2 py-1 text-xs font-semibold rounded-full bg-blue-500/20 text-blue-300">{{
                        configDetails.region
                      }}</span>
                    </p>
                  </div>
                  <div v-if="tabTenant" class="col-span-2">
                    <label class="text-slate-400">租户名称</label>
                    <p class="text-white font-medium">{{ tabTenant.name }}</p>
                  </div>
                  <div v-if="tabTenant" class="col-span-2">
                    <label class="text-slate-400">租户ID</label>
                    <p class="text-white font-mono text-xs break-all">{{ tabTenant.id }}</p>
                  </div>
                  <div v-if="tabTenant">
                    <label class="text-slate-400">主区域</label>
                    <p class="text-white">{{ tabTenant.homeRegionKey }}</p>
                  </div>
                  <div v-if="tabTenant && tabTenant.createTime">
                    <label class="text-slate-400">账户创建时间</label>
                    <p class="text-white text-xs">{{ tabTenant.createTime }}</p>
                  </div>
                  <div class="col-span-2">
                    <label class="text-slate-400">用户ID</label>
                    <p class="text-white font-mono text-xs break-all">{{ configDetails.userId }}</p>
                  </div>
                  <div>
                    <label class="text-slate-400">指纹</label>
                    <p class="text-white font-mono text-xs break-all">{{ configDetails.fingerprint }}</p>
                  </div>
                  <div>
                    <label class="text-slate-400">密钥文件</label>
                    <p class="text-white text-xs">{{ configDetails.keyPath }}</p>
                  </div>
                  <div>
                    <label class="text-slate-400">配置创建时间</label>
                    <p class="text-white text-xs">{{ configDetails.createTime }}</p>
                  </div>
                  <div v-if="tabTenant">
                    <label class="text-slate-400 mb-1 block">密码过期时间（天）</label>
                    <div class="flex items-center gap-2">
                      <input
                        v-if="editingPasswordExpiry"
                        v-model.number="passwordExpiryInput"
                        type="number"
                        min="0"
                        max="365"
                        class="input text-sm py-1 px-2 w-24"
                        placeholder="0"
                      />
                      <span v-else class="text-white">{{
                        tabTenant.passwordExpiresAfter === 0 ? '永不过期' : tabTenant.passwordExpiresAfter + ' 天'
                      }}</span>
                      <button
                        v-if="!editingPasswordExpiry"
                        class="text-blue-400 hover:text-blue-300 text-xs"
                        title="修改"
                        @click="startEditPasswordExpiry"
                      >
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                          />
                        </svg>
                      </button>
                      <div v-else class="flex gap-1">
                        <button
                          class="text-green-400 hover:text-green-300 text-xs"
                          :disabled="updatingPasswordExpiry"
                          title="保存"
                          @click="savePasswordExpiry"
                        >
                          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                          </svg>
                        </button>
                        <button
                          class="text-red-400 hover:text-red-300 text-xs"
                          :disabled="updatingPasswordExpiry"
                          title="取消"
                          @click="cancelEditPasswordExpiry"
                        >
                          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path
                              stroke-linecap="round"
                              stroke-linejoin="round"
                              stroke-width="2"
                              d="M6 18L18 6M6 6l12 12"
                            />
                          </svg>
                        </button>
                      </div>
                    </div>
                    <p class="text-xs text-slate-500 mt-1">设置为 0 表示永不过期</p>
                  </div>
                </div>

                <!-- 订阅区域 -->
                <div v-if="tabTenant && tabTenant.regions?.length" class="mt-4 pt-4 border-t border-slate-700">
                  <label class="text-slate-400 text-sm block mb-2">订阅区域 ({{ tabTenant.regions.length }})</label>
                  <div class="flex flex-wrap gap-2">
                    <span
                      v-for="region in tabTenant.regions"
                      :key="region"
                      class="px-2 py-1 text-xs rounded bg-blue-500/20 text-blue-300"
                    >{{ region }}</span>
                  </div>
                </div>
              </div>

              <!-- 用户列表卡片 -->
              <div v-if="tabTenant && tabTenant.userList?.length" class="card p-6">
                <h4 class="text-lg font-semibold mb-4 flex items-center justify-between">
                  <span class="flex items-center gap-2">
                    <svg class="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"
                      />
                    </svg>
                    用户列表 ({{ tabTenant.userList.length }})
                  </span>
                </h4>
                <div class="space-y-3">
                  <div
                    v-for="user in tabTenant.userList"
                    :key="user.id"
                    class="border border-slate-600 rounded-lg p-4 hover:border-blue-500/50 transition-all"
                  >
                    <div class="flex justify-between items-start mb-2">
                      <div class="flex-1">
                        <h5 class="text-white font-semibold">{{ user.name }}</h5>
                        <p v-if="user.email" class="text-slate-400 text-xs mt-1">{{ user.email }}</p>
                      </div>
                      <div class="flex gap-2 items-center ml-4">
                        <span
                          v-if="user.isMfaActivated"
                          class="text-xs px-2 py-0.5 rounded bg-green-500/20 text-green-300"
                        >MFA</span>
                        <span v-if="user.emailVerified" class="text-xs px-2 py-0.5 rounded bg-blue-500/20 text-blue-300">已验证</span>
                        <span
                          class="text-xs px-2 py-0.5 rounded"
                          :class="
                            user.state === 'ACTIVE' ? 'bg-green-500/20 text-green-300' : 'bg-red-500/20 text-red-300'
                          "
                        >{{ user.state }}</span>
                      </div>
                    </div>
                    <div class="text-xs text-slate-400 mb-3">
                      创建时间: {{ user.createTime }}
                      <span v-if="user.lastSuccessfulLoginTime" class="ml-4">最近登录: {{ user.lastSuccessfulLoginTime }}</span>
                    </div>
                    <div class="flex gap-2">
                      <button class="btn btn-primary text-xs py-1" title="编辑用户" @click="editUser(user)">
                        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                          />
                        </svg>
                        编辑
                      </button>
                      <button class="btn btn-warning text-xs py-1" title="重置密码" @click="resetUserPassword(user)">
                        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"
                          />
                        </svg>
                        重置密码
                      </button>
                      <button
                        v-if="user.isMfaActivated"
                        class="btn btn-secondary text-xs py-1"
                        title="清除MFA"
                        @click="clearUserMfa(user)"
                      >
                        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
                          />
                        </svg>
                        清除MFA
                      </button>
                      <button
                        class="btn btn-secondary text-xs py-1"
                        title="清除API密钥"
                        @click="clearUserApiKeys(user)"
                      >
                        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M8 11V7a4 4 0 118 0m-4 8v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2z"
                          />
                        </svg>
                        清除API
                      </button>
                      <button class="btn btn-danger text-xs py-1" title="删除用户" @click="deleteUser(user)">
                        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                          />
                        </svg>
                        删除
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 实例列表标签 -->
            <div v-show="activeTab === 'instances'">
              <div v-if="loadingTab" class="text-center py-12">
                <svg class="animate-spin h-12 w-12 mx-auto text-blue-500" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path
                    class="opacity-75"
                    fill="currentColor"
                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                  ></path>
                </svg>
                <p class="text-slate-400 mt-4">加载中...</p>
              </div>
              <div v-else-if="!tabInstances?.length" class="card p-12 text-center">
                <svg
                  class="w-16 h-16 mx-auto text-slate-600 mb-4"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"
                  />
                </svg>
                <p class="text-slate-400">暂无实例</p>
              </div>
              <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-4">
                <div
                  v-for="instance in tabInstances"
                  :key="instance.id"
                  class="card p-5 hover:border-blue-500/50 transition-all"
                >
                  <div class="flex justify-between items-start mb-3">
                    <div class="flex-1 min-w-0">
                      <h5 class="font-semibold text-white truncate">{{ instance.displayName }}</h5>
                      <p class="text-xs text-slate-400 font-mono truncate">{{ instance.id }}</p>
                    </div>
                    <span
                      class="badge ml-2 flex-shrink-0"
                      :class="{
                        'badge-success': instance.state === 'RUNNING',
                        'badge-danger': instance.state === 'STOPPED',
                        'badge-warning': !['RUNNING', 'STOPPED'].includes(instance.state)
                      }"
                    >{{ instance.state }}</span>
                  </div>
                  <div class="grid grid-cols-2 gap-2 text-sm mb-4">
                    <div>
                      <span class="text-slate-400">规格:</span><span class="ml-2 text-white">{{ instance.shape }}</span>
                    </div>
                    <div>
                      <span class="text-slate-400">CPU/内存:</span><span class="ml-2 text-white">{{ instance.ocpus }}核 / {{ instance.memory }}GB</span>
                    </div>
                    <div>
                      <span class="text-slate-400">引导卷:</span><span class="ml-2 text-white">{{ instance.bootVolumeSize || '-' }} GB</span>
                    </div>
                    <div>
                      <span class="text-slate-400">区域:</span><span class="ml-2 text-white">{{ instance.region }}</span>
                    </div>
                    <div class="col-span-2">
                      <span class="text-slate-400">公网IP:</span><span class="ml-2 text-white font-mono text-xs">{{
                        instance.publicIps?.join(', ') || '无'
                      }}</span>
                    </div>
                    <div v-if="instance.ipv6" class="col-span-2">
                      <span class="text-slate-400">IPv6:</span><span class="ml-2 text-blue-300 font-mono text-xs">{{ instance.ipv6 }}</span>
                    </div>
                    <div v-if="instance.imageName" class="col-span-2">
                      <span class="text-slate-400">镜像:</span><span class="ml-2 text-white text-xs">{{ instance.imageName }}</span>
                    </div>
                  </div>
                  <div class="space-y-2">
                    <div class="grid grid-cols-4 gap-2">
                      <button
                        class="btn btn-success text-xs py-1.5"
                        :disabled="instance.state === 'RUNNING' || instanceActionLoading[instance.id]"
                        title="启动"
                        @click="controlInstanceInDetails(instance.id, 'START')"
                      >
                        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"
                          />
                        </svg>
                        启动
                      </button>
                      <button
                        class="btn btn-warning text-xs py-1.5"
                        :disabled="instance.state !== 'RUNNING' || instanceActionLoading[instance.id]"
                        title="停止"
                        @click="controlInstanceInDetails(instance.id, 'STOP')"
                      >
                        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                          />
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M9 10a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z"
                          />
                        </svg>
                        停止
                      </button>
                      <button
                        class="btn btn-secondary text-xs py-1.5"
                        :disabled="instance.state !== 'RUNNING' || instanceActionLoading[instance.id]"
                        title="重启"
                        @click="controlInstanceInDetails(instance.id, 'SOFTRESET')"
                      >
                        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                          />
                        </svg>
                        重启
                      </button>
                      <button
                        class="btn btn-danger text-xs py-1.5"
                        :disabled="instanceActionLoading[instance.id]"
                        title="删除"
                        @click="terminateInstanceInDetails(instance.id)"
                      >
                        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                          />
                        </svg>
                        删除
                      </button>
                    </div>
                    <div class="grid grid-cols-3 gap-2">
                      <button
                        class="btn btn-info text-xs py-1.5"
                        :disabled="instanceActionLoading[instance.id]"
                        title="更改IP"
                        @click="changeIPInDetails(instance.id)"
                      >
                        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4"
                          />
                        </svg>
                        更改IP
                      </button>
                      <button
                        class="btn btn-info text-xs py-1.5"
                        :disabled="instanceActionLoading[instance.id]"
                        title="编辑配置"
                        @click="showEditConfigDialog(instance)"
                      >
                        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
                          />
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                          />
                        </svg>
                        编辑配置
                      </button>
                      <button
                        class="btn btn-info text-xs py-1.5"
                        :disabled="instanceActionLoading[instance.id]"
                        title="Cloud Shell"
                        @click="showCloudShellDialog(instance.id)"
                      >
                        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                          />
                        </svg>
                        Cloud Shell
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 引导卷列表标签 -->
            <div v-show="activeTab === 'volumes'">
              <div v-if="loadingTab" class="text-center py-12">
                <svg class="animate-spin h-12 w-12 mx-auto text-blue-500" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path
                    class="opacity-75"
                    fill="currentColor"
                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                  ></path>
                </svg>
                <p class="text-slate-400 mt-4">加载中...</p>
              </div>
              <div v-else-if="!tabVolumes?.length" class="card p-12 text-center">
                <svg
                  class="w-16 h-16 mx-auto text-slate-600 mb-4"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"
                  />
                </svg>
                <p class="text-slate-400">暂无引导卷</p>
              </div>
              <div v-else class="space-y-3">
                <div
                  v-for="volume in tabVolumes"
                  :key="volume.id"
                  class="border border-slate-600 rounded-lg p-4 hover:border-blue-500/50 transition-all"
                >
                  <div class="flex justify-between items-start mb-3">
                    <div class="flex items-center gap-3">
                      <div class="w-10 h-10 rounded-lg bg-blue-500/20 flex items-center justify-center">
                        <svg class="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
                        </svg>
                      </div>
                      <div>
                        <h5 class="font-semibold text-white">{{ volume.displayName }}</h5>
                        <p class="text-xs text-slate-400">{{ volume.id?.substring(0, 30) }}...</p>
                      </div>
                    </div>
                    <div class="flex items-center gap-2">
                      <span v-if="volume.attached" class="text-xs px-2 py-1 rounded bg-green-500/20 text-green-300">已附加</span>
                      <span v-else class="text-xs px-2 py-1 rounded bg-yellow-500/20 text-yellow-300">未附加</span>
                      <span
                        class="text-xs px-2 py-1 rounded"
                        :class="{
                          'bg-green-500/20 text-green-300': volume.state === 'AVAILABLE',
                          'bg-yellow-500/20 text-yellow-300': volume.state === 'PROVISIONING',
                          'bg-red-500/20 text-red-300': volume.state === 'FAULTY'
                        }"
                      >{{ volume.state }}</span>
                      <button 
                        class="btn btn-primary btn-sm" 
                        title="编辑引导卷"
                        @click="openVolumeEditDialog(volume)"
                      >
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                        </svg>
                      </button>
                    </div>
                  </div>
                  <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                    <div class="bg-slate-700/30 rounded-lg p-3">
                      <span class="text-slate-400 text-xs block mb-1">磁盘大小</span>
                      <span class="text-white font-semibold">{{ volume.sizeInGBs }} GB</span>
                    </div>
                    <div class="bg-slate-700/30 rounded-lg p-3">
                      <span class="text-slate-400 text-xs block mb-1">性能 (VPU/GB)</span>
                      <span class="text-white font-semibold">{{ volume.vpusPerGB || 10 }}</span>
                    </div>
                    <div v-if="volume.instanceName" class="bg-slate-700/30 rounded-lg p-3">
                      <span class="text-slate-400 text-xs block mb-1">附加实例</span>
                      <span class="text-blue-300 font-semibold">{{ volume.instanceName }}</span>
                    </div>
                    <div class="bg-slate-700/30 rounded-lg p-3">
                      <span class="text-slate-400 text-xs block mb-1">可用域</span>
                      <span class="text-white text-xs">{{ volume.availabilityDomain?.split(':').pop() || '-' }}</span>
                    </div>
                  </div>
                  <div v-if="volume.createTime" class="mt-3 text-xs text-slate-400">
                    创建时间: {{ volume.createTime }}
                  </div>
                </div>
              </div>
            </div>

            <!-- VCN列表标签 -->
            <div v-show="activeTab === 'vcns'">
              <div v-if="loadingTab" class="text-center py-12">
                <svg class="animate-spin h-12 w-12 mx-auto text-blue-500" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path
                    class="opacity-75"
                    fill="currentColor"
                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                  ></path>
                </svg>
                <p class="text-slate-400 mt-4">加载中...</p>
              </div>
              <div v-else-if="!tabVCNs?.length" class="card p-12 text-center">
                <svg
                  class="w-16 h-16 mx-auto text-slate-600 mb-4"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9"
                  />
                </svg>
                <p class="text-slate-400">暂无VCN</p>
              </div>
              <div v-else class="space-y-3">
                <div v-for="vcn in tabVCNs" :key="vcn.id" class="border border-slate-600 rounded-lg p-4">
                  <div class="flex justify-between items-start mb-2">
                    <h5 class="font-semibold text-white">{{ vcn.displayName }}</h5>
                    <span class="text-xs px-2 py-1 rounded bg-green-500/20 text-green-300">{{ vcn.state }}</span>
                  </div>
                  <div class="space-y-1 text-sm">
                    <div>
                      <span class="text-slate-400">CIDR:</span><span class="ml-2 text-white font-mono">{{ vcn.cidrBlock }}</span>
                    </div>
                    <div v-if="vcn.createTime">
                      <span class="text-slate-400">创建时间:</span><span class="ml-2 text-white">{{ vcn.createTime }}</span>
                    </div>
                    <div v-if="vcn.subnets?.length" class="mt-2">
                      <span class="text-slate-400 block mb-2">子网 ({{ vcn.subnets.length }}个):</span>
                      <div class="pl-4 space-y-1">
                        <div v-for="subnet in vcn.subnets" :key="subnet.id" class="text-xs bg-slate-700/50 rounded p-2">
                          <div class="flex justify-between">
                            <span class="text-white">{{ subnet.displayName }}</span>
                            <span :class="subnet.isPublic ? 'text-green-400' : 'text-yellow-400'">{{
                              subnet.isPublic ? '公有' : '私有'
                            }}</span>
                          </div>
                          <div class="text-slate-400">{{ subnet.cidrBlock }}</div>
                        </div>
                      </div>
                    </div>
                    <!-- VCN操作按钮 -->
                    <div class="flex gap-2 mt-3 pt-3 border-t border-slate-600">
                      <button 
                        class="btn btn-primary btn-sm" 
                        title="查看安全列表"
                        @click="viewSecurityList(vcn)"
                      >
                        <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                        </svg>
                        安全列表
                      </button>
                      <button 
                        class="btn btn-success btn-sm" 
                        :disabled="vcnActionLoading[vcn.id]"
                        title="一键放行所有端口"
                        @click="releaseVcnRules(vcn)"
                      >
                        <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 11V7a4 4 0 118 0m-4 8v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2z" />
                        </svg>
                        {{ vcnActionLoading[vcn.id] ? '处理中...' : '放行规则' }}
                      </button>
                      <button 
                        class="btn btn-danger btn-sm" 
                        :disabled="vcnActionLoading[vcn.id]"
                        title="删除VCN"
                        @click="deleteVcn(vcn)"
                      >
                        <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                        删除
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 流量统计标签 -->
            <div v-show="activeTab === 'traffic'">
              <div class="space-y-4">
                <div class="card p-4">
                  <h5 class="font-semibold text-white mb-3">查询条件</h5>
                  <div class="grid grid-cols-2 gap-4">
                    <div>
                      <label class="block text-sm text-slate-400 mb-1">选择实例</label>
                      <select v-model="trafficForm.instanceId" class="input text-sm">
                        <option value="">请选择实例</option>
                        <option v-for="inst in trafficCondition.instances" :key="inst.value" :value="inst.value">
                          {{ inst.label }}
                        </option>
                      </select>
                    </div>
                    <div>
                      <label class="block text-sm text-slate-400 mb-1">选择VNIC</label>
                      <select v-model="trafficForm.vnicId" class="input text-sm" :disabled="!trafficVnics.length">
                        <option value="">请选择VNIC</option>
                        <option v-for="vnic in trafficVnics" :key="vnic.value" :value="vnic.value">
                          {{ vnic.label }}
                        </option>
                      </select>
                    </div>
                    <div>
                      <label class="block text-sm text-slate-400 mb-1">开始时间</label>
                      <input
                        v-model="trafficForm.startTime"
                        type="text"
                        class="input text-sm"
                        placeholder="YYYY-MM-DD HH:mm:ss"
                      />
                    </div>
                    <div>
                      <label class="block text-sm text-slate-400 mb-1">结束时间</label>
                      <input
                        v-model="trafficForm.endTime"
                        type="text"
                        class="input text-sm"
                        placeholder="YYYY-MM-DD HH:mm:ss"
                      />
                    </div>
                  </div>
                  <button class="btn btn-primary mt-4" :disabled="loadingTraffic" @click="loadTrafficData">
                    {{ loadingTraffic ? '查询中...' : '查询流量' }}
                  </button>
                </div>

                <div v-if="loadingTraffic" class="text-center py-12">
                  <svg class="animate-spin h-12 w-12 mx-auto text-blue-500" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path
                      class="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    ></path>
                  </svg>
                  <p class="text-slate-400 mt-4">加载中...</p>
                </div>
                <div v-else-if="tabTraffic.time?.length" class="card p-4">
                  <h5 class="font-semibold text-white mb-3">流量数据 (单位: MB)</h5>
                  <div class="overflow-x-auto">
                    <table class="w-full text-sm">
                      <thead class="bg-slate-700/50">
                        <tr>
                          <th class="px-3 py-2 text-left text-slate-300">时间</th>
                          <th class="px-3 py-2 text-left text-green-400">入站 (MB)</th>
                          <th class="px-3 py-2 text-left text-blue-400">出站 (MB)</th>
                        </tr>
                      </thead>
                      <tbody class="divide-y divide-slate-700">
                        <tr v-for="(time, index) in tabTraffic.time" :key="index" class="hover:bg-slate-700/30">
                          <td class="px-3 py-2 text-slate-300">{{ time }}</td>
                          <td class="px-3 py-2 text-green-400">{{ tabTraffic.inbound[index] || '0' }}</td>
                          <td class="px-3 py-2 text-blue-400">{{ tabTraffic.outbound[index] || '0' }}</td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                </div>
                <div v-else class="card p-12 text-center">
                  <svg
                    class="w-16 h-16 mx-auto text-slate-600 mb-4"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
                    />
                  </svg>
                  <p class="text-slate-400">请选择实例和VNIC后查询流量数据</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 创建实例弹窗 -->
    <div
      v-if="showCreateInstanceModal"
      class="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50 p-4"
    >
      <div class="card max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <div class="p-6 border-b border-slate-700 flex justify-between items-center">
          <h3 class="text-xl font-bold">创建实例任务</h3>
          <button class="text-slate-400 hover:text-white" @click="closeInstanceModal">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <form class="p-6 space-y-4" @submit.prevent="submitInstanceTask">
          <div class="bg-blue-500/10 border border-blue-500/30 rounded-lg p-4 mb-4">
            <p class="text-sm text-blue-300">
              <svg class="w-4 h-4 inline mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fill-rule="evenodd"
                  d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                  clip-rule="evenodd"
                />
              </svg>
              为配置 <strong>{{ selectedConfigForInstance?.username }}</strong> 创建实例
            </p>
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">区域</label>
            <input
              v-model="instanceForm.ociRegion"
              type="text"
              class="input"
              placeholder="例: ap-singapore-1"
              required
            />
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">CPU核心数</label>
              <input v-model.number="instanceForm.ocpus" type="number" step="0.1" min="0.1" class="input" required />
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">内存(GB)</label>
              <input v-model.number="instanceForm.memory" type="number" step="0.1" min="0.1" class="input" required />
            </div>
          </div>

          <div class="grid grid-cols-3 gap-4">
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">磁盘(GB)</label>
              <input v-model.number="instanceForm.disk" type="number" min="50" class="input" required />
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">VPU/GB</label>
              <select v-model.number="instanceForm.bootVolumeVpu" class="input">
                <option v-for="vpu in [10,20,30,40,50,60,70,80,90,100,110,120]" :key="vpu" :value="vpu">{{ vpu }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">架构</label>
              <select v-model="instanceForm.architecture" class="input" required @change="onArchitectureChange">
                <option value="ARM">ARM</option>
                <option value="AMD">AMD</option>
              </select>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">操作系统类型</label>
              <select v-model="instanceForm.operationSystem" class="input" required @change="onOperationSystemChange">
                <option value="Ubuntu">Ubuntu</option>
                <option value="CentOS">CentOS</option>
                <option value="Oracle Linux">Oracle Linux</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">系统版本</label>
              <select v-model="instanceForm.imageId" class="input" :disabled="loadingImages">
                <option value="">{{ loadingImages ? '加载中...' : (filteredImages.length === 0 ? '无可用镜像' : '自动选择最新版本') }}</option>
                <option v-for="img in filteredImages" :key="img.id" :value="img.id">
                  {{ img.operatingSystem }} {{ img.operatingSystemVersion }}
                </option>
              </select>
              <p class="text-xs text-slate-400 mt-1">
                {{ filteredImages.length > 0 ? `共 ${filteredImages.length} 个可用镜像` : '' }}
              </p>
            </div>
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">SSH公钥</label>
            <select v-model="instanceForm.sshKeyId" class="input" required>
              <option value="">请选择SSH公钥</option>
              <option v-for="key in sshKeys" :key="key.id" :value="key.id">{{ key.name }}</option>
            </select>
            <p class="text-xs text-slate-400 mt-2">
              请先在 <router-link to="/keys" class="text-blue-400 hover:underline">密钥管理</router-link> 中添加SSH公钥
            </p>
          </div>

          <div class="flex items-center gap-3 py-2">
            <label class="flex items-center gap-2 cursor-pointer">
              <input v-model="instanceForm.isTaskMode" type="checkbox" class="checkbox" />
              <span class="text-sm font-medium text-slate-300">抢占实例任务</span>
            </label>
            <span class="text-xs text-slate-400">（开启后将持续尝试创建实例直到成功）</span>
          </div>

          <div v-if="instanceForm.isTaskMode">
            <label class="block text-sm font-medium text-slate-300 mb-2">执行间隔（秒）</label>
            <input v-model.number="instanceForm.interval" type="number" min="10" class="input" placeholder="60" />
            <p class="text-xs text-slate-400 mt-2">
              任务将每隔指定秒数尝试创建实例，创建成功后自动停止。最小10秒。
            </p>
          </div>

          <div class="flex gap-3 pt-4">
            <button type="button" class="btn btn-secondary flex-1" @click="closeInstanceModal">取消</button>
            <button type="submit" class="btn btn-primary flex-1" :disabled="submittingInstance">
              {{ submittingInstance ? '创建中...' : (instanceForm.isTaskMode ? '创建任务' : '创建实例') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- 批量创建实例弹窗 -->
    <div
      v-if="showBatchCreateModal"
      class="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50 p-4"
    >
      <div class="card max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <div class="p-6 border-b border-slate-700 flex justify-between items-center">
          <h3 class="text-xl font-bold">批量创建实例任务</h3>
          <button class="text-slate-400 hover:text-white" @click="closeBatchCreateModal">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <form class="p-6 space-y-4" @submit.prevent="submitBatchInstanceTask">
          <div class="bg-blue-500/10 border border-blue-500/30 rounded-lg p-4 mb-4">
            <p class="text-sm text-blue-300">
              <svg class="w-4 h-4 inline mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fill-rule="evenodd"
                  d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                  clip-rule="evenodd"
                />
              </svg>
              将为 <strong>{{ selectedConfigIds.length }}</strong> 个配置批量创建实例（每个配置使用其自身区域）
            </p>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">CPU核心数</label>
              <input v-model.number="instanceForm.ocpus" type="number" step="0.1" min="0.1" class="input" required />
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">内存(GB)</label>
              <input v-model.number="instanceForm.memory" type="number" step="0.1" min="0.1" class="input" required />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">磁盘(GB)</label>
              <input v-model.number="instanceForm.disk" type="number" min="50" class="input" required />
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">架构</label>
              <select v-model="instanceForm.architecture" class="input" required>
                <option value="ARM">ARM</option>
                <option value="AMD">AMD</option>
              </select>
            </div>
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">操作系统</label>
            <select v-model="instanceForm.operationSystem" class="input" required>
              <option value="Ubuntu">Ubuntu</option>
              <option value="CentOS">CentOS</option>
              <option value="Oracle Linux">Oracle Linux</option>
            </select>
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">SSH公钥</label>
            <select v-model="instanceForm.sshKeyId" class="input" required>
              <option value="">请选择SSH公钥</option>
              <option v-for="key in sshKeys" :key="key.id" :value="key.id">{{ key.name }}</option>
            </select>
            <p class="text-xs text-slate-400 mt-2">
              请先在 <router-link to="/keys" class="text-blue-400 hover:underline">密钥管理</router-link> 中添加SSH公钥
            </p>
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">执行间隔（秒）</label>
            <input v-model.number="instanceForm.interval" type="number" min="10" class="input" placeholder="60" />
            <p class="text-xs text-slate-400 mt-2">
              每个任务将每隔指定秒数尝试创建实例，创建成功后自动停止。最小10秒。
            </p>
          </div>

          <div class="flex gap-3 pt-4">
            <button type="button" class="btn btn-secondary flex-1" @click="closeBatchCreateModal">取消</button>
            <button type="submit" class="btn btn-primary flex-1" :disabled="submittingInstance">
              {{ submittingInstance ? '创建中...' : '批量创建' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- 编辑用户弹窗 -->
    <div
      v-if="showEditUserModal"
      class="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50 p-4"
    >
      <div class="card max-w-md w-full">
        <div class="p-6 border-b border-slate-700 flex justify-between items-center">
          <h3 class="text-xl font-bold">编辑用户信息</h3>
          <button class="text-slate-400 hover:text-white" @click="closeEditUserModal">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <form class="p-6 space-y-4" @submit.prevent="saveUserInfo">
          <div class="bg-blue-500/10 border border-blue-500/30 rounded-lg p-4 mb-4">
            <p class="text-sm text-blue-300">
              <svg class="w-4 h-4 inline mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fill-rule="evenodd"
                  d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                  clip-rule="evenodd"
                />
              </svg>
              编辑用户 <strong>{{ editingUser?.name }}</strong>
            </p>
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">用户名</label>
            <input v-model="userForm.dbUserName" type="text" class="input" placeholder="输入用户名" required />
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">邮箱</label>
            <input v-model="userForm.email" type="email" class="input" placeholder="输入邮箱地址" required />
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">描述（可选）</label>
            <textarea v-model="userForm.description" class="input min-h-[80px]" placeholder="输入用户描述"></textarea>
          </div>

          <div class="flex gap-3 pt-4">
            <button type="button" class="btn btn-secondary flex-1" @click="closeEditUserModal">取消</button>
            <button type="submit" class="btn btn-primary flex-1">保存</button>
          </div>
        </form>
      </div>
    </div>

    <!-- 编辑实例配置弹窗 -->
    <div
      v-if="editConfigDialogVisible"
      class="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50 p-4"
    >
      <div class="card max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <div class="p-6 border-b border-slate-700 flex justify-between items-center">
          <h3 class="text-xl font-bold">编辑实例配置</h3>
          <button class="text-slate-400 hover:text-white" @click="editConfigDialogVisible = false">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="p-6 space-y-6">
          <!-- 实例名称 -->
          <div class="card p-4 bg-slate-700/30">
            <h4 class="text-sm font-semibold text-slate-300 mb-3">修改实例名称</h4>
            <div class="flex gap-3">
              <input v-model="editConfigForm.displayName" type="text" class="input flex-1" placeholder="请输入实例名称" />
              <button class="btn btn-primary" :disabled="configUpdating" @click="updateInstanceName">
                {{ configUpdating ? '修改中...' : '修改名称' }}
              </button>
            </div>
          </div>

          <!-- CPU和内存 -->
          <div class="card p-4 bg-slate-700/30">
            <h4 class="text-sm font-semibold text-slate-300 mb-3">修改CPU和内存</h4>
            <div class="bg-yellow-500/10 border border-yellow-500/30 rounded-lg p-3 mb-3">
              <p class="text-xs text-yellow-300">
                <svg class="w-4 h-4 inline mr-1" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
                </svg>
                修改CPU和内存需要停止实例
              </p>
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-xs text-slate-400 mb-1">CPU核心数 (OCPUs)</label>
                <input v-model.number="editConfigForm.ocpus" type="number" min="1" max="64" class="input" />
              </div>
              <div>
                <label class="block text-xs text-slate-400 mb-1">内存 (GB)</label>
                <input v-model.number="editConfigForm.memoryInGBs" type="number" min="1" max="1024" class="input" />
              </div>
            </div>
            <button class="btn btn-primary mt-3 w-full" :disabled="configUpdating" @click="updateInstanceConfig">
              {{ configUpdating ? '修改中...' : '修改CPU和内存' }}
            </button>
          </div>

          <!-- 引导卷配置 -->
          <div class="card p-4 bg-slate-700/30">
            <h4 class="text-sm font-semibold text-slate-300 mb-3">修改引导卷大小及VPU</h4>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-xs text-slate-400 mb-1">引导卷大小 (GB)</label>
                <input v-model.number="editConfigForm.bootVolumeSize" type="number" min="50" max="32768" class="input" />
              </div>
              <div>
                <label class="block text-xs text-slate-400 mb-1">VPU/GB (性能)</label>
                <select v-model.number="editConfigForm.vpusPerGB" class="input">
                  <option :value="10">10 VPU - 平衡</option>
                  <option :value="20">20 VPU - 高性能</option>
                  <option :value="30">30 VPU - 更高性能</option>
                  <option :value="40">40 VPU - 极高性能</option>
                  <option :value="50">50 VPU - 超高性能</option>
                  <option :value="100">100+ VPU - 最高性能</option>
                  <option :value="120">120 VPU - 极致性能</option>
                </select>
              </div>
            </div>
            <button class="btn btn-primary mt-3 w-full" :disabled="configUpdating" @click="updateBootVolume">
              {{ configUpdating ? '修改中...' : '修改引导卷配置' }}
            </button>
          </div>

          <!-- IPv6 -->
          <div class="card p-4 bg-slate-700/30">
            <h4 class="text-sm font-semibold text-slate-300 mb-3">附加IPv6地址</h4>
            <div class="flex items-center justify-between">
              <div class="text-sm text-slate-400">
                <span v-if="editConfigForm.currentIpv6">当前IPv6: <span class="text-blue-300 font-mono">{{ editConfigForm.currentIpv6 }}</span></span>
                <span v-else>该实例未分配IPv6地址</span>
              </div>
              <button 
                class="btn btn-primary" 
                :disabled="configUpdating || editConfigForm.currentIpv6" 
                @click="attachIPv6"
              >
                {{ configUpdating ? '附加中...' : (editConfigForm.currentIpv6 ? '已有IPv6' : '附加IPv6') }}
              </button>
            </div>
          </div>

          <!-- 自动救援/缩小硬盘 -->
          <div class="card p-4 bg-slate-700/30">
            <h4 class="text-sm font-semibold text-slate-300 mb-3">自动救援/缩小硬盘</h4>
            <div class="space-y-3">
              <div class="bg-yellow-500/10 border border-yellow-500/30 rounded-lg p-3">
                <p class="text-xs text-yellow-300">
                  <svg class="w-4 h-4 inline mr-1" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
                  </svg>
                  此操作会关闭实例，备份引导卷，创建47GB新引导卷并重新启动。过程需要5-10分钟。
                </p>
              </div>
              <div class="flex items-center gap-4">
                <label class="flex items-center gap-2 text-sm text-slate-300 cursor-pointer">
                  <input v-model="editConfigForm.keepBackup" type="checkbox" class="checkbox" />
                  <span>保留原引导卷备份</span>
                </label>
                <button 
                  class="btn btn-warning" 
                  :disabled="configUpdating" 
                  @click="startAutoRescue"
                >
                  {{ configUpdating ? '执行中...' : '开始救援' }}
                </button>
              </div>
            </div>
          </div>

          <!-- 500Mbps下行带宽 -->
          <div class="card p-4 bg-slate-700/30">
            <h4 class="text-sm font-semibold text-slate-300 mb-3">500Mbps下行带宽 (仅AMD实例)</h4>
            <div class="space-y-3">
              <div class="bg-blue-500/10 border border-blue-500/30 rounded-lg p-3">
                <p class="text-xs text-blue-300">
                  <svg class="w-4 h-4 inline mr-1" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
                  </svg>
                  通过NAT网关和网络负载均衡器实现500Mbps下行带宽。仅支持AMD E2.1.Micro实例。
                </p>
              </div>
              <div class="flex items-center gap-2">
                <label class="text-sm text-slate-300">SSH端口:</label>
                <input v-model.number="editConfigForm.sshPort" type="number" min="1" max="65535" class="input w-24" placeholder="22" />
              </div>
              <div class="flex gap-2">
                <button 
                  class="btn btn-success flex-1" 
                  :disabled="configUpdating" 
                  @click="enable500Mbps"
                >
                  {{ configUpdating ? '执行中...' : '开启500Mbps' }}
                </button>
                <button 
                  class="btn btn-danger flex-1" 
                  :disabled="configUpdating" 
                  @click="disable500Mbps"
                >
                  {{ configUpdating ? '执行中...' : '关闭500Mbps' }}
                </button>
              </div>
              <div class="flex items-center gap-4 text-xs text-slate-400">
                <label class="flex items-center gap-1 cursor-pointer">
                  <input v-model="editConfigForm.retainNatGw" type="checkbox" class="checkbox checkbox-sm" />
                  <span>保留NAT网关</span>
                </label>
                <label class="flex items-center gap-1 cursor-pointer">
                  <input v-model="editConfigForm.retainNlb" type="checkbox" class="checkbox checkbox-sm" />
                  <span>保留负载均衡器</span>
                </label>
              </div>
            </div>
          </div>

          <!-- 关闭按钮 -->
          <div class="flex justify-end pt-2">
            <button class="btn btn-secondary" @click="editConfigDialogVisible = false">关闭</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Cloud Shell弹窗 -->
    <div
      v-if="cloudShellDialogVisible"
      class="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50 p-4"
    >
      <div class="card max-w-2xl w-full">
        <div class="p-6 border-b border-slate-700 flex justify-between items-center">
          <h3 class="text-xl font-bold">Cloud Shell 连接</h3>
          <button class="text-slate-400 hover:text-white" @click="cloudShellDialogVisible = false">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="p-6 space-y-4">
          <div class="bg-blue-500/10 border border-blue-500/30 rounded-lg p-4">
            <p class="text-sm text-blue-300">
              <svg class="w-4 h-4 inline mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fill-rule="evenodd"
                  d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                  clip-rule="evenodd"
                />
              </svg>
              请提供SSH公钥以创建Cloud Shell连接
            </p>
          </div>

          <div v-if="!cloudShellResult.connectionString">
            <label class="block text-sm font-medium text-slate-300 mb-2">SSH 公钥</label>
            <textarea
              v-model="cloudShellForm.publicKey"
              class="input min-h-[120px] font-mono text-xs"
              placeholder="ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC..."
              required
            ></textarea>
            <p class="text-xs text-slate-400 mt-2">粘贴您的 SSH 公钥（通常位于 ~/.ssh/id_rsa.pub）</p>
          </div>

          <div v-if="cloudShellResult.connectionString" class="space-y-3">
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">连接ID</label>
              <div class="flex gap-2">
                <input
                  :value="cloudShellResult.connectionId"
                  readonly
                  class="input flex-1 font-mono text-xs bg-slate-700/50"
                />
                <button class="btn btn-secondary" title="复制" @click="copyToClipboard(cloudShellResult.connectionId)">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
                    />
                  </svg>
                </button>
              </div>
            </div>

            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">连接字符串</label>
              <div class="flex gap-2">
                <input
                  :value="cloudShellResult.connectionString"
                  readonly
                  class="input flex-1 font-mono text-xs bg-slate-700/50"
                />
                <button
                  class="btn btn-secondary"
                  title="复制"
                  @click="copyToClipboard(cloudShellResult.connectionString)"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
                    />
                  </svg>
                </button>
              </div>
            </div>

            <div class="bg-green-500/10 border border-green-500/30 rounded-lg p-4">
              <p class="text-sm text-green-300">
                <svg class="w-4 h-4 inline mr-2" fill="currentColor" viewBox="0 0 20 20">
                  <path
                    fill-rule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                    clip-rule="evenodd"
                  />
                </svg>
                连接创建成功！请使用SSH客户端连接。
              </p>
            </div>
          </div>

          <div class="flex gap-3 pt-4">
            <button type="button" class="btn btn-secondary flex-1" @click="cloudShellDialogVisible = false">
              关闭
            </button>
            <button
              v-if="!cloudShellResult.connectionString"
              class="btn btn-primary flex-1"
              :disabled="cloudShellCreating || !cloudShellForm.publicKey.trim()"
              @click="createCloudShell"
            >
              {{ cloudShellCreating ? '创建中...' : '创建连接' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 引导卷编辑弹窗 -->
    <div
      v-if="volumeEditDialogVisible"
      class="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50 p-4"
    >
      <div class="card max-w-lg w-full">
        <div class="p-6 border-b border-slate-700 flex justify-between items-center">
          <h3 class="text-xl font-bold">编辑引导卷</h3>
          <button class="text-slate-400 hover:text-white" @click="volumeEditDialogVisible = false">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="p-6 space-y-4">
          <!-- 引导卷信息 -->
          <div class="bg-slate-700/30 rounded-lg p-4">
            <div class="flex items-center gap-3 mb-3">
              <div class="w-10 h-10 rounded-lg bg-blue-500/20 flex items-center justify-center">
                <svg class="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
                </svg>
              </div>
              <div>
                <h5 class="font-semibold text-white">{{ volumeEditForm.displayName }}</h5>
                <p class="text-xs text-slate-400">{{ volumeEditForm.volumeId?.substring(0, 40) }}...</p>
              </div>
            </div>
            <div v-if="volumeEditForm.instanceName" class="text-sm">
              <span class="text-slate-400">附加实例:</span>
              <span class="ml-2 text-blue-300">{{ volumeEditForm.instanceName }}</span>
            </div>
          </div>

          <!-- 磁盘大小 -->
          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">磁盘大小 (GB)</label>
            <div class="flex items-center gap-3">
              <input 
                v-model.number="volumeEditForm.sizeInGBs" 
                type="number" 
                min="50" 
                max="32768" 
                class="input flex-1" 
                placeholder="输入磁盘大小"
              />
              <span class="text-slate-400 text-sm">当前: {{ volumeEditForm.originalSize }} GB</span>
            </div>
            <p class="text-xs text-slate-400 mt-1">最小 50GB，最大 32768GB。只能增大，不能缩小。</p>
          </div>

          <!-- VPU性能 -->
          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">性能 (VPU/GB)</label>
            <select v-model.number="volumeEditForm.vpusPerGB" class="input">
              <option :value="10">10 VPU/GB - 平衡性能</option>
              <option :value="20">20 VPU/GB - 高性能</option>
              <option :value="30">30 VPU/GB - 更高性能</option>
              <option :value="40">40 VPU/GB - 高级性能</option>
              <option :value="50">50 VPU/GB - 极高性能</option>
              <option :value="60">60 VPU/GB - 超高性能</option>
              <option :value="70">70 VPU/GB - 顶级性能</option>
              <option :value="80">80 VPU/GB - 旗舰性能</option>
              <option :value="90">90 VPU/GB - 极致性能</option>
              <option :value="100">100 VPU/GB - 最高性能</option>
              <option :value="110">110 VPU/GB - 超旗舰</option>
              <option :value="120">120 VPU/GB - 极限性能</option>
            </select>
            <p class="text-xs text-slate-400 mt-1">VPU 越高，IOPS 和吞吐量越高。当前: {{ volumeEditForm.originalVpu }} VPU/GB</p>
          </div>

          <!-- 警告提示 -->
          <div class="bg-yellow-500/10 border border-yellow-500/30 rounded-lg p-3">
            <p class="text-xs text-yellow-300">
              <svg class="w-4 h-4 inline mr-1" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
              </svg>
              修改引导卷配置可能需要重启实例才能生效。磁盘大小只能增大，不能缩小。
            </p>
          </div>

          <!-- 操作按钮 -->
          <div class="flex gap-3 pt-2">
            <button type="button" class="btn btn-secondary flex-1" @click="volumeEditDialogVisible = false">
              取消
            </button>
            <button 
              class="btn btn-primary flex-1" 
              :disabled="volumeUpdating"
              @click="updateVolumeConfig"
            >
              {{ volumeUpdating ? '更新中...' : '保存修改' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 安全列表弹窗 -->
    <div
      v-if="securityListDialogVisible"
      class="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50 p-4"
    >
      <div class="card max-w-4xl w-full max-h-[90vh] flex flex-col">
        <div class="p-6 border-b border-slate-700 flex justify-between items-center">
          <h3 class="text-xl font-bold">安全列表 - {{ currentSecurityList?.displayName }}</h3>
          <button class="text-slate-400 hover:text-white" @click="securityListDialogVisible = false">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="p-6 overflow-y-auto flex-1">
          <div v-if="loadingSecurityList" class="text-center py-8">
            <svg class="animate-spin h-8 w-8 mx-auto text-blue-500" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            <p class="text-slate-400 mt-2">加载中...</p>
          </div>

          <div v-else-if="currentSecurityList">
            <!-- 入站规则 -->
            <div class="mb-6">
              <div class="flex justify-between items-center mb-3">
                <h4 class="text-lg font-semibold text-green-400">入站规则 (Ingress)</h4>
                <button class="btn btn-primary btn-sm" @click="showAddRuleForm('ingress')">
                  <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                  </svg>
                  添加规则
                </button>
              </div>
              <div class="overflow-x-auto">
                <table class="w-full text-sm">
                  <thead class="bg-slate-700/50">
                    <tr>
                      <th class="px-3 py-2 text-left text-slate-300">协议</th>
                      <th class="px-3 py-2 text-left text-slate-300">来源</th>
                      <th class="px-3 py-2 text-left text-slate-300">端口</th>
                      <th class="px-3 py-2 text-left text-slate-300">描述</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-700">
                    <tr v-for="(rule, index) in currentSecurityList.ingressRules" :key="'in-' + index" class="hover:bg-slate-700/30">
                      <td class="px-3 py-2 text-white">{{ rule.protocolName }}</td>
                      <td class="px-3 py-2 text-white font-mono text-xs">{{ rule.source }}</td>
                      <td class="px-3 py-2 text-white">
                        {{ formatPortRange(rule) }}
                      </td>
                      <td class="px-3 py-2 text-slate-400">{{ rule.description || '-' }}</td>
                    </tr>
                    <tr v-if="!currentSecurityList.ingressRules?.length">
                      <td colspan="4" class="px-3 py-4 text-center text-slate-400">暂无入站规则</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <!-- 出站规则 -->
            <div>
              <div class="flex justify-between items-center mb-3">
                <h4 class="text-lg font-semibold text-blue-400">出站规则 (Egress)</h4>
                <button class="btn btn-primary btn-sm" @click="showAddRuleForm('egress')">
                  <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                  </svg>
                  添加规则
                </button>
              </div>
              <div class="overflow-x-auto">
                <table class="w-full text-sm">
                  <thead class="bg-slate-700/50">
                    <tr>
                      <th class="px-3 py-2 text-left text-slate-300">协议</th>
                      <th class="px-3 py-2 text-left text-slate-300">目标</th>
                      <th class="px-3 py-2 text-left text-slate-300">端口</th>
                      <th class="px-3 py-2 text-left text-slate-300">描述</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-700">
                    <tr v-for="(rule, index) in currentSecurityList.egressRules" :key="'out-' + index" class="hover:bg-slate-700/30">
                      <td class="px-3 py-2 text-white">{{ rule.protocolName }}</td>
                      <td class="px-3 py-2 text-white font-mono text-xs">{{ rule.destination }}</td>
                      <td class="px-3 py-2 text-white">
                        {{ formatPortRange(rule) }}
                      </td>
                      <td class="px-3 py-2 text-slate-400">{{ rule.description || '-' }}</td>
                    </tr>
                    <tr v-if="!currentSecurityList.egressRules?.length">
                      <td colspan="4" class="px-3 py-4 text-center text-slate-400">暂无出站规则</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>

        <div class="p-6 border-t border-slate-700 flex justify-end">
          <button class="btn btn-secondary" @click="securityListDialogVisible = false">关闭</button>
        </div>
      </div>
    </div>

    <!-- 添加安全规则弹窗 -->
    <div
      v-if="addRuleDialogVisible"
      class="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50 p-4"
    >
      <div class="card max-w-lg w-full">
        <div class="p-6 border-b border-slate-700 flex justify-between items-center">
          <h3 class="text-xl font-bold">添加{{ addRuleForm.isIngress ? '入站' : '出站' }}规则</h3>
          <button class="text-slate-400 hover:text-white" @click="addRuleDialogVisible = false">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="p-6 space-y-4">
          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">协议</label>
            <select v-model="addRuleForm.protocol" class="input">
              <option value="all">所有协议</option>
              <option value="6">TCP</option>
              <option value="17">UDP</option>
              <option value="1">ICMP</option>
            </select>
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">
              {{ addRuleForm.isIngress ? '来源 CIDR' : '目标 CIDR' }}
            </label>
            <input 
              v-model="addRuleForm.cidr" 
              type="text" 
              class="input" 
              placeholder="0.0.0.0/0 或 ::/0"
            />
          </div>

          <div v-if="['6', '17'].includes(addRuleForm.protocol)" class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">端口范围(最小)</label>
              <input 
                v-model.number="addRuleForm.portMin" 
                type="number" 
                class="input" 
                placeholder="1"
                min="1"
                max="65535"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">端口范围(最大)</label>
              <input 
                v-model.number="addRuleForm.portMax" 
                type="number" 
                class="input" 
                placeholder="65535"
                min="1"
                max="65535"
              />
            </div>
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">描述(可选)</label>
            <input 
              v-model="addRuleForm.description" 
              type="text" 
              class="input" 
              placeholder="规则描述"
            />
          </div>

          <div class="flex gap-3 pt-2">
            <button class="btn btn-secondary flex-1" @click="addRuleDialogVisible = false">取消</button>
            <button 
              class="btn btn-primary flex-1" 
              :disabled="addingRule"
              @click="submitAddRule"
            >
              {{ addingRule ? '添加中...' : '添加规则' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, watch, onMounted, computed } from 'vue'
import api from '../utils/api'
import { toast } from '../utils/toast'

// 配置列表状态
const configs = ref([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = 10
const totalPages = ref(0)
const searchText = ref('')

// 批量选择状态
const selectedConfigIds = ref([])
const showBatchCreateModal = ref(false)

// 计算属性：全选状态
const isAllSelected = computed(() => {
  return configs.value.length > 0 && selectedConfigIds.value.length === configs.value.length
})

const isIndeterminate = computed(() => {
  return selectedConfigIds.value.length > 0 && selectedConfigIds.value.length < configs.value.length
})

// 弹窗状态
const showAddModal = ref(false)
const showCreateInstanceModal = ref(false)
const showConfigDetailsSidebar = ref(false)
const showEditUserModal = ref(false)
const editConfigDialogVisible = ref(false)
const cloudShellDialogVisible = ref(false)
const volumeEditDialogVisible = ref(false)
const securityListDialogVisible = ref(false)
const addRuleDialogVisible = ref(false)

// 编辑状态
const editingConfig = ref(null)
const selectedConfigForInstance = ref(null)
const configDetails = ref(null)
const editingUser = ref(null)

// 加载状态
const submitting = ref(false)
const submittingInstance = ref(false)
const loadingDetails = ref(false)
const loadingTab = ref(false)
const loadingTraffic = ref(false)
const configUpdating = ref(false)
const cloudShellCreating = ref(false)
const volumeUpdating = ref(false)
const instanceActionLoading = reactive({})
const loadingSecurityList = ref(false)
const addingRule = ref(false)
const vcnActionLoading = reactive({})

// 标签页状态
const activeTab = ref('basic')
const tabInstances = ref([])
const tabVolumes = ref([])
const tabVCNs = ref([])
const tabTenant = ref(null)
const tabTraffic = ref({ time: [], inbound: [], outbound: [] })

const detailTabs = [
  { key: 'basic', label: '基本信息' },
  { key: 'instances', label: '实例列表' },
  { key: 'volumes', label: '引导卷' },
  { key: 'vcns', label: 'VCN网络' },
  { key: 'traffic', label: '流量统计' }
]

// 密码过期时间编辑
const editingPasswordExpiry = ref(false)
const passwordExpiryInput = ref(0)
const updatingPasswordExpiry = ref(false)

// 流量查询
const trafficCondition = ref({ regions: [], instances: [] })
const trafficVnics = ref([])
const trafficForm = ref({ instanceId: '', vnicId: '', startTime: '', endTime: '' })

// 表单数据
const form = ref({ username: '', configContent: '', ociKeyPath: '' })
const uploadedFile = ref(null)
const isDragging = ref(false)

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

const sshKeys = ref([])
const availableImages = ref([])
const filteredImages = ref([])
const loadingImages = ref(false)

const userForm = ref({ email: '', dbUserName: '', description: '' })

const editConfigForm = reactive({ 
  instanceId: '', 
  displayName: '', 
  ocpus: 2, 
  memoryInGBs: 12,
  bootVolumeSize: 50,
  vpusPerGB: 10,
  currentIpv6: '',
  keepBackup: false,
  sshPort: 22,
  retainNatGw: false,
  retainNlb: false
})

const cloudShellForm = reactive({ instanceId: '', publicKey: '' })
const cloudShellResult = reactive({ connectionId: '', connectionString: '' })

const volumeEditForm = reactive({
  volumeId: '',
  displayName: '',
  instanceId: '',
  instanceName: '',
  sizeInGBs: 50,
  vpusPerGB: 10,
  originalSize: 50,
  originalVpu: 10
})

// VCN 安全列表相关
const currentVcnForSecurity = ref(null)
const currentSecurityList = ref(null)
const addRuleForm = reactive({
  isIngress: true,
  protocol: '6',
  cidr: '0.0.0.0/0',
  portMin: 1,
  portMax: 65535,
  description: ''
})

// 工具函数
const parseConfigContent = (content) => {
  const lines = content.split('\n')
  const config = {}
  lines.forEach((line) => {
    const trimmed = line.trim()
    if (trimmed && trimmed.includes('=')) {
      const [key, value] = trimmed.split('=').map((s) => s.trim())
      config[key] = value
    }
  })
  return {
    ociUserId: config.user || '',
    ociFingerprint: config.fingerprint || '',
    ociTenantId: config.tenancy || '',
    ociRegion: config.region || ''
  }
}

const formatFileSize = (bytes) => {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(2) + ' MB'
}

const formatDateTime = (date) => {
  const pad = (n) => n.toString().padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

const copyToClipboard = (text) => {
  navigator.clipboard
    .writeText(text)
    .then(() => {
      toast.success('已复制到剪贴板')
    })
    .catch(() => {
      toast.error('复制失败')
    })
}

// 配置列表操作
const loadConfigs = async (page = 1) => {
  loading.value = true
  try {
    const response = await api.post('/oci/userPage', { page, pageSize, username: searchText.value })
    configs.value = response.data.list || []
    currentPage.value = response.data.page
    totalPages.value = Math.ceil(response.data.total / pageSize)
  } catch (error) {
    toast.error(error.message || '加载失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => loadConfigs(1)

// 选择操作
const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedConfigIds.value = []
  } else {
    selectedConfigIds.value = configs.value.map(c => c.id)
  }
}

const toggleSelectConfig = (id) => {
  const index = selectedConfigIds.value.indexOf(id)
  if (index > -1) {
    selectedConfigIds.value.splice(index, 1)
  } else {
    selectedConfigIds.value.push(id)
  }
}

// 批量删除
const batchDeleteConfigs = async () => {
  if (selectedConfigIds.value.length === 0) {
    toast.warning('请先选择要删除的配置')
    return
  }
  if (!confirm(`确定要删除选中的 ${selectedConfigIds.value.length} 个配置吗？`)) return
  try {
    await api.post('/oci/removeCfg', { ids: selectedConfigIds.value })
    toast.success('批量删除成功')
    selectedConfigIds.value = []
    await loadConfigs(currentPage.value)
  } catch (error) {
    toast.error(error.message || '批量删除失败')
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
    toast.success(`已为 ${selectedConfigIds.value.length} 个配置创建定时任务，可在任务列表查看`)
    closeBatchCreateModal()
    selectedConfigIds.value = []
  } catch (error) {
    toast.error(error.message || '批量创建失败')
  } finally {
    submittingInstance.value = false
  }
}

const closeModal = () => {
  showAddModal.value = false
  editingConfig.value = null
  uploadedFile.value = null
  isDragging.value = false
  form.value = { username: '', configContent: '', ociKeyPath: '' }
}

const handleFileSelect = (event) => {
  const file = event.target.files[0]
  if (file) uploadedFile.value = file
}

const handleFileDrop = (event) => {
  isDragging.value = false
  const file = event.dataTransfer.files[0]
  if (file && (file.name.endsWith('.pem') || file.name.endsWith('.key'))) {
    uploadedFile.value = file
  } else {
    toast.error('请上传 .pem 或 .key 格式的文件')
  }
}

const clearFile = () => {
  uploadedFile.value = null
}

const submitForm = async () => {
  if (!uploadedFile.value && !editingConfig.value) {
    toast.error('请选择密钥文件')
    return
  }
  submitting.value = true
  try {
    let keyPath = form.value.ociKeyPath
    if (uploadedFile.value) {
      const formData = new FormData()
      formData.append('file', uploadedFile.value)
      const uploadResponse = await api.post('/oci/uploadKey', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
      keyPath = uploadResponse.data
    }
    const parsedConfig = parseConfigContent(form.value.configContent)
    const payload = {
      username: form.value.username,
      tenantName: form.value.username,
      ociTenantId: parsedConfig.ociTenantId,
      ociUserId: parsedConfig.ociUserId,
      ociFingerprint: parsedConfig.ociFingerprint,
      ociRegion: parsedConfig.ociRegion,
      ociKeyPath: keyPath
    }
    if (editingConfig.value) {
      await api.post('/oci/updateCfgName', {
        id: editingConfig.value.id,
        username: payload.username,
        ociKeyPath: uploadedFile.value ? keyPath : undefined
      })
      toast.success('更新成功')
    } else {
      await api.post('/oci/addCfg', payload)
      toast.success('添加成功')
    }
    closeModal()
    await loadConfigs(currentPage.value)
  } catch (error) {
    toast.error(error.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

const editConfig = (config) => {
  editingConfig.value = config
  uploadedFile.value = null
  form.value = {
    username: config.username,
    configContent: `user=${config.ociUserId}\nfingerprint=${config.ociFingerprint}\ntenancy=${config.ociTenantId}\nregion=${config.ociRegion}`,
    ociKeyPath: config.ociKeyPath
  }
  showAddModal.value = true
}

const deleteConfig = async (id) => {
  if (!confirm('确定要删除此配置吗？')) return
  try {
    await api.post('/oci/removeCfg', { ids: [id] })
    toast.success('删除成功')
    await loadConfigs(currentPage.value)
  } catch (error) {
    toast.error(error.message || '删除失败')
  }
}

// 加载SSH密钥列表
const loadSSHKeys = async () => {
  try {
    const response = await api.get('/key/standalone')
    sshKeys.value = response.data || []
  } catch (error) {
    console.error('加载SSH密钥失败:', error)
    sshKeys.value = []
  }
}

// 加载可用镜像列表
const loadImages = async (configId, region, architecture) => {
  if (!configId || !region || !architecture) return
  loadingImages.value = true
  try {
    const response = await api.post('/oci/images', { configId, region, architecture })
    availableImages.value = response.data || []
    filterImagesByOS()
  } catch (error) {
    console.error('加载镜像列表失败:', error)
    availableImages.value = []
    filteredImages.value = []
  } finally {
    loadingImages.value = false
  }
}

// 根据操作系统类型过滤镜像
const filterImagesByOS = () => {
  const os = instanceForm.value.operationSystem
  let osKeyword = ''
  if (os === 'Ubuntu') {
    osKeyword = 'Ubuntu'
  } else if (os === 'CentOS') {
    osKeyword = 'CentOS'
  } else if (os === 'Oracle Linux') {
    osKeyword = 'Oracle'
  }
  
  if (osKeyword) {
    filteredImages.value = availableImages.value.filter(img => 
      img.operatingSystem.toLowerCase().includes(osKeyword.toLowerCase())
    )
  } else {
    filteredImages.value = availableImages.value
  }
  
  // 默认选择第一个镜像
  if (filteredImages.value.length > 0) {
    instanceForm.value.imageId = filteredImages.value[0].id
  } else {
    instanceForm.value.imageId = ''
  }
}

// 架构或区域变化时重新加载镜像
const onArchitectureChange = async () => {
  if (selectedConfigForInstance.value) {
    await loadImages(
      selectedConfigForInstance.value.id,
      instanceForm.value.ociRegion,
      instanceForm.value.architecture
    )
  }
}

// 操作系统变化时过滤镜像
const onOperationSystemChange = () => {
  filterImagesByOS()
}

// 创建实例
const createInstance = async (config) => {
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
  instanceForm.value = {
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
  }
}

const submitInstanceTask = async () => {
  submittingInstance.value = true
  try {
    const payload = {
      userId: selectedConfigForInstance.value.id,
      ociRegion: instanceForm.value.ociRegion,
      ocpus: instanceForm.value.ocpus,
      memory: instanceForm.value.memory,
      disk: instanceForm.value.disk,
      bootVolumeVpu: instanceForm.value.bootVolumeVpu || 10,
      architecture: instanceForm.value.architecture,
      operationSystem: instanceForm.value.operationSystem,
      imageId: instanceForm.value.imageId,
      sshKeyId: instanceForm.value.sshKeyId,
      interval: instanceForm.value.interval || 60,
      executeOnce: !instanceForm.value.isTaskMode
    }
    const response = await api.post('/task/create', payload)
    if (instanceForm.value.isTaskMode) {
      toast.success('定时任务创建成功，可在任务列表查看执行状态')
    } else {
      toast.success(response.message || '实例创建请求已提交')
    }
    closeInstanceModal()
  } catch (error) {
    toast.error(error.message || '创建失败')
  } finally {
    submittingInstance.value = false
  }
}

// 配置详情
const viewConfigDetails = async (config) => {
  showConfigDetailsSidebar.value = true
  loadingDetails.value = true
  activeTab.value = 'basic'
  tabInstances.value = []
  tabVolumes.value = []
  tabVCNs.value = []
  tabTenant.value = null
  try {
    const response = await api.post('/oci/details', { configId: config.id })
    configDetails.value = response.data
    await loadTenant()
  } catch (error) {
    toast.error(error.message || '加载配置详情失败')
    closeConfigDetailsSidebar()
  } finally {
    loadingDetails.value = false
  }
}

const closeConfigDetailsSidebar = () => {
  showConfigDetailsSidebar.value = false
  configDetails.value = null
  loadingDetails.value = false
  activeTab.value = 'basic'
  tabInstances.value = []
  tabVolumes.value = []
  tabVCNs.value = []
  tabTenant.value = null
  tabTraffic.value = { time: [], inbound: [], outbound: [] }
  trafficCondition.value = { regions: [], instances: [] }
  trafficVnics.value = []
  trafficForm.value = { instanceId: '', vnicId: '', startTime: '', endTime: '' }
}

// 标签页数据加载
const loadInstances = async (clearCache = false) => {
  if (!configDetails.value) return
  loadingTab.value = true
  try {
    const response = await api.post('/oci/details/instances', { configId: configDetails.value.userId, clearCache })
    tabInstances.value = response.data || []
  } catch (error) {
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
  } catch (error) {
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
  } catch (error) {
    toast.error(error.message || '加载VCN列表失败')
    tabVCNs.value = []
  } finally {
    loadingTab.value = false
  }
}

const loadTenant = async (clearCache = false) => {
  if (!configDetails.value) return
  loadingTab.value = true
  try {
    const response = await api.post('/oci/tenant/info', { configId: configDetails.value.userId, clearCache })
    tabTenant.value = response.data
  } catch (error) {
    toast.error(error.message || '加载租户详情失败')
    tabTenant.value = null
  } finally {
    loadingTab.value = false
  }
}

const loadTrafficCondition = async () => {
  if (!configDetails.value) return
  try {
    const response = await api.get('/oci/traffic/condition', { params: { configId: configDetails.value.userId } })
    trafficCondition.value = response.data || { regions: [], instances: [] }
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
  } catch (error) {
    console.error('加载VNIC列表失败:', error)
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
  } catch (error) {
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

// 监听标签切换
watch(activeTab, () => {
  if (activeTab.value === 'basic' && !tabTenant.value) loadTenant()
  else if (activeTab.value === 'instances' && tabInstances.value.length === 0) loadInstances()
  else if (activeTab.value === 'volumes' && tabVolumes.value.length === 0) loadVolumes()
  else if (activeTab.value === 'vcns' && tabVCNs.value.length === 0) loadVCNs()
  else if (activeTab.value === 'traffic' && trafficCondition.value.instances.length === 0) loadTrafficCondition()
})

watch(
  () => trafficForm.value.instanceId,
  (newVal) => {
    if (newVal) loadInstanceVnics()
    else {
      trafficVnics.value = []
      trafficForm.value.vnicId = ''
    }
  }
)

// 密码过期时间编辑
const startEditPasswordExpiry = () => {
  passwordExpiryInput.value = tabTenant.value.passwordExpiresAfter || 0
  editingPasswordExpiry.value = true
}

const cancelEditPasswordExpiry = () => {
  editingPasswordExpiry.value = false
  passwordExpiryInput.value = 0
}

const savePasswordExpiry = async () => {
  updatingPasswordExpiry.value = true
  try {
    await api.post('/oci/tenant/updatePwdEx', {
      cfgId: configDetails.value.userId,
      passwordExpiresAfter: passwordExpiryInput.value
    })
    toast.success('密码过期时间更新成功')
    tabTenant.value.passwordExpiresAfter = passwordExpiryInput.value
    editingPasswordExpiry.value = false
  } catch (error) {
    toast.error(error.message || '更新失败')
  } finally {
    updatingPasswordExpiry.value = false
  }
}

// 用户管理
const editUser = (user) => {
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
      ociCfgId: configDetails.value.userId,
      userId: editingUser.value.id,
      email: userForm.value.email,
      dbUserName: userForm.value.dbUserName,
      description: userForm.value.description
    })
    toast.success('用户信息更新成功')
    closeEditUserModal()
    await loadTenant(true)
  } catch (error) {
    toast.error(error.message || '更新失败')
  }
}

const resetUserPassword = async (user) => {
  if (!confirm(`确定要重置用户 ${user.name} 的密码吗？`)) return
  try {
    await api.post('/oci/tenant/resetPassword', { ociCfgId: configDetails.value.userId, userId: user.id })
    toast.success('密码重置成功')
  } catch (error) {
    toast.error(error.message || '重置密码失败')
  }
}

const clearUserMfa = async (user) => {
  if (!confirm(`确定要清除用户 ${user.name} 的 MFA 设备吗？`)) return
  try {
    await api.post('/oci/tenant/deleteMfaDevice', { ociCfgId: configDetails.value.userId, userId: user.id })
    toast.success('MFA 设备清除成功')
    await loadTenant(true)
  } catch (error) {
    toast.error(error.message || 'MFA 清除失败')
  }
}

const clearUserApiKeys = async (user) => {
  if (!confirm(`确定要清除用户 ${user.name} 的所有 API 密钥吗？`)) return
  try {
    await api.post('/oci/tenant/deleteApiKey', { ociCfgId: configDetails.value.userId, userId: user.id })
    toast.success('API 密钥清除成功')
  } catch (error) {
    toast.error(error.message || 'API 密钥清除失败')
  }
}

const deleteUser = async (user) => {
  if (!confirm(`确定要删除用户 ${user.name} 吗？此操作不可恢复！`)) return
  try {
    await api.post('/oci/tenant/deleteUser', { ociCfgId: configDetails.value.userId, userId: user.id })
    toast.success('用户删除成功')
    await loadTenant(true)
  } catch (error) {
    toast.error(error.message || '删除用户失败')
  }
}

// 实例操作
const controlInstanceInDetails = async (instanceId, action) => {
  const actionMap = {
    START: { endpoint: '/instance/start', message: '启动' },
    STOP: { endpoint: '/instance/stop', message: '停止' },
    SOFTRESET: { endpoint: '/instance/reboot', message: '重启' }
  }
  instanceActionLoading[instanceId] = true
  try {
    await api.post(actionMap[action].endpoint, { userId: configDetails.value.userId, instanceId })
    toast.success(`${actionMap[action].message}操作已提交`)
    setTimeout(() => loadInstances(true), 3000)
  } catch (error) {
    toast.error(error.message || '操作失败')
  } finally {
    delete instanceActionLoading[instanceId]
  }
}

const terminateInstanceInDetails = async (instanceId) => {
  if (!confirm('确定要删除此实例吗？此操作不可恢复！')) return
  instanceActionLoading[instanceId] = true
  try {
    await api.post('/instance/terminate', { userId: configDetails.value.userId, instanceId })
    toast.success('删除操作已提交')
    setTimeout(() => loadInstances(true), 3000)
  } catch (error) {
    toast.error(error.message || '删除失败')
  } finally {
    delete instanceActionLoading[instanceId]
  }
}

const changeIPInDetails = async (instanceId) => {
  if (!confirm('确定要更改此实例的公网IP吗？')) return
  instanceActionLoading[instanceId] = true
  try {
    const response = await api.post('/instance/changeIP', { userId: configDetails.value.userId, instanceId })
    if (response.data && response.data.code === 200) {
      toast.success(`IP更改成功，新IP: ${response.data.data.newIP}`)
      setTimeout(() => loadInstances(true), 2000)
    } else {
      toast.error(response.data?.msg || 'IP更改失败')
    }
  } catch (error) {
    toast.error(error.message || 'IP更改失败')
  } finally {
    delete instanceActionLoading[instanceId]
  }
}

const showEditConfigDialog = (instance) => {
  editConfigForm.instanceId = instance.id
  editConfigForm.displayName = instance.displayName
  editConfigForm.ocpus = instance.ocpus || 2
  editConfigForm.memoryInGBs = instance.memory || 12
  editConfigForm.bootVolumeSize = instance.bootVolumeSize || 50
  editConfigForm.vpusPerGB = instance.bootVolumeVpu || 10
  editConfigForm.currentIpv6 = instance.ipv6 || ''
  editConfigDialogVisible.value = true
}

const updateInstanceConfig = async () => {
  if (!editConfigForm.instanceId) {
    toast.warning('缺少实例ID')
    return
  }
  configUpdating.value = true
  try {
    const response = await api.post('/instance/updateConfig', {
      userId: configDetails.value.userId,
      instanceId: editConfigForm.instanceId,
      ocpus: editConfigForm.ocpus,
      memoryInGBs: editConfigForm.memoryInGBs
    })
    if (response.data && response.data.code === 200) {
      toast.success('实例配置更新成功')
      setTimeout(() => loadInstances(true), 2000)
    } else {
      toast.error(response.data?.msg || '实例配置更新失败')
    }
  } catch (error) {
    toast.error(error.message || '实例配置更新失败')
  } finally {
    configUpdating.value = false
  }
}

const updateInstanceName = async () => {
  if (!editConfigForm.instanceId || !editConfigForm.displayName) {
    toast.warning('请输入实例名称')
    return
  }
  configUpdating.value = true
  try {
    const response = await api.post('/instance/updateName', {
      userId: configDetails.value.userId,
      instanceId: editConfigForm.instanceId,
      displayName: editConfigForm.displayName
    })
    if (response.data && response.data.code === 200) {
      toast.success('实例名称更新成功')
      setTimeout(() => loadInstances(true), 2000)
    } else {
      toast.error(response.data?.msg || '实例名称更新失败')
    }
  } catch (error) {
    toast.error(error.message || '实例名称更新失败')
  } finally {
    configUpdating.value = false
  }
}

const updateBootVolume = async () => {
  if (!editConfigForm.instanceId) {
    toast.warning('缺少实例ID')
    return
  }
  configUpdating.value = true
  try {
    const response = await api.post('/instance/updateBootVolume', {
      userId: configDetails.value.userId,
      instanceId: editConfigForm.instanceId,
      sizeInGBs: editConfigForm.bootVolumeSize,
      vpusPerGB: editConfigForm.vpusPerGB
    })
    if (response.data && response.data.code === 200) {
      toast.success('引导卷配置更新成功')
      setTimeout(() => loadInstances(true), 2000)
    } else {
      toast.error(response.data?.msg || '引导卷配置更新失败')
    }
  } catch (error) {
    toast.error(error.message || '引导卷配置更新失败')
  } finally {
    configUpdating.value = false
  }
}

const attachIPv6 = async () => {
  if (!editConfigForm.instanceId) {
    toast.warning('缺少实例ID')
    return
  }
  if (editConfigForm.currentIpv6) {
    toast.warning('该实例已有IPv6地址')
    return
  }
  if (!confirm('确定要为该实例附加IPv6地址吗？')) return
  configUpdating.value = true
  try {
    const response = await api.post('/instance/attachIPv6', {
      userId: configDetails.value.userId,
      instanceId: editConfigForm.instanceId
    })
    if (response.data && response.data.code === 200) {
      toast.success(`IPv6附加成功: ${response.data.data.ipv6}`)
      editConfigForm.currentIpv6 = response.data.data.ipv6
      setTimeout(() => loadInstances(true), 2000)
    } else {
      toast.error(response.data?.msg || 'IPv6附加失败')
    }
  } catch (error) {
    toast.error(error.message || 'IPv6附加失败')
  } finally {
    configUpdating.value = false
  }
}

// 自动救援
const startAutoRescue = async () => {
  if (!editConfigForm.instanceId) {
    toast.warning('缺少实例ID')
    return
  }
  if (!confirm('自动救援将会关闭实例，备份引导卷，创建47GB新引导卷并重新启动。此操作需要5-10分钟，确定继续吗？')) return
  configUpdating.value = true
  try {
    const response = await api.post('/instance/autoRescue', {
      userId: configDetails.value.userId,
      instanceId: editConfigForm.instanceId,
      instanceName: editConfigForm.displayName,
      keepBackup: editConfigForm.keepBackup
    })
    if (response.data && response.data.code === 200) {
      toast.success('自动救援任务已启动，请等待5-10分钟完成')
      editConfigDialogVisible.value = false
    } else {
      toast.error(response.data?.msg || '自动救援启动失败')
    }
  } catch (error) {
    toast.error(error.message || '自动救援启动失败')
  } finally {
    configUpdating.value = false
  }
}

// 开启500Mbps
const enable500Mbps = async () => {
  if (!editConfigForm.instanceId) {
    toast.warning('缺少实例ID')
    return
  }
  if (!confirm('此操作将创建NAT网关和网络负载均衡器来实现500Mbps下行带宽。仅AMD E2.1.Micro实例支持此功能，确定继续吗？')) return
  configUpdating.value = true
  try {
    const response = await api.post('/instance/enable500Mbps', {
      userId: configDetails.value.userId,
      instanceId: editConfigForm.instanceId,
      sshPort: editConfigForm.sshPort || 22
    })
    if (response.data && response.data.code === 200) {
      toast.success('500Mbps开启任务已启动，请稍后检查实例状态')
      editConfigDialogVisible.value = false
    } else {
      toast.error(response.data?.msg || '500Mbps开启失败')
    }
  } catch (error) {
    toast.error(error.message || '500Mbps开启失败')
  } finally {
    configUpdating.value = false
  }
}

// 关闭500Mbps
const disable500Mbps = async () => {
  if (!editConfigForm.instanceId) {
    toast.warning('缺少实例ID')
    return
  }
  if (!confirm('确定要关闭500Mbps下行带宽吗？这将删除相关的网络资源。')) return
  configUpdating.value = true
  try {
    const response = await api.post('/instance/disable500Mbps', {
      userId: configDetails.value.userId,
      instanceId: editConfigForm.instanceId,
      retainNatGw: editConfigForm.retainNatGw,
      retainNlb: editConfigForm.retainNlb
    })
    if (response.data && response.data.code === 200) {
      toast.success('500Mbps关闭任务已启动，请稍后检查实例状态')
      editConfigDialogVisible.value = false
    } else {
      toast.error(response.data?.msg || '500Mbps关闭失败')
    }
  } catch (error) {
    toast.error(error.message || '500Mbps关闭失败')
  } finally {
    configUpdating.value = false
  }
}

// 引导卷编辑
const openVolumeEditDialog = (volume) => {
  volumeEditForm.volumeId = volume.id
  volumeEditForm.displayName = volume.displayName
  volumeEditForm.instanceId = volume.instanceId || ''
  volumeEditForm.instanceName = volume.instanceName || ''
  volumeEditForm.sizeInGBs = volume.sizeInGBs || 50
  volumeEditForm.vpusPerGB = volume.vpusPerGB || 10
  volumeEditForm.originalSize = volume.sizeInGBs || 50
  volumeEditForm.originalVpu = volume.vpusPerGB || 10
  volumeEditDialogVisible.value = true
}

const updateVolumeConfig = async () => {
  if (!volumeEditForm.volumeId) {
    toast.warning('缺少引导卷ID')
    return
  }
  if (volumeEditForm.sizeInGBs < volumeEditForm.originalSize) {
    toast.warning('磁盘大小只能增大，不能缩小')
    return
  }
  if (!confirm('确定要修改引导卷配置吗？修改可能需要重启实例才能生效。')) return
  volumeUpdating.value = true
  try {
    const response = await api.post('/bootVolume/update', {
      userId: configDetails.value.userId,
      bootVolumeId: volumeEditForm.volumeId,
      sizeInGBs: volumeEditForm.sizeInGBs,
      vpusPerGB: volumeEditForm.vpusPerGB
    })
    if (response.data && response.data.code === 200) {
      toast.success('引导卷配置更新成功')
      volumeEditDialogVisible.value = false
      // 刷新引导卷列表
      loadVolumes()
    } else {
      toast.error(response.data?.msg || '引导卷配置更新失败')
    }
  } catch (error) {
    toast.error(error.message || '引导卷配置更新失败')
  } finally {
    volumeUpdating.value = false
  }
}

// Cloud Shell
const showCloudShellDialog = (instanceId) => {
  cloudShellForm.instanceId = instanceId
  cloudShellForm.publicKey = ''
  cloudShellResult.connectionId = ''
  cloudShellResult.connectionString = ''
  cloudShellDialogVisible.value = true
}

const createCloudShell = async () => {
  if (!cloudShellForm.publicKey.trim()) {
    toast.warning('请输入SSH公钥')
    return
  }
  cloudShellCreating.value = true
  try {
    const response = await api.post('/instance/createCloudShell', {
      userId: configDetails.value.userId,
      instanceId: cloudShellForm.instanceId,
      publicKey: cloudShellForm.publicKey
    })
    if (response.data && response.data.code === 200) {
      cloudShellResult.connectionId = response.data.data.connectionId
      cloudShellResult.connectionString = response.data.data.connectionString
      toast.success('Cloud Shell连接创建成功')
    } else {
      toast.error(response.data?.msg || 'Cloud Shell连接创建失败')
    }
  } catch (error) {
    toast.error(error.message || 'Cloud Shell连接创建失败')
  } finally {
    cloudShellCreating.value = false
  }
}

// VCN 安全列表相关方法
const formatPortRange = (rule) => {
  if (rule.protocolName === '所有协议' || rule.protocolName === 'all') {
    return '所有'
  }
  if (rule.protocolName === 'ICMP' || rule.protocolName === 'ICMPv6') {
    if (rule.icmpType !== undefined) {
      return `Type: ${rule.icmpType}${rule.icmpCode !== undefined ? ', Code: ' + rule.icmpCode : ''}`
    }
    return '所有'
  }
  if (rule.portMin && rule.portMax) {
    if (rule.portMin === rule.portMax) {
      return rule.portMin
    }
    return `${rule.portMin}-${rule.portMax}`
  }
  return '所有'
}

const viewSecurityList = async (vcn) => {
  currentVcnForSecurity.value = vcn
  currentSecurityList.value = null
  securityListDialogVisible.value = true
  loadingSecurityList.value = true
  
  try {
    const response = await api.post('/oci/vcn/securityList', {
      configId: configDetails.value.userId,
      vcnId: vcn.id
    })
    currentSecurityList.value = response.data
  } catch (error) {
    toast.error(error.message || '获取安全列表失败')
  } finally {
    loadingSecurityList.value = false
  }
}

const showAddRuleForm = (type) => {
  addRuleForm.isIngress = type === 'ingress'
  addRuleForm.protocol = '6'
  addRuleForm.cidr = '0.0.0.0/0'
  addRuleForm.portMin = 1
  addRuleForm.portMax = 65535
  addRuleForm.description = ''
  addRuleDialogVisible.value = true
}

const submitAddRule = async () => {
  if (!addRuleForm.cidr) {
    toast.warning('请输入CIDR地址')
    return
  }
  addingRule.value = true
  try {
    const params = {
      configId: configDetails.value.userId,
      vcnId: currentVcnForSecurity.value.id,
      isIngress: addRuleForm.isIngress,
      protocol: addRuleForm.protocol,
      portMin: addRuleForm.protocol === '6' || addRuleForm.protocol === '17' ? addRuleForm.portMin : undefined,
      portMax: addRuleForm.protocol === '6' || addRuleForm.protocol === '17' ? addRuleForm.portMax : undefined,
      description: addRuleForm.description || undefined
    }
    if (addRuleForm.isIngress) {
      params.source = addRuleForm.cidr
    } else {
      params.destination = addRuleForm.cidr
    }
    await api.post('/oci/vcn/addSecurityRule', params)
    toast.success('安全规则添加成功')
    addRuleDialogVisible.value = false
    viewSecurityList(currentVcnForSecurity.value)
  } catch (error) {
    toast.error(error.message || '添加安全规则失败')
  } finally {
    addingRule.value = false
  }
}

const releaseVcnRules = async (vcn) => {
  if (!confirm('确定要放行所有规则吗？这将允许所有入站和出站流量。')) return
  
  vcnActionLoading[vcn.id] = true
  try {
    await api.post('/oci/vcn/releaseSecurityRules', {
      configId: configDetails.value.userId,
      vcnId: vcn.id
    })
    toast.success('安全规则已放行')
  } catch (error) {
    toast.error(error.message || '放行安全规则失败')
  } finally {
    vcnActionLoading[vcn.id] = false
  }
}

const deleteVcn = async (vcn) => {
  if (!confirm(`确定要删除VCN "${vcn.displayName}" 吗？此操作将删除VCN及其所有子网、网关等资源，且不可恢复！`)) return
  
  vcnActionLoading[vcn.id] = true
  try {
    await api.post('/oci/vcn/delete', {
      configId: configDetails.value.userId,
      vcnId: vcn.id
    })
    toast.success('VCN删除成功')
    loadVCNs()
  } catch (error) {
    toast.error(error.message || '删除VCN失败')
  } finally {
    vcnActionLoading[vcn.id] = false
  }
}

onMounted(() => {
  loadConfigs()
})
</script>
