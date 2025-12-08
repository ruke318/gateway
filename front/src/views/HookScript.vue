<template>
  <div>
    <el-form :inline="true" :model="query" style="margin-bottom: 20px">
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
      <el-table-column prop="name" label="名称" width="150" />
      <el-table-column prop="description" label="描述" />
      <el-table-column prop="created_at" label="创建时间" width="180" />
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="showDialog(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑Hook脚本' : '新增Hook脚本'" width="700px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="脚本名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="脚本内容" required>
          <el-input
            :model-value="scriptPreview"
            type="textarea"
            :rows="3"
            placeholder="双击打开代码编辑器"
            readonly
            class="code-input"
            @dblclick="openCodeEditor"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 代码编辑器全屏弹窗 -->
    <el-dialog
      v-model="codeEditorVisible"
      title="编辑脚本"
      fullscreen
      :close-on-click-modal="false"
    >
      <div class="code-editor-container">
        <VAceEditor
          v-model:value="codeEditorContent"
          lang="javascript"
          theme="monokai"
          :options="editorOptions"
          class="code-editor"
        />
      </div>
      <template #footer>
        <el-button @click="codeEditorVisible = false">取消</el-button>
        <el-button type="primary" @click="saveCodeEditor">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { hookScriptApi } from '../api'
import { VAceEditor } from 'vue3-ace-editor'
import 'ace-builds/src-noconflict/mode-javascript'
import 'ace-builds/src-noconflict/theme-monokai'

const list = ref([])
const dialogVisible = ref(false)
const query = ref({ name: '' })
const form = ref({ name: '', description: '', script_content: '' })

// 代码编辑器
const codeEditorVisible = ref(false)
const codeEditorContent = ref('')
const editorOptions = {
  fontSize: 14,
  showPrintMargin: false,
  tabSize: 2,
  wrap: true
}

// 预览显示（截取前几行）
const scriptPreview = computed(() => {
  const content = form.value.script_content || ''
  if (!content) return ''
  const lines = content.split('\n').slice(0, 3).join('\n')
  return content.split('\n').length > 3 ? lines + '\n...' : lines
})

const openCodeEditor = () => {
  codeEditorContent.value = form.value.script_content || ''
  codeEditorVisible.value = true
}

const saveCodeEditor = () => {
  form.value.script_content = codeEditorContent.value
  codeEditorVisible.value = false
}

const fetchList = async () => {
  const res = await hookScriptApi.list(query.value)
  list.value = res.data || []
}

const resetQuery = () => {
  query.value = { name: '' }
  fetchList()
}

const showDialog = (row = null) => {
  form.value = row ? { ...row } : { name: '', description: '', script_content: '' }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (form.value.id) {
    await hookScriptApi.update(form.value.id, form.value)
    ElMessage.success('更新成功')
  } else {
    await hookScriptApi.create(form.value)
    ElMessage.success('创建成功')
  }
  dialogVisible.value = false
  fetchList()
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确定删除该Hook脚本？', '提示')
  await hookScriptApi.delete(row.id)
  ElMessage.success('删除成功')
  fetchList()
}

onMounted(fetchList)
</script>

<style scoped>
.code-input :deep(textarea) {
  background-color: #f5f7fa;
  cursor: pointer;
  font-family: monospace;
}

.code-input :deep(textarea):hover {
  background-color: #e6f7ff;
  border-color: #409eff;
}

.code-editor-container {
  height: calc(100vh - 150px);
}

.code-editor {
  height: 100%;
  width: 100%;
}
</style>
