<script setup lang="ts">
import { computed } from 'vue';
import { ChevronRight, Home, ArrowDownAZ, ArrowUpAZ, ArrowDown01, ArrowUp01, Shuffle, SortAsc, SortDesc } from 'lucide-vue-next';

const props = defineProps<{
  path: string;
  sortMode: string;
  isDescending: boolean;
  layoutMode: string;
}>();

const emit = defineEmits<{
  (e: 'navigate', path: string): void;
  (e: 'update:sortMode', mode: string): void;
  (e: 'toggleSortOrder'): void;
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

const sortModes = [
  { id: 'name', label: 'Name', icon: ArrowDownAZ },
  { id: 'natural', label: 'Natural', icon: ArrowDown01 },
  { id: 'size', label: 'Size', icon: ArrowDown01 }, // Using same icon for simplicity
  { id: 'random', label: 'Shuffle', icon: Shuffle },
];
</script>

<template>
  <nav class="flex items-center px-3 sm:px-6 py-2 bg-white dark:bg-dracula-700 border-b border-gray-200 dark:border-dracula-600 text-[10px] sm:text-xs overflow-x-auto no-scrollbar z-10 shadow-sm transition-colors">
    <div class="flex items-center space-x-1 flex-1 min-w-0 overflow-x-auto no-scrollbar">
      <button 
        @click="emit('navigate', '/')"
        class="p-1 sm:p-1.5 rounded-md hover:bg-blue-50 dark:hover:bg-dracula-purple/20 transition-all text-gray-400 dark:text-dracula-200 hover:text-blue-600 dark:hover:text-dracula-purple group focus-visible:ring-2 focus-visible:ring-blue-500/50 dark:focus-visible:ring-dracula-purple/50 outline-none flex-shrink-0"
        title="Root Library"
        aria-label="Root Library"
      >
        <Home class="w-3.5 h-3.5 sm:w-4 h-4 group-hover:fill-blue-50 dark:group-hover:fill-dracula-purple/30" />
      </button>

      <div v-for="seg in segments" :key="seg.path" class="flex items-center space-x-1 sm:space-x-2 shrink-0">
        <ChevronRight class="w-3 h-3 sm:w-3.5 sm:h-3.5 text-gray-300 dark:text-dracula-500" />
        <button 
          @click="emit('navigate', seg.path)"
          class="px-1.5 sm:px-2.5 py-1 sm:py-1.5 rounded-md hover:bg-blue-50 dark:hover:bg-dracula-purple/20 transition-all text-gray-600 dark:text-dracula-100 hover:text-blue-600 dark:hover:text-dracula-purple font-semibold whitespace-nowrap border border-transparent hover:border-blue-100 dark:hover:border-dracula-purple/30 focus-visible:ring-2 focus-visible:ring-blue-500/50 outline-none"
        >
          {{ seg.name }}
        </button>
      </div>
    </div>

    <!-- Sorting Controls -->
    <div v-if="layoutMode !== 'waterfall'" class="flex items-center ml-4 space-x-1 sm:space-x-2 shrink-0 border-l border-gray-100 dark:border-dracula-600 pl-4">
      <select 
        :value="sortMode" 
        @change="(e) => emit('update:sortMode', (e.target as HTMLSelectElement).value)"
        class="bg-transparent border-none text-gray-500 dark:text-dracula-200 focus:ring-0 cursor-pointer hover:text-blue-600 dark:hover:text-dracula-purple transition-colors font-medium outline-none text-[10px] sm:text-xs"
      >
        <option v-for="mode in sortModes" :key="mode.id" :value="mode.id" class="dark:bg-dracula-800">{{ mode.label }}</option>
      </select>

      <button 
        v-if="sortMode !== 'random'"
        @click="emit('toggleSortOrder')"
        class="p-1 sm:p-1.5 rounded-md hover:bg-blue-50 dark:hover:bg-dracula-purple/20 transition-all text-gray-400 dark:text-dracula-200 hover:text-blue-600 dark:hover:text-dracula-purple outline-none"
        :title="isDescending ? 'Sort Descending' : 'Sort Ascending'"
      >
        <SortDesc v-if="isDescending" class="w-3.5 h-3.5 sm:w-4 h-4" />
        <SortAsc v-else class="w-3.5 h-3.5 sm:w-4 h-4" />
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
