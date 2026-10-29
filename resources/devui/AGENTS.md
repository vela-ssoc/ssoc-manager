# DEVUI 前端项目说明（AI Agent 必读）

本文件是 `ssoc-manager/resources/devui` 前端项目的开发规范、风格与约束说明。
任何 AI Agent（或开发者）在改动本项目前，应先阅读本文件，保持与既有实现一致。

---

## 1. 项目定位与边界

- 本项目是一套 **全新 UI**，基于 **React + shadcn/ui**，用于 `ssoc-manager` 的管理控制台。
- 工作目录固定为 `ssoc-manager/resources/devui`。
- **只能修改工作目录内的文件**；其它目录（尤其后端 Go 代码、`resources` 下其它目录）严禁改动。
- 若发现后端 BUG 或建议，仅通过文字向用户反馈，禁止私自修改。
- 后端接口实现位于 `ssoc-manager/appv3/restapi/*.go`（路由注册与 handler 均在 `Restapi` 相关文件中）。

## 2. 技术栈与版本

- Vite 7 + React 19 + TypeScript 5.8
- Tailwind CSS v4（`@tailwindcss/vite` 插件，`@theme inline` 映射 CSS 变量）
- shadcn/ui 风格组件（自实现于 `src/components/ui/`，依赖 Radix UI）
- react-router-dom v7（`BrowserRouter` + `basename="/devui"`）
- monaco-editor + @monaco-editor/react（文件编辑/预览、接口调试报文）
- react-markdown + remark-gfm（markdown 渲染）
- zod（表单校验）、sonner（toast）、lucide-react（图标）
- @xterm/xterm + addon-fit + addon-unicode-graphemes（Web 终端）
- webdav（中心端/中继节点文件管理）
- qrcode.react（TOTP 绑定二维码）、input-otp（6 位验证码）
- 包管理器：`npm`（注意 Windows 下用 `npm.cmd`）

## 3. 常用命令

```bash
npm run dev     # 开发（vite，端口 8087，/api 代理到 https://127.0.0.1:8081）
npm run build   # tsc -b && vite build
npm run lint    # eslint（必须 0 error）
```

除非用户明确要求，不要在每次改动后运行 `npm run build`；如有必要，可单独运行 `npm run lint`。
当前 lint 基线为 **0 error / 3 warning**（`badge.tsx`、`button.tsx`、`file-icon.tsx` 的既有
react-refresh warning），不允许新增 error/warning。

## 4. 目录结构

```
src/
  components/
    ui/                  # shadcn 风格通用组件（button/input/select/dialog/card/table/switch/...）
    layout/              # 整体布局（sidebar + 展示区）
    app-toaster.tsx      # sonner Toaster（跟随主题，自定义配色）
    dark-embedded-frame.tsx # 旧版页面/大屏 iframe 嵌入（invert / overlay 两种模式）
    data-pagination.tsx  # 共享分页（每页条数、输入页码跳转）
    info-tooltip.tsx     # 问号图标 + 富文本提示
    logo.tsx             # 四色 Logo 与 SSOC 彩色文字
    theme-toggle.tsx
    tooltip-hint.tsx     # 统一悬浮提示
    web-terminal.tsx     # xterm + WebSocket 虚拟终端
  features/              # 按业务域组织
    auth/                # 旧版登录 API、咚咚扫码登录
    broker/              # 中继节点（列表、详情各 Tab、配置编辑、WebDAV、接口调试）
    buildinfo/           # 编译信息展示（中心端与中继节点复用）
    central/             # 中心端（监控、配置、定时任务、路由、buildinfo）
    minion/              # 探针节点类型与旧版 API 封装
    oauth-config/        # 身份验证（OAuth 配置 api + types）
    reverse-proxy/       # 反向代理（列表 + 创建/编辑/删除）
    server-events/       # SSE 事件流（连接、重连、订阅 hook、通知卡片、事件历史、铃铛面板）
    system-process/      # 系统进程（列表 + 连接/文件详情，中心端与中继节点复用）
    victoria-metrics/    # VictoriaMetrics 推送配置（api + types）
    zip-archive/         # 归档（列表 + 创建/编辑弹窗 + api + types）
    zip-file/            # 归档内部文件（文件树 + 预览 + 编辑 + api + types）
  pages/                 # 路由页面
    login.tsx            # 扫码 / 账密 + TOTP 登录
    legacy.tsx           # 旧版界面 iframe
    minions.tsx          # 探针节点
    brokers.tsx          # 中继节点列表
    broker-detail.tsx    # 中继节点详情（多 Tab）
    central.tsx          # 中心端（多 Tab）
    zip-archive.tsx      # 归档列表
    zip-archive-detail.tsx # 归档文件树 + 预览编辑
    reverse-proxies.tsx  # 反向代理
    oauth-config.tsx     # 身份验证（OAuth 配置）
    victoria-metrics.tsx # VictoriaMetrics 推送配置
    hone-dashboard.tsx   # 蜜罐大屏 iframe
    risk-dashboard.tsx   # 风险大屏 iframe
    oas3.tsx             # 接口文档 iframe
    settings.tsx         # 主题、强调色、Monaco 字体
  providers/             # theme-provider（主题/强调色/字体上下文）
  lib/                   # http / theme / format / toast / fonts / utils / auth-session
                         # use-picture-in-picture（Document PiP）等基础设施
  index.css              # 全局样式 + 主题变量（唯一配色来源）
```

