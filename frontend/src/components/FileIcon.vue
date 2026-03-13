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
  <Folder v-if="props.isDir" class="w-10 h-10 text-blue-500 fill-current" />
  <FileArchive v-else-if="isZip" class="w-10 h-10 text-amber-500 fill-current" />
  <FileVideo v-else-if="props.capabilities & CAP_STREAM" class="w-10 h-10 text-purple-500" />
  <FileImage v-else-if="props.capabilities & CAP_RENDER" class="w-10 h-10 text-rose-500" />
  <FileCode v-else-if="props.capabilities & CAP_EDIT" class="w-10 h-10 text-emerald-500" />
  <FileGeneric v-else class="w-10 h-10 text-gray-400" />
</template>
