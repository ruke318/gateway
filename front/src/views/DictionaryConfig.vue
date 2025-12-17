<template>
  <div>
    <el-form :inline="true" :model="query" style="margin-bottom: 20px">
      <el-form-item label="机构">
        <el-select v-model="query.org_id" placeholder="全部机构" clearable filterable>
          <el-option
            v-for="org in orgList"
            :key="org.code"
            :label="`${org.name} (${org.code})`"
            :value="org.code"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="字典类型">
        <el-input v-model="query.dict_type" placeholder="如: payment_method" clearable @keyup.enter="fetchList" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="fetchList">搜索</el-button>
        <el-button @click="resetQuery">重置</el-button>
        <el-button type="success" @click="showDialog()">新增</el-button>
        <el-button type="warning" @click="showBatchDialog">批量导入</el-button>
        <el-button type="info" @click="handleReload">重新加载字典</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="list" border stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="org_id" label="机构ID" min-width="100" />
      <el-table-column prop="dict_type" label="字典类型" min-width="120" />
      <el-table-column prop="dict_key" label="标准键" min-width="120" />
      <el-table-column prop="dict_value" label="机构值" min-width="120" />
      <el-table-column prop="description" label="说明" min-width="150" show-overflow-tooltip />
      <el-table-column prop="created_at" label="创建时间" min-width="150" />
      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="showDialog(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 新增/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑字典配置' : '新增字典配置'" width="600px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="机构ID" required>
          <el-select v-model="form.org_id" placeholder="选择机构" filterable>
            <el-option
              v-for="org in orgList"
              :key="org.code"
              :label="`${org.name} (${org.code})`"
              :value="org.code"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="字典类型" required>
          <el-input v-model="form.dict_type" placeholder="如: payment_method, order_status" />
          <div style="font-size: 12px; color: #999; margin-top: 5px">
            常用类型: payment_method(支付方式), order_status(订单状态), bank_code(银行编码)
          </div>
        </el-form-item>
        <el-form-item label="标准键" required>
          <el-input v-model="form.dict_key" placeholder="如: ALIPAY, WECHAT, SUCCESS" />
          <div style="font-size: 12px; color: #999; margin-top: 5px">
            标准键用于跨机构映射，建议使用大写英文
          </div>
        </el-form-item>
        <el-form-item label="机构值" required>
          <el-input v-model="form.dict_value" placeholder="如: 01, A001" />
          <div style="font-size: 12px; color: #999; margin-top: 5px">
            机构特定的值，由各机构自行定义
          </div>
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="可选，描述该字典的用途" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 批量导入对话框 -->
    <el-dialog v-model="batchDialogVisible" title="批量导入字典配置" width="800px">
      <div style="margin-bottom: 10px">
        <el-alert type="info" :closable="false">
          <div>批量导入格式（JSON数组）：</div>
          <pre style="margin-top: 10px; font-size: 12px">
[
  {
    "org_id": "org001",
    "dict_type": "payment_method",
    "dict_key": "ALIPAY",
    "dict_value": "01",
    "description": "支付宝"
  },
  {
    "org_id": "org001",
    "dict_type": "payment_method",
    "dict_key": "WECHAT",
    "dict_value": "02",
    "description": "微信支付"
  }
]</pre>
        </el-alert>
      </div>
      <el-input
        v-model="batchData"
        type="textarea"
        :rows="15"
        placeholder="粘贴JSON数据..."
        :class="{ 'json-error': batchJsonError }"
      />
      <span v-if="batchJsonError" class="json-error-tip">JSON格式错误</span>
      <template #footer>
        <el-button @click="batchDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleBatchImport">导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { dictionaryConfigApi, reloadDictionary, organizationApi } from '../api'

const list = ref([])
const orgList = ref([])
const dialogVisible = ref(false)
const batchDialogVisible = ref(false)
const query = ref({ org_id: '', dict_type: '' })
const form = ref({
  org_id: '',
  dict_type: '',
  dict_key: '',
  dict_value: '',
  description: ''
})

const batchData = ref('')
const batchJsonError = ref(false)

// 监听批量数据，校验JSON格式
watch(batchData, (val) => {
  if (!val) {
    batchJsonError.value = false
    return
  }
  try {
    JSON.parse(val)
    batchJsonError.value = false
  } catch (e) {
    batchJsonError.value = true
  }
})

// 加载机构列表
const fetchOrganizations = async () => {
  try {
    const res = await organizationApi.list()
    orgList.value = res.data || []
  } catch (err) {
    ElMessage.error(err.response?.data?.message || '加载机构列表失败')
  }
}

// 加载字典列表
const fetchList = async () => {
  try {
    const params = {}
    if (query.value.org_id) params.org_id = query.value.org_id
    if (query.value.dict_type) params.dict_type = query.value.dict_type

    const res = await dictionaryConfigApi.list(params)
    list.value = res.data || []
  } catch (err) {
    ElMessage.error(err.response?.data?.message || '加载失败')
  }
}

// 重置查询
const resetQuery = () => {
  query.value = { org_id: '', dict_type: '' }
  fetchList()
}

// 显示对话框
const showDialog = (row) => {
  if (row) {
    form.value = { ...row }
  } else {
    form.value = {
      org_id: '',
      dict_type: '',
      dict_key: '',
      dict_value: '',
      description: ''
    }
  }
  dialogVisible.value = true
}

// 显示批量导入对话框
const showBatchDialog = () => {
  batchData.value = ''
  batchJsonError.value = false
  batchDialogVisible.value = true
}

// 提交表单
const handleSubmit = async () => {
  if (!form.value.org_id || !form.value.dict_type || !form.value.dict_key || !form.value.dict_value) {
    ElMessage.warning('请填写所有必填字段')
    return
  }

  try {
    if (form.value.id) {
      await dictionaryConfigApi.update(form.value.id, form.value)
      ElMessage.success('更新成功')
    } else {
      await dictionaryConfigApi.create(form.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } catch (err) {
    ElMessage.error(err.response?.data?.message || '操作失败')
  }
}

// 批量导入
const handleBatchImport = async () => {
  if (!batchData.value) {
    ElMessage.warning('请输入JSON数据')
    return
  }

  if (batchJsonError.value) {
    ElMessage.error('JSON格式错误，请检查')
    return
  }

  try {
    const data = JSON.parse(batchData.value)
    if (!Array.isArray(data)) {
      ElMessage.error('数据格式错误，必须是数组')
      return
    }

    await dictionaryConfigApi.batchCreate(data)
    ElMessage.success(`成功导入 ${data.length} 条数据`)
    batchDialogVisible.value = false
    fetchList()
  } catch (err) {
    ElMessage.error(err.response?.data?.message || '导入失败')
  }
}

// 删除
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除字典配置 "${row.dict_type} - ${row.dict_key}" 吗？`, '提示', {
      type: 'warning'
    })
    await dictionaryConfigApi.delete(row.id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error(err.response?.data?.message || '删除失败')
    }
  }
}

// 重新加载字典
const handleReload = async () => {
  try {
    await reloadDictionary()
    ElMessage.success('字典重新加载成功，新配置已生效')
  } catch (err) {
    ElMessage.error(err.response?.data?.message || '重新加载失败')
  }
}

onMounted(() => {
  fetchOrganizations()
  fetchList()
})
</script>

<style scoped>
.json-input {
  cursor: pointer;
}

.json-error {
  border-color: red;
}

.json-error-tip {
  color: red;
  font-size: 12px;
}

pre {
  background: #f5f7fa;
  padding: 10px;
  border-radius: 4px;
  overflow-x: auto;
}
</style>