> 约定：新增业务功能放到 `features/<domain>/`，页面放 `pages/`，通用 UI 放 `components/ui/`。
> 所有分页列表优先复用 `DataPagination`，统一支持每页条数选择和页码输入跳转。卡片列表的条数选项用 12 的倍数，便于铺满响应式网格。

## 5. 路由与导航

路由定义在 `src/App.tsx`，除 `/login` 外均由 `RequireAuth` 保护；首页重定向到 `/legacy`。

| 路由                        | 页面                           | 说明                                                                              |
|-----------------------------|--------------------------------|-----------------------------------------------------------------------------------|
| `/login`                    | `pages/login.tsx`              | 默认扫码登录，可切换账密 + TOTP                                                   |
| `/legacy`                   | `pages/legacy.tsx`             | 旧版 `/` iframe，暗色下整体反色                                                   |
| `/minions`                  | `pages/minions.tsx`            | 探针节点卡片 + 分页 + 逻辑删除                                                    |
| `/brokers`                  | `pages/brokers.tsx`            | 中继节点行卡片：左侧信息，右侧最近 6 小时 CPU/内存/网络指标（一次拉全量，无分页） |
| `/brokers/:slug`            | `pages/broker-detail.tsx`      | 中继节点详情（多 Tab）                                                            |
| `/central`                  | `pages/central.tsx`            | 中心端（多 Tab）                                                                  |
| `/zip-archives`             | `pages/zip-archive.tsx`        | 归档卡片 + 分页                                                                   |
| `/zip-archives/:name`       | `pages/zip-archive-detail.tsx` | 文件树 + Monaco 预览编辑                                                          |
| `/reverse-proxies`          | `pages/reverse-proxies.tsx`    | 反向代理表格 CRUD                                                                 |
| `/configs/oauth`            | `pages/oauth-config.tsx`       | 身份验证（OAuth 登录配置）                                                        |
| `/configs/victoria-metrics` | `pages/victoria-metrics.tsx`   | VictoriaMetrics 推送配置                                                          |
| `/ssoc`                     | `pages/risk-dashboard.tsx`     | 风险大屏 iframe（overlay）                                                        |
| `/hone`                     | `pages/hone-dashboard.tsx`     | 蜜罐大屏 iframe（overlay）                                                        |
| `/oas3`                     | `pages/oas3.tsx`               | 接口文档 iframe                                                                   |
| `/settings`                 | `pages/settings.tsx`           | 主题 / 强调色 / 字体                                                              |

Sidebar 导航项定义在 `components/layout/app-layout.tsx`：旧版界面、探针节点、中继节点、中心端、
虚拟文件、反向代理、配置管理（可折叠：身份验证 / VictoriaMetrics）、大屏展示（可折叠：风险大屏 / 蜜罐大屏）、
接口文档；底部为设置入口与用户菜单。
顶栏 Logo 右侧为通知铃铛（未读徽标 + 通知面板，数据来自 SSE 事件历史；面板内可关闭「订阅通知」，关闭后显示静音铃铛）。
导航项可配置 `external`，hover 时显示「新标签打开」按钮。

## 6. 网络请求规范（重要）

- API BASE 固定为 **`/api/v3`**（`src/lib/http.ts` 的 `API_BASE`）；旧版接口 BASE 为 **`/api/v1`**（`LEGACY_API_BASE`）。
- 认证 token 从 `localStorage` 的 `key = "token"` 读取。
- `/login` 的账密流程使用旧版 `/api/v1`：`POST /auth/valid` 校验账号密码，未绑定 OTP 时调用 `POST /auth/totp` 并显示二维码，最后用
  `POST /auth/submit` 完成登录。扫码登录打开页面时先请求 `GET /api/v3/oauth/provider?origin=...`，返回 `client_id` /
  `redirect_uri` / `authorization_endpoint`；请求期间二维码区域显示同尺寸 loading 占位，iframe 加载时再叠加 loading
  层。出错、返回 null 或字段为空表示未开启扫码登录：页面自动切到账密模式，点击「咚咚扫码」时 toast 提示「当前暂未开启扫码登录」。扫码成功后调用
  `POST /api/v3/oauth/dong`（提交 `code` + `redirect_uri`）完成登录；二维码下方提供「咚咚统一登录认证」超链接，地址为
  `authorization_endpoint?client_id=...&redirect_uri=...`，用于无法扫码时走整页 OAuth 登录。返回 session token 写入
  `localStorage.token`，用户资料写入 `localStorage.auth-profile` 并显示在 Sidebar 用户区。业务路由由 `RequireAuth`
  保护，统一请求收到 401 会清除 token/资料并跳转登录页。
