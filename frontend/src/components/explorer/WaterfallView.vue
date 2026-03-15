<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useInfiniteScroll, useElementSize } from '@vueuse/core';
import { fetchRandom, getThumbUrl, getRawUrl } from '../../api';
import type { FileItem } from '../../types';
import { Loader2, ImageOff, RefreshCw } from 'lucide-vue-next';

const props = defineProps<{
  path: string;
}>();

const emit = defineEmits<{
  (e: 'click', item: FileItem, items: FileItem[]): void;
}>();

const items = ref<FileItem[]>([]);
const isLoading = ref(false);
const hasMore = ref(true);
const containerRef = ref<HTMLElement | null>(null);
const scrollRef = ref<HTMLElement | null>(null);
const { width } = useElementSize(containerRef);
const pageSize = 50;

// Dynamic column count based on width
const columnCount = computed(() => {
  if (width.value < 640) return 2;
  if (width.value < 768) return 3;
  if (width.value < 1024) return 4;
  if (width.value < 1280) return 5;
  return 6;
});

// Distribute items into columns for a stable masonry layout
const columns = computed(() => {
  const cols: FileItem[][] = Array.from({ length: columnCount.value }, () => []);
  items.value.forEach((item, index) => {
    cols[index % columnCount.value].push(item);
  });
  return cols;
});

const loadMore = async (reset = false) => {
  if (isLoading.value || (!hasMore.value && !reset)) return;
  
  if (reset) {
    items.value = [];
    hasMore.value = true;
  }

  isLoading.value = true;
  try {
    const newItems = await fetchRandom(props.path, pageSize);
    
    // Deduplicate
    const existingPaths = new Set(items.value.map(i => i.path));
    const uniqueNewItems = newItems.filter(i => !existingPaths.has(i.path));
    
    items.value.push(...uniqueNewItems);
    
    if (newItems.length < pageSize) {
      hasMore.value = false;
    }
  } catch (error) {
    console.error('Failed to fetch random items:', error);
  } finally {
    isLoading.value = false;
  }
};

useInfiniteScroll(
  scrollRef,
  () => {
    loadMore();
  },
  { distance: 600 }
);

onMounted(() => {
  loadMore();
});

// Reset when path changes
watch(() => props.path, () => {
  loadMore(true);
});

const handleItemClick = (item: FileItem) => {
  emit('click', item, items.value);
};
</script>

<template>
  <div ref="containerRef" class="h-full flex flex-col overflow-hidden">
    <!-- Header/Actions -->
    <div class="flex items-center justify-between px-4 sm:px-8 py-4 border-b border-slate-200/40 dark:border-dracula-600/40 bg-white/30 dark:bg-dracula-800/30 backdrop-blur-sm sticky top-0 z-10">
      <div class="flex items-center gap-2">
        <div class="w-2 h-2 rounded-full bg-blue-500 animate-pulse"></div>
        <h2 class="text-sm font-bold text-slate-600 dark:text-dracula-200 uppercase tracking-widest">Discovery</h2>
      </div>
      <button 
        @click="loadMore(true)"
        class="p-2 hover:bg-slate-100 dark:hover:bg-dracula-700 rounded-lg text-slate-500 dark:text-dracula-400 transition-all active:scale-95 group"
        title="Refresh Discovery"
      >
        <RefreshCw class="w-4 h-4 group-hover:rotate-180 transition-transform duration-500" />
      </button>
    </div>

    <div ref="scrollRef" class="p-4 sm:p-6 lg:p-8 flex-1 overflow-y-auto no-scrollbar" id="file-grid">
      <div v-if="items.length === 0 && !isLoading" class="flex flex-col items-center justify-center py-32 text-slate-400 dark:text-dracula-400">
        <div class="relative mb-6">
          <ImageOff class="w-16 h-16 opacity-10" />
          <div class="absolute inset-0 flex items-center justify-center">
            <div class="w-8 h-8 border-2 border-slate-200 dark:border-dracula-600 rounded-lg rotate-12"></div>
          </div>
        </div>
        <p class="font-medium text-slate-500 dark:text-dracula-300">No images found under this directory</p>
        <p class="text-xs mt-2 opacity-60">Try navigating to a different folder or archive</p>
      </div>

      <!-- Stable Masonry Layout -->
      <div class="flex gap-4 items-start">
        <div 
          v-for="(columnItems, colIndex) in columns" 
          :key="colIndex"
          class="flex-1 flex flex-col gap-4"
        >
          <div 
            v-for="item in columnItems" 
            :key="item.path"
            class="animate-in fade-in zoom-in-95 duration-500"
          >
            <button
              type="button"
              @click="handleItemClick(item)"
              class="relative w-full group overflow-hidden rounded-2xl bg-slate-100 dark:bg-dracula-800 border border-slate-200/50 dark:border-dracula-700/50 hover:border-blue-500/50 dark:hover:border-dracula-purple/50 transition-all duration-300 shadow-sm hover:shadow-2xl hover:shadow-blue-500/10 active:scale-[0.98] block"
              :data-pswp-src="getRawUrl(item.path)"
            >
              <!-- Aspect Ratio Placeholder -->
              <div class="relative w-full bg-slate-200 dark:bg-dracula-700 min-h-[100px]">
                <img 
                  :src="getThumbUrl(item.path)"
                  class="w-full h-auto block transition-transform duration-700 group-hover:scale-110"
                  loading="lazy"
                  :alt="item.name"
                >
              </div>
              
              <!-- Refined Hover Overlay -->
              <div class="absolute inset-0 bg-gradient-to-t from-black/80 via-black/20 to-transparent opacity-0 group-hover:opacity-100 transition-all duration-300 flex flex-col justify-end p-4 text-left translate-y-2 group-hover:translate-y-0">
                <p class="text-white text-xs font-bold truncate mb-1">
                  {{ item.name }}
                </p>
                <div class="flex items-center gap-2">
                  <span class="text-[10px] text-white/70 font-medium px-1.5 py-0.5 rounded bg-white/10 backdrop-blur-md">
                    {{ (item.size / 1024 / 1024).toFixed(1) }} MB
                  </span>
                </div>
              </div>
            </button>
          </div>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="isLoading" class="flex flex-col items-center justify-center py-16 gap-4">
        <Loader2 class="w-10 h-10 text-blue-500 animate-spin" />
        <p class="text-xs font-bold text-slate-400 dark:text-dracula-500 uppercase tracking-widest">Finding more gems...</p>
      </div>
      
      <!-- End of results -->
      <div v-if="!hasMore && items.length > 0" class="text-center py-20 text-slate-400 dark:text-dracula-500">
        <div class="flex items-center justify-center gap-4 mb-4 opacity-20">
          <div class="h-px w-12 bg-current"></div>
          <div class="w-2 h-2 rounded-full bg-current"></div>
          <div class="h-px w-12 bg-current"></div>
        </div>
        <p class="text-sm font-bold uppercase tracking-widest">End of the trail</p>
        <button 
          @click="loadMore(true)"
          class="mt-4 text-xs text-blue-500 hover:text-blue-600 font-bold underline underline-offset-4"
        >
          Discover New Random Set
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Ensure smooth transitions */
.animate-in {
  animation-fill-mode: forwards;
}
</style>
