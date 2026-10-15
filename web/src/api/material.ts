import client, { unwrap, type PageData } from './client'

export interface CategoryNode {
  id: number
  parentId: number
  name: string
  code: string
  sort: number
  enabled: number
  children?: CategoryNode[]
}

export interface Material {
  id: number
  categoryId: number
  title: string
  mediaType: 'image' | 'video'
  url: string
  coverUrl: string
  fileName: string
  mime: string
  sizeBytes: number
  durationMs: number
  width: number
  height: number
  tags: string[]
  remark: string
  productId?: number | null
  productSn: string
  sort: number
  status: number
  createdBy: number
  createdAt: string
  updatedAt: string
}

export interface DashboardStats {
  materialCount: number
  imageCount: number
  videoCount: number
  weekNewCount: number
  categoryCount: number
  byCategory: { categoryId: number; categoryName: string; count: number }[]
}

export interface MaterialListQuery {
  categoryId?: number
  includeChildren?: boolean
  mediaType?: string
  keyword?: string
  productSn?: string
  hasProduct?: string
  tag?: string
  page?: number
  pageSize?: number
}

export async function fetchDashboardStats(): Promise<DashboardStats> {
  const res = await client.get('/dashboard/stats')
  return unwrap(res)
}

export async function listCategories(): Promise<CategoryNode[]> {
  const res = await client.get('/categories')
  return unwrap(res) || []
}

export async function createCategory(body: {
  parentId?: number
  name: string
  sort?: number
  enabled?: number
}): Promise<CategoryNode> {
  const res = await client.post('/categories', body)
  return unwrap(res)
}

export async function updateCategory(
  id: number,
  body: Partial<{ parentId: number; name: string; sort: number; enabled: number }>,
): Promise<CategoryNode> {
  const res = await client.put(`/categories/${id}`, body)
  return unwrap(res)
}

export async function deleteCategory(id: number): Promise<void> {
  await client.delete(`/categories/${id}`)
}

export async function reorderCategories(
  items: { id: number; parentId: number; sort: number }[],
): Promise<void> {
  await client.put('/categories/reorder', { items })
}

export async function listMaterials(query: MaterialListQuery): Promise<PageData<Material>> {
  const res = await client.get('/materials', { params: query })
  return unwrap(res)
}

export async function getMaterial(id: number): Promise<Material> {
  const res = await client.get(`/materials/${id}`)
  return unwrap(res)
}

export async function createMaterial(body: {
  categoryId?: number
  title: string
  mediaType: string
  url: string
  coverUrl?: string
  fileName?: string
  mime?: string
  sizeBytes?: number
  tags?: string[]
  remark?: string
  productId?: number | null
  productSn?: string
}): Promise<Material> {
  const res = await client.post('/materials', body)
  return unwrap(res)
}

export async function updateMaterial(
  id: number,
  body: Partial<{
    categoryId: number
    title: string
    url: string
    coverUrl: string
    fileName: string
    mime: string
    sizeBytes: number
    width: number
    height: number
    tags: string[]
    remark: string
    productId: number | null
    clearProductId: boolean
    productSn: string
    sort: number
  }>,
): Promise<Material> {
  const res = await client.put(`/materials/${id}`, body)
  return unwrap(res)
}

/** 旋转并重写图片像素（degrees: 0=EXIF纠正，90/-90/180） */
export async function rotateMaterial(id: number, degrees: number): Promise<Material> {
  const res = await client.post(`/materials/${id}/rotate`, { degrees }, { timeout: 120000 })
  return unwrap(res)
}

export type WatermarkPosition =
  | 'top-left'
  | 'top-right'
  | 'bottom-left'
  | 'bottom-right'
  | 'center'
  | 'top-center'
  | 'bottom-center'
  | 'custom'

export interface WatermarkPayload {
  ids: number[]
  logoUrl: string
  position: WatermarkPosition
  offsetX?: number
  offsetY?: number
  margin?: number
  scale?: number
  opacity?: number
}

export interface WatermarkResult {
  ok: number
  fail: number
  success: Material[]
  failed: { id: number; error: string }[]
}

/** 批量/单张叠加水印（文字请先渲染为透明 PNG 再作为 logoUrl） */
export async function watermarkMaterials(body: WatermarkPayload): Promise<WatermarkResult> {
  const res = await client.post('/materials/watermark', body, { timeout: 300000 })
  return unwrap(res)
}

export async function deleteMaterial(id: number): Promise<void> {
  await client.delete(`/materials/${id}`)
}

export async function batchDeleteMaterials(ids: number[]): Promise<void> {
  await client.post('/materials/batch-delete', { ids })
}

/** flatten category tree for select options */
export function flattenCategories(
  nodes: CategoryNode[],
  depth = 0,
): { id: number; name: string; code: string; depth: number; disabled?: boolean }[] {
  const out: { id: number; name: string; code: string; depth: number }[] = []
  for (const n of nodes) {
    out.push({ id: n.id, name: n.name, code: n.code, depth })
    if (n.children?.length) {
      out.push(...flattenCategories(n.children, depth + 1))
    }
  }
  return out
}
