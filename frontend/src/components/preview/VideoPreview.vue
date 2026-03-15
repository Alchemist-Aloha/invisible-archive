<script setup lang="ts">
import { ref, watch } from 'vue';
import { getRawUrl } from '../../api';
import type { FileItem } from '../../types';

const props = defineProps<{
  item: FileItem;
  isSeeking: boolean;
  seekDelta: number;
}>();

const videoElement = ref<HTMLVideoElement | null>(null);

const playVideo = () => {
  if (videoElement.value) {
    videoElement.value.play().catch(err => console.warn('Autoplay prevented:', err));
  }
};

watch(() => props.item.path, () => {
  if (videoElement.value) {
    videoElement.value.load();
    playVideo();
  }
});

defineExpose({
  videoElement,
  playVideo
});
</script>

<template>
  <div class="w-full h-full flex items-center justify-center rounded-2xl overflow-hidden shadow-2xl bg-black relative">
    <video
      ref="videoElement"
      playsinline
      controls
      autoplay
      crossorigin="anonymous"
      class="w-full h-full"
    >
      <source :src="getRawUrl(item.path)" />
    </video>
    
    <!-- Seek Overlay -->
    <transition name="fade">
      <div v-if="isSeeking" class="absolute inset-0 flex items-center justify-center pointer-events-none z-10 bg-black/20 backdrop-blur-[2px]">
        <div class="px-8 py-4 bg-black/60 backdrop-blur-xl rounded-3xl border border-white/10 flex flex-col items-center gap-2 shadow-2xl">
          <div class="text-4xl font-black text-white tracking-tighter">
            {{ seekDelta > 0 ? '+' : '' }}{{ Math.round(seekDelta) }}s
          </div>
          <div class="text-sm font-bold text-blue-400 uppercase tracking-widest">
            Seeking
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.4s cubic-bezier(0.4, 0, 0.2, 1); }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
