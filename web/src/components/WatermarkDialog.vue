<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import MediaUploadField from './MediaUploadField.vue'
import type { MediaItem } from '../api/upload'
import { uploadMedia } from '../api/upload'
import {
  watermarkMaterials,
  type Material,
  type WatermarkPosition,
} from '../api/material'
import { renderTextWatermarkPng } from '../utils/textWatermark'

const props = defineProps<{
  modelValue: boolean
  materials: Material[]
}>()

const emit = defineEmits<{
  'update:modelValue': [boolean]
  done: []
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (v: boolean) => emit('update:modelValue', v),
})

const saving = ref(false)
const mode = ref<'text' | 'logo'>('text')
const logoMedia = ref<MediaItem[]>([])
const form = reactive({
  text: '素材中心',
  fontSize: 48,
  color: '#ffffff',
  bold: true,
  position: 'bottom-right' as WatermarkPosition,
  scale: 18, // %
  opacity: 70, // %
  margin: 3, // %
  offsetX: 70,
  offsetY: 80,
})

const imageTargets = computed(() => props.materials.filter((m) => m.mediaType === 'image'))
const preview = computed(() => imageTargets.value[0] || null)

const positionOptions: { value: WatermarkPosition; label: string }[] = [
  { value: 'top-left', label: '左上' },
  { value: 'top-center', label: '上中' },
  { value: 'top-right', label: '右上' },
  { value: 'center', label: '居中' },
  { value: 'bottom-left', label: '左下' },
  { value: 'bottom-center', label: '下中' },
  { value: 'bottom-right', label: '右下' },
  { value: 'custom', label: '自定义拖拽' },
]

const previewBox = ref<HTMLElement | null>(null)
const dragging = ref(false)

const overlayStyle = computed(() => {
  const scalePct = form.scale
  const base: Record<string, string> = {
    width: `${scalePct}%`,
    opacity: String(form.opacity / 100),
    maxWidth: '80%',
  }
  if (form.position === 'custom') {
    return {
      ...base,
      left: `${form.offsetX}%`,
      top: `${form.offsetY}%`,
      transform: 'none',
    }
  }
  const m = `${form.margin}%`
  switch (form.position) {
    case 'top-left':
      return { ...base, left: m, top: m }
    case 'top-right':
      return { ...base, right: m, top: m, left: 'auto' }
    case 'bottom-left':
      return { ...base, left: m, bottom: m, top: 'auto' }
    case 'bottom-right':
      return { ...base, right: m, bottom: m, top: 'auto', left: 'auto' }
    case 'top-center':
      return { ...base, left: '50%', top: m, transform: 'translateX(-50%)' }
    case 'bottom-center':
      return { ...base, left: '50%', bottom: m, top: 'auto', transform: 'translateX(-50%)' }
    case 'center':
      return { ...base, left: '50%', top: '50%', transform: 'translate(-50%, -50%)' }
    default:
      return { ...base, right: m, bottom: m, top: 'auto', left: 'auto' }
  }
})

const overlaySrc = computed(() => {
  if (mode.value === 'logo' && logoMedia.value[0]?.url) return logoMedia.value[0].url
  return ''
})

const textPreview = computed(() => form.text.trim() || '水印预览')

watch(visible, (v) => {
  if (v) {
    logoMedia.value = []
    if (form.position !== 'custom') {
      // keep
    }
  }
})

function onPreset(pos: WatermarkPosition) {
  form.position = pos
}

function onPointerDown(e: PointerEvent) {
  if (form.position !== 'custom') {
    form.position = 'custom'
  }
  dragging.value = true
  ;(e.currentTarget as HTMLElement).setPointerCapture(e.pointerId)
  moveTo(e)
}

function onPointerMove(e: PointerEvent) {
  if (!dragging.value) return
  moveTo(e)
}

function onPointerUp() {
  dragging.value = false
}

function moveTo(e: PointerEvent) {
  const box = previewBox.value
  if (!box) return
  const rect = box.getBoundingClientRect()
  const x = ((e.clientX - rect.left) / rect.width) * 100
  const y = ((e.clientY - rect.top) / rect.height) * 100
  form.offsetX = Math.min(95, Math.max(0, x))
  form.offsetY = Math.min(95, Math.max(0, y))
}

async function resolveLogoUrl(): Promise<string> {
  if (mode.value === 'logo') {
    const url = logoMedia.value[0]?.url
    if (!url) throw new Error('请上传 Logo 图片')
    return url
  }
  const blob = await renderTextWatermarkPng({
    text: form.text,
    fontSize: form.fontSize,
    color: form.color,
    bold: form.bold,
  })
  const file = new File([blob], `wm-text-${Date.now()}.png`, { type: 'image/png' })
  const up = await uploadMedia(file, 'watermarks')
  return up.url
}

