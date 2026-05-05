<script setup>
import { reactive, ref, computed, onMounted } from 'vue'
import { LoadSave, ApplySave, GetSavePath } from '../../wailsjs/go/main/App'
import { WindowMinimise, Quit } from '../../wailsjs/runtime/runtime'
import RollNumber from './RollNumber.vue'

// 存档数据（当前值显示用）
const saveData = reactive({
  rebirthLevel: 0,
  level: 0,
  gold: 0,
})

// 修改用的输入框数据
const editData = reactive({
  rebirthLevel: '',
  level: '',
  gold: '',
})

const isLoaded = ref(false)
const isLoading = ref(false)
const isApplying = ref(false)
const saveStatus = ref('')
const statusType = ref('') // 'success' | 'error'
const savePath = ref('')

// 输入值是否与已保存数据完全一致（未作任何修改）
const isUnchanged = computed(() => {
  return (
    parseInt(editData.rebirthLevel) === saveData.rebirthLevel &&
    parseInt(editData.level)        === saveData.level &&
    parseInt(editData.gold)         === saveData.gold
  )
})

// 获取存档路径显示
onMounted(() => {
  GetSavePath().then(p => { savePath.value = p }).catch(() => {})
})

// 读取存档
function loadSave() {
  isLoading.value = true
  saveStatus.value = ''
  LoadSave()
    .then((result) => {
      saveData.rebirthLevel = result.rebirthLevel
      saveData.level = result.level
      saveData.gold = result.gold
      editData.rebirthLevel = String(result.rebirthLevel)
      editData.level = String(result.level)
      editData.gold = String(result.gold)
      isLoaded.value = true
      isLoading.value = false
      loadDisplayText.value = LOAD_TARGET_LOADED
      showStatus('存档读取成功！', 'success')
    })
    .catch((err) => {
      isLoading.value = false
      showStatus('读取失败：' + (err || '未知错误'), 'error')
    })
}

// 应用修改
function applyChanges() {
  const rb = parseInt(editData.rebirthLevel)
  const lv = parseInt(editData.level)
  const gd = parseInt(editData.gold)

  if (isNaN(rb) || isNaN(lv) || isNaN(gd)) {
    showStatus('请输入有效的数字！', 'error')
    return
  }

  isApplying.value = true
  ApplySave({ rebirthLevel: rb, level: lv, gold: gd })
    .then(() => {
      saveData.rebirthLevel = rb
      saveData.level = lv
      saveData.gold = gd
      isApplying.value = false
      showStatus('修改已写入存档！', 'success')
    })
    .catch((err) => {
      isApplying.value = false
      showStatus('保存失败：' + (err || '未知错误'), 'error')
    })
}

// ── 读取按钮乱码动画 ──
const LOAD_TARGET_DEFAULT = '读取存档'
const LOAD_TARGET_LOADED  = '重新读取'

const loadDisplayText  = ref(LOAD_TARGET_DEFAULT)
const isLoadShineActive = ref(false)
let   loadScrambleTimer = null
let   loadScramblePos   = 0

function loadScrambleStart() {
  if (isLoading.value) return
  const target = isLoaded.value ? LOAD_TARGET_LOADED : LOAD_TARGET_DEFAULT
  isLoadShineActive.value = true
  loadScramblePos = 0
  clearInterval(loadScrambleTimer)
  const chars = [...target]
  loadScrambleTimer = setInterval(() => {
    loadDisplayText.value = chars
      .map((ch, i) => {
        if (loadScramblePos / CYCLES_PER_CHAR > i) return ch
        return CHARS[Math.floor(Math.random() * CHARS.length)]
      })
      .join('')
    loadScramblePos++
    if (loadScramblePos >= chars.length * CYCLES_PER_CHAR) loadScrambleStop()
  }, SHUFFLE_MS)
}

function loadScrambleStop() {
  clearInterval(loadScrambleTimer)
  loadDisplayText.value  = isLoaded.value ? LOAD_TARGET_LOADED : LOAD_TARGET_DEFAULT
  isLoadShineActive.value = false
}

