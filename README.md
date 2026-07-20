# MaterialCore — OSMS 素材中心

MaterialCore 是 **OSMS** 平台下的素材中心：管理商品宣传用图片/视频（询盘快发、预期效果、宣传海报、带货短视频等），支持**自定义分类**与**手机扫码上传**。

与 ProductCore 商品编辑里的 PIM 图文（主图 / `media_json` 白底图等）分离：本应用服务销售询盘「找得到、发得出」，**不强制关联商品**（可选填资料编码作弱关联）。

| 项 | 值 |
|----|-----|
| Go module | `materialcore` |
| API | `:8101` |
| Web | `:5185` |
| Docker 镜像 | `materialcore-api`、`materialcore-web` |
| UserCore app | `materialcore`（`material:read` / `material:write`） |
| 存储 | 本地 `./data/uploads` 或平台 MinIO bucket `materialcore` |
| 端口约定 | [deploy/docs/PORTS.md](../deploy/docs/PORTS.md) |
| 平台编排 | `/home/asialeaf/projects/deploy` |

## 本地开发

```bash
cp configs/config.example.yaml configs/config.yaml
# 编辑 postgres_dsn / jwt_secret；开发可用 sqlite（见 configs/config.yaml 示例）

go run ./cmd/api -config configs/config.yaml

cd web && npm i && npm run dev
```

默认分类种子（租户首次访问分类列表时写入）：询盘快发 / 预期效果 / 宣传海报 / 带货短视频 / 未分类（系统保留，不可删）。

## 功能概览

- **素材库**：图片/视频上传（本机 + 扫码）、宫格浏览、筛选、编辑、删除
- **分类管理**：树形自定义分类
- **询盘工作台**：多选 → 复制链接 / 打包 ZIP 下载
- **手机扫码上传**：管理端生成二维码 → 手机打开 `/m/photo-upload?token=…`

## API

- `GET /health`
- Admin（JWT）：`/api/v1/admin/categories`、`/materials`、`/upload`、`/photo-upload-sessions`、`/materials/export-zip`、`/dashboard/stats`
- Mobile（免登录短时 token）：`GET|POST /api/v1/mobile/photo-upload/:token`
