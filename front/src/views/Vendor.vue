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
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="code" label="编码" width="120" />
      <el-table-column prop="name" label="名称" width="150" />
      <el-table-column prop="base_url" label="基础URL" />
      <el-table-column prop="description" label="描述" />
      <el-table-column prop="created_at" label="创建时间" width="180" />
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="showDialog(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑厂商' : '新增厂商'" width="600px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="编码" required>
          <el-input v-model="form.code" placeholder="厂商唯一编码" />
        </el-form-item>
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="厂商名称" />
        </el-form-item>
        <el-form-item label="基础URL">
          <el-input v-model="form.base_url" placeholder="如 https://api.vendor.com" />
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
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { vendorApi } from '../api'

const list = ref([])
const dialogVisible = ref(false)
const query = ref({ code: '', name: '' })
const form = ref({ code: '', name: '', base_url: '', description: '' })

const fetchList = async () => {
  const res = await vendorApi.list(query.value)
  list.value = res.data || []
}

const resetQuery = () => {
  query.value = { code: '', name: '' }
  fetchList()
}

const showDialog = (row = null) => {
  form.value = row ? { ...row } : { code: '', name: '', base_url: '', description: '' }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (form.value.id) {
    await vendorApi.update(form.value.id, form.value)
    ElMessage.success('更新成功')
  } else {
    await vendorApi.create(form.value)
    ElMessage.success('创建成功')
  }
  dialogVisible.value = false
  fetchList()
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确定删除该厂商？', '提示')
  await vendorApi.delete(row.id)
  ElMessage.success('删除成功')
  fetchList()
}

onMounted(fetchList)
</script>
