<script setup lang="ts">
import { ref, computed } from 'vue';
import { useVirtualizer } from '@tanstack/vue-virtual';
import { CAP_BROWSE, CAP_RENDER, getThumbUrl } from '../api';
import type { FileItem } from '../api';
import FileIcon from './FileIcon.vue';

const props = defineProps<{
  items: FileItem[];
}>();

const emit = defineEmits<{
  (e: 'navigate', path: string): void;
  (e: 'preview', item: FileItem): void;
}>();

const containerRef = ref<HTMLElement | null>(null);

const columns = ref(6); // Should ideally be responsive

const rowCount = computed(() => Math.ceil(props.items.length / columns.value));

const virtualizer = useVirtualizer({
  count: rowCount.value,
  getScrollElement: () => containerRef.value,
  estimateSize: () => 140, // Height of a row
  overscan: 5,
});

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
</script>

<template>
  <div 
    ref="containerRef" 
    class="h-full overflow-auto p-4"
  >
    <div
      :style="{
        height: `${virtualizer.getTotalSize()}px`,
        width: '100%',
        position: 'relative',
      }"
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
        class="flex gap-4"
      >
        <div 
          v-for="item in getRowItems(virtualRow.index)"
          :key="item.path"
          @click="handleItemClick(item)"
          class="flex flex-col items-center p-2 rounded-lg hover:bg-blue-50 cursor-pointer w-32 group transition-colors"
        >
          <div class="relative p-4 bg-white rounded-xl shadow-sm group-hover:shadow-md transition-shadow border border-gray-100 flex items-center justify-center w-20 h-20 overflow-hidden">
            <img 
              v-if="item.capabilities & CAP_RENDER"
              :src="getThumbUrl(item.path)"
              class="absolute inset-0 w-full h-full object-cover rounded-xl"
              loading="lazy"
            >
            <FileIcon v-else :name="item.name" :isDir="item.is_dir" :capabilities="item.capabilities" />
          </div>
          <span class="mt-2 text-xs font-medium text-gray-700 text-center line-clamp-2 break-all px-1">
            {{ item.name }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>
