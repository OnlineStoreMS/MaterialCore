<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Picture, FolderOpened, ChatDotRound, VideoCamera } from '@element-plus/icons-vue'
import { fetchDashboardStats, type DashboardStats } from '../api/material'

const router = useRouter()
const stats = ref<DashboardStats>({
  materialCount: 0,
  imageCount: 0,
  videoCount: 0,
  weekNewCount: 0,
  categoryCount: 0,
  byCategory: [],
})
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    stats.value = await fetchDashboardStats()
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div v-loading="loading" class="dashboard">
    <h2 class="page-title">素材中心工作台</h2>
    <p class="desc">
      灵活管理商品宣传素材（询盘图、效果图、海报、带货短视频等），支持自定义分类与手机扫码上传。
      与 ProductCore 商品图文分离，不强制关联商品。
    </p>

    <div class="stat-row">
      <el-card shadow="never" class="stat"><div class="n">{{ stats.materialCount }}</div><div class="l">素材总数</div></el-card>
      <el-card shadow="never" class="stat"><div class="n">{{ stats.imageCount }}</div><div class="l">图片</div></el-card>
      <el-card shadow="never" class="stat"><div class="n">{{ stats.videoCount }}</div><div class="l">视频</div></el-card>
      <el-card shadow="never" class="stat"><div class="n">{{ stats.weekNewCount }}</div><div class="l">近 7 日新增</div></el-card>
      <el-card shadow="never" class="stat"><div class="n">{{ stats.categoryCount }}</div><div class="l">分类数</div></el-card>
    </div>

    <div class="card-grid">
      <el-card shadow="hover" class="action-card" @click="router.push('/materials')">
        <el-icon :size="32" color="#409eff"><Picture /></el-icon>
        <h3>素材库</h3>
        <p>上传、浏览与管理素材</p>
      </el-card>
      <el-card shadow="hover" class="action-card" @click="router.push('/categories')">
        <el-icon :size="32" color="#67c23a"><FolderOpened /></el-icon>
        <h3>分类管理</h3>
        <p>自定义树形分类</p>
      </el-card>
      <el-card shadow="hover" class="action-card" @click="router.push('/inquiry')">
        <el-icon :size="32" color="#e6a23c"><ChatDotRound /></el-icon>
        <h3>询盘工作台</h3>
        <p>快选复制链接 / 打包下载</p>
      </el-card>
      <el-card shadow="hover" class="action-card" @click="router.push('/materials')">
        <el-icon :size="32" color="#909399"><VideoCamera /></el-icon>
        <h3>扫码上传</h3>
        <p>在素材库内使用手机扫码</p>
      </el-card>
    </div>

    <el-card v-if="stats.byCategory?.length" shadow="never" class="by-cat">
      <template #header>各分类素材数</template>
      <div class="cat-list">
        <div v-for="c in stats.byCategory" :key="c.categoryId" class="cat-item">
          <span>{{ c.categoryName || `分类#${c.categoryId}` }}</span>
          <strong>{{ c.count }}</strong>
        </div>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.dashboard { width: 100%; }
.page-title { margin: 0 0 8px; font-size: 22px; }
.desc { color: #606266; margin: 0 0 24px; line-height: 1.6; max-width: 720px; }
.stat-row {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 12px;
  margin-bottom: 20px;
}
.stat { text-align: center; }
.stat .n { font-size: 28px; font-weight: 600; color: #303133; }
.stat .l { font-size: 13px; color: #909399; margin-top: 4px; }
.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}
.action-card {
  cursor: pointer;
  text-align: center;
  transition: transform 0.15s;
}
.action-card:hover { transform: translateY(-2px); }
.action-card h3 { margin: 12px 0 6px; font-size: 16px; }
.action-card p { margin: 0; color: #909399; font-size: 13px; }
.by-cat { margin-top: 8px; }
.cat-list { display: flex; flex-wrap: wrap; gap: 12px 24px; }
.cat-item { display: flex; gap: 8px; align-items: baseline; font-size: 14px; color: #606266; }
.cat-item strong { color: #303133; }
</style>
