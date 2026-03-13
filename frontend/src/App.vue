<script setup lang="ts">
import { ref, computed } from 'vue';
import { useQuery } from '@tanstack/vue-query';
import { Search, Loader2 } from 'lucide-vue-next';
import { fetchList, searchFiles } from './api';
import type { FileItem } from './api';
import Breadcrumbs from './components/Breadcrumbs.vue';
import FileGrid from './components/FileGrid.vue';

const currentPath = ref('/library');
const searchQuery = ref('');
const isSearching = ref(false);

const { data: items, isLoading, refetch } = useQuery({
  queryKey: ['files', currentPath],
  queryFn: () => fetchList(currentPath.value),
  enabled: !isSearching.value,
});

const handleNavigate = (path: string) => {
  searchQuery.value = '';
  isSearching.value = false;
  currentPath.value = path;
};

const handleSearch = async () => {
  if (!searchQuery.value) {
    isSearching.value = false;
    refetch();
    return;
  }
  isSearching.value = true;
  // In a real app, useQuery would handle search too
};

const { data: searchResults } = useQuery({
  queryKey: ['search', searchQuery],
  queryFn: () => searchFiles(searchQuery.value),
  enabled: isSearching,
});

const displayItems = computed(() => isSearching.value ? searchResults.value : items.value);

const handlePreview = (item: FileItem) => {
  console.log('Previewing:', item);
  // Phase 3 Next: Implement Quick Look Modal
};
</script>

<template>
  <div class="flex flex-col h-screen bg-gray-50 overflow-hidden">
    <!-- Header -->
    <header class="flex items-center justify-between px-6 py-3 bg-white border-b border-gray-200">
      <h1 class="text-xl font-bold bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-transparent">
        Invisible Archive
      </h1>
      
      <div class="relative w-96">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
        <input 
          v-model="searchQuery"
          @keyup.enter="handleSearch"
          type="text" 
          placeholder="Search files and archives..."
          class="w-full pl-10 pr-4 py-2 bg-gray-100 border-transparent focus:bg-white focus:border-blue-500 focus:ring-0 rounded-full text-sm transition-all"
        >
      </div>
      
      <div class="w-20"></div> <!-- Spacer -->
    </header>

    <!-- Navigation -->
    <Breadcrumbs :path="currentPath" @navigate="handleNavigate" />

    <!-- Main Content -->
    <main class="flex-1 relative">
      <div v-if="isLoading" class="absolute inset-0 flex items-center justify-center bg-white/50 backdrop-blur-sm z-10">
        <Loader2 class="w-8 h-8 text-blue-600 animate-spin" />
      </div>

      <FileGrid 
        v-if="displayItems"
        :items="displayItems" 
        @navigate="handleNavigate"
        @preview="handlePreview"
      />
      
      <div v-else-if="!isLoading" class="flex flex-col items-center justify-center h-full text-gray-400">
        <p class="text-lg font-medium">No items found</p>
      </div>
    </main>
  </div>
</template>
