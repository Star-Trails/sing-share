<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from "vue";
import { createQRSFrameStream, type QRSFrameStream } from "../lib/qrs";

const props = defineProps<{
  bytes: Uint8Array;
  filename: string;
  fps: number;
  sliceSize: number;
}>();

const emit = defineEmits<{
  error: [message: string];
}>();

const frameElement = ref<HTMLElement | null>(null);
let frameStream: QRSFrameStream | null = null;
let timer: number | null = null;

function stopScheduler(): void {
  if (timer !== null) {
    window.clearTimeout(timer);
    timer = null;
  }
}

function releaseStream(): void {
  stopScheduler();
  frameStream?.dispose();
  frameStream = null;
  if (frameElement.value) {
    frameElement.value.replaceChildren();
  }
}

function scheduleNextFrame(): void {
  stopScheduler();
  timer = window.setTimeout(renderFrame, 1000 / props.fps);
}

function renderFrame(): void {
  if (!frameStream || !frameElement.value) {
    return;
  }
  try {
    frameElement.value.innerHTML = frameStream.nextFrame();
    scheduleNextFrame();
  } catch (error) {
    releaseStream();
    emit("error", error instanceof Error ? error.message : String(error));
  }
}

function rebuildStream(): void {
  releaseStream();
  try {
    frameStream = createQRSFrameStream(props.bytes, props.filename, props.sliceSize);
    renderFrame();
  } catch (error) {
    emit("error", error instanceof Error ? error.message : String(error));
  }
}

onMounted(rebuildStream);
onBeforeUnmount(releaseStream);

watch(
  () => [props.bytes, props.filename, props.sliceSize] as const,
  rebuildStream,
);
watch(
  () => props.fps,
  () => {
    if (frameStream) {
      scheduleNextFrame();
    }
  },
);
</script>

<template>
  <div class="qr-shell">
    <div
      ref="frameElement"
      class="qr-frame"
      role="img"
      aria-label="Animated QRS code for the current sing-box profile"
    />
  </div>
</template>

<style scoped>
.qr-shell {
  width: min(100%, 352px, max(240px, calc(100vh - 348px)));
  aspect-ratio: 1;
  display: grid;
  place-items: center;
  padding: 16px;
  border: 1px solid #d8d5cc;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 8px 28px rgb(30 31 28 / 8%);
}

.qr-frame {
  width: 100%;
  height: 100%;
  display: grid;
  place-items: center;
  overflow: hidden;
  background: #fff;
}

.qr-frame :deep(svg) {
  display: block;
  width: 100%;
  height: 100%;
  color: #000;
  shape-rendering: crispEdges;
}
</style>
