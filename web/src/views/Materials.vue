<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Edit, Search, Iphone, Download, RefreshRight, RefreshLeft, Stamp } from '@element-plus/icons-vue'
import MediaUploadField from '../components/MediaUploadField.vue'
import WatermarkDialog from '../components/WatermarkDialog.vue'
import type { MediaItem } from '../api/upload'
import {
  listCategories,
  listMaterials,
  createMaterial,
  updateMaterial,
  deleteMaterial,
  batchDeleteMaterials,
  rotateMaterial,
  ensureMaterialThumbs,
  materialListSrc,
  flattenCategories,
  type CategoryNode,
  type Material,
} from '../api/material'

const loading = ref(false)
const categories = ref<CategoryNode[]>([])
const selectedCatId = ref<number | undefined>()
const materials = ref<Material[]>([])
const total = ref(0)
const selectedIds = ref<number[]>([])

const query = reactive({
  keyword: '',
  mediaType: '',
  hasProduct: '',
  page: 1,
  pageSize: 24,
})

const uploadVisible = ref(false)
const uploadMedia = ref<MediaItem[]>([])
const uploadForm = reactive({
  categoryId: 0 as number,
  titlePrefix: '',
  tags: '' as string,
  remark: '',
  productSn: '',
})

const editVisible = ref(false)
const editForm = reactive({
  id: 0,
  title: '',
  categoryId: 0,
  tags: '' as string,
  remark: '',
  productSn: '',
  coverUrl: '',
})

const rotateVisible = ref(false)
const rotateSaving = ref(false)
const rotateTarget = ref<Material | null>(null)
/** 预览累计角度（顺时针为正，与 CSS rotate 一致） */
const rotateDeg = ref(0)

const watermarkVisible = ref(false)
const watermarkTargets = ref<Material[]>([])

const selectedImageCount = computed(
  () => materials.value.filter((m) => selectedIds.value.includes(m.id) && m.mediaType === 'image').length,
)

const flatCats = computed(() => flattenCategories(categories.value))

const treeData = computed(() => [
  { id: 0, label: '全部分类', children: mapTree(categories.value) },
])

function mapTree(nodes: CategoryNode[]): { id: number; label: string; children?: any[] }[] {
  return nodes.map((n) => ({
    id: n.id,
    label: n.name,
    children: n.children?.length ? mapTree(n.children) : undefined,
  }))
}

async function loadCategories() {
  categories.value = await listCategories()
}

