<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from "vue";

const props = defineProps<{ siteKey: string }>();
const emit = defineEmits<{ "update:modelValue": [value: string] }>();

interface RecaptchaApi {
  ready(callback: () => void): void;
  render(container: HTMLElement, options: Record<string, unknown>): number;
  reset(widgetId?: number): void;
}

declare global {
  interface Window {
    grecaptcha?: RecaptchaApi;
  }
}

const container = ref<HTMLElement | null>(null);
let widgetId: number | undefined;

function loadRecaptcha(): Promise<RecaptchaApi> {
  if (window.grecaptcha) return Promise.resolve(window.grecaptcha);

  return new Promise((resolve, reject) => {
    const existing = document.querySelector<HTMLScriptElement>(
      "script[data-ocserv-recaptcha]",
    );
    const script = existing ?? document.createElement("script");
    const resolveApi = () =>
      window.grecaptcha
        ? resolve(window.grecaptcha)
        : reject(new Error("Captcha failed to load."));

    script.addEventListener("load", resolveApi, { once: true });
    script.addEventListener(
      "error",
      () => reject(new Error("Captcha failed to load.")),
      { once: true },
    );

    if (!existing) {
      script.src = "https://www.google.com/recaptcha/api.js?render=explicit";
      script.async = true;
      script.defer = true;
      script.dataset.ocservRecaptcha = "true";
      document.head.append(script);
    }
  });
}

async function renderCaptcha(): Promise<void> {
  if (!container.value || !props.siteKey) return;
  const recaptcha = await loadRecaptcha();
  recaptcha.ready(() => {
    if (!container.value || widgetId !== undefined) return;
    widgetId = recaptcha.render(container.value, {
      sitekey: props.siteKey,
      callback: (token: string) => emit("update:modelValue", token),
      "expired-callback": () => emit("update:modelValue", ""),
      "error-callback": () => emit("update:modelValue", ""),
    });
  });
}

watch(
  () => props.siteKey,
  () => void renderCaptcha(),
);
onMounted(() => void renderCaptcha());
onBeforeUnmount(() => {
  if (widgetId !== undefined) window.grecaptcha?.reset(widgetId);
});
</script>

<template>
  <div ref="container" class="min-h-19 overflow-hidden rounded-md" />
</template>
