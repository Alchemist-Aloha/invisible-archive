<script setup lang="ts">
import { computed } from 'vue';
import { ChevronRight, Home } from 'lucide-vue-next';

const props = defineProps<{
  path: string;
}>();

const emit = defineEmits<{
  (e: 'navigate', path: string): void;
}>();

const segments = computed(() => {
  const p = props.path.startsWith('/') ? props.path.slice(1) : props.path;
  if (p === '' || p === '.') return [];

  const parts = p.split('/');
  return parts.map((name, index) => {
    const segmentPath = '/' + parts.slice(0, index + 1).join('/');
    return {
      name,
      path: segmentPath,
    };
  });
});
</script>

<template>
  <nav class="flex items-center space-x-1 px-3 sm:px-6 py-2 bg-white dark:bg-dracula-900 border-b border-gray-200 dark:border-dracula-800 text-[10px] sm:text-xs overflow-x-auto no-scrollbar z-10 shadow-sm transition-colors">
    <button 
      @click="emit('navigate', '/')"
      class="p-1 sm:p-1.5 rounded-md hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-all text-gray-400 dark:text-dracula-500 hover:text-blue-600 dark:hover:text-blue-400 group focus-visible:ring-2 focus-visible:ring-blue-500/50 outline-none flex-shrink-0"
      title="Root Library"
      aria-label="Root Library"
    >
      <Home class="w-3.5 h-3.5 sm:w-4 h-4 group-hover:fill-blue-50 dark:group-hover:fill-blue-900/30" />
    </button>

    <div v-for="seg in segments" :key="seg.path" class="flex items-center space-x-1 sm:space-x-2 shrink-0">
      <ChevronRight class="w-3 h-3 sm:w-3.5 sm:h-3.5 text-gray-300 dark:text-dracula-700" />
      <button 
        @click="emit('navigate', seg.path)"
        class="px-1.5 sm:px-2.5 py-1 sm:py-1.5 rounded-md hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-all text-gray-600 dark:text-dracula-300 hover:text-blue-600 dark:hover:text-blue-400 font-semibold whitespace-nowrap border border-transparent hover:border-blue-100 dark:hover:border-blue-900/30"
      >
        {{ seg.name }}
      </button>
    </div>
  </nav>
</template>

<style scoped>
.no-scrollbar::-webkit-scrollbar {
  display: none;
}
.no-scrollbar {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
</style>
