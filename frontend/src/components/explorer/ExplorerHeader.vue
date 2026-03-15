<script setup lang="ts">
import { 
  Search, 
  X, 
  Moon, 
  Sun, 
  LayoutGrid, 
  List, 
  LayoutList,
  Info
} from 'lucide-vue-next';
import type { LayoutMode } from '../../composables/useLayout';

defineProps<{
  searchQuery: string;
  isDarkMode: boolean;
  layoutMode: LayoutMode;
}>();

const emit = defineEmits<{
  (e: 'update:searchQuery', value: string): void;
  (e: 'search'): void;
  (e: 'toggleDarkMode'): void;
  (e: 'setLayoutMode', mode: LayoutMode): void;
  (e: 'cycleLayout'): void;
  (e: 'navigate', path: string): void;
}>();
</script>

<template>
  <header class="flex items-center justify-between px-3 sm:px-8 py-3 sm:py-4 bg-white/80 dark:bg-dracula-800/90 backdrop-blur-xl border-b border-slate-200/60 dark:border-dracula-600/60 sticky top-0 z-30 shadow-[0_1px_2px_rgba(0,0,0,0.02)]">
    <div class="flex items-center gap-2 sm:gap-4 shrink-0">
      <button
        @click="emit('navigate', '/')"
        class="w-8 h-8 sm:w-10 sm:h-10 bg-gradient-to-tr from-blue-600 to-blue-500 rounded-lg sm:rounded-xl flex items-center justify-center shadow-lg shadow-blue-500/20 cursor-pointer hover:scale-105 active:scale-95 transition-transform focus-visible:ring-2 focus-visible:ring-blue-500/50 outline-none p-1.5 sm:p-2"
        aria-label="Go to root directory"
      >
        <svg viewBox="0 0 512 512" class="w-full h-full" xmlns="http://www.w3.org/2000/svg">
          <path d="M64 128 C 64 100, 80 96, 100 96 H 200 L 240 128 H 412 C 440 128, 448 144, 448 164 V 400 C 448 428, 432 432, 412 432 H 100 C 72 432, 64 416, 64 400 Z" fill="white" fill-opacity="0.9" />
          <rect x="160" y="180" width="192" height="160" rx="8" fill="#3b82f6" fill-opacity="0.2" />
          <rect x="196" y="210" width="120" height="12" rx="4" fill="#3b82f6" fill-opacity="0.8" />
          <rect x="196" y="245" width="120" height="12" rx="4" fill="#3b82f6" fill-opacity="0.8" />
          <rect x="196" y="280" width="120" height="12" rx="4" fill="#3b82f6" fill-opacity="0.8" />
          <path d="M64 200 C 64 172, 80 168, 100 168 H 412 C 440 168, 448 184, 448 204 V 400 C 448 428, 432 432, 412 432 H 100 C 72 432, 64 416, 64 400 Z" fill="white" fill-opacity="0.3" />
        </svg>
      </button>
      <div class="hidden md:block">
        <h1 class="text-base sm:text-lg font-extrabold tracking-tight text-slate-800 dark:text-dracula-200 leading-none">
          Archive
        </h1>
        <p class="text-[10px] font-bold text-slate-400 dark:text-dracula-500 uppercase tracking-widest mt-0.5">Invisible</p>
      </div>
    </div>
    
    <div class="relative flex-1 max-w-lg mx-2 sm:mx-4 group">
      <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-400 dark:text-dracula-500 group-focus-within:text-blue-500 transition-colors" />
      <input 
        :value="searchQuery"
        @input="emit('update:searchQuery', ($event.target as HTMLInputElement).value)"
        @keyup.enter="emit('search')"
        type="text" 
        placeholder="Search..."
        class="w-full pl-9 pr-8 py-2 bg-slate-100 dark:bg-dracula-800 border border-transparent focus:bg-white dark:focus:bg-dracula-700 focus:border-blue-500/30 focus:ring-4 focus:ring-blue-500/5 rounded-xl text-sm transition-all outline-none placeholder:text-slate-400 dark:text-dracula-300"
      >
      <button 
        v-if="searchQuery" 
        @click="emit('update:searchQuery', ''); emit('search')"
        class="absolute right-2 top-1/2 -translate-y-1/2 p-1 hover:bg-slate-200 dark:hover:bg-dracula-700 rounded-lg text-slate-400 transition-colors focus-visible:ring-2 focus-visible:ring-blue-500/50 outline-none"
        aria-label="Clear search"
      >
        <X class="w-3 h-3" />
      </button>
    </div>
    
    <div class="flex items-center gap-1 sm:gap-2 shrink-0">
      <!-- Layout Toggle (Desktop: Full, Mobile: Cycle) -->
      <div class="hidden sm:flex items-center bg-slate-100 dark:bg-dracula-800 p-1 rounded-xl border border-slate-200/50 dark:border-dracula-700/50">
        <button 
          @click="emit('setLayoutMode', 'grid')"
          :class="[
            'p-1.5 rounded-lg transition-all',
            layoutMode === 'grid' ? 'bg-white dark:bg-dracula-700 shadow-sm text-blue-600 dark:text-blue-400' : 'text-slate-400 hover:text-slate-600 dark:hover:text-dracula-200'
          ]"
          title="Grid View"
        >
          <LayoutGrid class="w-4 h-4" />
        </button>
        <button 
          @click="emit('setLayoutMode', 'list')"
          :class="[
            'p-1.5 rounded-lg transition-all',
            layoutMode === 'list' ? 'bg-white dark:bg-dracula-700 shadow-sm text-blue-600 dark:text-blue-400' : 'text-slate-400 hover:text-slate-600 dark:hover:text-dracula-200'
          ]"
          title="List View"
        >
          <List class="w-4 h-4" />
        </button>
        <button 
          @click="emit('setLayoutMode', 'details')"
          :class="[
            'p-1.5 rounded-lg transition-all',
            layoutMode === 'details' ? 'bg-white dark:bg-dracula-700 shadow-sm text-blue-600 dark:text-blue-400' : 'text-slate-400 hover:text-slate-600 dark:hover:text-dracula-200'
          ]"
          title="Details View"
        >
          <LayoutList class="w-4 h-4" />
        </button>
      </div>

      <!-- Mobile Layout Toggle -->
      <button 
        @click="emit('cycleLayout')"
        class="sm:hidden p-2 bg-slate-100 dark:bg-dracula-800 rounded-lg text-slate-500 dark:text-dracula-400 border border-slate-200/50 dark:border-dracula-700/50"
        aria-label="Cycle layout"
      >
        <LayoutGrid v-if="layoutMode === 'grid'" class="w-4 h-4" />
        <List v-else-if="layoutMode === 'list'" class="w-4 h-4" />
        <LayoutList v-else class="w-4 h-4" />
      </button>

      <!-- Theme Toggle -->
      <button 
        @click="emit('toggleDarkMode')"
        class="p-2 sm:p-2.5 bg-slate-100 dark:bg-dracula-800 hover:bg-slate-200 dark:hover:bg-dracula-700 rounded-lg sm:rounded-xl text-slate-500 dark:text-dracula-400 transition-colors border border-slate-200/50 dark:border-dracula-700/50"
        aria-label="Toggle dark mode"
      >
        <Sun v-if="isDarkMode" class="w-4 h-4" />
        <Moon v-else class="w-4 h-4" />
      </button>

      <button class="hidden lg:block p-2.5 hover:bg-slate-100 dark:hover:bg-dracula-800 rounded-xl text-slate-500 dark:text-dracula-400 focus-visible:ring-2 focus-visible:ring-blue-500/50 outline-none" aria-label="Information">
        <Info class="w-5 h-5" />
      </button>
    </div>
  </header>
</template>
