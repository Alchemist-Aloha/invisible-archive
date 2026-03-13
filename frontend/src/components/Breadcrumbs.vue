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
  if (props.path === '.' || props.path === '') return [];
  const parts = props.path.split('/');
  return parts.map((name, index) => ({
    name,
    path: parts.slice(0, index + 1).join('/'),
  }));
});
</script>

<template>
  <nav class="flex items-center space-x-2 px-4 py-2 bg-white border-b border-gray-200 text-sm overflow-x-auto no-scrollbar">
    <button 
      @click="emit('navigate', '.')"
      class="p-1 rounded hover:bg-gray-100 transition-colors text-gray-500 hover:text-blue-600"
    >
      <Home class="w-4 h-4" />
    </button>
    
    <div v-for="seg in segments" :key="seg.path" class="flex items-center space-x-2 shrink-0">
      <ChevronRight class="w-4 h-4 text-gray-400" />
      <button 
        @click="emit('navigate', seg.path)"
        class="px-2 py-1 rounded hover:bg-gray-100 transition-colors text-gray-700 hover:text-blue-600 font-medium whitespace-nowrap"
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
