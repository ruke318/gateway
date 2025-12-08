<template>
  <div>
    <el-form :inline="true" :model="query" style="margin-bottom: 20px">
      <el-form-item label="接口标识">
        <el-input v-model="query.service_id" placeholder="模糊搜索" clearable @keyup.enter="fetchList" />
      </el-form-item>
      <el-form-item label="名称">
        <el-input v-model="query.name" placeholder="模糊搜索" clearable @keyup.enter="fetchList" />
      </el-form-item>
      <el-form-item label="厂商">
        <el-select v-model="query.vendor_id" placeholder="全部" clearable style="width: 150px">
          <el-option v-for="v in vendors" :key="v.id" :label="v.name" :value="v.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="机构">
        <el-select v-model="query.org_id" placeholder="全部" clearable style="width: 150px">
          <el-option v-for="o in organizations" :key="o.id" :label="o.name" :value="o.id" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="fetchList">搜索</el-button>
        <el-button @click="resetQuery">重置</el-button>
        <el-button type="success" @click="showDialog()">新增</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="list" border stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="service_id" label="接口标识" width="120" />
      <el-table-column prop="name" label="名称" width="120" />
      <el-table-column label="机构" width="100">
        <template #default="{ row }">
          {{ row.organization?.name || row.org_id }}
        </template>
      </el-table-column>
      <el-table-column label="厂商" width="100">
        <template #default="{ row }">
          {{ row.vendor?.name || row.vendor_id }}
        </template>
      </el-table-column>
      <el-table-column prop="backend_method" label="方法" width="80" />
      <el-table-column prop="backend_url" label="后端URL" />
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="showDialog(row)">编辑</el-button>
          <el-button size="small" type="warning" @click="showHookDialog(row)">Hook</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 接口编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑接口' : '新增接口'" width="700px">
      <el-form :model="form" label-width="120px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="接口标识" required>
              <el-input v-model="form.service_id" placeholder="唯一标识" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="名称" required>
              <el-input v-model="form.name" placeholder="接口名称" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="机构" required>
              <el-select v-model="form.org_id" placeholder="选择机构" style="width: 100%">
                <el-option v-for="o in organizations" :key="o.id" :label="o.name" :value="o.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="厂商" required>
              <el-select v-model="form.vendor_id" placeholder="选择厂商" style="width: 100%">
                <el-option v-for="v in vendors" :key="v.id" :label="v.name" :value="v.id" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="后端URL">
          <el-input v-model="form.backend_url" placeholder="留空则使用厂商BaseURL" />
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="方法">
              <el-select v-model="form.backend_method" style="width: 100%">
                <el-option label="GET" value="GET" />
                <el-option label="POST" value="POST" />
                <el-option label="PUT" value="PUT" />
                <el-option label="DELETE" value="DELETE" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Body类型">
              <el-select v-model="form.body_type" style="width: 100%">
                <el-option label="JSON" value="json" />
                <el-option label="Form表单" value="form" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="后端路径">
          <el-input v-model="form.backend_path" placeholder="如 /api/v1/xxx?id={id}，{key}会从DSL结果取值" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="请求转换">
          <el-input
            v-model="requestTransformStr"
            type="textarea"
            :rows="3"
            placeholder="双击打开JSON编辑器"
            readonly
            class="json-input"
            :class="{ 'json-error': requestJsonError }"
            @dblclick="openJsonEditor('request')"
          />
          <span v-if="requestJsonError" class="json-error-tip">JSON格式错误</span>
        </el-form-item>
        <el-form-item label="响应转换">
          <el-input
            v-model="responseTransformStr"
            type="textarea"
            :rows="3"
            placeholder="双击打开JSON编辑器"
            readonly
            class="json-input"
            :class="{ 'json-error': responseJsonError }"
            @dblclick="openJsonEditor('response')"
          />
          <span v-if="responseJsonError" class="json-error-tip">JSON格式错误</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- JSON编辑器全屏弹窗 -->
    <el-dialog
      v-model="jsonEditorVisible"
      :title="jsonEditorTitle"
      fullscreen
      :close-on-click-modal="false"
    >
      <div class="json-editor-container">
        <JsonEditorVue
          v-model="jsonEditorContent"
          mode="text"
          :mainMenuBar="true"
          :statusBar="true"
          class="json-editor"
        />
      </div>
      <template #footer>
        <el-button @click="jsonEditorVisible = false">取消</el-button>
        <el-button type="primary" @click="saveJsonEditor">确定</el-button>
      </template>
    </el-dialog>

    <!-- Hook管理弹窗 -->
    <el-dialog v-model="hookDialogVisible" title="Hook管理" width="800px">
      <div style="margin-bottom: 15px">
        <span style="font-weight: bold">接口: {{ currentService?.name }}</span>
      </div>
      <el-table :data="serviceHooks" border stripe>
        <el-table-column prop="hook_point" label="Hook节点" width="180" />
        <el-table-column label="脚本" width="150">
          <template #default="{ row }">
            {{ row.script?.name || row.script_id }}
          </template>
        </el-table-column>
        <el-table-column prop="priority" label="优先级" width="80" />
        <el-table-column label="启用" width="100">
          <template #default="{ row }">
            <el-switch :model-value="row.status === 1" @change="handleToggleStatus(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150">
          <template #default="{ row }">
            <el-button size="small" @click="showHookEditDialog(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDeleteHook(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div style="margin-top: 15px">
        <el-button type="primary" @click="showHookEditDialog()">添加Hook</el-button>
      </div>
    </el-dialog>

    <!-- Hook编辑弹窗 -->
    <el-dialog v-model="hookEditVisible" :title="hookForm.id ? '编辑Hook' : '添加Hook'" width="500px">
      <el-form :model="hookForm" label-width="100px">
        <el-form-item label="Hook节点" required>
          <el-select v-model="hookForm.hook_point" style="width: 100%">
            <el-option v-for="h in hookPoints" :key="h" :label="h" :value="h" />
          </el-select>
        </el-form-item>
        <el-form-item label="Hook脚本" required>
          <el-select v-model="hookForm.script_id" style="width: 100%">
            <el-option v-for="s in hookScripts" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="优先级">
          <el-input-number v-model="hookForm.priority" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="hookEditVisible = false">取消</el-button>
        <el-button type="primary" @click="handleHookSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { serviceApi, vendorApi, organizationApi, hookScriptApi, serviceHookApi } from '../api'
