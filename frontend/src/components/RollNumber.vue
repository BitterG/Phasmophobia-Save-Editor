<script setup>
import { computed, ref, watch } from 'vue'

const props = defineProps({
  value: { type: Number, default: 0 },
  // 是否用千位分隔符（金币用）
  locale: { type: Boolean, default: false },
})

// 把数字拆成单独字符（含逗号）
const digits = computed(() => {
  const str = props.locale
    ? props.value.toLocaleString()
    : String(props.value)
  return str.split('')
})

// 每个字符的动画状态：idle | exit | enter
// key 变化时强制触发动画
const animKeys = ref(digits.value.map(() => 0))
const animStates = ref(digits.value.map(() => 'idle'))

let prevDigits = [...digits.value]

watch(digits, (next) => {
  // 找出变化位，逐位触发滚动
  const len = Math.max(next.length, prevDigits.length)
  // 重新对齐（新旧长度可能不同，从右对齐）
  const pad = (arr, l) => Array(l - arr.length).fill('').concat(arr)
  const nextPad = pad([...next], len)
  const prevPad = pad([...prevDigits], len)

  // 动画状态数组跟随新长度
  animStates.value = nextPad.map(() => 'idle')
  animKeys.value   = nextPad.map(() => 0)

  nextPad.forEach((ch, i) => {
    if (ch !== prevPad[i]) {
      animKeys.value[i]++
    }
  })

  prevDigits = [...next]
}, { deep: true })
</script>

<template>
  <div class="roll-number">
    <span
      v-for="(ch, i) in digits"
      :key="i + '_' + animKeys[i]"
      class="roll-digit"
      :class="{
        'roll-sep': ch === ',',
        'roll-animate': animKeys[i] > 0
      }"
    >{{ ch }}</span>
  </div>
</template>

<style scoped>
.roll-number {
  display: flex;
  align-items: baseline;
  justify-content: flex-end;
  gap: 0;
  overflow: hidden;
  line-height: 1;
}

.roll-digit {
  display: inline-block;
  line-height: inherit;
}

.roll-sep {
  margin: 0 1px;
  opacity: 0.5;
  font-size: 0.75em;
}

/* 触发动画：旧值上飞出，新值从下飞入 */
@keyframes roll-in {
  0%   { transform: translateY(60%); opacity: 0; }
  100% { transform: translateY(0%);  opacity: 1; }
}

.roll-animate {
  animation: roll-in 0.32s cubic-bezier(0.22, 1, 0.36, 1) both;
}
</style>
