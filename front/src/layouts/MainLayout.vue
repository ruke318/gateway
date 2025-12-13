<template>
  <el-container class="layout">
    <el-aside width="200px">
      <div class="logo">
        <h3>网关管理后台</h3>
      </div>
      <el-menu
        :default-active="$route.path"
        router
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409eff"
      >
        <el-menu-item index="/vendor">
          <el-icon><Shop /></el-icon>
          <span>厂商管理</span>
        </el-menu-item>
        <el-menu-item index="/organization">
          <el-icon><OfficeBuilding /></el-icon>
          <span>机构管理</span>
        </el-menu-item>
        <el-menu-item index="/service">
          <el-icon><Connection /></el-icon>
          <span>接口管理</span>
        </el-menu-item>
        <el-menu-item index="/script">
          <el-icon><Document /></el-icon>
          <span>公共函数库</span>
        </el-menu-item>
        <el-menu-item index="/hook-script">
          <el-icon><Files /></el-icon>
          <span>Hook脚本</span>
        </el-menu-item>
        <el-menu-item index="/user">
          <el-icon><User /></el-icon>
          <span>用户管理</span>
        </el-menu-item>
        <el-menu-item index="/operation-log">
          <el-icon><List /></el-icon>
          <span>操作日志</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header>
        <span class="title">网关管理系统</span>
        <div class="user-info">
          <el-dropdown>
            <span class="user-name">
              <el-icon><User /></el-icon>
              {{ username }}
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="handleChangePassword">
                  <el-icon><Lock /></el-icon>
                  修改密码
                </el-dropdown-item>
                <el-dropdown-item divided @click="handleLogout">
                  <el-icon><SwitchButton /></el-icon>
                  退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>

    <!-- 修改密码对话框 -->
    <el-dialog v-model="changePasswordVisible" title="修改密码" width="400px">
      <el-form :model="passwordForm" ref="passwordFormRef" label-width="80px">
        <el-form-item label="旧密码" prop="old_password">
          <el-input
            v-model="passwordForm.old_password"
            type="password"
            placeholder="请输入旧密码"
            show-password
          />
        </el-form-item>
        <el-form-item label="新密码" prop="new_password">
          <el-input
            v-model="passwordForm.new_password"
            type="password"
            placeholder="请输入新密码"
            show-password
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="changePasswordVisible = false">取消</el-button>
        <el-button type="primary" @click="handleChangePasswordSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>
  </el-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Shop,
  OfficeBuilding,
  Connection,
  Document,
  Files,
  User,
  List,
  Lock,
  SwitchButton
} from '@element-plus/icons-vue'
import { changePassword } from '../api'

const router = useRouter()
const username = ref(localStorage.getItem('username') || '未登录')

const changePasswordVisible = ref(false)
const passwordFormRef = ref(null)
const passwordForm = ref({
  old_password: '',
  new_password: ''
})

const handleChangePassword = () => {
  changePasswordVisible.value = true
}

const handleChangePasswordSubmit = async () => {
  if (!passwordForm.value.old_password || !passwordForm.value.new_password) {
    ElMessage.warning('请填写完整信息')
    return
  }

  try {
    await changePassword(passwordForm.value)
    ElMessage.success('密码修改成功，请重新登录')
    changePasswordVisible.value = false
    handleLogout()
  } catch (err) {
    ElMessage.error(err.message || '修改密码失败')
  }
}

const handleLogout = () => {
  ElMessageBox.confirm('确定要退出登录吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  })
    .then(() => {
      localStorage.clear()
      router.push('/login')
      ElMessage.success('已退出登录')
    })
    .catch(() => {})
}

onMounted(() => {
  // 可以在这里获取用户信息
})
</script>

<style scoped>
.layout {
  height: 100%;
  width: 100%;
}

.el-aside {
  background-color: #304156;
  height: 100%;
  overflow-y: auto;
}

.el-aside::-webkit-scrollbar {
  width: 0;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #2b3a4b;
}

.logo h3 {
  margin: 0;
  color: #fff;
  font-size: 16px;
  font-weight: 500;
}

.el-menu {
  border-right: none;
}

.el-header {
  background-color: #fff;
  border-bottom: 1px solid #e6e6e6;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.el-main {
  background-color: #f5f7fa;
  overflow-y: auto;
}

.title {
  font-size: 18px;
  font-weight: bold;
}

.user-info {
  display: flex;
  align-items: center;
}

.user-name {
  display: flex;
  align-items: center;
  gap: 5px;
  cursor: pointer;
  padding: 5px 10px;
  border-radius: 4px;
  transition: background-color 0.3s;
}

.user-name:hover {
  background-color: #f5f7fa;
}
</style>
