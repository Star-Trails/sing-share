<script setup lang="ts">
const props = defineProps<{
  fps: number;
  sliceSize: number;
  saving: boolean;
}>();

const emit = defineEmits<{
  "update:fps": [value: number];
  "update:sliceSize": [value: number];
  save: [];
  stop: [];
}>();

function updateFPS(event: Event): void {
  const value = Number((event.target as HTMLInputElement).value);
  emit("update:fps", Math.min(30, Math.max(1, value)));
}

function updateSliceSize(event: Event): void {
  const value = Number((event.target as HTMLInputElement).value);
  emit("update:sliceSize", Math.min(500, Math.max(20, value)));
}
</script>

<template>
  <div class="controls">
    <div class="actions">
      <button class="secondary-button" type="button" :disabled="saving" @click="emit('save')">
        {{ saving ? "Saving…" : "Save BPF" }}
      </button>
      <button class="text-button" type="button" @click="emit('stop')">Stop</button>
    </div>

    <div class="parameters">
      <label>
        <span>FPS</span>
        <input
          type="number"
          min="1"
          max="30"
          step="1"
          :value="props.fps"
          aria-label="Frames per second"
          @change="updateFPS"
        />
      </label>
      <label>
        <span>Slice Size</span>
        <input
          type="number"
          min="20"
          max="500"
          step="10"
          :value="props.sliceSize"
          aria-label="QRS slice size"
          @change="updateSliceSize"
        />
      </label>
    </div>
  </div>
</template>

<style scoped>
.controls {
  width: 100%;
  display: grid;
  gap: 16px;
}

.actions,
.parameters {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.parameters {
  padding-top: 14px;
  border-top: 1px solid #dedbd3;
}

label {
  display: flex;
  align-items: center;
  gap: 9px;
  color: #68665e;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.02em;
}

input {
  width: 66px;
  height: 32px;
  padding: 0 8px;
  border: 1px solid #cbc8be;
  border-radius: 7px;
  background: #fff;
  color: #292a26;
  font: inherit;
  font-variant-numeric: tabular-nums;
  text-align: right;
  outline: none;
}

input:focus {
  border-color: #275f53;
  box-shadow: 0 0 0 2px rgb(39 95 83 / 14%);
}
</style>
