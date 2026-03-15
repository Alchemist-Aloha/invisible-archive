import { ref, type Ref } from 'vue';
import type { FileItem } from '../types';
import { CAP_RENDER, CAP_BROWSE } from '../types';
import { getRawUrl, getThumbUrl } from '../api';
import PhotoSwipeLightbox from 'photoswipe/lightbox';

export function usePreview(displayItems: Ref<FileItem[] | undefined>) {
  const previewItem = ref<FileItem | null>(null);
  const showInfo = ref(false);
  const transitionName = ref('slide-next');
  const lightbox = ref<PhotoSwipeLightbox | null>(null);

  const initPhotoSwipe = () => {
    lightbox.value = new PhotoSwipeLightbox({
      gallery: '#file-grid',
      children: '[data-pswp-src]',
      pswpModule: () => import('photoswipe'),
      bgOpacity: 0.98,
      padding: { top: 20, bottom: 20, left: 20, right: 20 },
      initialZoomLevel: 'fit',
      secondaryZoomLevel: 1.5,
      maxZoomLevel: 8,
      wheelToZoom: true,
      pinchToClose: false,
      closeOnVerticalDrag: false,
    });

    lightbox.value.on('gettingData', (event) => {
      const { data } = event;
      if (data.src && (!data.width || data.width === 1)) {
        data.width = 3000;
        data.height = 2000;
        const img = new Image();
        img.src = data.src;
        img.onload = () => {
          if (img.width && img.height) {
            const changed = data.width !== img.width || data.height !== img.height;
            data.width = img.width;
            data.height = img.height;
            if (changed && lightbox.value?.pswp) {
              lightbox.value.pswp.refreshSlideContent(event.index);
            }
          }
        };
      }
    });

    lightbox.value.init();

    lightbox.value.on('close', () => {
      if (history.state?.pswp) {
        history.back();
      }
    });
  };

  const handlePreview = (item: FileItem) => {
    if ((item.capabilities & CAP_RENDER) && !item.name.toLowerCase().endsWith('.pdf')) {
      if (lightbox.value && displayItems.value) {
        const imageItems = displayItems.value.filter(i => (i.capabilities & CAP_RENDER) && !i.name.toLowerCase().endsWith('.pdf'));
        const index = imageItems.findIndex(i => i.path === item.path);
        
        if (index !== -1) {
          const dataSource = imageItems.map(i => ({
            src: getRawUrl(i.path),
            alt: i.name,
            msrc: getThumbUrl(i.path),
            element: document.querySelector(`[data-pswp-src="${getRawUrl(i.path)}"]`) as HTMLElement || undefined
          }));

          history.pushState({ pswp: true }, '');
          lightbox.value.loadAndOpen(index, dataSource);
        }
      }
      return;
    }
    
    history.pushState({ customPreview: true }, '');
    previewItem.value = item;
    showInfo.value = false;
  };

  const closePreview = (triggerBack = true) => {
    if (triggerBack && history.state?.customPreview) {
      history.back();
    }
    previewItem.value = null;
  };

  const navigatePreview = (direction: 'prev' | 'next') => {
    if (!previewItem.value || !displayItems.value) return;
    
    transitionName.value = direction === 'next' ? 'slide-next' : 'slide-prev';
    
    const list = displayItems.value;
    const currentIndex = list.findIndex(item => item.path === previewItem.value!.path);
    if (currentIndex === -1) return;

    const step = direction === 'next' ? 1 : -1;
    let nextIndex = currentIndex;
    
    for (let i = 0; i < list.length; i++) {
      nextIndex = (nextIndex + step + list.length) % list.length;
      const item = list[nextIndex];
      if (!item.is_dir && !(item.capabilities & CAP_BROWSE)) {
        handlePreview(item);
        return;
      }
    }
  };

  return {
    previewItem,
    showInfo,
    transitionName,
    lightbox,
    initPhotoSwipe,
    handlePreview,
    closePreview,
    navigatePreview,
  };
}
