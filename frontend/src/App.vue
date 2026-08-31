<script setup lang="ts">
import { Events } from "@wailsio/runtime";
import { onBeforeUnmount, onMounted, ref, shallowRef } from "vue";
import type { ShareProfile } from "../bindings/sing-share/services/models";
import * as ShareService from "../bindings/sing-share/services/shareservice";
import DropZone from "./components/DropZone.vue";
import QRShare from "./components/QRShare.vue";
import ShareControls from "./components/ShareControls.vue";
import { decodeWailsBytes } from "./lib/qrs";

type ProfileMetadata = Omit<ShareProfile, "data">;

const profile = shallowRef<ProfileMetadata | null>(null);
const bpfBytes = shallowRef<Uint8Array | null>(null);
const fps = ref(10);
const sliceSize = ref(500);
const busy = ref(false);
const saving = ref(false);
const errorMessage = ref("");
const statusMessage = ref("");

let bpfBase64: string | null = null;
let loadSequence = 0;
let removeDropListener: (() => void) | null = null;

function clearProfile(): void {
  loadSequence += 1;
  bpfBytes.value?.fill(0);
  bpfBytes.value = null;
  bpfBase64 = null;
  profile.value = null;
  statusMessage.value = "";
}

function acceptProfile(next: ShareProfile): void {
  if (!next.data) {
    throw new Error("The backend returned an empty BPF profile");
  }
  const decoded = decodeWailsBytes(next.data);
  bpfBytes.value?.fill(0);
  bpfBytes.value = decoded;
  bpfBase64 = next.data;
  profile.value = {
    name: next.name,
    filename: next.filename,
    size: next.size,
  };
  statusMessage.value = "";
}

async function runLoad(loader: () => Promise<ShareProfile | null>): Promise<void> {
  const sequence = ++loadSequence;
  busy.value = true;
  errorMessage.value = "";
  statusMessage.value = "";
  try {
    const next = await loader();
    if (sequence !== loadSequence || !next) {
      return;
    }
    acceptProfile(next);
  } catch (error) {
    if (sequence === loadSequence) {
      errorMessage.value = error instanceof Error ? error.message : String(error);
    }
  } finally {
    if (sequence === loadSequence) {
      busy.value = false;
    }
  }
}

function chooseConfig(): void {
  void runLoad(() => ShareService.OpenConfig());
}

function loadDroppedConfig(path: string): void {
  void runLoad(() => ShareService.LoadConfig(path));
}

async function saveBPF(): Promise<void> {
  if (!profile.value || !bpfBase64) {
    return;
  }
  saving.value = true;
  errorMessage.value = "";
  statusMessage.value = "";
  try {
    const saved = await ShareService.SaveBPF(
      bpfBase64,
      `${profile.value.name}.bpf`,
    );
    if (saved) {
      statusMessage.value = "BPF saved";
    }
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : String(error);
  } finally {
    saving.value = false;
  }
}

function handleQRSFailure(message: string): void {
  errorMessage.value = `Could not generate QRS frames: ${message}`;
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) {
    return `${bytes} B`;
  }
  return `${(bytes / 1024).toFixed(1)} KiB`;
}

onMounted(() => {
  removeDropListener = Events.On("config-dropped", (event) => {
    loadDroppedConfig(event.data);
  });
  void runLoad(() => ShareService.StartupProfile());
});

onBeforeUnmount(() => {
  removeDropListener?.();
  clearProfile();
});
</script>

