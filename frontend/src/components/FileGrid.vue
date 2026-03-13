<script setup lang="ts">
import { ref, computed } from 'vue';
import { useVirtualizer } from '@tanstack/vue-virtual';
import { useElementSize } from '@vueuse/core';
import { CAP_BROWSE, CAP_RENDER, getThumbUrl, getRawUrl } from '../api';
import type { FileItem } from '../api';
import FileIcon from './FileIcon.vue';

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

// Helper to truncate long names while preserving extension
const truncateName = (name: string, maxLength: number) => {
  if (name.length <= maxLength) return name;
  const ext = name.includes('.') ? name.split('.').pop() : '';
  const nameWithoutExt = name.includes('.') ? name.substring(0, name.lastIndexOf('.')) : name;
  
  if (ext && ext.length < 10) {
    const charsToShow = maxLength - ext.length - 3;
    if (charsToShow > 0) {
      return nameWithoutExt.substring(0, charsToShow) + '...' + ext;
    }
  }
  return name.substring(0, maxLength - 3) + '...';
};

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
      <span class="ml-8 pl-8 border-l border-slate-200 dark:border-dracula-600 text-slate-300 dark:text-dracula-500 whitespace-nowrap">
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
          <button 
            type="button"
            v-for="item in getRowItems(virtualRow.index)"
            :key="item.path"
            @click="handleItemClick(item)"
            class="flex flex-col items-center p-2 rounded-2xl hover:bg-white dark:hover:bg-dracula-600/50 hover:shadow-xl hover:shadow-blue-500/5 transition-all duration-300 cursor-pointer w-28 sm:w-36 group border border-transparent hover:border-blue-100/50 dark:hover:border-dracula-purple/30 focus-visible:ring-2 focus-visible:ring-blue-500/50 dark:focus-visible:ring-dracula-purple/50 outline-none text-left shrink-0"
            :data-pswp-src="(item.capabilities & CAP_RENDER) && !item.name.toLowerCase().endsWith('.pdf') ? getRawUrl(item.path) : undefined"
          >
            <div class="relative w-20 h-20 sm:w-28 sm:h-28 bg-white dark:bg-dracula-800 rounded-2xl shadow-sm overflow-hidden flex items-center justify-center group-hover:scale-[1.02] group-active:scale-95 transition-transform duration-300 ring-1 ring-black/5 dark:ring-white/5 shrink-0">
              <!-- Thumbnail with fallback -->
              <img 
                v-if="(item.capabilities & CAP_RENDER) && !thumbErrors[item.path]"
                :src="getThumbUrl(item.path)"
                @error="handleThumbError(item.path)"
                class="w-full h-full object-cover"
                loading="lazy"
              >
              <div v-else class="p-4 sm:p-6 w-full h-full flex items-center justify-center bg-gradient-to-br from-gray-50 to-white dark:from-dracula-700 dark:to-dracula-800">
                <FileIcon :name="item.name" :isDir="item.is_dir" :capabilities="item.capabilities" />
              </div>
              
              <!-- Type Badge for Archives -->
              <div v-if="item.name.toLowerCase().endsWith('.zip')" class="absolute top-1.5 right-1.5 px-1.5 py-0.5 bg-amber-500/90 dark:bg-dracula-orange/90 backdrop-blur-sm text-[9px] font-bold text-white rounded shadow-sm uppercase tracking-wider">
                Zip
              </div>

              <!-- Directory Overlay -->
              <div v-if="item.is_dir || (item.capabilities & CAP_BROWSE)" class="absolute inset-0 bg-blue-600/0 dark:bg-dracula-purple/0 group-hover:bg-blue-600/5 dark:group-hover:bg-dracula-purple/5 transition-colors duration-300"></div>
            </div>
            
            <div class="mt-3 w-full px-1 flex flex-col items-center">
              <div class="h-8 sm:h-10 w-full flex items-start justify-center overflow-hidden">
                <span 
                  class="block text-[11px] sm:text-[13px] font-semibold text-gray-700 dark:text-dracula-50 text-center line-clamp-2 break-all leading-tight group-hover:text-blue-600 dark:group-hover:text-dracula-purple transition-colors"
                  :title="item.name"
                >
                  {{ truncateName(item.name, 40) }}
                </span>
              </div>
              <span v-if="!item.is_dir" class="block mt-1 text-[10px] text-gray-400 dark:text-dracula-400 text-center font-medium opacity-0 group-hover:opacity-100 transition-opacity shrink-0">
                {{ (item.size / 1024 / 1024).toFixed(1) }} MB
              </span>
            </div>
          </button>
        </template>

        <!-- List / Details Layout -->
        <template v-else>
          <button 
            type="button"
            v-for="item in getRowItems(virtualRow.index)"
            :key="item.path"
            @click="handleItemClick(item)"
            class="flex items-center px-4 sm:px-8 py-2 hover:bg-white dark:hover:bg-dracula-600/20 transition-all duration-200 cursor-pointer group border-b border-transparent hover:border-slate-100 dark:hover:border-dracula-600/30 focus-visible:ring-2 focus-visible:ring-blue-500/50 dark:focus-visible:ring-dracula-purple/50 outline-none w-full text-left"
            :data-pswp-src="(item.capabilities & CAP_RENDER) && !item.name.toLowerCase().endsWith('.pdf') ? getRawUrl(item.path) : undefined"
          >
            <div class="w-8 h-8 flex-shrink-0 mr-4">
              <FileIcon :name="item.name" :isDir="item.is_dir" :capabilities="item.capabilities" />
            </div>
            
            <div class="flex-1 min-w-0 flex items-center justify-between gap-4">
              <span 
                class="text-sm font-medium text-slate-700 dark:text-dracula-50 truncate group-hover:text-blue-600 dark:group-hover:text-dracula-purple transition-colors"
                :title="item.name"
              >
                {{ item.name }}
              </span>

              <!-- Details specific info -->
              <div v-if="layout === 'details'" class="hidden sm:flex items-center gap-8 text-[11px] font-bold text-slate-400 dark:text-dracula-400 uppercase tracking-wider shrink-0">
                <span class="w-20 text-right">{{ item.is_dir ? '--' : (item.size / 1024 / 1024).toFixed(2) + ' MB' }}</span>
                <span class="w-32 text-right">{{ new Date(item.mod_time * 1000).toLocaleDateString() }}</span>
              </div>
            </div>
          </button>
        </template>
      </div>
    </div>
  </div>
</template>