import JsonEditorVue from 'json-editor-vue'

const list = ref([])
const vendors = ref([])
const organizations = ref([])
const hookScripts = ref([])
const dialogVisible = ref(false)
const query = ref({ service_id: '', name: '', vendor_id: null, org_id: null })
const form = ref({
  service_id: '', name: '', org_id: null, vendor_id: null,
  backend_url: '', backend_path: '', backend_method: 'POST', body_type: 'json',
  description: '', request_transform: null, response_transform: null
})

const hookPoints = [
  'BeforeAuth', 'AfterAuth',
  'BeforeRequestTransform', 'AfterRequestTransform',
  'BeforeForward', 'AfterForward',
  'BeforeResponseTransform', 'AfterResponseTransform',
  'OnError'
]

// JSON字符串和校验
const requestTransformStr = ref('')
const responseTransformStr = ref('')
const requestJsonError = ref(false)
const responseJsonError = ref(false)

// 监听字符串变化，校验并同步到 form
watch(requestTransformStr, (val) => {
  if (!val) {
    form.value.request_transform = null
    requestJsonError.value = false
    return
  }
  try {
    form.value.request_transform = JSON.parse(val)
    requestJsonError.value = false
  } catch (e) {
    requestJsonError.value = true
  }
})

watch(responseTransformStr, (val) => {
  if (!val) {
    form.value.response_transform = null
    responseJsonError.value = false
    return
  }
  try {
    form.value.response_transform = JSON.parse(val)
    responseJsonError.value = false
  } catch (e) {
    responseJsonError.value = true
  }
})

// JSON编辑器
const jsonEditorVisible = ref(false)
const jsonEditorTitle = ref('')
const jsonEditorContent = ref(null)
const jsonEditorType = ref('') // 'request' or 'response'

const openJsonEditor = (type) => {
  jsonEditorType.value = type
  if (type === 'request') {
    jsonEditorTitle.value = '编辑请求转换DSL'
    jsonEditorContent.value = form.value.request_transform || {}
  } else {
    jsonEditorTitle.value = '编辑响应转换DSL'
    jsonEditorContent.value = form.value.response_transform || {}
  }
  jsonEditorVisible.value = true
}

const saveJsonEditor = () => {
  // jsonEditorContent 可能是对象或字符串，统一处理
  let jsonObj = jsonEditorContent.value
  if (typeof jsonObj === 'string') {
    try {
      jsonObj = JSON.parse(jsonObj)
    } catch (e) {
      ElMessage.error('JSON格式错误')
      return
    }
  }

  if (jsonEditorType.value === 'request') {
    form.value.request_transform = jsonObj
    requestTransformStr.value = jsonObj ? JSON.stringify(jsonObj) : ''
    requestJsonError.value = false
  } else {
    form.value.response_transform = jsonObj
    responseTransformStr.value = jsonObj ? JSON.stringify(jsonObj) : ''
    responseJsonError.value = false
  }
  jsonEditorVisible.value = false
}

// Hook管理
const hookDialogVisible = ref(false)
const hookEditVisible = ref(false)
const currentService = ref(null)
const serviceHooks = ref([])
const hookForm = ref({ hook_point: 'BeforeForward', script_id: null, priority: 0 })

