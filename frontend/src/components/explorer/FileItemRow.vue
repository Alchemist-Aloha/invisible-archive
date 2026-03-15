<script setup lang="ts">
import { CAP_RENDER } from '../../types';
import type { FileItem } from '../../types';
import { getRawUrl } from '../../api';
import FileIcon from '../ui/FileIcon.vue';

defineProps<{
  item: FileItem;
  layout: 'list' | 'details';
}>();

const emit = defineEmits<{
  (e: 'click', item: FileItem): void;
}>();
</script>

<template>
  <button 
    type="button"
    @click="emit('click', item)"
    class="flex items-center px-4 sm:px-8 py-2 hover:bg-white dark:hover:bg-dracula-600/20 transition-all duration-200 cursor-pointer group border-b border-transparent hover:border-slate-100 dark:hover:border-dracula-600/30 focus-visible:ring-2 focus-visible:ring-blue-500/50 dark:focus-visible:ring-dracula-purple/50 outline-none w-full text-left"
    :data-pswp-src="(item.capabilities & CAP_RENDER) && !item.name.toLowerCase().endsWith('.pdf') ? getRawUrl(item.path) : undefined"
    :aria-label="'Open ' + item.name"
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
