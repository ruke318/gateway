<template>
  <div class="page-container">
    <div class="page-header">
      <h2>操作日志</h2>
    </div>

    <div class="search-bar">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="用户名">
          <el-input
            v-model="searchForm.username"
            placeholder="请输入用户名"
            clearable
            style="width: 200px"
          />
        </el-form-item>
        <el-form-item label="操作类型">
          <el-select
            v-model="searchForm.operation"
            placeholder="请选择操作类型"
            clearable
            style="width: 150px"
          >
            <el-option label="创建" value="create" />
            <el-option label="更新" value="update" />
            <el-option label="删除" value="delete" />
          </el-select>
        </el-form-item>
        <el-form-item label="资源类型">
          <el-select
            v-model="searchForm.resource"
            placeholder="请选择资源类型"
            clearable
            style="width: 150px"
          >
            <el-option label="用户" value="user" />
            <el-option label="厂商" value="vendor" />
            <el-option label="机构" value="organization" />
            <el-option label="接口" value="service" />
            <el-option label="脚本" value="script" />
            <el-option label="Hook脚本" value="hook_script" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <el-table :data="tableData" border style="width: 100%">
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="username" label="操作用户" min-width="100" />
      <el-table-column prop="operation" label="操作类型" width="90">
        <template #default="{ row }">
          <el-tag
            :type="
              row.operation === 'create'
                ? 'success'
                : row.operation === 'update'
                ? 'warning'
                : 'danger'
            "
          >
            {{
              row.operation === 'create'
                ? '创建'
                : row.operation === 'update'
                ? '更新'
                : '删除'
            }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="resource" label="资源类型" min-width="120">
        <template #default="{ row }">
          {{ formatResource(row.resource) }}
        </template>
      </el-table-column>
      <el-table-column prop="ip" label="IP地址" min-width="110" />
      <el-table-column prop="created_at" label="操作时间" min-width="150">
        <template #default="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="80" align="center" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="handleViewDetail(row)">
            详情
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[20, 50, 100, 200]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchData"
        @current-change="fetchData"
      />
    </div>

    <!-- 详情对话框 -->
    <el-dialog v-model="detailVisible" title="操作详情" width="700px">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="ID" :span="1">
          {{ currentLog.id }}
        </el-descriptions-item>
        <el-descriptions-item label="操作时间" :span="1">
          {{ formatDate(currentLog.created_at) }}
        </el-descriptions-item>
        <el-descriptions-item label="操作用户" :span="1">
          {{ currentLog.username }}
        </el-descriptions-item>
        <el-descriptions-item label="用户ID" :span="1">
          {{ currentLog.user_id }}
        </el-descriptions-item>
        <el-descriptions-item label="操作类型" :span="1">
          <el-tag
            :type="
              currentLog.operation === 'create'
                ? 'success'
                : currentLog.operation === 'update'
                ? 'warning'
                : 'danger'
            "
          >
            {{
              currentLog.operation === 'create'
                ? '创建'
                : currentLog.operation === 'update'
                ? '更新'
                : '删除'
            }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="资源类型" :span="1">
          {{ formatResource(currentLog.resource) }}
        </el-descriptions-item>
        <el-descriptions-item label="资源ID" :span="1">
          {{ currentLog.resource_id }}
        </el-descriptions-item>
        <el-descriptions-item label="IP地址" :span="1">
          {{ currentLog.ip }}
        </el-descriptions-item>
      </el-descriptions>

      <div class="detail-section" v-if="currentLog.after_data || currentLog.before_data">
        <h4 class="detail-title">
          <el-icon><Document /></el-icon>
          操作内容
        </h4>

        <!-- 新增操作：只显示新增的数据 -->
        <div v-if="currentLog.operation === 'create'" class="detail-card create-card">
          <div class="card-header">
            <el-tag type="success" size="small">新增数据</el-tag>
          </div>
          <el-table :data="getDetailFields(currentLog.after_data)" border size="small">
            <el-table-column prop="field" label="字段" width="180">
              <template #default="{ row }">
                <span class="field-name">{{ row.field }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="value" label="值">
              <template #default="{ row }">
                <span class="field-value">{{ row.value }}</span>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <!-- 更新操作：显示修改前后对比 -->
        <div v-else-if="currentLog.operation === 'update'" class="detail-card update-card">
          <div class="card-header">
            <el-tag type="warning" size="small">修改内容</el-tag>
            <span class="tip-text">对比修改前后的差异</span>
          </div>
          <el-table :data="getCompareFields(currentLog.before_data, currentLog.after_data)" border size="small">
            <el-table-column prop="field" label="字段" width="150">
              <template #default="{ row }">
                <span class="field-name">{{ row.field }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="before" label="修改前">
              <template #default="{ row }">
                <span class="field-value" :class="{ 'text-muted': row.changed }">
                  {{ row.before }}
                </span>
              </template>
            </el-table-column>
            <el-table-column prop="after" label="修改后" width="250">
              <template #default="{ row }">
                <el-tag v-if="row.changed" type="warning" effect="plain" size="small">
                  {{ row.after }}
                </el-tag>
                <span v-else class="field-value">{{ row.after }}</span>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <!-- 删除操作：只显示删除的数据 -->
        <div v-else-if="currentLog.operation === 'delete'" class="detail-card delete-card">
          <div class="card-header">
            <el-tag type="danger" size="small">删除数据</el-tag>
          </div>
          <el-table :data="getDetailFields(currentLog.before_data)" border size="small">
            <el-table-column prop="field" label="字段" width="180">
              <template #default="{ row }">
                <span class="field-name">{{ row.field }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="value" label="删除的值">
              <template #default="{ row }">
                <span class="field-value deleted">{{ row.value }}</span>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <!-- 原始JSON展示 -->
        <div class="raw-json">
          <el-collapse>
            <el-collapse-item title="查看原始JSON" name="1">
              <div v-if="currentLog.before_data" class="json-section">
                <div class="json-label">修改前：</div>
                <pre class="json-content">{{ formatDetail(currentLog.before_data) }}</pre>
              </div>
              <div v-if="currentLog.after_data" class="json-section">
                <div class="json-label">修改后：</div>
                <pre class="json-content">{{ formatDetail(currentLog.after_data) }}</pre>
              </div>
            </el-collapse-item>
          </el-collapse>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Document } from '@element-plus/icons-vue'
import { getOperationLogs, getOperationLog } from '../api'

const tableData = ref([])
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)

const searchForm = ref({
  username: '',
  operation: '',
  resource: ''
})

const detailVisible = ref(false)
const currentLog = ref({})

const fetchData = async () => {
  try {
    const res = await getOperationLogs({
      page: currentPage.value,
      size: pageSize.value,
      ...searchForm.value
    })
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  } catch (err) {
    ElMessage.error('获取数据失败')
  }
}

const handleSearch = () => {
  currentPage.value = 1
  fetchData()
}

const handleReset = () => {
  searchForm.value = {
    username: '',
    operation: '',
    resource: ''
  }
  currentPage.value = 1
  fetchData()
}

const handleViewDetail = async (row) => {
  try {
    const res = await getOperationLog(row.id)
    currentLog.value = res.data
    detailVisible.value = true
  } catch (err) {
    ElMessage.error('获取详情失败')
  }
}

const formatResource = (resource) => {
  const map = {
    user: '用户',
    vendor: '厂商',
    organization: '机构',
    service: '接口',
    script: '脚本库',
    hook_script: 'Hook脚本',
    service_hook: '接口Hook关联'
  }
  return map[resource] || resource
}

const formatDate = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN')
}

const formatDetail = (detail) => {
  if (!detail) return '无'
  try {
    return JSON.stringify(JSON.parse(detail), null, 2)
  } catch {
    return detail
  }
}

// 将JSON详情转换为字段数组，便于表格展示
const getDetailFields = (detail) => {
  if (!detail) return []
  try {
    const obj = JSON.parse(detail)
    return Object.keys(obj).map(key => {
      let value = obj[key]
      // 格式化值
      if (value === null || value === undefined) {
        value = '-'
      } else if (typeof value === 'object') {
        value = JSON.stringify(value)
      } else {
        value = String(value)
      }

      return {
        field: key,  // 直接使用原始字段名
        value: value
      }
    })
  } catch {
    return []
  }
}

// 对比修改前后的数据，生成对比字段数组
const getCompareFields = (beforeData, afterData) => {
  if (!beforeData || !afterData) return []

  try {
    const before = JSON.parse(beforeData)
    const after = JSON.parse(afterData)

    // 合并所有字段
    const allKeys = new Set([...Object.keys(before), ...Object.keys(after)])

    return Array.from(allKeys).map(key => {
      // 格式化值
      const formatValue = (val) => {
        if (val === null || val === undefined) return '-'
        if (typeof val === 'object') return JSON.stringify(val)
        return String(val)
      }

      const beforeValue = formatValue(before[key])
      const afterValue = formatValue(after[key])

      return {
        field: key,  // 直接使用原始字段名
        before: beforeValue,
        after: afterValue,
        changed: beforeValue !== afterValue
      }
    })
  } catch {
    return []
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.page-container {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 500;
}

.search-bar {
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

/* 详情区域 */
.detail-section {
  margin-top: 20px;
}

.detail-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 0 16px 0;
  padding-bottom: 12px;
  border-bottom: 2px solid #e4e7ed;
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

.detail-title .el-icon {
  font-size: 18px;
  color: #409eff;
}

/* 详情卡片 */
.detail-card {
  margin-bottom: 16px;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
}

.tip-text {
  font-size: 12px;
  color: #909399;
}

.create-card {
  border: 1px solid #67c23a;
}

.create-card .card-header {
  background: #f0f9ff;
  border-bottom-color: #67c23a;
}

.update-card {
  border: 1px solid #e6a23c;
}

.update-card .card-header {
  background: #fdf6ec;
  border-bottom-color: #e6a23c;
}

.delete-card {
  border: 1px solid #f56c6c;
}

.delete-card .card-header {
  background: #fef0f0;
  border-bottom-color: #f56c6c;
}

/* 字段样式 */
.field-name {
  font-weight: 500;
  color: #606266;
}

.field-value {
  color: #303133;
  word-break: break-all;
}

.field-value.deleted {
  text-decoration: line-through;
  color: #909399;
}

/* 原始JSON */
.raw-json {
  margin-top: 16px;
}

.raw-json :deep(.el-collapse-item__header) {
  font-size: 13px;
  color: #909399;
  background: #fafafa;
  padding: 0 12px;
  border-radius: 4px;
}

.json-content {
  margin: 0;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
  color: #606266;
  max-height: 300px;
  overflow-y: auto;
}

/* 表格样式优化 */
.detail-card :deep(.el-table) {
  font-size: 13px;
}

.detail-card :deep(.el-table th) {
  background: #fafafa;
  color: #606266;
  font-weight: 500;
}

.detail-card :deep(.el-table td) {
  padding: 10px 0;
}

.detail-card :deep(.el-table .cell) {
  line-height: 1.6;
}

/* 对比展示相关样式 */
.text-muted {
  color: #909399;
  text-decoration: line-through;
}

.json-section {
  margin-bottom: 12px;
}

.json-section:last-child {
  margin-bottom: 0;
}

.json-label {
  font-size: 13px;
  font-weight: 500;
  color: #606266;
  margin-bottom: 8px;
  padding-left: 4px;
  border-left: 3px solid #409eff;
}

</style>
