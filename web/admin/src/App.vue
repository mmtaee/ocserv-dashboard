<script setup lang="ts">
import AppSidebar from '@/components/AppSidebar.vue'
import SiteHeader from '@/components/SiteHeader.vue'
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar'
</script>

<template>
  <div class="[--header-height:calc(--spacing(14))]">
    <SidebarProvider class="flex flex-col">
      <SiteHeader />
      <div class="flex flex-1">
        <AppSidebar />
        <SidebarInset>
          <main class="flex flex-1 flex-col gap-6 p-4 lg:p-6">
            <div>
              <p class="text-sm text-muted-foreground">Infrastructure overview</p>
              <h1 class="text-2xl font-semibold tracking-tight">OCServ Dashboard</h1>
            </div>
            <div class="grid gap-4 md:grid-cols-3">
              <section
                v-for="metric in ['Connected users', 'Active sessions', 'Managed servers']"
                :key="metric"
                class="rounded-xl border bg-card p-5 text-card-foreground shadow-sm"
              >
                <p class="text-sm font-medium text-muted-foreground">{{ metric }}</p>
                <p class="mt-3 text-3xl font-semibold tracking-tight">—</p>
              </section>
            </div>
            <section class="min-h-80 rounded-xl border bg-card p-6 shadow-sm">
              <h2 class="font-semibold">Service activity</h2>
              <p class="mt-1 text-sm text-muted-foreground">
                Connect a generated API service to populate dashboard data.
              </p>
            </section>
          </main>
        </SidebarInset>
      </div>
    </SidebarProvider>
  </div>
</template>
