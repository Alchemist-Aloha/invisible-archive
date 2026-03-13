<script setup lang="ts">
import { ref, computed, onMounted, watch, onUnmounted } from 'vue';
import { useQuery } from '@tanstack/vue-query';
import { useSwipe } from '@vueuse/core';
import { 
  Search, 
  Loader2, 
  X, 
  AlertCircle, 
  ChevronLeft, 
  ChevronRight, 
  Download, 
  Info, 
  FileText 
} from 'lucide-vue-next';
import { fetchList, searchFiles, getRawUrl, getThumbUrl, fetchText, CAP_STREAM, CAP_RENDER, CAP_EDIT, CAP_BROWSE } from './api';
import type { FileItem } from './api';
import Breadcrumbs from './components/Breadcrumbs.vue';
import FileGrid from './components/FileGrid.vue';
import Plyr from 'plyr';
import 'plyr/dist/plyr.css';
import PhotoSwipeLightbox from 'photoswipe/lightbox';
import 'photoswipe/style.css';

// Path state with URL and localStorage sync
const getInitialPath = () => {
  const hash = window.location.hash.slice(1);
  if (hash) {
    const decoded = decodeURIComponent(hash);
    return decoded.startsWith('/') ? decoded : '/' + decoded;
  }
  return localStorage.getItem('lastPath') || '/';
};

const currentPath = ref(getInitialPath());
const searchQuery = ref('');
const isSearching = ref(false);
const previewItem = ref<FileItem | null>(null);
const showInfo = ref(false);
const videoElement = ref<HTMLVideoElement | null>(null);
const previewStage = ref<HTMLElement | null>(null);
const transitionName = ref('slide-next');
let player: Plyr | null = null;
const lightbox = ref<PhotoSwipeLightbox | null>(null);

// Persistence and Navigation
const handleNavigate = (path: string, pushState = true) => {
  searchQuery.value = '';
  isSearching.value = false;
  currentPath.value = path;
  localStorage.setItem('lastPath', path);
  
  if (pushState) {
    history.pushState({ path }, '', '#' + encodeURIComponent(path));
  }
};

// Handle browser back/forward and mobile swipe gestures
onMounted(() => {
  // Initialize PhotoSwipe for robust pinch-zoom and dynamic scroll
  lightbox.value = new PhotoSwipeLightbox({
    gallery: '#file-grid',
    children: '[data-pswp-src]',
    pswpModule: () => import('photoswipe'),
    bgOpacity: 0.98,
    padding: { top: 20, bottom: 20, left: 20, right: 20 },
    // Robust zoom settings
    initialZoomLevel: 'fit',
    secondaryZoomLevel: 1.5,
    maxZoomLevel: 8,
    wheelToZoom: true,
    // Prevent accidental closes during complex gestures
    pinchToClose: false,
    closeOnVerticalDrag: false,
  });

  // Dynamically fix aspect ratio when image loads
  lightbox.value.on('gettingData', (event) => {
    const { data } = event;
    if (data.src && (!data.width || data.width === 1)) {
      // Set temporary large dimensions to allow zooming immediately
      data.width = 3000;
      data.height = 2000;

      const img = new Image();
      img.src = data.src;
      img.onload = () => {
        if (img.width && img.height) {
          const changed = data.width !== img.width || data.height !== img.height;
          data.width = img.width;
          data.height = img.height;
          
          // Only refresh if dimensions changed and the gallery is active
          if (changed && lightbox.value?.pswp) {
            lightbox.value.pswp.refreshSlideContent(event.index);
          }
        }
      };
    }
  });

  lightbox.value.init();

  // Handle closing via browser Back button/gesture
  window.addEventListener('popstate', (event) => {
    // 1. Handle PhotoSwipe
    if (lightbox.value?.pswp) {
      lightbox.value.pswp.close();
      return;
    }

    // 2. Handle Custom Preview (Videos/Text)
    if (previewItem.value) {
      closePreview(false); // Close without triggering history.back()
      return;
    }

    const path = event.state?.path || window.location.hash.slice(1) || '/';
    handleNavigate(decodeURIComponent(path), false);
  });

  // Sync history when PhotoSwipe is closed via UI/Esc
  lightbox.value.on('close', () => {
    if (history.state?.pswp) {
      history.back();
    }
  });

  window.addEventListener('keydown', handleKeydown);

  // Build history stack if we started at a deep path
  // This allows mobile "back" swipe to move to parent instead of closing the app
  const startPath = currentPath.value;
  if (startPath !== '/') {
    const segments = startPath.split('/').filter(Boolean);
    let cumulative = '';
    
    // Replace current entry with root
    history.replaceState({ path: '/' }, '', '#/');
    
    // Push parents
    for (let i = 0; i < segments.length - 1; i++) {
      cumulative += '/' + segments[i];
      history.pushState({ path: cumulative }, '', '#' + encodeURIComponent(cumulative));
    }
    
    // Push the actual current path
    history.pushState({ path: startPath }, '', '#' + encodeURIComponent(startPath));
  } else {
    // Just ensure the hash is correct for root
    if (!window.location.hash) {
      history.replaceState({ path: '/' }, '', '#/');
    }
  }
});

