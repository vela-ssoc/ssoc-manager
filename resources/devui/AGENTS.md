# DEVUI 前端项目说明（AI Agent 必读）

本文件是 `ssoc-manager/resources/devui` 前端项目的开发规范、风格与约束说明。
任何 AI Agent（或开发者）在改动本项目前，应先阅读本文件，保持与既有实现一致。

---

## 1. 项目定位与边界

- 本项目是一套 **全新 UI**，基于 **React + shadcn/ui**，用于 `ssoc-manager` 的管理控制台。
- 工作目录固定为 `ssoc-manager/resources/devui`。
- **只能修改工作目录内的文件**；其它目录（尤其后端 Go 代码、`resources` 下其它目录）严禁改动。
- 若发现后端 BUG 或建议，仅通过文字向用户反馈，禁止私自修改。

## 2. 技术栈与版本

- Vite 7 + React 19 + TypeScript 5.8
- Tailwind CSS v4（`@tailwindcss/vite` 插件，`@theme inline` 映射 CSS 变量）
- shadcn/ui 风格组件（自实现于 `src/components/ui/`，依赖 Radix UI）
- react-router-dom v7（`BrowserRouter` + `basename="/devui"`）
- monaco-editor + @monaco-editor/react（文件编辑/预览）
- react-markdown + remark-gfm（markdown 渲染）
- zod（表单校验）、sonner（toast）、lucide-react（图标）
- 包管理器：`npm`（注意 Windows 下用 `npm.cmd`）

## 3. 常用命令

```bash
npm run dev     # 开发（vite，端口见 vite.config.ts）
npm run build   # tsc -b && vite build（按需执行）
npm run lint    # eslint（必须 0 error）
```

除非用户明确要求，不要在每次改动后运行 `npm run build`；如有必要，可单独运行 `npm run lint`（允许 badge/button/file-icon 的 3 个既有 react-refresh warning，不允许新增 error/warning）。

## 4. 目录结构

```
src/
  components/
    ui/            # shadcn 风格通用组件（button/input/select/dialog/card/table/...）
    layout/        # 整体布局（sidebar + 展示区）
    data-pagination.tsx # 共享分页（每页条数、输入页码跳转）
    logo.tsx       # 四色 Logo 与 SSOC 彩色文字
    theme-toggle.tsx
  features/        # 按业务域组织
    zip-archive/   # 归档（列表圈 + 创建/编辑弹窗 + api + types）
    zip-file/      # 归档内部文件（文件树 + 预览 + 编辑 + api + types）
    reverse-proxy/ # 反向代理（列表 + 创建/编辑/删除）
    broker/        # 中继节点（v3 API 列表 + 新增）
    minion/        # 探针节点类型、旧版 API 封装和新建表单
  pages/           # 路由页面
    settings.tsx   # 主题与强调色设置
  providers/       # theme-provider
  lib/             # http / theme / format / toast / utils 等基础设施
    fonts.ts       # 可选字体注册表与持久化配置
  index.css        # 全局样式 + 主题变量（唯一配色来源）
```

> 约定：新增业务功能放到 `features/<domain>/`，页面放 `pages/`，通用 UI 放 `components/ui/`。
> 所有分页列表优先复用 `DataPagination`，统一支持每页条数选择和页码输入跳转。卡片列表的条数选项用 12 的倍数，便于铺满响应式网格。

## 5. 网络请求规范（重要）

