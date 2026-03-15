<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue';
import { useQuery } from '@tanstack/vue-query';
import { AlertCircle } from 'lucide-vue-next';
import { fetchList, searchFiles } from './api';
import { useTheme } from './composables/useTheme';
import { useLayout } from './composables/useLayout';
import { useNavigation } from './composables/useNavigation';
import { usePreview } from './composables/usePreview';
import ExplorerHeader from './components/explorer/ExplorerHeader.vue';
import Breadcrumbs from './components/explorer/Breadcrumbs.vue';
import FileGrid from './components/explorer/FileGrid.vue';
import FilePreview from './components/preview/FilePreview.vue';
import 'photoswipe/style.css';

// Logic orchestration
const { isDarkMode, toggleDarkMode, updateTheme } = useTheme();
const { layoutMode, setLayoutMode, cycleLayout } = useLayout();
const { currentPath, handleNavigate, goBack, initHistoryStack } = useNavigation();

const searchQuery = ref('');
const isSearching = ref(false);

const { data: listData, isLoading, error, refetch } = useQuery({
  queryKey: ['files', currentPath],
  queryFn: async () => {
    const res = await fetchList(currentPath.value);
    if (res.effective_path !== currentPath.value && !isSearching.value) {
      handleNavigate(res.effective_path, false);
    }
    return res;
  },
  enabled: computed(() => !isSearching.value),
  retry: 1,
});

const { data: searchResults, isLoading: isSearchLoading } = useQuery({
  queryKey: ['search', searchQuery],
  queryFn: () => searchFiles(searchQuery.value),
  enabled: isSearching,
});

const displayItems = computed(() => isSearching.value ? searchResults.value : listData.value?.items);
const showLoading = computed(() => isLoading.value || (isSearching.value && isSearchLoading.value));

const { 
  previewItem, 
  transitionName, 
  initPhotoSwipe, 
  handlePreview, 
  closePreview, 
  navigatePreview,
  lightbox
} = usePreview(displayItems);

const handleSearch = () => {
  if (!searchQuery.value) {
    isSearching.value = false;
    refetch();
    return;
  }
  isSearching.value = true;
};

// Global handlers
const handleKeydown = (e: KeyboardEvent) => {
  if (!previewItem.value) {
    if (e.key === 'Backspace' && !isSearching.value) goBack();
    return;
  }

  switch (e.key) {
    case 'Escape': closePreview(); break;
    case 'ArrowLeft': navigatePreview('prev'); break;
    case 'ArrowRight': navigatePreview('next'); break;
  }
};

onMounted(() => {
  updateTheme();
  initPhotoSwipe();
  initHistoryStack(currentPath.value);
  window.addEventListener('keydown', handleKeydown);
  
  window.addEventListener('popstate', (event) => {
    if (lightbox.value?.pswp) {
      lightbox.value.pswp.close();
      return;
    }
    if (previewItem.value) {
      closePreview(false);
      return;
    }
    const path = event.state?.path || window.location.hash.slice(1) || '/';
    handleNavigate(decodeURIComponent(path), false);
  });
});

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown);
});

// Watch currentPath to clear search
watch(currentPath, () => {
  searchQuery.value = '';
  isSearching.value = false;
});
</script>

