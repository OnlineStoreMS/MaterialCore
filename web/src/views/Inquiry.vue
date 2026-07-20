<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { CopyDocument, Download, Search } from '@element-plus/icons-vue'
import axios from 'axios'
import { getToken } from '../utils/auth'
import {
  listCategories,
  listMaterials,
  flattenCategories,
  type CategoryNode,
  type Material,
} from '../api/material'

const loading = ref(false)
const exporting = ref(false)
const categories = ref<CategoryNode[]>([])
const materials = ref<Material[]>([])
const total = ref(0)
const selected = ref<Material[]>([])

const query = reactive({
  categoryId: undefined as number | undefined,
  keyword: '',
  mediaType: '',
  page: 1,
  pageSize: 36,
})

const flatCats = computed(() => flattenCategories(categories.value))
const selectedIds = computed(() => selected.value.map((m) => m.id))

function isSelected(id: number) {
  return selected.value.some((m) => m.id === id)
}

function toggle(m: Material) {
  const i = selected.value.findIndex((x) => x.id === m.id)
  if (i >= 0) selected.value.splice(i, 1)
  else selected.value.push(m)
}

function clearSelection() {
  selected.value = []
}

async function loadCategories() {
  categories.value = await listCategories()
}

async function load() {
  loading.value = true
  try {
    const res = await listMaterials({
      categoryId: query.categoryId,
      includeChildren: true,
      keyword: query.keyword || undefined,
      mediaType: query.mediaType || undefined,
      page: query.page,
      pageSize: query.pageSize,
    })
    materials.value = res.list || []
    total.value = res.total || 0
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

async function copyLinks() {
  if (!selected.value.length) {
    ElMessage.warning('请先选择素材')
    return
  }
  const text = selected.value.map((m) => `${m.title}\n${m.url}`).join('\n\n')
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success(`已复制 ${selected.value.length} 条链接`)
  } catch {
    ElMessage.error('复制失败')
  }
}

async function downloadZip() {
  if (!selected.value.length) {
    ElMessage.warning('请先选择素材')
    return
  }
  exporting.value = true
  try {
    const res = await axios.post(
      '/api/v1/admin/materials/export-zip',
      { ids: selectedIds.value },
      {
        responseType: 'blob',
        timeout: 120000,
        headers: {
          Authorization: `Bearer ${getToken() || ''}`,
          'Content-Type': 'application/json',
        },
      },
    )
    const ct = String(res.headers['content-type'] ?? '')
    if (ct.includes('application/json')) {
      const text = await (res.data as Blob).text()
      const body = JSON.parse(text)
      throw new Error(body.message || '打包失败')
    }
    const blob = res.data as Blob
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `materials_${Date.now()}.zip`
    a.click()
    URL.revokeObjectURL(url)
    ElMessage.success('打包下载已开始')
  } catch (e) {
    ElMessage.error((e as Error).message || '打包失败')
  } finally {
    exporting.value = false
  }
}

onMounted(async () => {
  try {
    await loadCategories()
    await load()
  } catch (e) {
    ElMessage.error((e as Error).message || '初始化失败')
  }
})
</script>

<template>
  <div class="page">
    <div class="hd">
      <div>
        <h2>询盘工作台</h2>
        <p>按分类快选素材，一键复制链接发给顾客，或打包下载后发送。</p>
      </div>
      <div class="hd-actions">
        <el-tag v-if="selected.length" type="primary">已选 {{ selected.length }}</el-tag>
        <el-button v-if="selected.length" text @click="clearSelection">清空</el-button>
        <el-button type="primary" plain :icon="CopyDocument" :disabled="!selected.length" @click="copyLinks">
          复制链接
        </el-button>
        <el-button type="primary" :icon="Download" :loading="exporting" :disabled="!selected.length" @click="downloadZip">
          打包下载
        </el-button>
      </div>
    </div>

    <div class="toolbar">
      <el-select
        v-model="query.categoryId"
        clearable
        placeholder="全部分类"
        style="width: 200px"
        @change="() => { query.page = 1; load() }"
      >
        <el-option
          v-for="c in flatCats"
          :key="c.id"
          :label="'　'.repeat(c.depth) + c.name"
          :value="c.id"
        />
      </el-select>
      <el-select
        v-model="query.mediaType"
        clearable
        placeholder="类型"
        style="width: 110px"
        @change="() => { query.page = 1; load() }"
      >
        <el-option label="图片" value="image" />
        <el-option label="视频" value="video" />
      </el-select>
      <el-input
        v-model="query.keyword"
        clearable
        placeholder="关键词"
        style="width: 200px"
        :prefix-icon="Search"
        @keyup.enter="load"
      />
      <el-button type="primary" @click="load">查询</el-button>
    </div>

    <div v-loading="loading" class="grid">
      <div
        v-for="m in materials"
        :key="m.id"
        class="card"
        :class="{ selected: isSelected(m.id) }"
        @click="toggle(m)"
      >
        <div class="check" v-if="isSelected(m.id)">✓</div>
        <el-image v-if="m.mediaType === 'image'" :src="m.url" fit="cover" class="thumb" />
        <div v-else class="thumb video">▶</div>
        <div class="title">{{ m.title }}</div>
      </div>
      <el-empty v-if="!loading && !materials.length" description="没有匹配的素材" />
    </div>

    <div class="pager">
      <el-pagination
        v-model:current-page="query.page"
        v-model:page-size="query.pageSize"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="load"
      />
    </div>

    <el-card v-if="selected.length" shadow="never" class="preview">
      <template #header>已选预览</template>
      <div class="preview-list">
        <div v-for="m in selected" :key="m.id" class="preview-item">
          <el-image v-if="m.mediaType === 'image'" :src="m.url" fit="cover" class="p-thumb" />
          <div v-else class="p-thumb video">视频</div>
          <div class="p-meta">
            <div class="p-title">{{ m.title }}</div>
            <div class="p-url">{{ m.url }}</div>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.page { display: flex; flex-direction: column; gap: 14px; }
.hd { display: flex; justify-content: space-between; align-items: flex-start; gap: 16px; flex-wrap: wrap; }
.hd h2 { margin: 0 0 6px; font-size: 20px; }
.hd p { margin: 0; color: #909399; font-size: 13px; }
.hd-actions { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.toolbar { display: flex; flex-wrap: wrap; gap: 8px; }
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 10px;
  min-height: 160px;
}
.card {
  position: relative; background: #fff; border: 2px solid #ebeef5; border-radius: 8px;
  overflow: hidden; cursor: pointer; transition: border-color .15s;
}
.card.selected { border-color: #409eff; }
.check {
  position: absolute; z-index: 1; right: 6px; top: 6px;
  width: 22px; height: 22px; border-radius: 50%; background: #409eff; color: #fff;
  display: flex; align-items: center; justify-content: center; font-size: 12px; font-weight: 700;
}
.thumb { width: 100%; aspect-ratio: 1; display: block; background: #f5f7fa; }
.thumb.video, .p-thumb.video {
  display: flex; align-items: center; justify-content: center;
  background: #1a1a1a; color: #fff; font-size: 20px;
}
.title {
  padding: 6px 8px; font-size: 12px; color: #303133;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.pager { display: flex; justify-content: flex-end; }
.preview-list { display: flex; flex-direction: column; gap: 10px; max-height: 280px; overflow: auto; }
.preview-item { display: flex; gap: 10px; align-items: center; }
.p-thumb { width: 56px; height: 56px; border-radius: 6px; flex-shrink: 0; object-fit: cover; }
.p-title { font-size: 13px; color: #303133; }
.p-url { font-size: 12px; color: #909399; word-break: break-all; }
</style>