- 普通请求：token 自动写入 `Authorization` Header， **无 `Bearer` Scheme**（这一行为在 `http.ts` 中已处理好，业务代码用
  `http.get/post/put/delete` 即可，无需手动加）。旧版接口用 `legacyHttp`（同样自动携带 token）。
- **GET 下载类接口**：token 通过 **query 参数 `authorization`** 传递，交给浏览器直接下载。
    - 使用 `buildAuthUrl(path, query)` 生成带认证的完整 URL，配合 `<a href>` 直接下载（不要用 blob 方式下载 GET 下载接口）。
    - 已有封装：`zipArchiveApi.downloadUrl(name)`、`zipFileApi.contentUrl(name, fullName)`、
      `davDownloadUrl(endpoint, path)`。
- 文件内容预览（blob 读取）用 `download(path, query)` —— 它走 Header 认证 + `response.blob()`。
- 上传 multipart：`http.post(path, formData)`（`http.ts` 已自动识别 `FormData`，不强制 JSON；`File` 上传用 `FormData` 的
  `form.set(...)`）。
- **接口调试面板**使用 `rawRequest(path, init)`：返回 `{ status, body, contentType, headers, imageUrl }`，
  图片响应自动转 `blob:` URL；`rawRequest` 固定走 v3 Base。
- **WebSocket**：token 通过 query `authorization` 传递，例如
  `${ws|wss}://${host}/api/v3/system/pty?authorization=${token}`；消息格式为
  `{ type: "stdin" | "resize" | "ping", data }` 发送，`{ type: "stdout" | "error", data }` 接收。
- **WebDAV**：`features/broker/webdav.ts` 用 `webdav` 客户端，`remoteBasePath = "/api/v3/webdav"`，
  Header 携带 `Authorization: token`；目录列表、下载 URL、面包屑拆分统一走该文件。
- **SSE 事件流**：登录后由 `AppLayout` 通过 `features/server-events/` 统一订阅 `GET /api/v3/sse/event`
  （`fetch` + `ReadableStream`，Header 需带 `Accept: text/event-stream`、`Cache-Control: no-cache`、`Authorization: token`），
  断线按 1s → 30s 指数退避重连。事件结构为 `{ event, data }`，当前有
  `broker.online` / `broker.offline` / `broker.conflict`（
  `data: { slug, name, inet, version, text?, connected_at?, disconnected_at? }`）。事件提示使用
  `features/server-events/event-notifier.tsx` 的独立通知栈
  （右上角、状态色卡片、6s 自动关闭、hover 暂停、同 key 覆盖、最多 3 条），与 sonner 的普通 toast 区分；
  页面可用 `useServerEvent(事件名或数组, handler)` 自行响应。
- **事件展示定制**：`features/server-events/event-renderers.tsx` 是事件渲染注册表，每个事件返回
  `{ tone, icon, title, content, action?, href?, key?, duration? }`（`tone` 为
  `success | warning | danger | info | neutral`，
  `content` / `action` 可放任意 JSX；`href` 会渲染「查看」跳转；`key` 相同的通知会互相覆盖；`duration` 可单独调整停留时长）。
  新增事件时先在 `client.ts` 的 `ServerEventMap` 补充字段类型，再在 `eventRenderers` 注册渲染函数；
  未注册的事件会以 `info` 风格兜底显示事件名与原始 JSON。
  中继节点事件的标题格式为 `[中继节点] {name} 上线了 / 下线了 / 连接冲突`；
  `href` 指向 `/brokers/{slug}`，弹出通知与铃铛面板均可一键跳转详情。
  离线事件会展示后端计算的在线时长（`connected_at` → `disconnected_at`）；`broker.conflict` 使用红色告警样式，
  额外展示 `text` 中的冲突原因，停留 8 秒。
- **事件历史**：所有事件同时写入 `localStorage` 的 `event-history:<用户ID>`（`features/server-events/history.ts`，
  最新在前、最多 200 条、含 `read` 状态、超出配额时自动减半重试）。Sidebar 顶栏 Logo 右侧的铃铛
  （`features/server-events/notification-center.tsx`）显示未读数；面板支持「全部已读」、逐条标记已读/未读与单条删除，
  清空发生在退出登录时。
- **事件订阅开关**：面板底部「订阅通知」开关，状态持久化在 `localStorage` 的 `event-subscription:<用户ID>`
  （`features/server-events/preferences.ts`）；关闭后 `useServerEvents` 直接 `stopServerEvents()`，
  不再建立/重连 SSE，铃铛图标切换为静音的 `BellOff`，历史记录仍可查看与标记。
  两个存储键都通过 `features/server-events/scope.ts` 按 `auth-profile.id` 隔离，同一浏览器切换账号不会串数据；
  旧版无后缀的 `event-history` / `event-subscription` 键在模块加载时清除，`auth-profile` 变化也会让缓存失效。

### 后端错误约定