// ── 应用按钮乱码动画 ──
const APPLY_TARGET   = '应用修改'
const CHARS          = '!@#$%^&*():{};|,.<>/?'
const CYCLES_PER_CHAR = 2
const SHUFFLE_MS      = 50

const applyDisplayText = ref(APPLY_TARGET)
const isShineActive    = ref(false)
let   scrambleTimer    = null
let   scramblePos      = 0

function scrambleStart() {
  if (isUnchanged.value || isApplying.value) return
  isShineActive.value = true
  scramblePos = 0
  clearInterval(scrambleTimer)
  const chars = [...APPLY_TARGET]
  scrambleTimer = setInterval(() => {
    applyDisplayText.value = chars
      .map((ch, i) => {
        if (scramblePos / CYCLES_PER_CHAR > i) return ch
        return CHARS[Math.floor(Math.random() * CHARS.length)]
      })
      .join('')
    scramblePos++
    if (scramblePos >= chars.length * CYCLES_PER_CHAR) scrambleStop()
  }, SHUFFLE_MS)
}

function scrambleStop() {
  clearInterval(scrambleTimer)
  applyDisplayText.value = APPLY_TARGET
  isShineActive.value    = false
}

// 重新读取（先重置到未加载状态再调用 loadSave）
function resetSave() {
  isLoaded.value = false
  loadDisplayText.value = LOAD_TARGET_DEFAULT
  loadSave()
}

function showStatus(msg, type) {
  saveStatus.value = msg
  statusType.value = type
  setTimeout(() => { saveStatus.value = '' }, 3000)
}
</script>

