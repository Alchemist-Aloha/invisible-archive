<script setup lang="ts">
import { computed } from 'vue';
import { useQuery } from '@tanstack/vue-query';
import { FileText, Loader2 } from 'lucide-vue-next';
import { fetchText } from '../../api';
import type { FileItem } from '../../types';

const props = defineProps<{
  item: FileItem;
}>();

const { data: textContent, isLoading } = useQuery({
  queryKey: ['text', computed(() => props.item.path)],
  queryFn: () => fetchText(props.item.path),
});
</script>

<template>
  <div class="w-full max-w-5xl h-full bg-slate-900 dark:bg-dracula-900 rounded-2xl border border-white/10 dark:border-white/5 overflow-hidden flex flex-col shadow-2xl">
    <div class="px-4 py-3 bg-white/5 border-b border-white/10 flex items-center justify-between">
      <div class="flex items-center gap-2">
        <FileText class="w-4 h-4 text-slate-400 dark:text-dracula-400" />
        <span class="text-xs font-bold text-slate-300 dark:text-dracula-200 uppercase tracking-widest">Document Preview</span>
      </div>
      <div v-if="isLoading" class="flex items-center gap-2">
        <Loader2 class="w-3 h-3 text-blue-500 animate-spin" />
        <span class="text-[10px] text-slate-500 dark:text-dracula-500 font-bold uppercase tracking-tighter">Loading content...</span>
      </div>
    </div>
    <div class="flex-1 overflow-auto p-6 sm:p-10 font-mono text-sm leading-relaxed text-slate-300 dark:text-dracula-200 selection:bg-blue-500/30">
      <pre v-if="textContent" class="whitespace-pre-wrap break-all text-left">{{ textContent }}</pre>
      <div v-else-if="isLoading" class="h-full flex items-center justify-center">
        <div class="space-y-4 w-full max-w-md">
          <div class="h-4 bg-white/5 rounded-full w-3/4 animate-pulse"></div>
          <div class="h-4 bg-white/5 rounded-full w-full animate-pulse"></div>
          <div class="h-4 bg-white/5 rounded-full w-5/6 animate-pulse"></div>
          <div class="h-4 bg-white/5 rounded-full w-2/3 animate-pulse"></div>
        </div>
      </div>
    </div>
  </div>
</template>