后端错误遵循 RFC7807 `problem+json`，`http.ts` 统一解析为 `ApiError`（含 `status`/`detail`/`kind`）；网络连接失败与请求取消也分别归类为
`network` / `aborted`。业务层捕获后用
`toast.error(err.message)` 提示即可。所有 toast 从 `@/lib/toast` 导入，统一由 `AppToaster` 渲染。

## 7. 后端接口对照（appv3）

Base：`/api/v3`，旧版接口为 `/api/v1`。后端路由以 `ssoc-manager/appv3/restapi/*.go` 为准。

### 归档与文件

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

### 反向代理

| 方法   | 路径                 | 用途                       | 前端封装                 |
|--------|----------------------|----------------------------|--------------------------|
| GET    | `/reverse-proxies`   | 反向代理列表               | `reverseProxyApi.list`   |
| POST   | `/reverse-proxy`     | 新建反向代理               | `reverseProxyApi.create` |
| PUT    | `/reverse-proxy`     | 更新反向代理               | `reverseProxyApi.update` |
| DELETE | `/reverse-proxy`     | 删除反向代理（`location`） | `reverseProxyApi.remove` |
| ANY    | `/reverse/:location` | 中心端反向代理转发入口     | 前端只生成跳转地址       |

create/update 使用 JSON body，字段为 `location`、`name`、`upstream`、`header`、`manager`、`broker`。
Header 在表单中通过键值对编辑，提交为 JSON map。

### 中继节点

| 方法   | 路径                           | 用途                             | 前端封装                     |
|--------|--------------------------------|----------------------------------|------------------------------|
| GET    | `/brokers`                     | 中继节点列表（`{ records }`）    | `brokerApi.list`             |
| GET    | `/broker`                      | 详情（`slug`）                   | `brokerApi.info`             |
| POST   | `/broker`                      | 新增                             | `brokerApi.create`           |
| PUT    | `/broker`                      | 更新                             | `brokerApi.update`           |
| GET    | `/broker/connect-histories`    | 连接历史分页（`slug/page/size`） | `brokerApi.connectHistories` |
| GET    | `/broker/substreams`           | 活跃子流（`slug`）               | `brokerApi.substreams`       |
| DELETE | `/broker/stream`               | 结束主连接（`slug`）             | `brokerApi.closeStream`      |
| DELETE | `/broker/substream`            | 结束单个子流（`slug/id`）        | `brokerApi.killSubstream`    |
| ANY    | `/br/:slug`、`/br/:slug/*path` | 中继节点侧 API 代理              | 拼接路径直接请求             |

中继节点侧接口（经 `/br/{slug}` 代理到对端 `/api/v3`）：

| 方法            | 路径                                      | 前端封装                       |
|-----------------|-------------------------------------------|--------------------------------|
| GET             | `/itself/config`                          | `brokerApi.runningConfig`      |
| GET             | `/itself/buildinfo`                       | `brokerApi.buildInfo`          |
| GET             | `/itself/route/{managers,agents,exposes}` | `brokerApi.routes`             |
| GET/POST/DELETE | `/itself/exit`                            | `brokerApi.softExit*`          |
| GET             | `/itself/crontab/jobs`                    | `brokerApi.crontabJobs`        |
| DELETE          | `/itself/crontab/job`                     | `brokerApi.crontabRemove`      |
| GET             | `/system/pty`（WebSocket）                | `WebTerminal`                  |
| GET             | `/system/processes`                       | `systemProcessApi.list`        |
| GET             | `/system/process/connections`             | `systemProcessApi.connections` |
| GET             | `/system/process/openfiles`               | `systemProcessApi.openFiles`   |
| ANY             | `/webdav`、`/webdav/*`                    | `brokerDavEndpoint`            |

### 中心端