async function submit() {
  if (!imageTargets.value.length) {
    ElMessage.warning('请选择至少一张图片素材')
    return
  }
  saving.value = true
  try {
    const logoUrl = await resolveLogoUrl()
    const res = await watermarkMaterials({
      ids: imageTargets.value.map((m) => m.id),
      logoUrl,
      position: form.position,
      offsetX: form.offsetX,
      offsetY: form.offsetY,
      margin: form.margin / 100,
      scale: form.scale / 100,
      opacity: form.opacity / 100,
    })
    if (res.fail && res.failed?.length) {
      ElMessage.warning(`成功 ${res.ok}，失败 ${res.fail}：${res.failed[0]?.error || ''}`)
    } else {
      ElMessage.success(`已为 ${res.ok} 张图片加水印`)
    }
    visible.value = false
    emit('done')
  } catch (e) {
    ElMessage.error((e as Error).message || '加水印失败')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <el-dialog
    v-model="visible"
    title="图片水印"
    width="720px"
    destroy-on-close
    :close-on-click-modal="!saving"
  >
    <p class="hint">
      将对选中的 <b>{{ imageTargets.length }}</b> 张图片写入水印（像素级保存）。支持文字或 Logo；可固定角落/居中，或自定义拖拽位置。
    </p>

    <el-form label-width="88px">
      <el-form-item label="类型">
        <el-radio-group v-model="mode">
          <el-radio-button label="text">文字水印</el-radio-button>
          <el-radio-button label="logo">Logo 图片</el-radio-button>
        </el-radio-group>
      </el-form-item>

      <template v-if="mode === 'text'">
        <el-form-item label="文字">
          <el-input v-model="form.text" type="textarea" :rows="2" maxlength="80" show-word-limit />
        </el-form-item>
        <el-form-item label="字号">
          <el-slider v-model="form.fontSize" :min="24" :max="120" :step="2" show-input />
        </el-form-item>
        <el-form-item label="颜色">
          <el-color-picker v-model="form.color" show-alpha />
          <el-checkbox v-model="form.bold" style="margin-left: 12px">加粗</el-checkbox>
        </el-form-item>
      </template>

      <el-form-item v-else label="Logo">
        <MediaUploadField v-model="logoMedia" subdir="watermarks" :max-count="1" />
      </el-form-item>

      <el-form-item label="位置">
        <div class="pos-grid">
          <el-check-tag
            v-for="p in positionOptions"
            :key="p.value"
            :checked="form.position === p.value"
            @click="onPreset(p.value)"
          >
            {{ p.label }}
          </el-check-tag>
        </div>
      </el-form-item>

      <el-form-item label="大小">
        <el-slider v-model="form.scale" :min="5" :max="50" :step="1" show-input />
        <span class="unit">占图宽 %</span>
      </el-form-item>
      <el-form-item label="透明度">
        <el-slider v-model="form.opacity" :min="10" :max="100" :step="5" show-input />
      </el-form-item>
      <el-form-item v-if="form.position !== 'custom'" label="边距">
        <el-slider v-model="form.margin" :min="0" :max="12" :step="0.5" show-input />
        <span class="unit">%</span>
      </el-form-item>
    </el-form>

    <div class="preview-block">
      <div class="preview-label">预览（可拖拽水印到自定义位置）</div>
      <div
        ref="previewBox"
        class="preview-box"
        @pointerdown="onPointerDown"
        @pointermove="onPointerMove"
        @pointerup="onPointerUp"
        @pointercancel="onPointerUp"
      >
        <img v-if="preview" :src="preview.url" class="preview-img" alt="预览" draggable="false" />
        <div v-else class="preview-empty">请先选择图片素材</div>
        <div class="wm-layer" :style="overlayStyle">
          <img v-if="overlaySrc" :src="overlaySrc" class="wm-logo" alt="logo" draggable="false" />
          <div v-else class="wm-text" :style="{ color: form.color, fontWeight: form.bold ? 700 : 600 }">
            {{ textPreview }}
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <el-button :disabled="saving" @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="saving" @click="submit">应用到选中图片</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.hint {
  margin: 0 0 12px;
  font-size: 13px;
  color: #909399;
  line-height: 1.5;
}
.pos-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.unit {
  margin-left: 8px;
  color: #909399;
  font-size: 12px;
}
.preview-block { margin-top: 8px; }
.preview-label {
  font-size: 13px;
  color: #606266;
  margin-bottom: 8px;
}
.preview-box {
  position: relative;
  width: 100%;
  aspect-ratio: 16 / 10;
  background: #1a1a1a;
  border-radius: 8px;
  overflow: hidden;
  cursor: crosshair;
  user-select: none;
  touch-action: none;
}
.preview-img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  display: block;
  pointer-events: none;
}
.preview-empty {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #909399;
}
.wm-layer {
  position: absolute;
  pointer-events: none;
  line-height: 1.2;
}
.wm-logo {
  width: 100%;
  height: auto;
  display: block;
}
.wm-text {
  font-size: clamp(14px, 3.2vw, 28px);
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.45);
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