- API BASE 固定为 **`/api/v3`**（`src/lib/http.ts` 的 `API_BASE`）。
- 认证 token 从 `localStorage` 的 `key = "token"` 读取。
- `/login` 使用旧版 `/api/v1` 登录流程：`POST /auth/valid` 校验账号密码，未绑定 OTP 时调用 `POST /auth/totp` 并显示二维码，最后用 `POST /auth/submit` 完成登录。返回 session token 写入 `localStorage.token`，用户资料写入 `localStorage.auth-profile` 并显示在 Sidebar 用户区。业务路由由 `RequireAuth` 保护，统一请求收到 401 会清除 token/资料并跳转登录页。
- 中继节点页 `/brokers` 使用 v3 接口：`GET /api/v3/brokers`（`{ records }`）、`GET /api/v3/broker?slug=...`（详情）、`POST /api/v3/broker`（新增）、`PUT /api/v3/broker`（更新）和 `GET /api/v3/config`（中心端配置，用于展示继承值）。新增弹窗只提供服务配置，完整配置在详情页 `/brokers/:slug` 编辑。`config` 的 `server`/`logger`/`database` 可分别为 `null`，表示继承中心端配置。
- 探针节点页 `/minions` 使用旧版 `/api/v1/minions`（动态 status filter 通过 `filters` JSON query）和 `DELETE /minion` 逻辑删除。前端暂不提供新增探针功能。
- 普通请求：token 自动写入 `Authorization` Header， **无 `Bearer` Scheme**（这一行为在 `http.ts` 中已处理好，业务代码用
  `http.get/post/put/delete` 即可，无需手动加）。
- **GET 下载类接口**：token 通过 **query 参数 `authorization`** 传递，交给浏览器直接下载。
    - 使用 `buildAuthUrl(path, query)` 生成带认证的完整 URL，配合 `<a href>` 直接下载（不要用 blob 方式下载 GET 下载接口）。
    - 已有封装：`zipArchiveApi.downloadUrl(name)`、`zipFileApi.contentUrl(name, fullName)`。
- 文件内容预览（blob 读取）用 `download(path, query)` —— 它走 Header 认证 + `response.blob()`。
- 上传 multipart：`http.post(path, formData)`（`http.ts` 已自动识别 `FormData`，不强制 JSON；`File` 上传用 `FormData` 的
  `form.set(...)`）。

### 后端错误约定

后端错误遵循 RFC7807 `problem+json`，`http.ts` 统一解析为 `ApiError`（含 `status`/`detail`/`kind`）；网络连接失败与请求取消也分别归类为 `network` / `aborted`。业务层捕获后用
`toast.error(err.message)` 提示即可。所有 toast 从 `@/lib/toast` 导入，统一由 `AppToaster` 渲染。

## 6. 后端接口对照（appv3）

Base：`/api/v3`

| 方法   | 路径                    | 用途                                                                        | 前端封装                            |
|--------|-------------------------|-----------------------------------------------------------------------------|-------------------------------------|
| GET    | `/zip-archives`         | 归档分页列表（`page/size/q`）                                               | `zipArchiveApi.page`                |
| POST   | `/zip-archive`          | 手动创建空归档                                                              | `zipArchiveApi.create`              |
| PUT    | `/zip-archive`          | 更新归档（`{id,...}`）                                                      | `zipArchiveApi.update`              |
| GET    | `/zip-archive/info`     | 归档详情（`name`）                                                          | `zipArchiveApi.info`                |
| GET    | `/zip-archive/download` | 下载整个虚拟 zip（`name`）                                                  | `zipArchiveApi.downloadUrl`         |
| GET    | `/zip-file/readdir`     | 归档内文件列表（`name`）                                                    | `zipFileApi.list`                   |
| GET    | `/zip-file/content`     | 下载/预览单文件内容（`name`+`full_name`）                                   | `zipFileApi.content` / `contentUrl` |
| POST   | `/zip-file/content`     | 全量更新文件内容（multipart：`name`,`full_name`,`file`）                    | `zipFileApi.update` / `create`      |
| POST   | `/zip-file/unzip`       | 上传 zip 摄入（multipart：`name`,`description`,`method`,`password`,`file`） | `zipFileApi.unzip`                  |
| POST   | `/zip-file/mkdir`       | 创建目录（`name`,`full_name`）                                              | `zipFileApi.mkdir`                  |
| DELETE | `/zip-file/content`     | 删除文件/目录（`name`,`full_name`）                                         | `zipFileApi.remove`                 |
| GET    | `/reverse-proxies`      | 反向代理列表                                                                | `reverseProxyApi.list`              |
| POST   | `/reverse-proxy`        | 新建反向代理                                                                | `reverseProxyApi.create`            |
| PUT    | `/reverse-proxy`        | 更新反向代理                                                                | `reverseProxyApi.update`            |
| DELETE | `/reverse-proxy`        | 删除反向代理（`location`）                                                  | `reverseProxyApi.remove`            |

