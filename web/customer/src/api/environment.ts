export const isTestMode =
  import.meta.env.MODE === "test" || import.meta.env.VITE_USE_MOCKS === "true";