<template>
  <!-- 整个窗口可拖动，内容区取消拖动 -->
  <div class="app-window">

    <!-- ── 自定义标题栏（拖动区）── -->
    <div class="titlebar" style="--wails-draggable:drag">
      <div class="titlebar-left">
        <span class="titlebar-icon">👻</span>
        <span class="titlebar-title">Phasmophobia 存档修改器</span>
        <transition name="fade">
          <span v-if="saveStatus" class="titlebar-status" :class="statusType">
            {{ statusType === 'success' ? '●' : '✕' }} {{ saveStatus }}
          </span>
        </transition>
      </div>
      <div class="titlebar-controls" style="--wails-draggable:no-drag">
        <button class="win-btn minimize" @click="WindowMinimise" title="最小化">
          <svg width="10" height="1" viewBox="0 0 10 1"><rect width="10" height="1.5" rx="0.75" fill="currentColor"/></svg>
        </button>
        <button class="win-btn close" @click="Quit" title="关闭">
          <svg width="10" height="10" viewBox="0 0 10 10"><line x1="1" y1="1" x2="9" y2="9" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/><line x1="9" y1="1" x2="1" y2="9" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
        </button>
      </div>
    </div>

    <!-- ── 内容区（不可拖动）── -->
    <main class="container" style="--wails-draggable:no-drag">
    <!-- 存档路径 -->
    <div v-if="savePath" class="path-bar">
      <span class="path-icon">📁</span>
      <span class="path-text" :title="savePath">{{ savePath }}</span>
    </div>

    <!-- 读取存档按钮 -->
    <div class="load-section">
      <button
        class="btn-load"
        :class="{ loaded: isLoaded }"
        @click="isLoaded ? resetSave() : loadSave()"
        @mouseenter="loadScrambleStart"
        @mouseleave="loadScrambleStop"
        :disabled="isLoading"
      >
        <div class="btn-load-inner">
          <svg v-if="isLoading" class="btn-load-icon spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
          </svg>
          <svg v-else-if="isLoaded" class="btn-load-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 2v6h-6"/><path d="M3 12a9 9 0 0 1 15-6.7L21 8"/><path d="M3 22v-6h6"/><path d="M21 12a9 9 0 0 1-15 6.7L3 16"/>
          </svg>
          <svg v-else class="btn-load-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
          </svg>
          <span class="btn-load-text">{{ isLoading ? '读取中...' : loadDisplayText }}</span>
        </div>
        <span class="btn-load-shine" :class="{ active: isLoadShineActive && !isLoading }"></span>
      </button>
    </div>

    <!-- 数据展示与编辑区 -->
    <transition name="slide-up">
      <div v-if="isLoaded" class="data-panel">

        <!-- 转生等级 -->
        <div class="data-card rebirth">
          <div class="card-header">
            <span class="card-icon">🌟</span>
            <span class="card-label">转生等级</span>
            <span class="card-range">0 ~ 20</span>
          </div>
          <div class="card-value">
            <RollNumber :value="saveData.rebirthLevel" />
          </div>
          <div class="card-edit">
            <input
              v-model="editData.rebirthLevel"
              type="number"
              min="0"
              max="20"
              class="edit-input"
              placeholder="0 ~ 20"
            />
          </div>
        </div>

        <!-- 等级 -->
        <div class="data-card level">
          <div class="card-header">
            <span class="card-icon">⚡</span>
            <span class="card-label">等级</span>
            <span class="card-range">1 ~ 9999</span>
          </div>
          <div class="card-value">
            <RollNumber :value="saveData.level" />
          </div>
          <div class="card-edit">
            <input
              v-model="editData.level"
              type="number"
              min="1"
              max="9999"
              class="edit-input"
              placeholder="1 ~ 9999"
            />
          </div>
        </div>

        <!-- 金币 -->
        <div class="data-card gold">
          <div class="card-header">
            <span class="card-icon">💰</span>
            <span class="card-label">金币</span>
            <span class="card-range">0 ~ 9,999,999</span>
          </div>
          <div class="card-value">
            <RollNumber :value="saveData.gold" :locale="true" />
          </div>
          <div class="card-edit">
            <input
              v-model="editData.gold"
              type="number"
              min="0"
              max="9999999"
              class="edit-input"
              placeholder="0 ~ 9999999"
            />
          </div>
        </div>

        <!-- 应用修改按钮 -->
        <div class="apply-section">
          <button
            class="btn-apply"
            :class="{ unchanged: isUnchanged }"
            @click="applyChanges"
            @mouseenter="scrambleStart"
            @mouseleave="scrambleStop"
            :disabled="isApplying || isUnchanged"
          >
            <div class="btn-apply-inner">
              <svg class="btn-apply-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>
              </svg>
              <span class="btn-apply-text">{{ isApplying ? '写入中...' : isUnchanged ? '未作修改' : applyDisplayText }}</span>
            </div>
            <span class="btn-apply-shine" :class="{ active: isShineActive }"></span>
          </button>
        </div>

      </div>
    </transition>

    <!-- 未读取存档时的占位提示 -->
    <transition name="fade">
      <div v-if="!isLoaded && !isLoading" class="placeholder">
        <div class="placeholder-icon">🗂️</div>
        <p>点击上方按钮读取存档数据</p>
        <p class="placeholder-tip">请先在游戏主界面退出到桌面，<br>再使用本工具修改存档</p>
      </div>
    </transition>

  </main>
  </div>
</template>

<style scoped>
/* ── 窗口外壳 ── */
.app-window {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
  background-color: rgba(27, 38, 54, 1);
  border-radius: 10px;          /* 无边框时圆角更好看 */
  box-shadow: 0 8px 40px rgba(0, 0, 0, 0.6);
}

/* ── 自定义标题栏 ── */
.titlebar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 38px;
  padding: 0 6px 0 14px;
  background: rgba(18, 26, 38, 0.95);
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  flex-shrink: 0;
  user-select: none;
}

.titlebar-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.titlebar-icon {
  font-size: 1rem;
  line-height: 1;
}

.titlebar-title {
  font-size: 0.8rem;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.55);
  letter-spacing: 0.5px;
}

/* 窗口控制按钮区 */
.titlebar-controls {
  display: flex;
  align-items: center;
  gap: 2px;
}

.win-btn {
  width: 32px;
  height: 28px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: rgba(255, 255, 255, 0.45);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s, color 0.15s;
}

.win-btn.minimize:hover {
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.9);
}

.win-btn.close:hover {
  background: rgba(239, 68, 68, 0.8);
  color: #fff;
}

