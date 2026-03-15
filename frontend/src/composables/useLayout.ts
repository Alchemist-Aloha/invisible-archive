import { ref } from 'vue';

export type LayoutMode = 'grid' | 'list' | 'details';

export function useLayout() {
  const layoutMode = ref<LayoutMode>((localStorage.getItem('layoutMode') as LayoutMode) || 'grid');

  const setLayoutMode = (mode: LayoutMode) => {
    layoutMode.value = mode;
    localStorage.setItem('layoutMode', mode);
  };

  const cycleLayout = () => {
    const modes: LayoutMode[] = ['grid', 'list', 'details'];
    const currentIndex = modes.indexOf(layoutMode.value);
    setLayoutMode(modes[(currentIndex + 1) % modes.length]);
  };

  return {
    layoutMode,
    setLayoutMode,
    cycleLayout,
  };
}
