/** 用浏览器字体把文字画成透明 PNG，供后端当 Logo 合成（支持中文） */

export interface TextWatermarkOpts {
  text: string
  fontSize?: number
  color?: string
  fontFamily?: string
  bold?: boolean
}

export async function renderTextWatermarkPng(opts: TextWatermarkOpts): Promise<Blob> {
  const text = opts.text.trim()
  if (!text) throw new Error('请输入水印文字')
  const fontSize = opts.fontSize || 48
  const color = opts.color || 'rgba(255,255,255,0.95)'
  const family = opts.fontFamily || 'PingFang SC, Microsoft YaHei, sans-serif'
  const weight = opts.bold ? '700' : '600'
  const font = `${weight} ${fontSize}px ${family}`

  const measure = document.createElement('canvas')
  const mctx = measure.getContext('2d')
  if (!mctx) throw new Error('canvas unavailable')
  mctx.font = font
  const lines = text.split(/\n/).map((s) => s.trim()).filter(Boolean)
  if (!lines.length) throw new Error('请输入水印文字')
  const lineH = Math.ceil(fontSize * 1.35)
  let maxW = 0
  for (const line of lines) {
    maxW = Math.max(maxW, mctx.measureText(line).width)
  }
  const padX = Math.ceil(fontSize * 0.35)
  const padY = Math.ceil(fontSize * 0.3)
  const canvas = document.createElement('canvas')
  canvas.width = Math.max(8, Math.ceil(maxW) + padX * 2)
  canvas.height = Math.max(8, lineH * lines.length + padY * 2)
  const ctx = canvas.getContext('2d')
  if (!ctx) throw new Error('canvas unavailable')
  ctx.clearRect(0, 0, canvas.width, canvas.height)
  ctx.font = font
  ctx.fillStyle = color
  ctx.textBaseline = 'top'
  // 浅描边提升浅色底上的可读性
  ctx.strokeStyle = 'rgba(0,0,0,0.35)'
  ctx.lineWidth = Math.max(1, Math.round(fontSize / 18))
  lines.forEach((line, i) => {
    const x = padX
    const y = padY + i * lineH
    ctx.strokeText(line, x, y)
    ctx.fillText(line, x, y)
  })
  const blob = await new Promise<Blob | null>((resolve) => canvas.toBlob(resolve, 'image/png'))
  if (!blob) throw new Error('生成文字水印失败')
  return blob
}
