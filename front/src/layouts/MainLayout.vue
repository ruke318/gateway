<template>
  <el-container class="layout">
    <el-aside width="200px">
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
      </el-menu>
    </el-aside>
    <el-container>
      <el-header>
        <span class="title">网关管理后台</span>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { Shop, OfficeBuilding, Connection, Document, Files } from '@element-plus/icons-vue'
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
.el-menu {
  border-right: none;
}
.el-header {
  background-color: #fff;
  border-bottom: 1px solid #e6e6e6;
  display: flex;
  align-items: center;
}
.el-main {
  background-color: #f5f7fa;
  overflow-y: auto;
}
.title {
  font-size: 18px;
  font-weight: bold;
}
</style>
