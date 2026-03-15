import { ref } from 'vue';

export function useNavigation() {
  const getInitialPath = () => {
    const hash = window.location.hash.slice(1);
    if (hash) {
      const decoded = decodeURIComponent(hash);
      return decoded.startsWith('/') ? decoded : '/' + decoded;
    }
    return localStorage.getItem('lastPath') || '/';
  };

  const currentPath = ref(getInitialPath());

  const handleNavigate = (path: string, pushState = true) => {
    currentPath.value = path;
    localStorage.setItem('lastPath', path);
    
    if (pushState) {
      history.pushState({ path }, '', '#' + encodeURIComponent(path));
    }
  };

  const goBack = () => {
    if (currentPath.value === '/') return;
    const parts = currentPath.value.split('/');
    parts.pop();
    handleNavigate(parts.join('/') || '/');
  };

  // Build history stack if we started at a deep path
  const initHistoryStack = (startPath: string) => {
    if (startPath !== '/') {
      const segments = startPath.split('/').filter(Boolean);
      let cumulative = '';
      
      history.replaceState({ path: '/' }, '', '#/');
      
      for (let i = 0; i < segments.length - 1; i++) {
        cumulative += '/' + segments[i];
        history.pushState({ path: cumulative }, '', '#' + encodeURIComponent(cumulative));
      }
      
      history.pushState({ path: startPath }, '', '#' + encodeURIComponent(startPath));
    } else if (!window.location.hash) {
      history.replaceState({ path: '/' }, '', '#/');
    }
  };

  return {
    currentPath,
    handleNavigate,
    goBack,
    initHistoryStack,
  };
}
