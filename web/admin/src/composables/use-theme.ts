import { useDark } from "@vueuse/core";

const isDark = useDark({
  storageKey: "ocserv-dashboard-theme",
});

export function useTheme() {
  const toggleTheme = () => {
    isDark.value = !isDark.value;
  };

  return {
    isDark,
    toggleTheme,
  };
}
