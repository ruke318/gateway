<template>
  <div>
    <el-form :inline="true" :model="query" style="margin-bottom: 20px">
      <el-form-item label="编码">
        <el-input v-model="query.code" placeholder="精确匹配" clearable @keyup.enter="fetchList" />
      </el-form-item>
      <el-form-item label="名称">
        <el-input v-model="query.name" placeholder="模糊搜索" clearable @keyup.enter="fetchList" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="fetchList">搜索</el-button>
        <el-button @click="resetQuery">重置</el-button>
        <el-button type="success" @click="showDialog()">新增</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="list" border stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="code" label="编码" min-width="100" />
      <el-table-column prop="name" label="名称" min-width="120" />
      <el-table-column prop="description" label="描述" min-width="150" />
      <el-table-column prop="created_at" label="创建时间" min-width="150" />
      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="showDialog(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑机构' : '新增机构'" width="600px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="编码" required>
          <el-input v-model="form.code" placeholder="机构唯一编码" />
        </el-form-item>
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="机构名称" />
        </el-form-item>
        <el-form-item label="配置">
          <el-input
            v-model="configStr"
            type="textarea"
            :rows="3"
            placeholder="双击打开JSON编辑器"
            readonly
            class="json-input"
            :class="{ 'json-error': configJsonError }"
            @dblclick="openJsonEditor"
          />
          <span v-if="configJsonError" class="json-error-tip">JSON格式错误</span>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" />
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
      title="编辑配置"
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
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { organizationApi } from '../api'
import JsonEditorVue from 'json-editor-vue'

const list = ref([])
const dialogVisible = ref(false)
const query = ref({ code: '', name: '' })
const form = ref({ code: '', name: '', config: null, description: '' })

// JSON字符串和校验
const configStr = ref('')
const configJsonError = ref(false)

// 监听字符串变化，校验并同步到 form
watch(configStr, (val) => {
  if (!val) {
    form.value.config = null
    configJsonError.value = false
    return
  }
  try {
    form.value.config = JSON.parse(val)
    configJsonError.value = false
  } catch (e) {
    configJsonError.value = true
  }
})

// JSON编辑器
const jsonEditorVisible = ref(false)
const jsonEditorContent = ref(null)

const openJsonEditor = () => {
  jsonEditorContent.value = form.value.config || {}
  jsonEditorVisible.value = true
}

const saveJsonEditor = () => {
  let jsonObj = jsonEditorContent.value
  if (typeof jsonObj === 'string') {
    try {
      jsonObj = JSON.parse(jsonObj)
    } catch (e) {
      ElMessage.error('JSON格式错误')
      return
    }
  }
  form.value.config = jsonObj
  configStr.value = jsonObj ? JSON.stringify(jsonObj) : ''
  configJsonError.value = false
  jsonEditorVisible.value = false
}

// 解析 JSON 字段
const parseJsonField = (val) => {
  if (!val) return null
  if (typeof val === 'object') return val
  if (typeof val === 'string') {
    try {
      return JSON.parse(val)
    } catch (e) {
      return null
    }
  }
  return null
}

const fetchList = async () => {
  const res = await organizationApi.list(query.value)
  list.value = res.data || []
}

const resetQuery = () => {
  query.value = { code: '', name: '' }
  fetchList()
}

const showDialog = (row = null) => {
  if (row) {
    form.value = { ...row }
    form.value.config = parseJsonField(row.config)
    configStr.value = form.value.config ? JSON.stringify(form.value.config) : ''
  } else {
    form.value = { code: '', name: '', config: null, description: '' }
    configStr.value = ''
  }
  configJsonError.value = false
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (configJsonError.value) {
    ElMessage.error('请修正JSON格式错误')
    return
  }
  const submitData = {
    ...form.value,
    config: form.value.config || null
  }
  if (form.value.id) {
    await organizationApi.update(form.value.id, submitData)
    ElMessage.success('更新成功')
  } else {
    await organizationApi.create(submitData)
    ElMessage.success('创建成功')
  }
  dialogVisible.value = false
  fetchList()
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确定删除该机构？', '提示')
  await organizationApi.delete(row.id)
  ElMessage.success('删除成功')
  fetchList()
}

onMounted(fetchList)
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