async function loadMaterials() {
  loading.value = true
  try {
    const res = await listMaterials({
      categoryId: selectedCatId.value || undefined,
      includeChildren: true,
      keyword: query.keyword || undefined,
      mediaType: query.mediaType || undefined,
      hasProduct: query.hasProduct || undefined,
      page: query.page,
      pageSize: query.pageSize,
    })
    materials.value = res.list || []
    total.value = res.total || 0
    void ensureThumbsForPage()
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

/** 当前页缺缩略图的图片后台补齐，完成后刷新列表用小图 */
async function ensureThumbsForPage() {
  const need = materials.value
    .filter((m) => m.mediaType === 'image' && !m.thumbUrl)
    .map((m) => m.id)
  if (!need.length) return
  try {
    const r = await ensureMaterialThumbs(need)
    if (r.ok > 0) {
      const res = await listMaterials({
        categoryId: selectedCatId.value || undefined,
        includeChildren: true,
        keyword: query.keyword || undefined,
        mediaType: query.mediaType || undefined,
        hasProduct: query.hasProduct || undefined,
        page: query.page,
        pageSize: query.pageSize,
      })
      materials.value = res.list || []
      total.value = res.total || 0
    }
  } catch {
    // 缩略图失败不影响浏览（仍显示原图）
  }
}

function onTreeClick(data: { id: number }) {
  selectedCatId.value = data.id || undefined
  query.page = 1
  loadMaterials()
}

function openUpload() {
  uploadMedia.value = []
  uploadForm.categoryId = selectedCatId.value || flatCats.value[0]?.id || 0
  uploadForm.titlePrefix = ''
  uploadForm.tags = ''
  uploadForm.remark = ''
  uploadForm.productSn = ''
  uploadVisible.value = true
}

async function submitUpload() {
  if (!uploadMedia.value.length) {
    ElMessage.warning('请先上传文件')
    return
  }
  try {
    const tags = uploadForm.tags
      .split(/[,，\s]+/)
      .map((t) => t.trim())
      .filter(Boolean)
    for (let i = 0; i < uploadMedia.value.length; i++) {
      const m = uploadMedia.value[i]
      const title =
        uploadForm.titlePrefix ||
        (uploadMedia.value.length === 1 ? '素材' : `素材 ${i + 1}`)
      await createMaterial({
        categoryId: uploadForm.categoryId || undefined,
        title: uploadMedia.value.length > 1 && uploadForm.titlePrefix
          ? `${uploadForm.titlePrefix} ${i + 1}`
          : title,
        mediaType: m.mediaType,
        url: m.url,
        tags,
        remark: uploadForm.remark,
        productSn: uploadForm.productSn,
      })
    }
    ElMessage.success(`已入库 ${uploadMedia.value.length} 个素材`)
    uploadVisible.value = false
    await loadMaterials()
  } catch (e) {
    ElMessage.error((e as Error).message || '保存失败')
  }
}

function openEdit(m: Material) {
  editForm.id = m.id
  editForm.title = m.title
  editForm.categoryId = m.categoryId
  editForm.tags = (m.tags || []).join(', ')
  editForm.remark = m.remark || ''
  editForm.productSn = m.productSn || ''
  editForm.coverUrl = m.coverUrl || ''
  editVisible.value = true
}

async function submitEdit() {
  try {
    const tags = editForm.tags
      .split(/[,，\s]+/)
      .map((t) => t.trim())
      .filter(Boolean)
    await updateMaterial(editForm.id, {
      title: editForm.title,
      categoryId: editForm.categoryId,
      tags,
      remark: editForm.remark,
      productSn: editForm.productSn,
      coverUrl: editForm.coverUrl,
    })
    ElMessage.success('已保存')
    editVisible.value = false
    await loadMaterials()
  } catch (e) {
    ElMessage.error((e as Error).message || '保存失败')
  }
}

function openRotate(m: Material) {
  if (m.mediaType !== 'image') {
    ElMessage.warning('仅图片支持旋转')
    return
  }
  rotateTarget.value = m
  rotateDeg.value = 0
  rotateVisible.value = true
}

function nudgeRotate(delta: number) {
  rotateDeg.value = ((rotateDeg.value + delta) % 360 + 360) % 360
}

async function submitRotate() {
  const m = rotateTarget.value
  if (!m) return
  rotateSaving.value = true
  try {
    // 0 = 仅按 EXIF 纠正写入像素；90/180/270 = 再旋转
    const degrees = rotateDeg.value
    await rotateMaterial(m.id, degrees)
    ElMessage.success(degrees === 0 ? '已按 EXIF 纠正并保存' : '已旋转并保存')
    rotateVisible.value = false
    await loadMaterials()
  } catch (e) {
    ElMessage.error((e as Error).message || '旋转保存失败')
  } finally {
    rotateSaving.value = false
  }
}

function openWatermark(targets?: Material[]) {
  const list = (targets && targets.length
    ? targets
    : materials.value.filter((m) => selectedIds.value.includes(m.id))
  ).filter((m) => m.mediaType === 'image')
  if (!list.length) {
    ElMessage.warning('请先选择至少一张图片')
    return
  }
  watermarkTargets.value = list
  watermarkVisible.value = true
}

async function removeOne(m: Material) {
  await ElMessageBox.confirm(`确定删除「${m.title}」？`, '删除素材', { type: 'warning' })
  await deleteMaterial(m.id)
  ElMessage.success('已删除')
  await loadMaterials()
}

async function removeSelected() {
  if (!selectedIds.value.length) return
  await ElMessageBox.confirm(`确定删除选中的 ${selectedIds.value.length} 个素材？`, '批量删除', {
    type: 'warning',
  })
  await batchDeleteMaterials(selectedIds.value)
  selectedIds.value = []
  ElMessage.success('已删除')
  await loadMaterials()
}

function toggleSelect(id: number) {
  const i = selectedIds.value.indexOf(id)
  if (i >= 0) selectedIds.value.splice(i, 1)
  else selectedIds.value.push(id)
}

function isSelected(id: number) {
  return selectedIds.value.includes(id)
}

function copyUrl(url: string) {
  navigator.clipboard.writeText(url).then(
    () => ElMessage.success('链接已复制'),
    () => ElMessage.error('复制失败'),
  )
}

function guessFilename(m: Material): string {
  if (m.fileName?.trim()) return m.fileName.trim()
  try {
    const path = new URL(m.url).pathname
    const base = path.split('/').pop()
    if (base) return decodeURIComponent(base)
  } catch {
    /* ignore */
  }
  const ext = m.mediaType === 'video' ? 'mp4' : 'jpg'
  return `${m.title || 'material'}_${m.id}.${ext}`
}

async function downloadOne(m: Material) {
  const name = guessFilename(m)
  try {
    const res = await fetch(m.url)
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const blob = await res.blob()
    const href = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = href
    a.download = name
    a.rel = 'noopener'
    document.body.appendChild(a)
    a.click()
    a.remove()
    URL.revokeObjectURL(href)
    ElMessage.success('已开始下载')
  } catch {
    // 跨域无法 blob 时回退为新窗口打开，由浏览器自行保存
    const a = document.createElement('a')
    a.href = m.url
    a.download = name
    a.target = '_blank'
    a.rel = 'noopener'
    document.body.appendChild(a)
    a.click()
    a.remove()
    ElMessage.success('已打开文件，请另存为')
  }
}

async function downloadSelected() {
  if (!selectedIds.value.length) return
  const list = materials.value.filter((m) => selectedIds.value.includes(m.id))
  for (const m of list) {
    await downloadOne(m)
  }
}

watch(
  () => [query.keyword, query.mediaType, query.hasProduct],
  () => {
    query.page = 1
  },
)

onMounted(async () => {
  try {
    await loadCategories()
    await loadMaterials()
  } catch (e) {
    ElMessage.error((e as Error).message || '初始化失败')
  }
})
</script>

<template>
  <div class="page">
    <div class="left">
      <div class="left-hd">分类</div>
      <el-tree
        :data="treeData"
        node-key="id"
        default-expand-all
        highlight-current
        :props="{ label: 'label', children: 'children' }"
        @node-click="onTreeClick"
      />
    </div>

    <div class="right">
      <div class="toolbar">
        <el-input
          v-model="query.keyword"
          clearable
          placeholder="搜索标题/备注/资料编码"
          style="width: 220px"
          :prefix-icon="Search"
          @keyup.enter="loadMaterials"
        />
        <el-select v-model="query.mediaType" clearable placeholder="类型" style="width: 110px" @change="loadMaterials">
          <el-option label="图片" value="image" />
          <el-option label="视频" value="video" />
        </el-select>
        <el-select v-model="query.hasProduct" clearable placeholder="商品关联" style="width: 120px" @change="loadMaterials">
          <el-option label="已关联" value="yes" />
          <el-option label="未关联" value="no" />
        </el-select>
        <el-button type="primary" @click="loadMaterials">查询</el-button>
        <div class="spacer" />
        <el-button :disabled="!selectedIds.length" plain :icon="Download" @click="downloadSelected">
          批量下载
        </el-button>
        <el-button
          :disabled="!selectedImageCount"
          plain
          type="warning"
          :icon="Stamp"
          @click="openWatermark()"
        >
          批量水印{{ selectedImageCount ? ` (${selectedImageCount})` : '' }}
        </el-button>
        <el-button :disabled="!selectedIds.length" type="danger" plain :icon="Delete" @click="removeSelected">
          批量删除
        </el-button>
        <el-button type="primary" :icon="Plus" @click="openUpload">上传素材</el-button>
      </div>

      <div v-loading="loading" class="grid">
        <div
          v-for="m in materials"
          :key="m.id"
          class="card"
          :class="{ selected: isSelected(m.id) }"
          @click="toggleSelect(m.id)"
        >
          <div class="thumb-wrap">
            <el-image
              v-if="m.mediaType === 'image'"
              :src="materialListSrc(m)"
              fit="cover"
              class="thumb"
              lazy
              :preview-src-list="[m.url]"
              preview-teleported
              @click.stop
            />
            <a v-else :href="m.url" target="_blank" class="thumb video" @click.stop>
              <span>▶ 视频</span>
            </a>
            <el-tag size="small" class="type-tag" :type="m.mediaType === 'video' ? 'warning' : 'info'">
              {{ m.mediaType === 'video' ? '视频' : '图片' }}
            </el-tag>
          </div>
          <div class="meta">
            <div class="title" :title="m.title">{{ m.title }}</div>
            <div class="sub">{{ m.productSn || '未关联商品' }}</div>
          </div>
          <div class="ops" @click.stop>
            <el-button text type="primary" size="small" @click="copyUrl(m.url)">复制链接</el-button>
            <el-button text type="primary" size="small" :icon="Download" @click="downloadOne(m)">下载</el-button>
            <el-button
              v-if="m.mediaType === 'image'"
              text
              type="primary"
              size="small"
              :icon="RefreshRight"
              @click="openRotate(m)"
            >
              旋转
            </el-button>
            <el-button
              v-if="m.mediaType === 'image'"
              text
              type="warning"
              size="small"
              :icon="Stamp"
              @click="openWatermark([m])"
            >
              水印
            </el-button>
            <el-button text size="small" :icon="Edit" @click="openEdit(m)" />
            <el-button text type="danger" size="small" :icon="Delete" @click="removeOne(m)" />
          </div>
        </div>
        <el-empty v-if="!loading && !materials.length" description="暂无素材，点击右上角上传" />
      </div>

      <div class="pager">
        <el-pagination
          v-model:current-page="query.page"
          v-model:page-size="query.pageSize"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="loadMaterials"
        />
      </div>
    </div>

    <el-dialog v-model="uploadVisible" title="上传素材" width="560px" destroy-on-close>
      <el-form label-width="88px">
        <el-form-item label="分类">
          <el-select v-model="uploadForm.categoryId" style="width: 100%">
            <el-option
              v-for="c in flatCats"
              :key="c.id"
              :label="'　'.repeat(c.depth) + c.name"
              :value="c.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="标题前缀">
          <el-input v-model="uploadForm.titlePrefix" placeholder="可选；多文件时自动加序号" />
        </el-form-item>
        <el-form-item label="标签">
          <el-input v-model="uploadForm.tags" placeholder="逗号分隔，如：询盘,白底" />
        </el-form-item>
        <el-form-item label="资料编码">
          <el-input v-model="uploadForm.productSn" placeholder="可选弱关联" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="uploadForm.remark" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="文件">
          <MediaUploadField v-model="uploadMedia" subdir="materials" :max-count="20" />
          <div class="hint"><el-icon><Iphone /></el-icon> 支持本机与手机扫码上传图片/视频</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="uploadVisible = false">取消</el-button>
        <el-button type="primary" @click="submitUpload">入库</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="editVisible" title="编辑素材" width="480px" destroy-on-close>
      <el-form label-width="88px">
        <el-form-item label="标题">
          <el-input v-model="editForm.title" />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="editForm.categoryId" style="width: 100%">
            <el-option
              v-for="c in flatCats"
              :key="c.id"
              :label="'　'.repeat(c.depth) + c.name"
              :value="c.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="标签">
          <el-input v-model="editForm.tags" placeholder="逗号分隔" />
        </el-form-item>
        <el-form-item label="资料编码">
          <el-input v-model="editForm.productSn" />
        </el-form-item>
        <el-form-item label="封面 URL">
          <el-input v-model="editForm.coverUrl" placeholder="视频封面可选" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="editForm.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" @click="submitEdit">保存</el-button>
      </template>
    </el-dialog>

    <WatermarkDialog
      v-model="watermarkVisible"
      :materials="watermarkTargets"
      @done="loadMaterials"
    />

    <el-dialog
      v-model="rotateVisible"
      title="旋转图片"
      width="560px"
      destroy-on-close
      :close-on-click-modal="!rotateSaving"
    >
      <p class="rotate-hint">
        保存时会按当前角度重写图片像素并去掉错误的 EXIF 方向信息，各端显示一致。
        若不点旋转直接保存，则仅按 EXIF 纠正方向。
      </p>
      <div class="rotate-preview">
        <img
          v-if="rotateTarget"
          :src="rotateTarget.url"
          :style="{ transform: `rotate(${rotateDeg}deg)` }"
          alt="预览"
        />
      </div>
      <div class="rotate-actions">
        <el-button :icon="RefreshLeft" :disabled="rotateSaving" @click="nudgeRotate(-90)">左转 90°</el-button>
        <el-button :icon="RefreshRight" :disabled="rotateSaving" @click="nudgeRotate(90)">右转 90°</el-button>
        <el-button :disabled="rotateSaving || rotateDeg === 0" @click="rotateDeg = 0">重置预览</el-button>
      </div>
      <p class="rotate-deg">当前角度：{{ rotateDeg }}°</p>
      <template #footer>
        <el-button :disabled="rotateSaving" @click="rotateVisible = false">取消</el-button>
        <el-button type="primary" :loading="rotateSaving" @click="submitRotate">保存到素材</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page { display: flex; gap: 16px; min-height: calc(100vh - 100px); }
.left {
  width: 220px; flex-shrink: 0; background: #fff; border-radius: 8px;
  padding: 12px; border: 1px solid #ebeef5;
}
.left-hd { font-weight: 600; margin-bottom: 8px; color: #303133; }
.right { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 12px; }
.toolbar { display: flex; flex-wrap: wrap; gap: 8px; align-items: center; }
.spacer { flex: 1; }
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 12px;
  min-height: 200px;
}
.card {
  background: #fff; border: 1px solid #ebeef5; border-radius: 8px; overflow: hidden;
  cursor: pointer; transition: box-shadow .15s, border-color .15s;
}
.card:hover { box-shadow: 0 2px 12px rgba(0,0,0,.06); }
.card.selected { border-color: #409eff; box-shadow: 0 0 0 1px #409eff inset; }
.thumb-wrap { position: relative; aspect-ratio: 1; background: #f5f7fa; }
.thumb { width: 100%; height: 100%; display: block; }
a.thumb.video {
  display: flex; align-items: center; justify-content: center;
  width: 100%; height: 100%; text-decoration: none; color: #409eff; font-weight: 600;
  background: #1a1a1a;
}
.type-tag { position: absolute; left: 6px; top: 6px; }
.meta { padding: 8px 10px 0; }
.title {
  font-size: 13px; color: #303133; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.rotate-hint {
  margin: 0 0 12px;
  font-size: 13px;
  color: #909399;
  line-height: 1.5;
}
.rotate-preview {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 280px;
  max-height: 420px;
  background: #1a1a1a;
  border-radius: 8px;
  overflow: hidden;
}
.rotate-preview img {
  max-width: 92%;
  max-height: 400px;
  object-fit: contain;
  transition: transform 0.2s ease;
}
.rotate-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: center;
  margin-top: 14px;
}
.rotate-deg {
  text-align: center;
  margin: 10px 0 0;
  font-size: 13px;
  color: #606266;
}
.sub { font-size: 12px; color: #909399; margin-top: 2px; }
.ops {
  display: flex; align-items: center; flex-wrap: wrap; gap: 0;
  padding: 4px 4px 8px;
}
.pager { display: flex; justify-content: flex-end; }
.hint { margin-top: 6px; font-size: 12px; color: #909399; display: flex; align-items: center; gap: 4px; }
</style>
