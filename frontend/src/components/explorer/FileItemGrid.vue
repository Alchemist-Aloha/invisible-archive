<script setup lang="ts">
import { CAP_RENDER, CAP_BROWSE } from '../../types';
import type { FileItem } from '../../types';
import { getThumbUrl, getRawUrl } from '../../api';
import FileIcon from '../ui/FileIcon.vue';

defineProps<{
  item: FileItem;
  thumbError?: boolean;
}>();

const emit = defineEmits<{
  (e: 'click', item: FileItem): void;
  (e: 'thumbError', path: string): void;
}>();

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
</script>

<template>
  <button 
    type="button"
    @click="emit('click', item)"
    class="flex flex-col items-center p-2 rounded-2xl hover:bg-white dark:hover:bg-dracula-600/50 hover:shadow-xl hover:shadow-blue-500/5 transition-all duration-300 cursor-pointer w-28 sm:w-36 group border border-transparent hover:border-blue-100/50 dark:hover:border-dracula-purple/30 focus-visible:ring-2 focus-visible:ring-blue-500/50 dark:focus-visible:ring-dracula-purple/50 outline-none text-left shrink-0"
    :data-pswp-src="(item.capabilities & CAP_RENDER) && !item.name.toLowerCase().endsWith('.pdf') ? getRawUrl(item.path) : undefined"
    :aria-label="'Open ' + item.name"
  >
    <div class="relative w-20 h-20 sm:w-28 sm:h-28 bg-white dark:bg-dracula-800 rounded-2xl shadow-sm overflow-hidden flex items-center justify-center group-hover:scale-[1.02] group-active:scale-95 transition-transform duration-300 ring-1 ring-black/5 dark:ring-white/5 shrink-0">
      <!-- Thumbnail with fallback -->
      <img 
        v-if="(item.capabilities & CAP_RENDER) && !thumbError"
        :src="getThumbUrl(item.path)"
        @error="emit('thumbError', item.path)"
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
