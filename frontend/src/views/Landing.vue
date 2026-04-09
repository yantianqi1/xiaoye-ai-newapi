<script setup>
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

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
</script>

<template>
<div class="landing">

  <!-- 导航栏 -->
  <nav class="l-nav">
    <div class="l-nav-inner">
      <div class="l-brand">
        <img src="/images/jmlogo.png" alt="" class="l-brand-icon" />
        <span class="l-brand-text">{{ $t('landing.navBrand') }}</span>
      </div>
      <button class="l-cta-btn" @click="handleEnter">
        {{ $t('landing.enterPlatform') }}
      </button>
    </div>
  </nav>

  <main class="l-main">

    <!-- 公告板 -->
    <section class="l-section">
      <div class="l-card">
        <h2 class="l-announce-heading">{{ $t('landing.announcementTitle') }}</h2>

        <!-- 蛋糕 API · NovelAI 中转站 -->
        <a
          class="l-promo"
          href="https://keai.keai.shop"
          target="_blank"
          rel="noopener noreferrer"
        >
          <div class="l-promo-badge">NovelAI 中转</div>
          <div class="l-promo-body">
            <div class="l-promo-title">蛋糕 API · 文生图中转站</div>
            <p class="l-promo-desc">
              提供 NovelAI 官方接口形式的文生图服务，可在 SillyTavern、Tavo
              以及任何支持正则与 NovelAI 官方接入方式的项目中使用（非 100% 适配）。
              欢迎前往站内注册、充值并创建密钥。
            </p>
            <span class="l-promo-link">
              keai.keai.shop
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M7 17L17 7"/><path d="M8 7h9v9"/></svg>
            </span>
          </div>
        </a>

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
              <tr><td><code>params</code></td><td>object</td><td>否</td><td>额外参数，如 <code>{"size":"1K"}</code></td></tr>
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
</template>

<style scoped>
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

.l-cta-btn {
  background: #7c3aed;
  color: #fff;
  border: none;
  border-radius: 6px;
  padding: 8px 20px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s;
}

.l-cta-btn:hover {
  background: #6d28d9;
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

.l-announce-item:first-child {
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

/* ===== Promo Card ===== */
.l-promo {
  display: flex;
  gap: 18px;
  align-items: flex-start;
  margin: 4px 0 22px;
  padding: 18px 20px;
  border-radius: 12px;
  background: linear-gradient(135deg, #faf5ff 0%, #f5f3ff 100%);
  border: 1px solid #e9d5ff;
  text-decoration: none;
  color: inherit;
  transition: transform 0.15s ease, box-shadow 0.15s ease, border-color 0.15s ease;
}

.l-promo:hover {
  transform: translateY(-1px);
  border-color: #c4b5fd;
  box-shadow: 0 8px 24px -12px rgba(124, 58, 237, 0.35);
}

.l-promo-badge {
  flex-shrink: 0;
  background: #7c3aed;
  color: #fff;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.3px;
  padding: 6px 10px;
  border-radius: 6px;
  margin-top: 2px;
}

.l-promo-body {
  flex: 1;
  min-width: 0;
}

.l-promo-title {
  font-size: 15px;
  font-weight: 700;
  color: #1a1a1a;
  margin-bottom: 6px;
}

.l-promo-desc {
  font-size: 13px;
  color: #4b5563;
  line-height: 1.7;
  margin: 0 0 10px;
}

.l-promo-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  font-weight: 600;
  color: #7c3aed;
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
</style>