反向代理页面路由为 `/reverse-proxies`。create/update 使用 JSON body，字段为 `location`、`name`、`upstream`、`header`、`manager`、`broker`。Header 在表单中通过键值对编辑，提交为 JSON map。

### 关键字段

- `ZipArchive`：`id / name / description / method / password / version / created_at / updated_at`
- `method`：加密方式（ **整数**）：0-不加密，1-ZipCrypto，2-AES-128，3-AES-192，4-AES-256。`methodLabel(method)` 转显示文案。
- `ZipFile`：
  `id / archive_id / grid_id / directory / file_name / full_name / file_size / mime_type / is_directory / version / created_at / updated_at`

### 后端已知 BUG/约定（前端已做规避）

- 删除目录：后端尚存 BUG，前端点击删除目录仅提示「删除目录功能存在 BUG，还在开发中...」（`zip-archive-detail.tsx` 的
  `handleDelete`）。
- 删除归档：功能未实现，列表卡片删除按钮仅提示「删除归档功能还在开发中...」。
- `is_directory=true` 的节点 **没有内容**，前端预览时直接显示目录占位， **不发起 `GET /zip-file/content`**（否则触发 GridFS
  "file not found"）。

## 7. 目录树 / 文件预览逻辑

- 文件列表接口返回 **扁平**数组，`buildFileTree()` 构建树。
- 文件树默认 **只展开根级目录**，深层折叠（`src/features/zip-file/file-tree.tsx` 的展开状态用 `Set<string>` 统一管理）。
- 树节点图标来自 `public/icons/*.svg`，命名规范：`xxx.svg`（light）/ `xxx_dark.svg`（dark）。
    - 图标映射集中在 `src/features/zip-file/file-icon.tsx` 的 `fileIconBase()` 与 `FileIcon` / `CollapseIcon`。
    - 新增图标后在此文件补充扩展名→图标名映射。
- 预览（`content-preview.tsx`）：
    - 文本：Monaco 编辑器，跟随主题（`useIsDark()` → `near-black`（暗） / `light`（亮））。
    - `md/markdown`：分屏（左 Monaco + 右实时预览），默认分屏，可切换。
    - `csv/tsv`：默认表格渲染，可切换源码编辑。
    - 图片：`<img>` 预览； **svg** 需先读文本再用 `image/svg+xml` blob 渲染（后端 MIME 判定不可靠）。
    - 其它（不支持类型）：显示「该类型文件不支持预览」+ 标题栏（含下载按钮）。
    - **大于 1MB** 的文本文件：不加载内容，显示「文件过大，暂不支持在线编辑」。
- Markdown 预览必须带 `h-full overflow-auto`（否则不滚动）。

## 8. 主题 / 配色 / 风格

- 主题：支持 dark/light，逻辑在 `src/lib/theme.ts` + `providers/theme-provider.tsx`。
    - 判断当前是否暗色：复用 `src/features/zip-file/file-icon.tsx` 的 `useIsDark()`（监听 `html.dark`）。
- **默认强调色：东方财富橙**；用户可在 `/settings` 设置页选择 Green/Teal/Blue/Indigo/Purple/Yellow/Orange/Red/Pink。
    - 主题设置（系统/亮色/暗色）与强调色持久化在 localStorage 的 `theme` / `accent-color`。
    - `primary`、`ring`、`sidebar-primary` 和 `destructive` 等强调色语义变量由 `src/lib/theme.ts` 按当前选择动态应用，避免在业务组件中硬编码主题色。
