<script setup lang="ts">
import { ref, watch } from 'vue';
import { usePointerSwipe } from '@vueuse/core';
import { 
  X, 
  ChevronLeft, 
  ChevronRight, 
  Download, 
  AlertCircle
} from 'lucide-vue-next';
import { getRawUrl } from '../../api';
import { CAP_STREAM, CAP_RENDER, CAP_EDIT } from '../../types';
import type { FileItem } from '../../types';
import FileIcon from '../ui/FileIcon.vue';
import VideoPreview from './VideoPreview.vue';
import TextPreview from './TextPreview.vue';

const props = defineProps<{
  item: FileItem;
  transitionName: string;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'prev'): void;
  (e: 'next'): void;
}>();

const videoPreviewRef = ref<InstanceType<typeof VideoPreview> | null>(null);
const previewStage = ref<HTMLElement | null>(null);

// Seek variables for video
const isSeeking = ref(false);
const seekDelta = ref(0);
const initialSeekTime = ref(0);
const shouldIgnoreSwipe = ref(false);

const { distanceX, isSwiping } = usePointerSwipe(previewStage, {
  onSwipeStart() {
    shouldIgnoreSwipe.value = false;
  },
  onSwipe() {
    if (shouldIgnoreSwipe.value) return;

    const videoElement = videoPreviewRef.value?.videoElement;
    if (props.item && (props.item.capabilities & CAP_STREAM) && videoElement) {
      if (!isSeeking.value && Math.abs(distanceX.value) > 10) {
        isSeeking.value = true;
        initialSeekTime.value = videoElement.currentTime;
      }

      if (isSeeking.value) {
        const scrubAmount = Math.min(videoElement.duration || 0, 90);
        const deltaX = -distanceX.value;
        seekDelta.value = (deltaX / window.innerWidth) * scrubAmount;
        
        let newTime = initialSeekTime.value + seekDelta.value;
        videoElement.currentTime = Math.max(0, Math.min(newTime, videoElement.duration || 0));
      }
    }
  },
  onSwipeEnd() {
    if (shouldIgnoreSwipe.value) {
      shouldIgnoreSwipe.value = false;
      return;
    }

    if (isSeeking.value) {
      isSeeking.value = false;
      seekDelta.value = 0;
    }
  },
});

watch(isSwiping, (val) => {
  if (!val && isSeeking.value) {
    isSeeking.value = false;
    seekDelta.value = 0;
  }
});

const handlePlayVideo = () => {
  videoPreviewRef.value?.playVideo();
};
</script>