// Preload Engine: Fetch upcoming images in the background
watch(previewItem, (item) => {
  if (!item || !displayItems.value) return;
  
  const list = displayItems.value;
  const currentIndex = list.findIndex(i => i.path === item.path);
  if (currentIndex === -1) return;

  // Preload next 2 and previous 1 images
  const indicesToPreload = [
    (currentIndex + 1) % list.length,
    (currentIndex + 2) % list.length,
    (currentIndex - 1 + list.length) % list.length
  ];

  indicesToPreload.forEach(idx => {
    const nextItem = list[idx];
    if (nextItem && (nextItem.capabilities & CAP_RENDER)) {
      const img = new Image();
      img.src = getRawUrl(nextItem.path);
    }
  });
});

// Swipe navigation
useSwipe(previewStage, {
  onSwipeEnd(_e, direction) {
    if (direction === 'left') navigatePreview('next');
    if (direction === 'right') navigatePreview('prev');
  },
});

const { data: items, isLoading, error, refetch } = useQuery({
  queryKey: ['files', currentPath],
  queryFn: () => fetchList(currentPath.value),
  enabled: !isSearching.value,
  retry: 1,
});

const goBack = () => {
  if (currentPath.value === '/') return;
  const parts = currentPath.value.split('/');
  parts.pop();
  handleNavigate(parts.join('/') || '/');
};

const handleSearch = async () => {
  if (!searchQuery.value) {
    isSearching.value = false;
    refetch();
    return;
  }
  isSearching.value = true;
};

const { data: searchResults, isLoading: isSearchLoading } = useQuery({
  queryKey: ['search', searchQuery],
  queryFn: () => searchFiles(searchQuery.value),
  enabled: isSearching,
});

const displayItems = computed(() => isSearching.value ? searchResults.value : items.value);
const showLoading = computed(() => isLoading.value || (isSearching.value && isSearchLoading.value));

// Text Content Query
const { data: textContent, isLoading: isTextLoading } = useQuery({
  queryKey: ['text', computed(() => previewItem.value?.path)],
  queryFn: () => fetchText(previewItem.value!.path),
  enabled: computed(() => !!previewItem.value && (previewItem.value.capabilities & CAP_EDIT) !== 0),
});

const handlePreview = (item: FileItem) => {
  // If it's an image (and not a PDF), PhotoSwipe handles it.
  if ((item.capabilities & CAP_RENDER) && !item.name.toLowerCase().endsWith('.pdf')) {
    if (lightbox.value && displayItems.value) {
      // Find all items that are images
      const imageItems = displayItems.value.filter(i => (i.capabilities & CAP_RENDER) && !i.name.toLowerCase().endsWith('.pdf'));
      const index = imageItems.findIndex(i => i.path === item.path);
      
      if (index !== -1) {
        // Map all image items to PhotoSwipe data source format
        const dataSource = imageItems.map(i => ({
          src: getRawUrl(i.path),
          // We omit width/height here; gettingData event fixes them on load
          alt: i.name,
          msrc: getThumbUrl(i.path), // Provide thumbnail for faster initial display
          element: document.querySelector(`[data-pswp-src="${getRawUrl(i.path)}"]`) as HTMLElement || undefined
        }));

        // Push a state so "Back" button closes the gallery
        history.pushState({ pswp: true }, '');
        lightbox.value.loadAndOpen(index, dataSource);
      }
    }
    return;
  }
  
  // Custom Preview (Video/Text)
  history.pushState({ customPreview: true }, '');
  previewItem.value = item;
  showInfo.value = false;
};

