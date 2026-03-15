<script setup lang="ts">
import { ref, computed } from 'vue';
import { useVirtualizer } from '@tanstack/vue-virtual';
import { useElementSize } from '@vueuse/core';
import { CAP_BROWSE } from '../../types';
import type { FileItem } from '../../types';
import FileItemGrid from './FileItemGrid.vue';
import FileItemRow from './FileItemRow.vue';

const props = withDefaults(defineProps<{
  items: FileItem[];
  layout?: 'grid' | 'list' | 'details';
}>(), {
  layout: 'grid',
});

const emit = defineEmits<{
  (e: 'navigate', path: string): void;
  (e: 'preview', item: FileItem): void;
}>();

const containerRef = ref<HTMLElement | null>(null);
const { width: containerWidth } = useElementSize(containerRef, { width: 1000, height: 0 });

// Precise column calculation for a balanced grid
const columns = computed(() => {
  if (props.layout !== 'grid') return 1;
  const w = containerWidth.value;
  if (w < 360) return 2;
  if (w < 520) return 3;
  if (w < 768) return 4;
  if (w < 1024) return 5;
  if (w < 1280) return 6;
  if (w < 1536) return 8;
  return 10;
});

const rowCount = computed(() => Math.ceil(props.items.length / columns.value));

const virtualizer = useVirtualizer(
  computed(() => ({
    count: rowCount.value,
    getScrollElement: () => containerRef.value,
    estimateSize: () => props.layout === 'grid' ? 150 : 48,
    overscan: 20,
  }))
);

const getRowItems = (rowIndex: number) => {
  const start = rowIndex * columns.value;
  return props.items.slice(start, start + columns.value);
};

const handleItemClick = (item: FileItem) => {
  if (item.is_dir || (item.capabilities & CAP_BROWSE)) {
    emit('navigate', item.path);
  } else {
    emit('preview', item);
  }
};

const thumbErrors = ref<Record<string, boolean>>({});
const handleThumbError = (path: string) => {
  thumbErrors.value[path] = true;
};
</script>

<template>
  <div 
    ref="containerRef" 
    id="file-grid"
    class="h-full overflow-y-auto overflow-x-hidden bg-gray-50/30 dark:bg-dracula-900/30 relative transition-colors duration-300"
  >
    <!-- Details Layout Header (Embedded) -->
    <div v-if="layout === 'details' && items.length > 0" class="flex items-center px-4 sm:px-8 py-2 sticky top-0 z-20 bg-white/95 dark:bg-dracula-800/95 backdrop-blur border-b border-slate-200 dark:border-dracula-600 text-[10px] font-bold text-slate-400 dark:text-dracula-300 uppercase tracking-wider">
      <span class="w-8 mr-4"></span>
      <span class="flex-1">Name</span>
      <div class="hidden sm:flex items-center gap-8 shrink-0">
        <span class="w-20 text-right">Size</span>
        <span class="w-32 text-right">Modified</span>
      </div>
      <!-- Inline Item Count for Details -->
      <span class="ml-8 pl-8 border-l border-slate-200 dark:border-dracula-600 text-slate-300 dark:text-dracula-50 whitespace-nowrap">
        {{ items.length }} items
      </span>
    </div>

    <!-- Floating Item Count Badge (For non-details layouts) -->
    <div v-if="items.length > 0 && layout !== 'details'" class="fixed bottom-6 right-6 z-30 pointer-events-none">
      <span class="px-3 py-1.5 bg-white/90 dark:bg-dracula-800/90 backdrop-blur-md border border-slate-200 dark:border-dracula-600 rounded-full text-[10px] font-bold text-slate-500 dark:text-dracula-200 shadow-xl shadow-blue-500/10 uppercase tracking-widest pointer-events-auto">
        {{ items.length }} items
      </span>
    </div>

    <div
      :style="{
        height: `${virtualizer.getTotalSize()}px`,
        width: '100%',
        position: 'relative',
      }"
      :class="layout === 'grid' ? 'p-2 sm:p-6' : 'p-0'"
    >
      <div
        v-for="virtualRow in virtualizer.getVirtualItems()"
        :key="virtualRow.index"
        :style="{
          position: 'absolute',
          top: 0,
          left: 0,
          width: '100%',
          height: `${virtualRow.size}px`,
          transform: `translateY(${virtualRow.start}px)`,
        }"
        :class="[
          'flex px-2 sm:px-4',
          layout === 'grid' ? 'gap-4 sm:gap-10 justify-center' : 'flex-col gap-0 justify-start'
        ]"
      >
        <!-- Grid Layout -->
        <template v-if="layout === 'grid'">
          <FileItemGrid
            v-for="item in getRowItems(virtualRow.index)"
            :key="item.path"
            :item="item"
            :thumb-error="thumbErrors[item.path]"
            @click="handleItemClick"
            @thumb-error="handleThumbError"
          />
        </template>

        <!-- List / Details Layout -->
        <template v-else>
          <FileItemRow
            v-for="item in getRowItems(virtualRow.index)"
            :key="item.path"
            :item="item"
            :layout="layout"
            @click="handleItemClick"
          />
        </template>
      </div>
    </div>
  </div>
</template>