| 方法     | 路径                          | 用途                                                                                                                                 | 前端封装                                           |
|----------|-------------------------------|--------------------------------------------------------------------------------------------------------------------------------------|----------------------------------------------------|
| GET      | `/itself/config/effect`       | 当前生效配置（含 boot 数据库信息）                                                                                                   | `centralApi.effectConfig` `brokerApi.effectConfig` |
| GET      | `/itself/config`              | 已保存但可能未生效的最新配置                                                                                                         | `centralApi.latestConfig`                          |
| POST     | `/itself/config`              | 修改/新增配置                                                                                                                        | `centralApi.updateConfig`                          |
| GET      | `/itself/buildinfo`           | 编译信息                                                                                                                             | `centralApi.buildInfo`                             |
| GET      | `/itself/exit`                | 定时退出截止时间                                                                                                                     | `centralApi.exitDeadline`                          |
| POST     | `/itself/exit`                | 定时退出（`seconds`）                                                                                                                | `centralApi.exit`                                  |
| DELETE   | `/itself/exit`                | 取消退出                                                                                                                             | `centralApi.exitCancel`                            |
| GET      | `/itself/crontab/jobs`        | 定时任务列表                                                                                                                         | `centralApi.crontabJobs`                           |
| DELETE   | `/itself/crontab/job`         | 删除定时任务（`id`）                                                                                                                 | `centralApi.crontabRemove`                         |
| GET      | `/itself/route/brokers`       | 中继节点侧接口列表                                                                                                                   | `centralApi.brokerRoutes`                          |
| GET      | `/itself/route/exposes`       | 用户侧接口列表                                                                                                                       | `centralApi.exposeRoutes`                          |
| GET      | `/vmp/api/v1/query`           | VictoriaMetrics 即时查询                                                                                                             | `instantQuery`                                     |
| GET      | `/vmp/api/v1/query_range`     | VictoriaMetrics 区间查询                                                                                                             | `rangeQuery`                                       |
| ANY      | `/webdav`、`/webdav/*`        | 中心端文件管理（WebDAV）                                                                                                             | `centralDavEndpoint`                               |
| GET      | `/system/pty`（WebSocket）    | 中心端虚拟终端                                                                                                                       | `WebTerminal`                                      |
| GET      | `/system/processes`           | 系统进程列表                                                                                                                         | `systemProcessApi.list`                            |
| GET      | `/system/process/connections` | 进程打开的连接（`pid`）                                                                                                              | `systemProcessApi.connections`                     |
| GET      | `/system/process/openfiles`   | 进程关联的文件描述符（`pid`）                                                                                                        | `systemProcessApi.openFiles`                       |
| GET/POST | `/oauth/config`               | OAuth 身份验证配置（`enabled,client_id,client_secret,authorization_endpoint,token_endpoint,userinfo_endpoint,redirect_uris,scopes`） | `oauthConfigApi.get/update`                        |
| GET      | `/oauth/provider`             | 登录页 OAuth 提供商信息（`origin`，返回 `client_id` / `redirect_uri` / `authorization_endpoint`）                                    | `oauthProviderApi.get`                             |
| GET      | `/sse/event`                  | SSE 事件流订阅（`Accept: text/event-stream`）                                                                                        | `features/server-events/client.ts`                 |
| POST     | `/oauth/dong`                 | 扫码 code 登录（`code` + `redirect_uri`）                                                                                            | `oauthApi.dong`                                    |
| GET/POST | `/victoria-metrics-config`    | VictoriaMetrics 推送配置（`{enabled,address,tenant_id,username,password,headers}`）                                                  | `victoriaMetricsApi.get/update`                    |

### 探针节点（旧版 `/api/v1`）

| 方法   | 路径       | 用途                                                | 前端封装                  |
|--------|------------|-----------------------------------------------------|---------------------------|
| GET    | `/minions` | 分页（`current/size/keyword`，可选 `filters` JSON） | `minionApi.page`          |
| GET    | `/minion`  | 详情（`id`，已封装但未使用）                        | `minionApi.detail`        |
| DELETE | `/minion`  | 逻辑删除（`filters` JSON）                          | `minionApi.logicalDelete` |

### 登录（旧版 `/api/v1`）

| 方法   | 路径           | 用途              | 前端封装                 |
|--------|----------------|-------------------|--------------------------|
| POST   | `/auth/valid`  | 校验账号密码      | `legacyAuthApi.validate` |
| POST   | `/auth/totp`   | 获取绑定二维码    | `legacyAuthApi.totp`     |
| POST   | `/auth/submit` | 提交 OTP 完成登录 | `legacyAuthApi.submit`   |
| DELETE | `/logout`      | 退出登录          | `legacyAuthApi.logout`   |

### 关键字段

- `ZipArchive`：`id / name / description / method / password / version / created_at / updated_at`
- `method`：加密方式（ **整数**）：0-不加密，1-ZipCrypto，2-AES-128，3-AES-192，4-AES-256。`methodLabel(method)` 转显示文案。
- `ZipFile`：
  `id / archive_id / grid_id / directory / file_name / full_name / file_size / mime_type / is_directory / version / created_at / updated_at`
- `Broker`：
  `id / slug / name / exposes / config / status / tunnel_stat / execute_stat / active_streams / heartbeat_at / disconnect_at / created_at / updated_at`
- `BrokerConfig`：`server` / `logger` / `database` 可分别为 `null`，表示继承中心端配置。
- `ManagerConfig`：中心端生效配置（`server.addr`、`database.uri/dsn`、`logger`），用于展示继承值与配置 diff。
- `BrokerConnectHistory`：`id / broker_id / execute_stat / tunnel_stat / seconds / disconnected_at`
- `BrokerSubstream`：`id（大整数字符串，比较时用 BigInt）/ rx / tx / established_at`
- `CronJobInfo`：`id / name / expr / timeout（纳秒）/ immediate / next / entry_id`
- `CentralConfig`：`id / server{addr, static{path:{spa,dir}}} / logger / created_at / updated_at`
- `CentralConfigResponse`：`{ boot: { database: { uri, dsn } }, config: CentralConfig | null }`
- `Minion`：
  `id / inet / goos / edition / status / heartbeat_at / cpu_core / mem_total / mem_free / idc / ibu / comment / broker_name / unload / uptime / unstable / tags`
- `MinionStatus`：1-未激活，2-离线，3-在线，4-已删除（`MINION_STATUSES`）
- `ProcessInfo`：
  `pid / name / cwd / exe / cmdline / username / cpu_percent / num_threads / num_fds / created_at / memory{rss,vms,hwm,data,stack,locked,swap}`
