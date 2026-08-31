<script setup lang="ts">
defineProps<{
  busy: boolean;
}>();

const emit = defineEmits<{
  choose: [];
}>();
</script>

<template>
  <section
    id="config-drop-zone"
    class="drop-zone"
    data-file-drop-target
    aria-labelledby="drop-title"
  >
    <div class="file-mark" aria-hidden="true">
      <svg viewBox="0 0 32 38" fill="none">
        <path d="M5 1h14l8 8v28H5z" stroke="currentColor" stroke-width="2" />
        <path d="M19 1v8h8M10 17h12M10 23h12M10 29h8" stroke="currentColor" stroke-width="2" />
      </svg>
    </div>
    <h1 id="drop-title">Drop a sing-box config here</h1>
    <p>or</p>
    <button class="primary-button" type="button" :disabled="busy" @click="emit('choose')">
      {{ busy ? "Opening…" : "Choose Config" }}
    </button>
    <span class="file-hint">JSON files only · processed locally</span>
  </section>
</template>

<style scoped>
.drop-zone {
  min-height: 430px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 48px 28px;
  border: 1.5px dashed #bbb8ae;
  border-radius: 16px;
  background: rgb(255 255 255 / 56%);
  text-align: center;
  transition: border-color 120ms ease, background-color 120ms ease;
}

.drop-zone.file-drop-target-active {
  border-color: #275f53;
  background: #eef5f1;
}

.file-mark {
  width: 42px;
  color: #275f53;
  margin-bottom: 8px;
}

.file-mark svg {
  display: block;
  width: 100%;
}

h1 {
  margin: 0;
  color: #252622;
  font-size: 20px;
  font-weight: 620;
  letter-spacing: -0.02em;
}

p {
  margin: -2px 0 0;
  color: #858278;
  font-size: 13px;
}

.file-hint {
  margin-top: 9px;
  color: #8a877e;
  font-size: 12px;
}
</style>
