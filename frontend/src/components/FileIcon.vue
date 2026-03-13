<script setup lang="ts">
import { 
  Folder, 
  FileArchive, 
  FileVideo, 
  FileImage, 
  FileCode, 
  File as FileGeneric 
} from 'lucide-vue-next';
import { CAP_STREAM, CAP_RENDER, CAP_EDIT } from '../api';

const props = defineProps<{
  name: string;
  isDir: boolean;
  capabilities: number;
}>();

const isZip = props.name.toLowerCase().endsWith('.zip');
</script>

<template>
  <Folder v-if="props.isDir" class="w-full h-full text-blue-500 fill-blue-500/10 dark:fill-blue-400/20 transition-colors" />
  <FileArchive v-else-if="isZip" class="w-full h-full text-amber-500 fill-amber-500/10 dark:fill-amber-400/20 transition-colors" />
  <FileVideo v-else-if="props.capabilities & CAP_STREAM" class="w-full h-full text-indigo-500 fill-indigo-500/10 dark:fill-indigo-400/20 transition-colors" />
  <FileImage v-else-if="props.capabilities & CAP_RENDER" class="w-full h-full text-rose-500 fill-rose-500/10 dark:fill-rose-400/20 transition-colors" />
  <FileCode v-else-if="props.capabilities & CAP_EDIT" class="w-full h-full text-emerald-500 fill-emerald-500/10 dark:fill-emerald-400/20 transition-colors" />
  <FileGeneric v-else class="w-full h-full text-slate-400 fill-slate-400/5 dark:fill-dracula-400/10 transition-colors" />
</template>