- `ProcessConnection`：`fd / family / type / localaddr{ip,port} / remoteaddr{ip,port} / status / uids / pid`
- `ProcessOpenFile`：`path / fd`
- `ReverseProxy`：`location / name / manager / broker / header / upstream / created_at / updated_at`
- `VictoriaMetricsConfig`：`id / enabled / address / tenant_id / username / password / headers / created_at / updated_at`
    - 该配置用于管理中心端向 VictoriaMetrics 推送自身运行指标；`enabled=false` 时不推送。
    - 集群版 `tenant_id` 非空时，后端自动拼接 `insert/{tenant}/prometheus` 写入路径。
- `OAuthConfig`：
  `enabled / client_id / client_secret / authorization_endpoint / token_endpoint / userinfo_endpoint / redirect_uris / scopes / created_at / updated_at`
    - 用于登录页的 OAuth 授权登录；`redirect_uris` 至少 1 个，`scopes` 最多 10 个。
    - 点击编辑时，三个端点为空会自动填充咚咚默认端点（login / token / userinfo）；
      回调地址为空时自动填充 `window.location.origin`；首次配置（尚无已保存配置）进入编辑时，启用开关默认打开。

### 后端已知 BUG/约定（前端已做规避）

- 删除目录：后端尚存 BUG，前端点击删除目录仅提示「删除目录功能存在 BUG，还在开发中...」（`zip-archive-detail.tsx` 的
  `handleDelete`）。
- 删除归档：功能未实现，列表卡片删除按钮仅提示「删除归档功能还在开发中...」。
- `is_directory=true` 的节点 **没有内容**，前端预览时直接显示目录占位， **不发起 `GET /zip-file/content`**（否则触发 GridFS
  "file not found"）。
- `GET /reverse-proxy` 在后端被错误地绑定到 `creator`（创建）处理，前端未调用该接口；如需单个详情请先与后端确认。

## 8. 目录树 / 文件预览逻辑

- 文件列表接口返回 **扁平**数组，`buildFileTree()` 构建树。
- 文件树默认 **只展开根级目录**，深层折叠（`src/features/zip-file/file-tree.tsx` 的展开状态用 `Set<string>` 统一管理）。
- 树节点图标来自 `public/icons/*.svg`，命名规范：`xxx.svg`（light）/ `xxx_dark.svg`（dark）。
    - 图标映射集中在 `src/features/zip-file/file-icon.tsx` 的 `fileIconBase()` 与 `FileIcon` / `CollapseIcon`。
    - 新增图标后在此文件补充扩展名→图标名映射。
- 预览（`content-preview.tsx`）：中心端/中继节点 WebDAV 与虚拟文件共用同一套预览逻辑，
  通过 `source.load` / `source.downloadUrl` / 可选 `source.save` 适配不同数据源；无 `save` 时为只读。
    - 先判断大小： **任何文件大于 1MB 都不预览**，显示「文件过大，暂不支持预览」。
    - 文本：Monaco 编辑器，跟随主题（`useIsDark()` → `near-black`（暗） / `light`（亮））。
    - `md/markdown`：默认直接渲染，可切换为分屏（左 Monaco + 右实时预览），滚动比例联动。
    - `csv/tsv`：默认表格渲染，可切换源码查看。
    - 图片：`<img>` 预览； **svg** 需先读文本再用 `image/svg+xml` blob 渲染（后端 MIME 判定不可靠）。
    - 其它（不支持类型）：显示「该类型文件不支持预览」+ 标题栏（含下载按钮）。
- Markdown 预览必须带 `h-full overflow-auto`（否则不滚动）。
- WebDAV 文件管理（`features/broker/webdav-explorer.tsx`）：左侧为 Windows 中图标网格
  （类型图标 + 名称，悬浮显示大小/修改时间，选中高亮）+ 右侧预览；预览直接复用
  `zip-file/content-preview.tsx`（只读，无保存按钮）。支持前进/后退/上级/刷新、
  浏览器支持时可用 Document PiP 画中画打开。

## 9. 主题 / 配色 / 风格

- 主题：支持 dark/light/system，逻辑在 `src/lib/theme.ts` + `providers/theme-provider.tsx`。
    - 判断当前是否暗色：复用 `src/features/zip-file/file-icon.tsx` 的 `useIsDark()`（监听 `html.dark`），
      或在组件内用 `useTheme().isDark`。
- **默认强调色：东方财富橙**；用户可在 `/settings` 设置页选择 Green/Teal/Blue/Indigo/Purple/Yellow/Orange/Red/Pink。
    - 主题设置（系统/亮色/暗色）与强调色持久化在 localStorage 的 `theme` / `accent-color`。
    - `primary`、`ring`、`sidebar-primary` 和 `destructive` 等强调色语义变量由 `src/lib/theme.ts`
      按当前选择动态应用，避免在业务组件中硬编码主题色。
