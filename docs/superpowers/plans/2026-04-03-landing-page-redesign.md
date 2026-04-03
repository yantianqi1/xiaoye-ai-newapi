# Landing Page Redesign (语画姬) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the flashy video-heavy "小野 AI" landing page with a clean, light-themed "语画姬" page showing an announcement board, 3-step usage guide, and API docs.

**Architecture:** Single file rewrite of `Landing.vue` (script + template + scoped style). i18n text updated in zh.json / en.json. Router default title and SEO keys updated to "语画姬". No new routes, no backend changes.

**Tech Stack:** Vue 3 (Composition API), Naive UI (NButton only), vue-i18n, vue-router, scoped CSS

---

## File Map

| Action | File | Responsibility |
|--------|------|---------------|
| Modify | `frontend/src/views/Landing.vue` | Complete rewrite — new script, template, styles |
| Modify | `frontend/src/locales/zh.json` | Update `landing.*` + `seo.landing.*` keys |
| Modify | `frontend/src/locales/en.json` | Same keys, English copy |
| Modify | `frontend/src/router/index.js` | Change fallback title `'小野 AI'` → `'语画姬'` |

---

## Task 1: Update i18n — zh.json landing keys

**Files:**
- Modify: `frontend/src/locales/zh.json`

- [ ] **Step 1: Replace all `landing.*` keys in zh.json**

Find the block of `"landing.*"` keys and replace the entire block with:

```json
"landing.navBrand": "语画姬",
"landing.enterPlatform": "进入平台",
"landing.announcementTitle": "📢 公告",
"landing.announcement1Date": "2026-04-03",
"landing.announcement1Title": "语画姬正式上线",
"landing.announcement1Content": "欢迎使用语画姬 AI 创作平台！支持图像生成、视频生成等功能，注册即可获得体验积分。",
"landing.stepsTitle": "三步开始创作",
"landing.step1Title": "注册账号",
"landing.step1Desc": "访问平台，使用邮箱完成注册，登录后即可使用所有功能。",
"landing.step2Title": "获取密钥",
"landing.step2Desc": "登录后进入账户页，兑换密钥或充值积分，积分用于调用 AI 生成接口。",
"landing.step3Title": "开始创作",
"landing.step3Desc": "在生成页输入提示词，选择模型，点击生成，几秒内获得 AI 创作结果。",
"landing.apiTitle": "API 文档",
"landing.apiSubtitle": "所有接口均为 JSON 格式，Base URL 替换为你的部署地址。",
"landing.apiAuthTitle": "1. 登录获取 Token",
"landing.apiAuthDesc": "使用邮箱和密码登录，返回 JWT Token，后续接口均需在 Header 中携带。",
"landing.apiModelsTitle": "2. 获取模型列表",
"landing.apiModelsDesc": "获取当前平台支持的全部 AI 模型及对应积分定价，无需登录。",
"landing.apiGenerateTitle": "3. 生成图像/视频",
"landing.apiGenerateDesc": "统一生成接口，type 为 image 或 video，需携带 Authorization Header。",
```

- [ ] **Step 2: Update SEO keys**

Find `"seo.landing.title"`, `"seo.landing.description"`, `"seo.landing.keywords"` and replace their values:

```json
"seo.landing.title": "语画姬 - AI 创作平台",
"seo.landing.description": "语画姬 AI 平台，支持图像生成、视频生成。",
"seo.landing.keywords": "语画姬,AI图像生成,AI视频生成,文生图,文生视频",
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/locales/zh.json
git commit -m "i18n: update landing page text for 语画姬"
```

---

## Task 2: Update i18n — en.json landing keys

**Files:**
- Modify: `frontend/src/locales/en.json`

- [ ] **Step 1: Replace all `landing.*` keys in en.json**

Find the `"landing.*"` block and replace with English equivalents:

```json
"landing.navBrand": "语画姬",
"landing.enterPlatform": "Enter Platform",
"landing.announcementTitle": "📢 Announcements",
"landing.announcement1Date": "2026-04-03",
"landing.announcement1Title": "语画姬 is now live",
"landing.announcement1Content": "Welcome to 语画姬 AI platform! Supports image and video generation. Register now to receive trial credits.",
"landing.stepsTitle": "Get Started in 3 Steps",
"landing.step1Title": "Create Account",
"landing.step1Desc": "Visit the platform, register with your email, and log in to access all features.",
"landing.step2Title": "Get API Key",
"landing.step2Desc": "Go to your account page after login, redeem a key or top up credits to use AI generation.",
"landing.step3Title": "Start Creating",
"landing.step3Desc": "Enter a prompt in the generate page, pick a model, and get your AI-generated result in seconds.",
"landing.apiTitle": "API Documentation",
"landing.apiSubtitle": "All endpoints use JSON. Replace the Base URL with your deployment address.",
"landing.apiAuthTitle": "1. Login to Get Token",
"landing.apiAuthDesc": "Login with email and password to receive a JWT Token. Include it in the Authorization header for subsequent requests.",
"landing.apiModelsTitle": "2. Get Model List",
"landing.apiModelsDesc": "Retrieve all supported AI models and their credit pricing. No authentication required.",
"landing.apiGenerateTitle": "3. Generate Image / Video",
"landing.apiGenerateDesc": "Unified generation endpoint. Set type to image or video. Requires Authorization header.",
```

- [ ] **Step 2: Update SEO keys**

```json
"seo.landing.title": "语画姬 - AI Creation Platform",
"seo.landing.description": "语画姬 AI platform for image and video generation.",
"seo.landing.keywords": "AI image generation,AI video generation,text to image,text to video",
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/locales/en.json
git commit -m "i18n: update landing page English text for 语画姬"
```

---

## Task 3: Update router default title

**Files:**
- Modify: `frontend/src/router/index.js:181`

- [ ] **Step 1: Change the fallback title**

On line 181, change:

```js
document.title = resolveMetaText(to.meta, 'title', 'titleKey', '小野 AI')
```

to:

```js
document.title = resolveMetaText(to.meta, 'title', 'titleKey', '语画姬')
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/router/index.js
git commit -m "chore: update default document title to 语画姬"
```

---

## Task 4: Rewrite Landing.vue — script section

**Files:**
- Modify: `frontend/src/views/Landing.vue`

- [ ] **Step 1: Replace the entire `<script setup>` block**

Replace everything between `<script setup>` and `</script>` with:

```js
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { NButton } from 'naive-ui'

const { t } = useI18n()
const router = useRouter()

const announcements = [
  {
    date: t('landing.announcement1Date'),
    title: t('landing.announcement1Title'),
    content: t('landing.announcement1Content'),
  },
]

const steps = [
  { num: '01', title: t('landing.step1Title'), desc: t('landing.step1Desc') },
  { num: '02', title: t('landing.step2Title'), desc: t('landing.step2Desc') },
  { num: '03', title: t('landing.step3Title'), desc: t('landing.step3Desc') },
]

const handleEnter = () => {
  router.push('/inspiration')
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/views/Landing.vue
git commit -m "refactor: rewrite Landing.vue script — remove video/animation logic"
```

---

## Task 5: Rewrite Landing.vue — template

**Files:**
- Modify: `frontend/src/views/Landing.vue`

- [ ] **Step 1: Replace the entire `<template>` block**

Replace everything between `<template>` and `</template>` with:

```html
<div class="landing">

  <!-- 导航栏 -->
  <nav class="l-nav">
    <div class="l-nav-inner">
      <div class="l-brand">
        <img src="/images/icon.png" alt="" class="l-brand-icon" />
        <span class="l-brand-text">{{ $t('landing.navBrand') }}</span>
      </div>
      <NButton type="primary" @click="handleEnter" strong>
        {{ $t('landing.enterPlatform') }}
      </NButton>
    </div>
  </nav>

  <main class="l-main">

    <!-- 公告板 -->
    <section class="l-section">
      <div class="l-card">
        <h2 class="l-announce-heading">{{ $t('landing.announcementTitle') }}</h2>
        <div
          v-for="(item, i) in announcements"
          :key="i"
          class="l-announce-item"
        >
          <div class="l-announce-meta">
            <span class="l-announce-title">{{ item.title }}</span>
            <span class="l-announce-date">{{ item.date }}</span>
          </div>
          <p class="l-announce-content">{{ item.content }}</p>
        </div>
      </div>
    </section>

    <!-- 使用步骤 -->
    <section class="l-section l-section-gray">
      <div class="l-container">
        <h2 class="l-section-title">{{ $t('landing.stepsTitle') }}</h2>
        <div class="l-steps">
          <div v-for="(s, i) in steps" :key="i" class="l-step-card">
            <div class="l-step-num">{{ s.num }}</div>
            <h3 class="l-step-title">{{ s.title }}</h3>
            <p class="l-step-desc">{{ s.desc }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- API 文档 -->
    <section class="l-section">
      <div class="l-container">
        <h2 class="l-section-title">{{ $t('landing.apiTitle') }}</h2>
        <p class="l-section-sub">{{ $t('landing.apiSubtitle') }}</p>

        <!-- 接口 1: 登录 -->
        <div class="l-api-block">
          <h3 class="l-api-title">{{ $t('landing.apiAuthTitle') }}</h3>
          <p class="l-api-desc">{{ $t('landing.apiAuthDesc') }}</p>
          <pre class="l-code"><code>curl -X POST https://your-domain/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"you@example.com","password":"yourpassword"}'

# 响应
{"token":"eyJhbGci..."}</code></pre>
        </div>

        <!-- 接口 2: 模型列表 -->
        <div class="l-api-block">
          <h3 class="l-api-title">{{ $t('landing.apiModelsTitle') }}</h3>
          <p class="l-api-desc">{{ $t('landing.apiModelsDesc') }}</p>
          <pre class="l-code"><code>curl https://your-domain/api/models</code></pre>
        </div>

        <!-- 接口 3: 生成 -->
        <div class="l-api-block">
          <h3 class="l-api-title">{{ $t('landing.apiGenerateTitle') }}</h3>
          <p class="l-api-desc">{{ $t('landing.apiGenerateDesc') }}</p>
          <table class="l-param-table">
            <thead>
              <tr><th>字段</th><th>类型</th><th>必填</th><th>说明</th></tr>
            </thead>
            <tbody>
              <tr><td><code>type</code></td><td>string</td><td>是</td><td>image 或 video</td></tr>
              <tr><td><code>prompt</code></td><td>string</td><td>是</td><td>文字描述</td></tr>
              <tr><td><code>model</code></td><td>string</td><td>否</td><td>模型 ID，见模型列表</td></tr>
              <tr><td><code>params</code></td><td>object</td><td>否</td><td>额外参数，如 {"size":"1K"}</td></tr>
            </tbody>
          </table>
          <pre class="l-code"><code>curl -X POST https://your-domain/api/generate \
  -H "Authorization: Bearer &lt;token&gt;" \
  -H "Content-Type: application/json" \
  -d '{"type":"image","prompt":"一只橘猫坐在窗台上","model":"nanobanana-pro"}'</code></pre>
        </div>
      </div>
    </section>

  </main>

  <!-- 页脚 -->
  <footer class="l-footer">
    <a href="https://beian.miit.gov.cn/" target="_blank" rel="noopener noreferrer">苏ICP备2025223139号-1</a>
    <span class="l-footer-divider">|</span>
    <a href="https://www.beian.gov.cn/portal/registerSystemInfo?recordcode=32050602014047" target="_blank" rel="noopener noreferrer">苏公网安备32050602014047号</a>
  </footer>

</div>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/views/Landing.vue
git commit -m "feat: rewrite Landing.vue template for 语画姬"
```

---

## Task 6: Rewrite Landing.vue — styles

**Files:**
- Modify: `frontend/src/views/Landing.vue`

- [ ] **Step 1: Replace the entire `<style scoped>` block**

Replace everything between `<style scoped>` and `</style>` with:

```css
/* ===== Layout ===== */
.landing {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #fff;
  color: #1a1a1a;
  font-family: -apple-system, BlinkMacSystemFont, 'PingFang SC', 'Microsoft YaHei', sans-serif;
}

/* ===== Nav ===== */
.l-nav {
  position: sticky;
  top: 0;
  z-index: 100;
  background: #fff;
  border-bottom: 1px solid #e5e7eb;
}

.l-nav-inner {
  max-width: 860px;
  margin: 0 auto;
  padding: 12px 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.l-brand {
  display: flex;
  align-items: center;
  gap: 8px;
}

.l-brand-icon {
  width: 28px;
  height: 28px;
  object-fit: contain;
}

.l-brand-text {
  font-size: 18px;
  font-weight: 700;
  color: #1a1a1a;
}

/* ===== Main ===== */
.l-main {
  flex: 1;
}

.l-section {
  padding: 48px 24px;
}

.l-section-gray {
  background: #f8f9fa;
}

.l-container {
  max-width: 860px;
  margin: 0 auto;
}

.l-section-title {
  font-size: 22px;
  font-weight: 700;
  color: #1a1a1a;
  margin: 0 0 8px;
}

.l-section-sub {
  font-size: 14px;
  color: #6b7280;
  margin: 0 0 28px;
}

/* ===== Announcement ===== */
.l-card {
  max-width: 760px;
  margin: 0 auto;
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 24px 28px;
}

.l-announce-heading {
  font-size: 18px;
  font-weight: 700;
  margin: 0 0 20px;
  color: #1a1a1a;
}

.l-announce-item {
  padding: 16px 0;
  border-top: 1px solid #f3f4f6;
}

.l-announce-item:first-of-type {
  border-top: none;
  padding-top: 0;
}

.l-announce-meta {
  display: flex;
  align-items: baseline;
  gap: 12px;
  margin-bottom: 6px;
}

.l-announce-title {
  font-size: 15px;
  font-weight: 600;
  color: #1a1a1a;
}

.l-announce-date {
  font-size: 12px;
  color: #9ca3af;
}

.l-announce-content {
  font-size: 14px;
  color: #4b5563;
  margin: 0;
  line-height: 1.7;
}

/* ===== Steps ===== */
.l-steps {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
  margin-top: 28px;
}

.l-step-card {
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 24px 20px;
}

.l-step-num {
  font-size: 28px;
  font-weight: 800;
  color: #7c3aed;
  margin-bottom: 12px;
  line-height: 1;
}

.l-step-title {
  font-size: 16px;
  font-weight: 600;
  color: #1a1a1a;
  margin: 0 0 8px;
}

.l-step-desc {
  font-size: 13px;
  color: #6b7280;
  margin: 0;
  line-height: 1.6;
}

/* ===== API Docs ===== */
.l-api-block {
  margin-bottom: 36px;
  padding-bottom: 36px;
  border-bottom: 1px solid #e5e7eb;
}

.l-api-block:last-child {
  border-bottom: none;
  margin-bottom: 0;
  padding-bottom: 0;
}

.l-api-title {
  font-size: 16px;
  font-weight: 600;
  color: #1a1a1a;
  margin: 0 0 6px;
}

.l-api-desc {
  font-size: 13px;
  color: #6b7280;
  margin: 0 0 14px;
  line-height: 1.6;
}

.l-code {
  background: #1e1e2e;
  color: #cdd6f4;
  border-radius: 8px;
  padding: 16px 20px;
  font-size: 13px;
  line-height: 1.6;
  overflow-x: auto;
  margin: 0;
}

.l-param-table {
  width: 100%;
  border-collapse: collapse;
  margin-bottom: 14px;
  font-size: 13px;
}

.l-param-table th,
.l-param-table td {
  text-align: left;
  padding: 8px 12px;
  border: 1px solid #e5e7eb;
}

.l-param-table th {
  background: #f8f9fa;
  font-weight: 600;
  color: #374151;
}

.l-param-table td {
  color: #4b5563;
}

.l-param-table code {
  background: #f3f4f6;
  padding: 1px 5px;
  border-radius: 4px;
  font-family: monospace;
  font-size: 12px;
  color: #7c3aed;
}

/* ===== Footer ===== */
.l-footer {
  padding: 20px 24px;
  text-align: center;
  font-size: 12px;
  color: #9ca3af;
  border-top: 1px solid #e5e7eb;
}

.l-footer a {
  color: #9ca3af;
  text-decoration: none;
}

.l-footer a:hover {
  color: #6b7280;
}

.l-footer-divider {
  margin: 0 10px;
}

/* ===== Responsive ===== */
@media (max-width: 640px) {
  .l-steps {
    grid-template-columns: 1fr;
  }

  .l-nav-inner {
    padding: 10px 16px;
  }

  .l-section {
    padding: 32px 16px;
  }
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/views/Landing.vue
git commit -m "style: add clean light-theme styles for 语画姬 landing page"
```

---

## Task 7: Verify and final cleanup

**Files:**
- Read: `frontend/src/views/Landing.vue`

- [ ] **Step 1: Start the dev server and verify visually**

```bash
cd frontend && npm run dev
```

Open the root URL (`/`) in a browser and verify:
- [ ] Brand shows "语画姬", no "小野 AI" visible
- [ ] Page is light-themed (white background)
- [ ] No video elements, no external CDN requests
- [ ] Announcement card renders correctly
- [ ] 3 step cards are side by side (desktop) / stacked (mobile)
- [ ] All 3 API blocks render with code examples
- [ ] Footer shows beian numbers
- [ ] "进入平台" button navigates to `/inspiration`

- [ ] **Step 2: Grep for any remaining "小野 AI" in landing-related files**

```bash
grep -rn "小野 AI" frontend/src/views/Landing.vue frontend/src/locales/zh.json frontend/src/locales/en.json frontend/src/router/index.js
```

Expected: no output (zero matches).

- [ ] **Step 3: Final commit if any cleanup needed**

```bash
git add -p
git commit -m "chore: remove remaining 小野 AI references from landing"
```