<template>
  <div class="app-shell">
    <header class="app-header">
      <div class="brand-mark" aria-hidden="true"><span /><span /><span /></div>
      <span class="brand-name">sing-share</span>
      <span class="local-badge">LOCAL</span>
    </header>

    <main>
      <DropZone v-if="!profile" :busy="busy" @choose="chooseConfig" />

      <section
        v-else
        id="active-share"
        class="share-view"
        data-file-drop-target
        aria-labelledby="profile-name"
      >
        <div class="profile-heading">
          <div>
            <h1 id="profile-name">{{ profile.name }}</h1>
            <p>{{ profile.filename }} · {{ formatBytes(profile.size) }} BPF</p>
          </div>
          <div class="stream-indicator"><span /> LIVE</div>
        </div>

        <QRShare
          v-if="bpfBytes"
          :bytes="bpfBytes"
          :filename="`${profile.name}.bpf`"
          :fps="fps"
          :slice-size="sliceSize"
          @error="handleQRSFailure"
        />

        <p class="scan-instruction">Scan continuously until transfer completes</p>

        <ShareControls
          :fps="fps"
          :slice-size="sliceSize"
          :saving="saving"
          @update:fps="fps = $event"
          @update:slice-size="sliceSize = $event"
          @save="saveBPF"
          @stop="clearProfile"
        />
      </section>

      <p v-if="errorMessage" class="message error" role="alert">{{ errorMessage }}</p>
      <p v-else-if="statusMessage" class="message success" role="status">{{ statusMessage }}</p>
    </main>

    <footer>
      <span>No uploads. No telemetry.</span>
      <span>QRS fountain transfer</span>
    </footer>
  </div>
</template>

<style scoped>
.app-shell {
  height: 100%;
  display: grid;
  grid-template-rows: auto minmax(0, 1fr) auto;
  overflow: hidden;
}

.app-header {
  height: 62px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 24px;
  border-bottom: 1px solid #dedbd3;
}

.brand-mark {
  width: 24px;
  height: 24px;
  display: grid;
  grid-template-columns: repeat(3, 4px);
  align-items: end;
  justify-content: center;
  gap: 2px;
  padding: 4px;
  border-radius: 6px;
  background: #275f53;
}

.brand-mark span {
  display: block;
  background: #fff;
  border-radius: 1px;
}

.brand-mark span:nth-child(1) { height: 7px; }
.brand-mark span:nth-child(2) { height: 13px; }
.brand-mark span:nth-child(3) { height: 10px; }

.brand-name {
  color: #22231f;
  font-size: 15px;
  font-weight: 680;
  letter-spacing: -0.01em;
}

.local-badge {
  margin-left: auto;
  padding: 3px 6px;
  border: 1px solid #c9c6bc;
  border-radius: 4px;
  color: #747168;
  font-size: 9px;
  font-weight: 700;
  letter-spacing: 0.11em;
}

main {
  min-height: 0;
  padding: 22px 24px 18px;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: transparent transparent;
}

main:hover {
  scrollbar-color: rgb(76 78 72 / 34%) transparent;
}

main::-webkit-scrollbar {
  width: 5px;
}

main::-webkit-scrollbar-track {
  background: transparent;
}

main::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: transparent;
}

main:hover::-webkit-scrollbar-thumb {
  background: rgb(76 78 72 / 28%);
}

main::-webkit-scrollbar-thumb:hover {
  background: rgb(39 95 83 / 55%);
}

.share-view {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.share-view.file-drop-target-active::after {
  content: "Drop to replace profile";
  position: fixed;
  inset: 62px 0 38px;
  z-index: 5;
  display: grid;
  place-items: center;
  background: rgb(238 245 241 / 94%);
  color: #275f53;
  font-size: 16px;
  font-weight: 650;
}

.profile-heading {
  width: 100%;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}

.profile-heading h1 {
  margin: 0;
  color: #242520;
  font-size: 21px;
  font-weight: 680;
  letter-spacing: -0.025em;
}

.profile-heading p {
  margin: 4px 0 0;
  color: #77746b;
  font-size: 12px;
}

.stream-indicator {
  display: flex;
  align-items: center;
  gap: 5px;
  padding-top: 5px;
  color: #427468;
  font-size: 9px;
  font-weight: 750;
  letter-spacing: 0.09em;
}

.stream-indicator span {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #3a8a75;
}

.scan-instruction {
  margin: -2px 0 2px;
  color: #68665e;
  font-size: 12px;
}

.message {
  margin: 12px 0 0;
  padding: 9px 11px;
  border-radius: 7px;
  font-size: 12px;
}

.error {
  border: 1px solid #dfc4be;
  background: #f9efed;
  color: #8b352a;
}

.success {
  border: 1px solid #bfd6cc;
  background: #edf6f2;
  color: #285f52;
}

footer {
  height: 38px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  border-top: 1px solid #e1ded6;
  color: #8a877e;
  font-size: 10px;
}
</style>