- 字体选项由 `src/lib/fonts.ts` 的 `FONT_OPTIONS` 统一维护，只影响 Monaco Editor/终端；当前含 Space Mono、IBM Plex
  Mono、JetBrains Mono、Maple Mono、Ubuntu Mono、Comic Mono 和系统等宽。Monaco 的 CJK 字体回退顺序为 NSimSun、Microsoft
  YaHei、PingFang SC、系统默认。新增字体资源放在 `public/fonts/`，在 `index.css` 注册 `@font-face`
  ，并在字体注册表中新增选项。网页其它区域不随该设置改变字体。
- 暗色下选中/悬停高亮用 **中性灰**（`--accent` 灰色），不用橙色（橙色高亮在暗色下突兀）。
- 暗色背景纯黑 `hsl(0 0% 0%)`。
- 圆角：`--radius: 0.375rem`，控件统一 `rounded-md`。
- 字体：网页全局采用系统默认等宽字体；中文按浏览器/操作系统默认回退，不应用 Monaco 专属中文字体顺序。
- 滚动条：全局默认透明隐藏，悬停时淡入（`index.css` 中已有 `::-webkit-scrollbar` 规则）。
- 时间格式：统一 `yyyy-MM-dd HH:mm:ss`， **一律用 `lib/format.ts` 的 `formatDateTime()`**，禁止手写 `toLocaleString`/自定义
  datePart。
- 鼠标可交互控件统一显示手型（`index.css` 中已有 button/[role] cursor 规则）。
- UI 要简洁：右侧区域尽量留给数据展示，减少无意义文字说明与控制按钮。
- 危险提示统一使用 `border-destructive/30 bg-destructive/5 text-destructive` 的浅底警告块 + `TriangleAlert` 图标。
- 成功/生效状态用 `text-green-600 dark:text-green-400`，异常用 `text-destructive`。

## 10. 交互约定

- 右键菜单：`src/components/ui/context-menu.tsx`（`ContextMenu`/`ContextMenuItem`），选项 hover 高亮。
- 所有悬浮提示统一使用 `src/components/tooltip-hint.tsx` 的 `TooltipHint` 或 `InfoTooltip`；禁止使用 HTML `title` 原生提示。
- 删除确认：用自定义 `Dialog`， **禁止**原生 `alert/confirm/prompt`。
- 开关统一使用 `src/components/ui/switch.tsx` 的 `Switch`（受控 `checked` + `onCheckedChange`，`label` 必填用于
  aria-label）。
- 危险操作（重连、重启、结束子流等）：除确认弹窗外还需 **3 秒倒计时**，复用
  `features/broker/use-confirm-countdown.ts` 的 `useConfirmCountdown(open)`，按钮文案为 `确定 (n)`。
- 新建归档弹窗：顶部 tab 切换「上传 ZIP」/「手动创建」；加密方式选了非「不加密」时 **必须填密码**（zod `superRefine`）。
- 密码框：右侧内嵌「随机生成」按钮（`a-zA-Z0-9` 10 位）。
- 归档列表卡片（`pages/zip-archive.tsx`）：hover 显示 下载/编辑/删除 三个操作；点击卡片进入详情页 `/zip-archives/:name`。
- 详情页顶部：返回 + 名称 + 下载 + 编辑 + 右侧统计（文件数/目录数/总大小）。
- 多 Tab 页面（中心端、中继节点详情）：顶部 `border-b` 容器 + 下划线选中态
  （`border-b-2` + `border-primary`），内容区 `min-h-0 flex-1 overflow-y-auto p-4`。
- 中继节点详情中「在线才可用」的 Tab 需置灰并在点击时 toast 提示「中继节点离线，无法访问该功能」。
- 列表/卡片 hover 操作：桌面端默认 `opacity-0`，`group-hover`/`group-focus-within` 显现；移动端可常显。
- 大表格使用 `table-fixed` + 固定列宽 + `truncate` + `TooltipHint`，避免内容撑破布局。
- 系统进程（中心端 / 中继节点「文件管理」之后）：列表支持搜索与刷新；点击进程行弹出详情，
  再按 `pid` 懒加载「打开的连接」和「打开的文件」；中继节点 Tab 仅在线可用。
- 运行监控（`/central`）：默认显示最近 24 小时；时间范围支持相对预设与自定义绝对区间（本地时间）；刷新间隔可选
  `关闭 / 10 秒 / 30 秒 / 1 分钟`；相对区间每次刷新按当前时间滑动。
  X 轴始终对应用户选择的时间窗口，无数据的区间留空（曲线断开），不压缩有数据的片段。
  自定义时间区间使用面板内的日历与时间输入，不使用原生 `datetime-local` 控件；
  第一次点击日期选择开始、第二次点击选择结束。
  在任意指标图表上横向拖拽可放大到所选时间区间（Grafana 式 zoom）。
  中继节点详情的「运行监控」Tab 复用 `MetricsMonitoring`，查询标签为
  `appname="ssoc-broker",instance="<slug>"`；中心端使用 `appname="ssoc-manager"`。
  顶部「虚拟通道」统计 `tunnel_receive_bytes` / `tunnel_transmit_bytes` /
  `tunnel_cumulative_substreams` / `tunnel_active_substreams`：中心端查询
  `appname="ssoc-broker"` 并保留 `instance` 作为图例；中继节点详情额外附加
  `instance="<slug>"`，图例不重复显示实例。接收/发送合并为一张流量图；
  累计子流按 `rate(...[5m])` 展示增长率，与活跃子流放在同一张图。
