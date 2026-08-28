<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";

import AppSidebar from "@/components/AppSidebar.vue";
import SiteHeader from "@/components/SiteHeader.vue";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { Skeleton } from "@/components/ui/skeleton";
import { isRtlLocale } from "@/locales";

const { locale, t } = useI18n({ useScope: "global" });
const sidebarSide = computed(() =>
  isRtlLocale(locale.value) ? "right" : "left",
);
const summaryCards = computed(() => [
  t("dashboard.connectedUsers"),
  t("dashboard.activeSessions"),
  t("dashboard.managedServers"),
]);
</script>

<template>
  <div class="flex flex-1 [--header-height:calc(--spacing(14))]">
    <SidebarProvider class="!min-h-0 flex-1 flex-col">
      <SiteHeader />
      <div class="flex min-h-0 flex-1">
        <AppSidebar :side="sidebarSide" collapsible="icon" />
        <SidebarInset>
          <main
            class="flex flex-1 flex-col gap-6 p-4 pb-14 lg:p-6 lg:pb-14"
            aria-busy="true"
          >
            <div>
              <h1 class="text-2xl font-semibold tracking-tight">
                {{ t("dashboard.overview") }}
              </h1>
              <p class="text-sm text-muted-foreground">
                {{ t("dashboard.activityDescription") }}
              </p>
            </div>

            <section
              class="grid gap-4 md:grid-cols-3"
              :aria-label="t('dashboard.overview')"
            >
              <Card v-for="label in summaryCards" :key="label">
                <CardHeader class="gap-2 pb-2">
                  <span class="text-sm font-medium text-muted-foreground">{{
                    label
                  }}</span>
                  <Skeleton class="h-8 w-20" />
                </CardHeader>
                <CardContent>
                  <Skeleton class="h-3 w-28" />
                </CardContent>
              </Card>
            </section>

            <Card class="min-h-72 flex-1">
              <CardHeader>
                <div class="grid gap-2">
                  <span class="font-semibold">{{
                    t("dashboard.activityTitle")
                  }}</span>
                  <Skeleton class="h-4 w-64 max-w-full" />
                </div>
              </CardHeader>
              <CardContent class="grid gap-4">
                <Skeleton class="h-40 w-full" />
                <div class="grid gap-3 sm:grid-cols-3">
                  <Skeleton v-for="item in 3" :key="item" class="h-12 w-full" />
                </div>
              </CardContent>
            </Card>
          </main>
        </SidebarInset>
      </div>
    </SidebarProvider>
  </div>
</template>
