<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue';
import { useIntersectionObserver } from '@vueuse/core';
import { getThumbUrl, getRawUrl } from '../../api';
import type { FileItem } from '../../types';

const props = defineProps<{
  item: FileItem;
}>();

const emit = defineEmits<{
  (e: 'click', item: FileItem): void;
}>();

const containerRef = ref<HTMLElement | null>(null);
const isNearViewport = ref(false);
const isLoaded = ref(false);
const thumbUrl = getThumbUrl(props.item.path);
const rawUrl = getRawUrl(props.item.path);

// Intersection Observer with a healthy margin to pre-load
const { stop } = useIntersectionObserver(
  containerRef,
  ([{ isIntersecting }]) => {
    isNearViewport.value = isIntersecting;
  },
  { rootMargin: '400px' }
);

// Preload original image when near the viewport
watch(isNearViewport, (near) => {
  if (near && !isLoaded.value) {
    const img = new Image();
    img.src = rawUrl;
    img.onload = () => {
      isLoaded.value = true;
    };
  }
});

onUnmounted(() => {
  stop();
});
</script>

<template>
  <div ref="containerRef" class="w-full">
    <button
      type="button"
      @click="emit('click', item)"
      class="relative w-full group overflow-hidden rounded-[24px] bg-white dark:bg-dracula-800 border border-slate-200/60 dark:border-dracula-700/60 hover:border-blue-500 dark:hover:border-dracula-purple transition-all duration-500 shadow-[0_4px_20px_-10px_rgba(0,0,0,0.1)] hover:shadow-[0_20px_50px_-20px_rgba(0,0,0,0.15)] active:scale-[0.97] block"
      :data-pswp-src="rawUrl"
    >
      <!-- Glassmorphism Aspect Ratio Placeholder -->
      <div class="relative w-full bg-slate-100 dark:bg-dracula-900/40 min-h-[120px] overflow-hidden">
        <!-- Always-on Thumbnail (Tier 1) -->
        <img 
          :src="thumbUrl"
          class="w-full h-auto block transition-transform duration-[1.5s] ease-out group-hover:scale-110"
          loading="lazy"
          :alt="item.name"
        >
        
        <!-- High-Res Original (Tier 2 - Swap & Unload) -->
        <transition name="fade">
          <img 
            v-if="isNearViewport && isLoaded"
            :src="rawUrl"
            class="absolute inset-0 w-full h-full object-cover block transition-transform duration-[1.5s] ease-out group-hover:scale-110"
            :alt="item.name"
          >
        </transition>

        <!-- Subtle Gradient Overlay (always visible for contrast) -->
        <div class="absolute inset-0 bg-gradient-to-t from-black/20 to-transparent pointer-events-none"></div>
      </div>
      
      <!-- Polished Hover Overlay -->
      <div class="absolute inset-0 bg-gradient-to-t from-black/90 via-black/30 to-transparent opacity-0 group-hover:opacity-100 transition-all duration-500 flex flex-col justify-end p-5 text-left translate-y-4 group-hover:translate-y-0">
        <p class="text-white text-xs font-black tracking-tight line-clamp-2 mb-2 leading-snug">
          {{ item.name }}
        </p>
        <div class="flex items-center gap-3">
          <span class="text-[9px] text-white/90 font-black uppercase tracking-[0.15em] px-2 py-1 rounded-full bg-white/10 backdrop-blur-xl border border-white/10">
            {{ (item.size / 1024 / 1024).toFixed(1) }} MB
          </span>
        </div>
      </div>
    </button>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.5s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