const closePreview = (triggerBack = true) => {
  if (player) {
    player.destroy();
    player = null;
  }
  
  // If we closed via UI/Esc and there's a history entry to pop
  if (triggerBack && history.state?.customPreview) {
    history.back();
  }
  
  previewItem.value = null;
};

// Navigation logic for preview
const navigatePreview = (direction: 'prev' | 'next') => {
  if (!previewItem.value || !displayItems.value) return;
  
  transitionName.value = direction === 'next' ? 'slide-next' : 'slide-prev';
  
  const list = displayItems.value;
  const currentIndex = list.findIndex(item => item.path === previewItem.value!.path);
  if (currentIndex === -1) return;

  const step = direction === 'next' ? 1 : -1;
  let nextIndex = currentIndex;
  
  // Find next non-directory item
  for (let i = 0; i < list.length; i++) {
    nextIndex = (nextIndex + step + list.length) % list.length;
    const item = list[nextIndex];
    // Skip folders and things that should be browsed (archives)
    if (!item.is_dir && !(item.capabilities & CAP_BROWSE)) {
      handlePreview(item);
      return;
    }
  }
};

watch(previewItem, (newItem, oldItem) => {
  // If we changed items while previewing, clean up old player
  if (player && newItem?.path !== oldItem?.path) {
    player.destroy();
    player = null;
  }

  if (newItem && (newItem.capabilities & CAP_STREAM)) {
    setTimeout(() => {
      if (videoElement.value) {
        player = new Plyr(videoElement.value, {
          autoplay: true,
          hideControls: true,
        });
      }
    }, 100);
  }
});

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

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown);
  if (lightbox.value) {
    lightbox.value.destroy();
    lightbox.value = null;
  }
});
</script>