- 中继节点列表（`/brokers`）：每行一张卡片，左侧节点信息，右侧上下排列最近 6 小时的
  CPU、内存、网络三张紧凑指标图；图例可点击隐藏曲线，点击图表区域仍进入详情，点击图例不跳转。
  指标通过 `appname="ssoc-broker"` 一次聚合查询全部节点（按 `instance` 分组），不按卡片重复请求。
  收到 `broker.online` / `broker.offline` 事件后会静默刷新列表（400ms 防抖，不触发整页 loading）；
  中继节点详情页仅同步自身 `slug` 的在线状态。
- 接口调试（`features/broker/route-api.tsx`）：方法/路径下拉 + 路径参数/查询参数/Header/报文分栏 + Monaco 响应查看，
  JSON 支持格式化切换，Header 名称提供联想。
- 虚拟终端与 WebDAV 支持 Document Picture-in-Picture（仅 Chromium 系浏览器），通过
  `lib/use-picture-in-picture.ts` 实现；终端在画中画窗口中同样支持复制粘贴。
- 虚拟终端快捷键：选中文字后 `Ctrl/Cmd + C` 复制选中内容；`Ctrl/Cmd + V` 粘贴剪贴板内容；
  未选中文字时 `Ctrl + C` 仍发送中断信号；复制成功后自动清除选区，复制失败保留选区。
  实现位于 `components/web-terminal.tsx` 的
  xterm `attachCustomKeyEventHandler`，剪贴板操作会使用终端当前所在的 document/window
  （包括画中画窗口）。
- Sidebar 用户信息区域可打开菜单；退出登录调用旧版 `DELETE /api/v1/logout`，无论接口结果如何都清理本地 token/profile
  并返回登录页。

## 11. 代码风格

- 缩进统一使用 **2 个空格**；`eslint.config.js` 的 `indent` 规则会校验此约定。
- TypeScript 严格模式，`verbatimModuleSyntax`（类型导入用 `import type`）。
- 不使用 `enum` 或命名空间（`erasableSyntaxOnly`），用字符串字面量联合 / `as const`。
- 组件文件若导出非组件（常量/函数），需加 `// eslint-disable-next-line react-refresh/only-export-components`（或拆到单独文件）。
- **不加注释**（除非用户要求）。
- 复用已有工具函数，勿重复造轮子：
    - `cn()`（`@/lib/utils`）
    - `formatDateTime` / `formatDuration` / `formatTimeout`（`@/lib/format`）
    - `formatFileSize`（`@/features/zip-file/types`）
    - `formatBytes`（`@/features/broker/types`，字节数换算，最多保留 2 位小数并去掉末尾 0）
    - `buildAuthUrl` / `download` / `rawRequest`（`@/lib/http`）
    - `useIsDark`（`@/features/zip-file/file-icon`）
    - `TooltipHint` / `InfoTooltip`
    - `DataPagination`
    - `useConfirmCountdown`
- 表单校验统一用 `zod`；错误信息通过 `safeParse` 收集到 `Record<string, string>` 后渲染在字段下方。
- 分页参数命名：v3 归档用 `page/size/q`，v1 探针用 `current/size/keyword`，不要混用。
- 页面级数据加载用 `useCallback` 包 `load`，`useEffect(() => void load(), [load])`；并发请求用递增 `requestId` 防止旧响应覆盖新响应。

## 12. 明确禁止事项

- 禁止修改工作目录之外的任何文件（尤其后端 Go 代码）。
- 禁止用原生 `alert/confirm`。
- 禁止手写时间格式化（用 `formatDateTime`）。
- 禁止 GET 下载接口用 blob 下载（用 `<a href=buildAuthUrl(...)>` 交给浏览器）。
- 不要把 token 写死 / 不要引入 Scheme `Bearer ` 前缀（后端约定无 Scheme）。
- 不要提交 commit（除非用户明确要求）。
- 不要把业务配色写死为橙色；一律使用主题变量（`primary` / `accent` / `muted` 等）。

---

## 13. 当前已知差异 / 待办

- 探针节点目前只有列表、筛选与逻辑删除；`minionApi.detail` 已封装但尚无详情页，也 **没有新建探针入口**。
- 搜索交互不统一：中继节点/反向代理为输入即筛选（前端过滤），归档/探针需回车或点击「搜索」（服务端查询）。
- `zip-archive-detail.tsx` 顶部存在一处残留空行（仅格式问题）。
- 删除归档、删除目录两个功能仍依赖后端修复。

---

（如后端接口有增改，请以 `ssoc-manager/appv3/restapi/*.go` 当前实现为准，并及时更新本文件。）