/* ── 内容滚动区 ── */
.container {
  flex: 1;
  overflow-y: auto;
  max-width: 480px;
  width: 100%;
  margin: 0 auto;
  padding: 20px 20px 40px;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  /* 隐藏滚动条 */
  scrollbar-width: none;
  -ms-overflow-style: none;
}
.container::-webkit-scrollbar {
  display: none;
}

/* ── 存档路径 ── */
.path-bar {
  width: 100%;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.path-icon {
  font-size: 0.9rem;
  flex-shrink: 0;
}

.path-text {
  font-size: 0.72rem;
  color: rgba(255, 255, 255, 0.4);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-family: 'Courier New', monospace;
}

/* ── 读取按钮 ── */
.load-section {
  width: 100%;
}

.btn-load {
  position: relative;
  overflow: hidden;
  width: 100%;
  padding: 13px 0;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 14px;
  font-size: 1rem;
  font-weight: 700;
  font-family: 'Courier New', monospace;
  letter-spacing: 2px;
  text-transform: uppercase;
  cursor: pointer;
  background: rgba(40, 48, 64, 1);
  color: rgba(255, 255, 255, 0.7);
  transition: color 0.2s, border-color 0.2s, transform 0.15s;
}

.btn-load:not(:disabled):hover {
  color: #67e8f9;
  border-color: rgba(103, 232, 249, 0.45);
  transform: scale(1.025);
}

.btn-load:not(:disabled):active {
  transform: scale(0.975);
}

.btn-load:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

/* 已读取状态边框色调变化 */
.btn-load.loaded {
  border-color: rgba(255, 255, 255, 0.12);
  color: rgba(255, 255, 255, 0.5);
}

.btn-load.loaded:not(:disabled):hover {
  color: #67e8f9;
  border-color: rgba(103, 232, 249, 0.45);
}

.btn-load-inner {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.btn-load-icon {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
}

.btn-load-text {
  min-width: 6em;
  text-align: left;
}

/* 旋转动画（读取中） */
.spin {
  animation: icon-spin 0.8s linear infinite;
}

@keyframes icon-spin {
  from { transform: rotate(0deg); }
  to   { transform: rotate(360deg); }
}

/* 扫光层（青色调，与应用按钮区分） */
.btn-load-shine {
  position: absolute;
  inset: 0;
  z-index: 0;
  transform: scaleY(1.25) translateY(100%);
  background: linear-gradient(
    to top,
    rgba(103, 232, 249, 0) 40%,
    rgba(103, 232, 249, 0.7) 50%,
    rgba(103, 232, 249, 0) 60%
  );
  opacity: 0;
  transition: opacity 0.3s;
  animation: shine-sweep 1s linear infinite;
  animation-play-state: paused;
}

.btn-load-shine.active {
  opacity: 1;
  animation-play-state: running;
}

/* ── 标题栏状态徽章 ── */
.titlebar-status {
  font-size: 0.68rem;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 20px;
  letter-spacing: 0.3px;
  white-space: nowrap;
}

.titlebar-status.success {
  color: #4ade80;
  background: rgba(34, 197, 94, 0.15);
}

.titlebar-status.error {
  color: #f87171;
  background: rgba(239, 68, 68, 0.15);
}

/* ── 数据面板 ── */
.data-panel {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* ── 数据卡片 ── */
.data-card {
  border-radius: 16px;
  padding: 16px 18px;
  display: grid;
  grid-template-columns: 1fr auto;
  grid-template-rows: auto auto;
  align-items: center;
  gap: 10px 12px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.data-card:hover {
  transform: translateY(-2px);
}

.data-card.rebirth {
  background: linear-gradient(135deg, rgba(124, 58, 237, 0.25) 0%, rgba(249, 212, 35, 0.1) 100%);
  box-shadow: 0 4px 20px rgba(124, 58, 237, 0.18);
}

.data-card.level {
  background: linear-gradient(135deg, rgba(79, 142, 247, 0.25) 0%, rgba(34, 211, 238, 0.1) 100%);
  box-shadow: 0 4px 20px rgba(79, 142, 247, 0.18);
}

.data-card.gold {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.25) 0%, rgba(249, 212, 35, 0.1) 100%);
  box-shadow: 0 4px 20px rgba(245, 158, 11, 0.18);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 7px;
  grid-column: 1;
  grid-row: 1;
}

.card-icon {
  font-size: 1.2rem;
}

.card-label {
  font-size: 0.88rem;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.65);
  letter-spacing: 1px;
}

.card-range {
  font-size: 0.68rem;
  color: rgba(255, 255, 255, 0.3);
  margin-left: 2px;
}

.card-value {
  font-size: 1.65rem;
  font-weight: 800;
  color: #fff;
  text-align: right;
  grid-column: 2;
  grid-row: 1;
  text-shadow: 0 0 16px rgba(255, 255, 255, 0.25);
  letter-spacing: 1px;
}

.card-edit {
  grid-column: 1 / -1;
  grid-row: 2;
}

.edit-input {
  width: 100%;
  box-sizing: border-box;
  padding: 8px 14px;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.15);
  background: rgba(255, 255, 255, 0.07);
  color: #fff;
  font-size: 0.95rem;
  font-family: inherit;
  outline: none;
  transition: border-color 0.2s, background 0.2s;
}