<template>
  <div class="flex flex-col h-screen bg-[#f8fafc] overflow-hidden text-slate-900 font-sans selection:bg-blue-100">
    <!-- Modern Header -->
    <header class="flex items-center justify-between px-4 sm:px-8 py-4 bg-white/80 backdrop-blur-xl border-b border-slate-200/60 sticky top-0 z-30 shadow-[0_1px_2px_rgba(0,0,0,0.02)]">
      <div class="flex items-center gap-2 sm:gap-4 shrink-0">
        <button
          @click="handleNavigate('/')"
          class="w-9 h-9 sm:w-10 sm:h-10 bg-gradient-to-tr from-blue-600 to-blue-500 rounded-xl flex items-center justify-center shadow-lg shadow-blue-500/20 cursor-pointer hover:scale-105 active:scale-95 transition-transform focus-visible:ring-2 focus-visible:ring-blue-500/50 outline-none"
          aria-label="Go to root directory"
        >
          <div class="w-4 h-4 sm:w-5 sm:h-5 border-[2.5px] border-white rounded-[3px] rotate-45"></div>
        </button>
        <div class="hidden min-[480px]:block">
          <h1 class="text-base sm:text-lg font-extrabold tracking-tight text-slate-800 leading-none">
            Archive
          </h1>
          <p class="text-[10px] font-bold text-slate-400 uppercase tracking-widest mt-0.5">Invisible</p>
        </div>
        <div v-if="displayItems" class="min-[480px]:hidden flex flex-col justify-center">
          <span class="text-[10px] font-black text-blue-600 bg-blue-50 px-1.5 py-0.5 rounded-md leading-none">{{ displayItems.length }}</span>
        </div>
      </div>
      
      <div class="relative w-full max-w-lg mx-4 group">
        <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 group-focus-within:text-blue-500 transition-colors" />
        <input 
          v-model="searchQuery"
          @keyup.enter="handleSearch"
          type="text" 
          placeholder="Search items..."
          class="w-full pl-10 pr-10 py-2.5 bg-slate-100 border border-transparent focus:bg-white focus:border-blue-500/30 focus:ring-4 focus:ring-blue-500/5 rounded-2xl text-sm transition-all outline-none placeholder:text-slate-400"
        >
        <button 
          v-if="searchQuery" 
          @click="searchQuery = ''; handleSearch()"
          class="absolute right-3 top-1/2 -translate-y-1/2 p-1.5 hover:bg-slate-200 rounded-lg text-slate-400 transition-colors focus-visible:ring-2 focus-visible:ring-blue-500/50 outline-none"
          aria-label="Clear search"
        >
          <X class="w-3.5 h-3.5" />
        </button>
      </div>
      
      <div class="flex items-center gap-3 shrink-0">
        <div class="hidden sm:flex flex-col items-end">
          <div class="flex items-center gap-1.5 px-2.5 py-1 bg-green-50 text-[10px] font-bold text-green-600 rounded-full border border-green-100">
            <span class="w-1.5 h-1.5 bg-green-500 rounded-full animate-pulse"></span>
            CONNECTED
          </div>
        </div>
        <button class="sm:hidden p-2 hover:bg-slate-100 rounded-xl text-slate-500 focus-visible:ring-2 focus-visible:ring-blue-500/50 outline-none" aria-label="Information">
          <Info class="w-5 h-5" />
        </button>
      </div>
    </header>

    <!-- Contextual Actions Bar -->
    <div class="flex items-center px-4 sm:px-8 py-2 bg-white border-b border-slate-200/60 z-20">
      <button 
        v-if="currentPath !== '/'"
        @click="goBack"
        class="mr-2 p-1.5 hover:bg-slate-100 rounded-lg text-slate-500 transition-colors group focus-visible:ring-2 focus-visible:ring-blue-500/50 outline-none"
        title="Go Back"
        aria-label="Go back"
      >
        <ChevronLeft class="w-5 h-5 group-active:-translate-x-1 transition-transform" />
      </button>
      <Breadcrumbs :path="currentPath" @navigate="handleNavigate" class="flex-1 border-none bg-transparent shadow-none px-0" />
    </div>

    <!-- Main Content Area -->
    <main class="flex-1 relative overflow-hidden flex flex-col">
      <!-- State Transitions Overlay -->
      <transition name="fade">
        <div v-if="showLoading" class="absolute inset-0 flex items-center justify-center bg-white/40 backdrop-blur-[1px] z-10">
          <div class="flex flex-col items-center gap-4 p-8 bg-white/80 rounded-3xl shadow-xl border border-white">
            <div class="relative">
              <div class="w-12 h-12 border-4 border-blue-100 rounded-full"></div>
              <div class="w-12 h-12 border-4 border-t-blue-600 rounded-full animate-spin absolute top-0"></div>
            </div>
            <p class="text-xs font-bold text-slate-500 uppercase tracking-widest">Accessing VFS</p>
          </div>
        </div>
      </transition>

      <!-- Connection Error UI -->
      <div v-if="error" class="flex-1 flex items-center justify-center p-6 bg-slate-50">
        <div class="max-w-md w-full p-8 bg-white rounded-[32px] shadow-2xl shadow-slate-200/50 border border-slate-100 text-center transform transition-all duration-500 hover:scale-[1.01]">
          <div class="w-20 h-20 bg-rose-50 rounded-3xl flex items-center justify-center mx-auto mb-6">
            <AlertCircle class="w-10 h-10 text-rose-500" />
          </div>
          <h2 class="text-xl font-black text-slate-800 mb-3">Service Unavailable</h2>
          <p class="text-slate-500 mb-8 text-sm leading-relaxed">The archive engine is currently unreachable. This might be due to a connection drop or server maintenance.</p>
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
          @navigate="handleNavigate"
          @preview="handlePreview"
        />
        
        <!-- Empty State UI -->
        <div v-else-if="!showLoading && !error" class="h-full flex flex-col items-center justify-center p-12 text-center">
          <div class="w-32 h-32 bg-slate-100 rounded-[40px] flex items-center justify-center mb-8 relative">
            <Search class="w-12 h-12 text-slate-300" />
            <div class="absolute -bottom-2 -right-2 w-12 h-12 bg-white rounded-2xl shadow-sm flex items-center justify-center border border-slate-100">
              <div class="w-6 h-1 bg-slate-200 rounded-full"></div>
            </div>
          </div>
          <h3 class="text-lg font-bold text-slate-700 mb-2">No items found</h3>
          <p class="text-slate-400 text-sm max-w-[240px] leading-relaxed">We couldn't find anything matching your request in this directory.</p>
          <button 
            @click="handleNavigate('/')"
            class="mt-8 px-6 py-2.5 bg-white border border-slate-200 text-slate-600 rounded-xl text-xs font-bold hover:bg-slate-50 transition-colors"
          >
            Back to Library
          </button>
        </div>
      </div>
    </main>

    <!-- Immersive Media Preview -->
    <transition name="preview-zoom">
      <div 
        v-if="previewItem" 
        class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/95 backdrop-blur-2xl"
        @click="closePreview()"
      >
        <!-- Navigation Buttons -->
        <button 
          @click.stop="navigatePreview('prev')"
          class="hidden sm:flex absolute left-4 top-1/2 -translate-y-1/2 w-14 h-14 items-center justify-center bg-white/5 hover:bg-white/10 text-white rounded-full transition-all z-50 border border-white/5 focus-visible:ring-2 focus-visible:ring-blue-500/50 outline-none"
          aria-label="Previous item"
        >
          <ChevronLeft class="w-8 h-8" />
        </button>
        <button 
          @click.stop="navigatePreview('next')"
          class="hidden sm:flex absolute right-4 top-1/2 -translate-y-1/2 w-14 h-14 items-center justify-center bg-white/5 hover:bg-white/10 text-white rounded-full transition-all z-50 border border-white/5 focus-visible:ring-2 focus-visible:ring-blue-500/50 outline-none"
          aria-label="Next item"
        >
          <ChevronRight class="w-8 h-8" />
        </button>

        <div class="w-full h-full flex flex-col sm:p-6" @click.stop>
          <!-- Preview Top Bar -->
          <div class="flex items-center justify-between p-4 sm:px-2">
            <div class="flex items-center gap-3 max-w-[60%] sm:max-w-[80%]">
              <div class="p-2 bg-white/10 rounded-lg text-white">
                <FileIcon :name="previewItem.name" :isDir="false" :capabilities="previewItem.capabilities" class="w-5 h-5" />
              </div>
              <div class="truncate text-left">
                <h4 class="text-white text-sm font-bold truncate">{{ previewItem.name }}</h4>
                <p class="text-[10px] text-slate-400 font-bold uppercase tracking-wider">{{ (previewItem.size / 1024 / 1024).toFixed(2) }} MB</p>
              </div>
            </div>
            
            <div class="flex items-center gap-2">
              <a 
                :href="getRawUrl(previewItem.path)" 
                download
                class="hidden sm:flex items-center gap-2 px-4 py-2 bg-white/10 hover:bg-white/20 text-white rounded-xl text-xs font-bold transition-all"
              >
                <Download class="w-4 h-4" />
                Raw File
              </a>
              <button 
                @click="closePreview()"
                class="p-2.5 bg-white/10 hover:bg-rose-500 text-white rounded-xl transition-all focus-visible:ring-2 focus-visible:ring-blue-500/50 outline-none"
                aria-label="Close preview"
              >
                <X class="w-5 h-5" />
              </button>
            </div>
          </div>

          <!-- Main Stage -->
          <div ref="previewStage" class="flex-1 flex items-center justify-center p-2 sm:p-4 overflow-hidden relative">
            <Transition :name="transitionName" mode="out-in">
              <!-- Image Stage -->
              <div v-if="previewItem.capabilities & CAP_RENDER" :key="previewItem.path" class="w-full h-full flex items-center justify-center">
                <img 
                  :src="getRawUrl(previewItem.path)"
                  class="max-w-full max-h-full object-contain shadow-[0_32px_64px_rgba(0,0,0,0.5)] rounded-sm transition-opacity duration-300"
                >
              </div>

              <!-- Video Stage -->
              <div v-else-if="previewItem.capabilities & CAP_STREAM" :key="previewItem.path + '-v'" class="w-full max-w-5xl rounded-2xl overflow-hidden shadow-2xl bg-black aspect-video">
                <video 
                  ref="videoElement"
                  playsinline 
                  controls
                  class="w-full h-full"
                >
                  <source :src="getRawUrl(previewItem.path)" type="video/mp4" />
                </video>
              </div>

              <!-- Text Stage -->
              <div v-else-if="previewItem.capabilities & CAP_EDIT" :key="previewItem.path + '-t'" class="w-full max-w-5xl h-full bg-slate-900 rounded-2xl border border-white/10 overflow-hidden flex flex-col shadow-2xl">
                <div class="px-4 py-3 bg-white/5 border-b border-white/10 flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <FileText class="w-4 h-4 text-slate-400" />
                    <span class="text-xs font-bold text-slate-300 uppercase tracking-widest">Document Preview</span>
                  </div>
                  <div v-if="isTextLoading" class="flex items-center gap-2">
                    <Loader2 class="w-3 h-3 text-blue-500 animate-spin" />
                    <span class="text-[10px] text-slate-500 font-bold uppercase tracking-tighter">Loading content...</span>
                  </div>
                </div>
                <div class="flex-1 overflow-auto p-6 sm:p-10 font-mono text-sm leading-relaxed text-slate-300 selection:bg-blue-500/30">
                  <pre v-if="textContent" class="whitespace-pre-wrap break-all text-left">{{ textContent }}</pre>
                  <div v-else-if="isTextLoading" class="h-full flex items-center justify-center">
                    <div class="space-y-4 w-full max-w-md">
                      <div class="h-4 bg-white/5 rounded-full w-3/4 animate-pulse"></div>
                      <div class="h-4 bg-white/5 rounded-full w-full animate-pulse"></div>
                      <div class="h-4 bg-white/5 rounded-full w-5/6 animate-pulse"></div>
                      <div class="h-4 bg-white/5 rounded-full w-2/3 animate-pulse"></div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Fallback Stage -->
              <div v-else :key="'fallback'" class="p-12 bg-white/5 rounded-3xl border border-white/10 text-center max-w-sm">
                <div class="w-20 h-20 bg-white/10 rounded-2xl flex items-center justify-center mx-auto mb-6">
                  <AlertCircle class="w-10 h-10 text-slate-400" />
                </div>
                <h3 class="text-white font-bold text-lg mb-2">No Preview</h3>
                <p class="text-slate-400 text-xs mb-8 leading-relaxed">This file type requires external software to view. You can download the raw data below.</p>
                <a 
                  :href="getRawUrl(previewItem.path)" 
                  class="inline-flex items-center gap-2 px-8 py-3 bg-blue-600 hover:bg-blue-500 text-white rounded-2xl text-xs font-bold transition-all shadow-lg shadow-blue-600/20"
                >
                  Download Data
                </a>
              </div>
            </Transition>
          </div>

          <!-- Mobile Actions -->
          <div class="sm:hidden p-4 grid grid-cols-2 gap-3">
            <button 
              @click.stop="navigatePreview('prev')"
              class="flex items-center justify-center gap-2 py-4 bg-white/10 text-white rounded-2xl text-sm font-bold focus-visible:ring-2 focus-visible:ring-blue-500/50 outline-none"
              aria-label="Previous item"
            >
              <ChevronLeft class="w-5 h-5" />
              Prev
            </button>
            <button 
              @click.stop="navigatePreview('next')"
              class="flex items-center justify-center gap-2 py-4 bg-white/10 text-white rounded-2xl text-sm font-bold focus-visible:ring-2 focus-visible:ring-blue-500/50 outline-none"
              aria-label="Next item"
            >
              Next
              <ChevronRight class="w-5 h-5" />
            </button>
            <a 
              :href="getRawUrl(previewItem.path)" 
              download
              class="col-span-2 flex items-center justify-center gap-2 py-4 bg-blue-600 text-white rounded-2xl text-sm font-bold"
            >
              <Download class="w-5 h-5" />
              Download Raw
            </a>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<style>
