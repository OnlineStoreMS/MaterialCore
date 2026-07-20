<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Edit } from '@element-plus/icons-vue'
import {
  listCategories,
  createCategory,
  updateCategory,
  deleteCategory,
  type CategoryNode,
} from '../api/material'

const loading = ref(false)
const categories = ref<CategoryNode[]>([])

const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const form = reactive({
  id: 0,
  parentId: 0,
  name: '',
  sort: 0,
  enabled: 1,
})

const treeData = computed(() => mapTree(categories.value))

function mapTree(nodes: CategoryNode[]): any[] {
  return nodes.map((n) => ({
    ...n,
    label: n.name,
    children: n.children?.length ? mapTree(n.children) : undefined,
  }))
}

async function load() {
  loading.value = true
  try {
    categories.value = await listCategories()
  } catch (e) {
    ElMessage.error((e as Error).message || '加载失败')
  } finally {
    loading.value = false
  }
}

function openCreate(parentId = 0) {
  dialogMode.value = 'create'
  form.id = 0
  form.parentId = parentId
  form.name = ''
  form.sort = 0
  form.enabled = 1
  dialogVisible.value = true
}

function openEdit(node: CategoryNode) {
  dialogMode.value = 'edit'
  form.id = node.id
  form.parentId = node.parentId
  form.name = node.name
  form.sort = node.sort
  form.enabled = node.enabled
  dialogVisible.value = true
}

async function submit() {
  if (!form.name.trim()) {
    ElMessage.warning('请输入分类名称')
    return
  }
  try {
    if (dialogMode.value === 'create') {
      await createCategory({
        parentId: form.parentId,
        name: form.name.trim(),
        sort: form.sort,
        enabled: form.enabled,
      })
      ElMessage.success('已创建')
    } else {
      await updateCategory(form.id, {
        parentId: form.parentId,
        name: form.name.trim(),
        sort: form.sort,
        enabled: form.enabled,
      })
      ElMessage.success('已保存')
    }
    dialogVisible.value = false
    await load()
  } catch (e) {
    ElMessage.error((e as Error).message || '操作失败')
  }
}

async function remove(node: CategoryNode) {
  if (node.code === 'uncategorized') {
    ElMessage.warning('系统分类「未分类」不可删除')
    return
  }
  await ElMessageBox.confirm(
    `确定删除分类「${node.name}」？若有子分类或素材将无法删除。`,
    '删除分类',
    { type: 'warning' },
  )
  try {
    await deleteCategory(node.id)
    ElMessage.success('已删除')
    await load()
  } catch (e) {
    ElMessage.error((e as Error).message || '删除失败')
  }
}

onMounted(load)
</script>

<template>
  <div v-loading="loading" class="page">
    <div class="hd">
      <div>
        <h2>分类管理</h2>
        <p>自定义树形分类；预置询盘快发 / 预期效果 / 宣传海报 / 带货短视频 / 未分类，均可按需调整。</p>
      </div>
      <el-button type="primary" :icon="Plus" @click="openCreate(0)">新建根分类</el-button>
    </div>

    <el-card shadow="never">
      <el-tree
        :data="treeData"
        node-key="id"
        default-expand-all
        :props="{ label: 'label', children: 'children' }"
      >
        <template #default="{ data }">
          <div class="node">
            <span class="name">
              {{ data.name }}
              <el-tag v-if="data.code === 'uncategorized'" size="small" type="info">系统</el-tag>
              <el-tag v-else-if="data.code" size="small" type="success">{{ data.code }}</el-tag>
            </span>
            <span class="acts">
              <el-button text type="primary" size="small" :icon="Plus" @click.stop="openCreate(data.id)">
                子分类
              </el-button>
              <el-button text size="small" :icon="Edit" @click.stop="openEdit(data)" />
              <el-button
                text
                type="danger"
                size="small"
                :icon="Delete"
                :disabled="data.code === 'uncategorized'"
                @click.stop="remove(data)"
              />
            </span>
          </div>
        </template>
      </el-tree>
      <el-empty v-if="!loading && !categories.length" description="暂无分类" />
    </el-card>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新建分类' : '编辑分类'"
      width="420px"
      destroy-on-close
    >
      <el-form label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="form.name" maxlength="64" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" :max="9999" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.enabled" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page { max-width: 900px; }
.hd {
  display: flex; justify-content: space-between; align-items: flex-start;
  margin-bottom: 16px; gap: 16px;
}
.hd h2 { margin: 0 0 6px; font-size: 20px; }
.hd p { margin: 0; color: #909399; font-size: 13px; line-height: 1.5; }
.node {
  flex: 1; display: flex; align-items: center; justify-content: space-between;
  padding-right: 8px; gap: 12px;
}
.name { display: inline-flex; align-items: center; gap: 6px; }
.acts { opacity: 0.85; }
</style>