.edit-input::placeholder {
  color: rgba(255, 255, 255, 0.22);
}

.edit-input:focus {
  border-color: rgba(255, 255, 255, 0.4);
  background: rgba(255, 255, 255, 0.12);
}

.edit-input::-webkit-outer-spin-button,
.edit-input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

/* ── 应用修改按钮 ── */
.apply-section {
  margin-top: 4px;
}

.btn-apply {
  position: relative;
  overflow: hidden;
  width: 100%;
  padding: 13px 0;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 14px;
  font-size: 1rem;
  font-weight: 700;
  font-family: 'Courier New', monospace;
  letter-spacing: 2px;
  text-transform: uppercase;
  cursor: pointer;
  background: rgba(40, 48, 64, 1);
  color: rgba(255, 255, 255, 0.7);
  transition: color 0.2s, border-color 0.2s, transform 0.15s;
}

.btn-apply:not(:disabled):hover {
  color: #a5b4fc;
  border-color: rgba(165, 180, 252, 0.5);
  transform: scale(1.025);
}

.btn-apply:not(:disabled):active {
  transform: scale(0.975);
}

.btn-apply:disabled {
  cursor: not-allowed;
}

/* 未修改时的暗色禁用态 */
.btn-apply.unchanged {
  background: rgba(255, 255, 255, 0.06);
  box-shadow: none;
  color: rgba(255, 255, 255, 0.2);
  border-color: rgba(255, 255, 255, 0.06);
}

/* 按钮内容层 */
.btn-apply-inner {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.btn-apply-icon {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
}

.btn-apply-text {
  min-width: 6em;
  text-align: left;
}

/* 扫光层：从下往上循环移动，hover 时显现 */
.btn-apply-shine {
  position: absolute;
  inset: 0;
  z-index: 0;
  transform: scaleY(1.25);
  background: linear-gradient(
    to top,
    rgba(99, 102, 241, 0) 40%,
    rgba(99, 102, 241, 0.85) 50%,
    rgba(99, 102, 241, 0) 60%
  );
  opacity: 0;
  transition: opacity 0.3s;
  animation: shine-sweep 1s linear infinite;
  animation-play-state: paused;
}

.btn-apply-shine.active {
  opacity: 1;
  animation-play-state: running;
}

@keyframes shine-sweep {
  0%   { transform: scaleY(1.25) translateY(100%); }
  100% { transform: scaleY(1.25) translateY(-100%); }
}

/* ── 占位提示 ── */
.placeholder {
  margin-top: 20px;
  color: rgba(255, 255, 255, 0.25);
  text-align: center;
  font-size: 0.88rem;
  line-height: 1.8;
}

.placeholder-icon {
  font-size: 3rem;
  margin-bottom: 8px;
  opacity: 0.4;
}

.placeholder-tip {
  font-size: 0.78rem;
  color: rgba(255, 255, 255, 0.18);
  margin-top: 8px;
}

/* ── 过渡动画 ── */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.slide-up-enter-active {
  transition: all 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94);
}

.slide-up-enter-from {
  opacity: 0;
  transform: translateY(24px);
}
</style>
