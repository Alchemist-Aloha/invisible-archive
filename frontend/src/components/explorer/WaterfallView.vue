<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useInfiniteScroll, useElementSize } from '@vueuse/core';
import { fetchRandom } from '../../api';
import type { FileItem } from '../../types';
import { Loader2, ImageOff, RefreshCw } from 'lucide-vue-next';
import WaterfallItem from './WaterfallItem.vue';

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
  const count = columnCount.value;
  const cols: FileItem[][] = Array.from({ length: count }, () => []);
  
  // Since we don't have heights, we'll use a balanced count-based distribution
  // but we'll try to randomize slightly for variety or just stick to a stable one.
  items.value.forEach((item, index) => {
    cols[index % count].push(item);
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
  <div ref="containerRef" class="h-full flex flex-col overflow-hidden bg-slate-50/50 dark:bg-dracula-700/50">
    <!-- Polished Discovery Header -->
    <div class="flex items-center justify-between px-6 sm:px-10 py-5 border-b border-slate-200/50 dark:border-dracula-600/40 bg-white/40 dark:bg-dracula-800/40 backdrop-blur-xl sticky top-0 z-10">
      <div class="flex items-center gap-3">
        <div class="relative">
          <div class="absolute inset-0 bg-blue-500 blur-md opacity-20 animate-pulse"></div>
          <div class="w-3 h-3 rounded-full bg-blue-500 shadow-lg shadow-blue-500/50 relative"></div>
        </div>
        <div>
          <h2 class="text-xs font-black text-slate-400 dark:text-dracula-400 uppercase tracking-[0.2em]">Discovery</h2>
          <p class="text-[10px] text-slate-400/60 dark:text-dracula-500 font-bold uppercase tracking-widest mt-0.5">Random gems from your library</p>
        </div>
      </div>
      <button 
        @click="loadMore(true)"
        class="flex items-center gap-2 px-3 py-2 bg-white/50 dark:bg-dracula-700/50 hover:bg-white dark:hover:bg-dracula-600 border border-slate-200/50 dark:border-dracula-600/50 rounded-xl text-xs font-bold text-slate-600 dark:text-dracula-200 transition-all active:scale-95 group shadow-sm hover:shadow-md"
      >
        <RefreshCw :class="['w-3.5 h-3.5 group-hover:rotate-180 transition-transform duration-700', isLoading ? 'animate-spin' : '']" />
        <span>Reshuffle</span>
      </button>
    </div>

    <div ref="scrollRef" class="p-4 sm:p-8 lg:p-12 flex-1 overflow-y-auto no-scrollbar" id="file-grid">
      <div v-if="items.length === 0 && !isLoading" class="flex flex-col items-center justify-center py-40 text-slate-400 dark:text-dracula-400 animate-in fade-in slide-in-from-bottom-4 duration-700">
        <div class="relative mb-8">
          <div class="absolute inset-0 bg-slate-200 dark:bg-dracula-600 blur-3xl opacity-20 scale-150"></div>
          <ImageOff class="w-20 h-20 opacity-10 relative" />
        </div>
        <h3 class="text-lg font-black text-slate-800 dark:text-dracula-100 mb-2">The trail is empty</h3>
        <p class="text-sm text-slate-500 dark:text-dracula-400 font-medium opacity-60">No images were found in the selected explorer path.</p>
        <button @click="loadMore(true)" class="mt-8 px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white rounded-2xl text-xs font-black uppercase tracking-widest shadow-xl shadow-blue-500/20 transition-all active:scale-95">
          Try Again
        </button>
      </div>

      <!-- Stable Masonry Layout -->
      <div class="flex gap-4 sm:gap-6 lg:gap-8 items-start max-w-[2000px] mx-auto">
        <div 
          v-for="(columnItems, colIndex) in columns" 
          :key="colIndex"
          class="flex-1 flex flex-col gap-4 sm:gap-6 lg:gap-8"
        >
          <div 
            v-for="(item, itemIndex) in columnItems" 
            :key="item.path"
            class="animate-in fade-in zoom-in-95 duration-700"
            :style="{ animationDelay: `${itemIndex * 50}ms` }"
          >
            <WaterfallItem 
              :item="item" 
              @click="handleItemClick(item)" 
            />
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
