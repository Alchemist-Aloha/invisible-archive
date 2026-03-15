import { ref } from 'vue';

export function useTheme() {
  const isDarkMode = ref(localStorage.getItem('theme') === 'dark');

  const updateTheme = () => {
    if (isDarkMode.value) {
      document.documentElement.classList.add('dark');
      document.documentElement.style.colorScheme = 'dark';
    } else {
      document.documentElement.classList.remove('dark');
      document.documentElement.style.colorScheme = 'light';
    }
  };

  const toggleDarkMode = () => {
    isDarkMode.value = !isDarkMode.value;
    localStorage.setItem('theme', isDarkMode.value ? 'dark' : 'light');
    updateTheme();
  };

  return {
    isDarkMode,
    updateTheme,
    toggleDarkMode,
  };
}