const fetchList = async () => {
  const res = await serviceApi.list(query.value)
  list.value = res.data || []
}

const resetQuery = () => {
  query.value = { service_id: '', name: '', vendor_id: null, org_id: null }
  fetchList()
}

const fetchOptions = async () => {
  const [v, o, h] = await Promise.all([vendorApi.list(), organizationApi.list(), hookScriptApi.list()])
  vendors.value = v.data || []
  organizations.value = o.data || []
  hookScripts.value = h.data || []
}

const showDialog = (row = null) => {
  if (row) {
    form.value = { ...row }
    // 处理 JSON 显示，确保正确解析
    form.value.request_transform = parseJsonField(row.request_transform)
    form.value.response_transform = parseJsonField(row.response_transform)
    requestTransformStr.value = form.value.request_transform ? JSON.stringify(form.value.request_transform) : ''
    responseTransformStr.value = form.value.response_transform ? JSON.stringify(form.value.response_transform) : ''
  } else {
    form.value = {
      service_id: '', name: '', org_id: null, vendor_id: null,
      backend_url: '', backend_path: '', backend_method: 'POST', body_type: 'json',
      description: '', request_transform: null, response_transform: null
    }
    requestTransformStr.value = ''
    responseTransformStr.value = ''
  }
  requestJsonError.value = false
  responseJsonError.value = false
  dialogVisible.value = true
}

// 解析 JSON 字段，处理字符串形式的 JSON（支持双重转义）
const parseJsonField = (val) => {
  if (!val) return null
  if (typeof val === 'object') return val
  if (typeof val === 'string') {
    try {
      let parsed = JSON.parse(val)
      // 如果解析后还是字符串，再解析一次（处理双重转义）
      if (typeof parsed === 'string') {
        parsed = JSON.parse(parsed)
      }
      return parsed
    } catch (e) {
      return null
    }
  }
  return null
}

const handleSubmit = async () => {
  if (requestJsonError.value || responseJsonError.value) {
    ElMessage.error('请修正JSON格式错误')
    return
  }
  // 构建提交数据，JSON 字段直接传对象，后端会自动处理
  const submitData = {
    ...form.value,
    request_transform: form.value.request_transform || null,
    response_transform: form.value.response_transform || null
  }
  if (form.value.id) {
    await serviceApi.update(form.value.id, submitData)
    ElMessage.success('更新成功')
  } else {
    await serviceApi.create(submitData)
    ElMessage.success('创建成功')
  }
  dialogVisible.value = false
  fetchList()
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确定删除该接口？', '提示')
  await serviceApi.delete(row.id)
  ElMessage.success('删除成功')
  fetchList()
}

// Hook相关
const showHookDialog = async (row) => {
  currentService.value = row
  const res = await serviceHookApi.listByService(row.id)
  serviceHooks.value = res.data || []
  hookDialogVisible.value = true
}

const showHookEditDialog = (row = null) => {
  hookForm.value = row ? { ...row } : {
    service_pk: currentService.value.id,
    hook_point: 'BeforeForward',
    script_id: null,
    priority: 0
  }
  hookEditVisible.value = true
}

const handleToggleStatus = async (row) => {
  const newStatus = row.status === 1 ? 0 : 1
  await serviceHookApi.update(row.id, { status: newStatus })
  row.status = newStatus
  ElMessage.success(newStatus === 1 ? '已启用' : '已禁用')
}

const handleHookSubmit = async () => {
  hookForm.value.service_pk = currentService.value.id
  if (hookForm.value.id) {
    await serviceHookApi.update(hookForm.value.id, hookForm.value)
    ElMessage.success('更新成功')
  } else {
    await serviceHookApi.create(hookForm.value)
    ElMessage.success('创建成功')
  }
  hookEditVisible.value = false
  showHookDialog(currentService.value)
}

const handleDeleteHook = async (row) => {
  await ElMessageBox.confirm('确定删除该Hook关联？', '提示')
  await serviceHookApi.delete(row.id)
  ElMessage.success('删除成功')
  showHookDialog(currentService.value)
}

onMounted(() => {
  fetchList()
  fetchOptions()
})
</script>

<style scoped>
.json-error :deep(textarea) {
  border-color: #f56c6c !important;
  background-color: #fef0f0 !important;
}

.json-error-tip {
  color: #f56c6c;
  font-size: 12px;
  margin-top: 4px;
  display: block;
}

.json-input :deep(textarea) {
  background-color: #f5f7fa;
  cursor: pointer;
}

.json-input :deep(textarea):hover {
  background-color: #e6f7ff;
  border-color: #409eff;
}

.json-input :deep(textarea)::placeholder {
  color: #909399;
}

.json-editor-container {
  height: calc(100vh - 150px);
}

.json-editor {
  height: 100%;
}
</style>