<template>
  <div class="flex flex-col h-screen bg-[#f8fafc] dark:bg-dracula-700 overflow-hidden text-slate-900 dark:text-dracula-50 font-sans selection:bg-blue-100 dark:selection:bg-blue-900/30 transition-colors duration-300">
    <ExplorerHeader
      v-model:searchQuery="searchQuery"
      :isDarkMode="isDarkMode"
      :layoutMode="layoutMode"
      @search="handleSearch"
      @toggleDarkMode="toggleDarkMode"
      @setLayoutMode="setLayoutMode"
      @cycleLayout="cycleLayout"
      @navigate="handleNavigate"
    />

    <!-- Contextual Actions Bar -->
    <div class="flex items-center px-4 sm:px-8 py-2 bg-white dark:bg-dracula-800/50 border-b border-slate-200/60 dark:border-dracula-600/60 z-20 transition-colors">
      <button 
        v-if="currentPath !== '/'"
        @click="goBack"
        class="mr-2 p-1.5 hover:bg-slate-100 dark:hover:bg-dracula-600 rounded-lg text-slate-500 dark:text-dracula-200 transition-colors group focus-visible:ring-2 focus-visible:ring-blue-500/50 outline-none"
        title="Go Back"
        aria-label="Go back"
      >
        <svg class="w-5 h-5 group-active:-translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
      </button>
      <Breadcrumbs :path="currentPath" @navigate="handleNavigate" class="flex-1 border-none bg-transparent shadow-none px-0 dark:text-dracula-200" />
    </div>

    <!-- Main Content Area -->
    <main class="flex-1 relative overflow-hidden flex flex-col">
      <!-- Non-blocking Loading Bar -->
      <div class="absolute top-0 left-0 right-0 h-0.5 sm:h-1 z-30 overflow-hidden pointer-events-none">
        <div 
          v-show="showLoading" 
          class="w-full h-full bg-blue-500 dark:bg-dracula-purple origin-left loading-bar"
        ></div>
      </div>

      <!-- Connection Error UI -->
      <div v-if="error" class="flex-1 flex items-center justify-center p-6 bg-slate-50 dark:bg-dracula-700">
        <div class="max-w-md w-full p-8 bg-white dark:bg-dracula-800 rounded-[32px] shadow-2xl shadow-slate-200/50 dark:shadow-black/50 border border-slate-100 dark:border-dracula-600 text-center transform transition-all duration-500">
          <div class="w-20 h-20 bg-rose-50 dark:bg-rose-900/20 rounded-3xl flex items-center justify-center mx-auto mb-6">
            <AlertCircle class="w-10 h-10 text-rose-500" />
          </div>
          <h2 class="text-xl font-black text-slate-800 dark:text-dracula-100 mb-3">Service Unavailable</h2>
          <p class="text-slate-500 dark:text-dracula-400 mb-8 text-sm leading-relaxed">The archive engine is currently unreachable.</p>
          <button @click="() => refetch()" class="w-full py-4 bg-blue-600 hover:bg-blue-700 active:bg-blue-800 text-white rounded-2xl text-sm font-bold shadow-lg shadow-blue-500/25 transition-all">
            Reconnect to Engine
          </button>
        </div>
      </div>

      <!-- Scrollable Content -->
      <div class="flex-1 min-h-0 overflow-hidden">
        <FileGrid 
          v-if="displayItems && displayItems.length > 0"
          :items="displayItems" 
          :layout="layoutMode"
          @navigate="handleNavigate"
          @preview="handlePreview"
        />
        
        <!-- Empty State UI -->
        <div v-else-if="!showLoading && !error" class="h-full flex flex-col items-center justify-center p-12 text-center">
          <h3 class="text-lg font-bold text-slate-700 dark:text-dracula-200 mb-2">No items found</h3>
          <button 
            @click="handleNavigate('/')"
            class="mt-8 px-6 py-2.5 bg-white dark:bg-dracula-800 border border-slate-200 dark:border-dracula-600 text-slate-600 dark:text-dracula-200 rounded-xl text-xs font-bold hover:bg-slate-50 dark:hover:bg-dracula-700 transition-colors"
          >
            Back to Library
          </button>
        </div>
      </div>
    </main>

    <!-- File Preview -->
    <FilePreview
      v-if="previewItem"
      :item="previewItem"
      :transition-name="transitionName"
      @close="closePreview"
      @prev="navigatePreview('prev')"
      @next="navigatePreview('next')"
    />
  </div>
</template>

<style>
.loading-bar {
  animation: loading 2s infinite ease-in-out;
}

@keyframes loading {
  0% { transform: translateX(-100%); }
  50% { transform: translateX(0); }
  100% { transform: translateX(100%); }
}

@supports (font-variation-settings: normal) {
  :root { font-family: 'Inter var', sans-serif; }
}

.no-scrollbar::-webkit-scrollbar { display: none; }
.no-scrollbar { -ms-overflow-style: none; scrollbar-width: none; }
</style>