<template>
  <transition name="preview-zoom">
    <div 
      class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/95 dark:bg-dracula-950/95 backdrop-blur-2xl"
      @click="emit('close')"
    >
      <!-- Navigation Buttons -->
      <button 
        @click.stop="emit('prev')"
        class="hidden sm:flex absolute left-4 top-1/2 -translate-y-1/2 w-14 h-14 items-center justify-center bg-white/5 hover:bg-white/10 text-white rounded-full transition-all z-50 border border-white/5 focus-visible:ring-2 focus-visible:ring-blue-500/50 outline-none"
        aria-label="Previous item"
      >
        <ChevronLeft class="w-8 h-8" />
      </button>
      <button 
        @click.stop="emit('next')"
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
              <FileIcon :name="item.name" :isDir="false" :capabilities="item.capabilities" class="w-5 h-5" />
            </div>
            <div class="truncate text-left">
              <h4 class="text-white text-sm font-bold truncate">{{ item.name }}</h4>
              <p class="text-[10px] text-slate-400 dark:text-dracula-300 font-bold uppercase tracking-wider">{{ (item.size / 1024 / 1024).toFixed(2) }} MB</p>
            </div>
          </div>
          
          <div class="flex items-center gap-2">
            <a 
              :href="getRawUrl(item.path, true)" 
              download
              class="hidden sm:flex items-center gap-2 px-4 py-2 bg-white/10 hover:bg-white/20 text-white rounded-xl text-xs font-bold transition-all"
            >
              <Download class="w-4 h-4" />
              Raw File
            </a>
            <button 
              @click="emit('close')"
              class="p-2.5 bg-white/10 hover:bg-rose-500 text-white rounded-xl transition-all focus-visible:ring-2 focus-visible:ring-blue-500/50 outline-none"
              aria-label="Close preview"
            >
              <X class="w-5 h-5" />
            </button>
          </div>
        </div>

        <!-- Main Stage -->
        <div 
          ref="previewStage" 
          class="flex-1 flex items-center justify-center p-2 sm:p-4 overflow-hidden relative"
          :class="{ 'touch-none': item.capabilities & CAP_STREAM }"
        >
          <Transition :name="transitionName" mode="out-in" @after-enter="handlePlayVideo">
            <!-- Image Stage -->
            <div v-if="item.capabilities & CAP_RENDER" :key="item.path" class="w-full h-full flex items-center justify-center">
              <img 
                :src="getRawUrl(item.path)"
                class="max-w-full max-h-full object-contain shadow-[0_32px_64px_rgba(0,0,0,0.5)] dark:shadow-black/80 rounded-sm transition-opacity duration-300"
              >
            </div>

            <!-- Video Stage -->
            <VideoPreview
              v-else-if="item.capabilities & CAP_STREAM"
              ref="videoPreviewRef"
              :key="item.path + '-v'"
              :item="item"
              :isSeeking="isSeeking"
              :seekDelta="seekDelta"
            />

            <!-- Text Stage -->
            <TextPreview
              v-else-if="item.capabilities & CAP_EDIT"
              :key="item.path + '-t'"
              :item="item"
            />

            <!-- Fallback Stage -->
            <div v-else :key="'fallback'" class="p-12 bg-white/5 rounded-3xl border border-white/10 text-center max-w-sm">
              <div class="w-20 h-20 bg-white/10 rounded-2xl flex items-center justify-center mx-auto mb-6">
                <AlertCircle class="w-10 h-10 text-slate-400 dark:text-dracula-400" />
              </div>
              <h3 class="text-white font-bold text-lg mb-2">No Preview</h3>
              <p class="text-slate-400 dark:text-dracula-400 text-xs mb-8 leading-relaxed">This file type requires external software to view. You can download the raw data below.</p>
              <a 
                :href="getRawUrl(item.path, true)" 
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
            @click.stop="emit('prev')"
            class="flex items-center justify-center gap-2 py-4 bg-white/10 text-white rounded-2xl text-sm font-bold focus-visible:ring-2 focus-visible:ring-blue-500/50 outline-none"
            aria-label="Previous item"
          >
            <ChevronLeft class="w-5 h-5" />
            Prev
          </button>
          <button 
            @click.stop="emit('next')"
            class="flex items-center justify-center gap-2 py-4 bg-white/10 text-white rounded-2xl text-sm font-bold focus-visible:ring-2 focus-visible:ring-blue-500/50 outline-none"
            aria-label="Next item"
          >
            Next
            <ChevronRight class="w-5 h-5" />
          </button>
          <a 
            :href="getRawUrl(item.path, true)" 
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
</template>

<style scoped>
.preview-zoom-enter-active, .preview-zoom-leave-active { transition: all 0.4s cubic-bezier(0.16, 1, 0.3, 1); }
.preview-zoom-enter-from { opacity: 0; transform: scale(0.95); }
.preview-zoom-leave-to { opacity: 0; transform: scale(1.05); }

/* Slide Transitions */
.slide-next-enter-active, .slide-next-leave-active,
.slide-prev-enter-active, .slide-prev-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.slide-next-enter-from { opacity: 0; transform: translateX(100%); }
.slide-next-leave-to { opacity: 0; transform: translateX(-100%); }

.slide-prev-enter-from { opacity: 0; transform: translateX(-100%); }
.slide-prev-leave-to { opacity: 0; transform: translateX(100%); }
</style>
