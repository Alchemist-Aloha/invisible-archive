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
  <nav class="flex items-center space-x-2 px-6 py-2.5 bg-white border-b border-gray-200 text-xs overflow-x-auto no-scrollbar z-10 shadow-sm">
    <button 
      @click="emit('navigate', '/')"
      class="p-1.5 rounded-md hover:bg-blue-50 transition-all text-gray-400 hover:text-blue-600 group"
      title="Root Library"
    >
      <Home class="w-4 h-4 group-hover:fill-blue-50" />
    </button>

    <div v-for="seg in segments" :key="seg.path" class="flex items-center space-x-2 shrink-0">
      <ChevronRight class="w-3.5 h-3.5 text-gray-300" />
      <button 
        @click="emit('navigate', seg.path)"
        class="px-2.5 py-1.5 rounded-md hover:bg-blue-50 transition-all text-gray-600 hover:text-blue-600 font-semibold whitespace-nowrap border border-transparent hover:border-blue-100"
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