.fade-enter-active, .fade-leave-active { transition: opacity 0.4s cubic-bezier(0.4, 0, 0.2, 1); }
.fade-enter-from, .fade-leave-to { opacity: 0; }

.preview-zoom-enter-active, .preview-zoom-leave-active { transition: all 0.4s cubic-bezier(0.16, 1, 0.3, 1); }
.preview-zoom-enter-from { opacity: 0; transform: scale(0.95); }
.preview-zoom-leave-to { opacity: 0; transform: scale(1.05); }

/* Slide Transitions */
.slide-next-enter-active, .slide-next-leave-active,
.slide-prev-enter-active, .slide-prev-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.slide-next-enter-from { opacity: 0; transform: translateX(30px); }
.slide-next-leave-to { opacity: 0; transform: translateX(-30px); }

.slide-prev-enter-from { opacity: 0; transform: translateX(-30px); }
.slide-prev-leave-to { opacity: 0; transform: translateX(30px); }

/* Custom Font Utilities */
@supports (font-variation-settings: normal) {
  :root { font-family: 'Inter var', sans-serif; }
}

/* Plyr Theming */
:root {
  --plyr-color-main: #2563eb;
  --plyr-video-background: transparent;
  --plyr-border-radius: 16px;
}

.plyr--video {
  height: 100%;
}

.no-scrollbar::-webkit-scrollbar { display: none; }
.no-scrollbar { -ms-overflow-style: none; scrollbar-width: none; }
</style>
