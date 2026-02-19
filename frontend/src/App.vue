<script setup>
import { computed, reactive } from 'vue'
import { RunScan, SelectInputFile } from '../wailsjs/go/main/App'

const form = reactive({
  inputFilePath: '',
  concurrency: 30,
  timeoutSeconds: 5,
  followRedirect: true,
})

const state = reactive({
  running: false,
  error: '',
  reportPath: '',
  total200Lines: 0,
  totalUrls: 0,
  succeeded: 0,
  failed: 0,
  rows: [],
})

const canStart = computed(() => !state.running && form.inputFilePath.trim() !== '')
const hasRows = computed(() => state.rows.length > 0)

function normalizeError(err) {
  if (!err) {
    return '未知错误'
  }
  if (typeof err === 'string') {
    return err
  }
  if (err.message) {
    return err.message
  }
  return String(err)
}

function formatComponents(components) {
  if (!Array.isArray(components) || components.length === 0) {
    return '无'
  }
  return components.join(', ')
}

async function browseFile() {
  state.error = ''
  try {
    const filePath = await SelectInputFile()
    if (filePath) {
      form.inputFilePath = filePath
    }
  } catch (err) {
    state.error = normalizeError(err)
  }
}

async function startScan() {
  if (!canStart.value) {
    return
  }

  state.running = true
  state.error = ''

  try {
    const response = await RunScan({
      inputFilePath: form.inputFilePath.trim(),
      concurrency: Number(form.concurrency),
      timeoutSeconds: Number(form.timeoutSeconds),
      followRedirect: Boolean(form.followRedirect),
    })

    state.reportPath = response.reportPath || ''
    state.total200Lines = response.total200Lines || 0
    state.totalUrls = response.totalUrls || 0
    state.succeeded = response.succeeded || 0
    state.failed = response.failed || 0
    state.rows = Array.isArray(response.rows) ? response.rows : []
  } catch (err) {
    state.error = normalizeError(err)
  } finally {
    state.running = false
  }
}
</script>

<template>
  <main class="page">
    <section class="hero">
      <h1>URL 扫描报告生成器</h1>
      <p>读取文本文件中的 200 状态行 URL，提取网页标题与组件信息，并生成 Markdown 报告。</p>
    </section>

    <section class="layout">
      <article class="card">
        <h2>扫描配置</h2>

        <div class="row">
          <label for="inputFilePath">输入文件</label>
          <div class="inline">
            <input
              id="inputFilePath"
              v-model="form.inputFilePath"
              class="input"
              type="text"
              placeholder="请选择源文本文件路径"
            />
            <button class="btn btn-secondary" :disabled="state.running" @click="browseFile">浏览</button>
          </div>
        </div>

        <div class="grid">
          <div class="row">
            <label for="concurrency">并发数</label>
            <input id="concurrency" v-model.number="form.concurrency" class="input" type="number" min="1" max="100" />
          </div>
          <div class="row">
            <label for="timeoutSeconds">超时时间（秒）</label>
            <input id="timeoutSeconds" v-model.number="form.timeoutSeconds" class="input" type="number" min="1" max="120" />
          </div>
          <div class="row checkbox-row">
            <label>
              <input v-model="form.followRedirect" type="checkbox" />
              跟随重定向
            </label>
          </div>
        </div>

        <div class="actions">
          <button class="btn btn-primary" :disabled="!canStart" @click="startScan">
            {{ state.running ? '扫描中...' : '开始扫描' }}
          </button>
        </div>

        <p v-if="state.error" class="error">{{ state.error }}</p>
      </article>

      <article class="card status-card">
        <h2>运行状态</h2>
        <p><strong>状态：</strong>{{ state.running ? '正在扫描' : '空闲' }}</p>
        <p><strong>报告路径：</strong>{{ state.reportPath || '尚未生成' }}</p>
        <p><strong>命中 200 行：</strong>{{ state.total200Lines }}</p>
        <p><strong>提取 URL：</strong>{{ state.totalUrls }}</p>
        <p><strong>成功：</strong>{{ state.succeeded }}</p>
        <p><strong>失败：</strong>{{ state.failed }}</p>
      </article>
    </section>

    <section class="card table-card">
      <h2>扫描结果预览</h2>
      <div v-if="hasRows" class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>URL</th>
              <th>标题</th>
              <th>组件信息</th>
              <th>错误信息</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in state.rows" :key="row.url">
              <td>{{ row.url }}</td>
              <td>{{ row.title || '无' }}</td>
              <td>{{ formatComponents(row.components) }}</td>
              <td>{{ row.error || '-' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-else class="empty">
        暂无结果，请先选择文件并开始扫描。
      </div>
    </section>
  </main>
</template>

<style scoped>
.page {
  max-width: 1200px;
  margin: 24px auto;
  padding: 0 20px 24px;
  box-sizing: border-box;
}

.hero {
  background: linear-gradient(120deg, #f0f7ff, #f7fbff);
  border: 1px solid #dbeafe;
  border-radius: 14px;
  padding: 20px 24px;
  margin-bottom: 16px;
  color: #1e3a8a;
}

.hero h1 {
  margin: 0 0 8px;
}

.hero p {
  margin: 0;
  color: #334155;
}

.layout {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 16px;
  margin-bottom: 16px;
}

.card {
  background: #ffffff;
  color: #1f2937;
  border-radius: 12px;
  padding: 20px;
  border: 1px solid #e2e8f0;
  box-shadow: 0 6px 20px rgba(148, 163, 184, 0.16);
  text-align: left;
}

h1,
h2 {
  margin: 0 0 12px;
}

.grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.row {
  margin-bottom: 12px;
}

label {
  display: block;
  font-weight: 600;
  margin-bottom: 6px;
}

.inline {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 8px;
}

.input {
  width: 100%;
  min-height: 38px;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 8px 10px;
  box-sizing: border-box;
  color: #1f2937;
  background: #ffffff;
}

.checkbox-row {
  display: flex;
  align-items: end;
}

.checkbox-row label {
  margin-bottom: 8px;
  font-weight: 500;
}

.actions {
  margin-top: 8px;
}

.btn {
  min-height: 38px;
  border: none;
  border-radius: 8px;
  padding: 0 14px;
  cursor: pointer;
}

.btn:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.btn-primary {
  background: #3b82f6;
  color: #ffffff;
}

.btn-secondary {
  background: #eff6ff;
  color: #1d4ed8;
  border: 1px solid #bfdbfe;
}

.error {
  margin-top: 14px;
  color: #b91c1c;
  font-weight: 600;
}

.status-card p {
  margin: 8px 0;
  color: #334155;
}

.table-card {
  overflow: hidden;
}

.table-wrap {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th,
td {
  border: 1px solid #e5e7eb;
  padding: 8px;
  vertical-align: top;
  font-size: 14px;
  line-height: 1.4;
}

th {
  background: #f8fafc;
}

.empty {
  min-height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px dashed #cbd5e1;
  border-radius: 8px;
  color: #64748b;
  background: #f8fafc;
}

@media (max-width: 900px) {
  .layout {
    grid-template-columns: 1fr;
  }

  .grid {
    grid-template-columns: 1fr;
  }
}
</style>