- 字体选项由 `src/lib/fonts.ts` 的 `FONT_OPTIONS` 统一维护，只影响 Monaco Editor；当前含 Space Mono、IBM Plex Mono、JetBrains Mono、Maple Mono、Ubuntu Mono、Comic Mono 和系统等宽。Monaco 的 CJK 字体回退顺序为 NSimSun、Microsoft YaHei、PingFang SC、系统默认。新增字体资源放在 `public/fonts/`，在 `index.css` 注册 `@font-face`，并在字体注册表中新增选项。网页其它区域不随该设置改变字体。
    - 暗色下选中/悬停高亮用 **中性灰**（`--accent` 灰色），不用橙色（橙色高亮在暗色下突兀）。
    - 暗色背景纯黑 `hsl(0 0% 0%)`。
- 圆角：`--radius: 0.375rem`，控件统一 `rounded-md`。
- 字体：网页全局采用系统默认等宽字体；中文按浏览器/操作系统默认回退，不应用 Monaco 专属中文字体顺序。
- 滚动条：全局默认透明隐藏，悬停时淡入（`index.css` 中已有 `::-webkit-scrollbar` 规则）。
- 时间格式：统一 `yyyy-MM-dd HH:mm:ss`， **一律用 `lib/format.ts` 的 `formatDateTime()`**，禁止手写 `toLocaleString`/自定义
  datePart。
- 鼠标可交互控件统一显示手型（`index.css` 中已有 button/[role] cursor 规则）。
- UI 要简洁：右侧区域尽量留给数据展示，减少无意义文字说明与控制按钮。

## 9. 交互约定

- 右键菜单：`src/components/ui/context-menu.tsx`（`ContextMenu`/`ContextMenuItem`），选项 hover 高亮。
- 所有悬浮提示统一使用 `src/components/tooltip-hint.tsx` 的 `TooltipHint` 或 `InfoTooltip`；禁止使用 HTML `title` 原生提示。
- 删除确认：用自定义 `Dialog`， **禁止**原生 `alert/confirm/prompt`。
- 新建归档弹窗：顶部 tab 切换「上传 ZIP」/「手动创建」；加密方式选了非「不加密」时 **必须填密码**（zod `superRefine`）。
- 密码框：右侧内嵌「随机生成」按钮（`a-zA-Z0-9` 10 位）。
- 归档列表卡片（`pages/zip-archive.tsx`）：hover 显示 下载/编辑/删除 三个操作；点击卡片进入详情页 `/zip-archives/:name`。
- 详情页顶部：返回 + 名称 + 下载 + 编辑 + 右侧统计（文件数/目录数/总大小）。
- Sidebar 用户信息区域可打开菜单；退出登录调用旧版 `DELETE /api/v1/logout`，无论接口结果如何都清理本地 token/profile 并返回登录页。

## 10. 代码风格

- 缩进统一使用 **2 个空格**；`eslint.config.js` 的 `indent` 规则会校验此约定。
- TypeScript 严格模式，`verbatimModuleSyntax`（类型导入用 `import type`）。
- 不使用 `enum` 或命名空间（`erasableSyntaxOnly`），用字符串字面量联合 / `as const`。
- 组件文件若导出非组件（常量/函数），需加 `// eslint-disable-next-line react-refresh/only-export-components`（或拆到单独文件）。
- **不加注释**（除非用户要求）。
- 复用已有工具函数，勿重复造轮子：
    - `cn()`（`@/lib/utils`）
    - `formatDateTime`（`@/lib/format`）
    - `formatFileSize`（`@/features/zip-file/types`）
    - `buildAuthUrl`（`@/lib/http`）
    - `useIsDark`（`@/features/zip-file/file-icon`）

## 11. 明确禁止事项

- 禁止修改工作目录之外的任何文件（尤其后端 Go 代码）。
- 禁止用原生 `alert/confirm`。
- 禁止手写时间格式化（用 `formatDateTime`）。
- 禁止 GET 下载接口用 blob 下载（用 `<a href=buildAuthUrl(...)>` 交给浏览器）。
- 不要把 token 写死 / 不要引入 Scheme `Bearer ` 前缀（后端约定无 Scheme）。
- 不要提交 commit（除非用户明确要求）。

---

（如后端接口有增改，请以 `ssoc-manager/appv3/manager/webapi/*.go` 当前实现为准，并及时更新本文件。）
