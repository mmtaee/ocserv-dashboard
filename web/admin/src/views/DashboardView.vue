<script setup lang="ts">
import AppSidebar from '@/components/AppSidebar.vue'
import DashboardHeader from '@/components/DashboardHeader.vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar'
</script>

<template>
  <div class="[--header-height:calc(--spacing(14))]">
    <SidebarProvider class="flex flex-col">
      <DashboardHeader />
      <div class="flex flex-1">
        <AppSidebar />
        <SidebarInset>
          <main class="flex flex-1 flex-col gap-6 p-4 lg:p-6">
            <div>
              <p class="text-sm text-muted-foreground">Infrastructure overview</p>
              <h1 class="text-2xl font-semibold tracking-tight">Dashboard</h1>
            </div>
            <div class="grid gap-4 md:grid-cols-3">
              <Card v-for="metric in ['Connected users', 'Active sessions', 'Managed servers']" :key="metric">
                <CardHeader class="pb-2">
                  <CardTitle class="text-sm font-medium text-muted-foreground">{{ metric }}</CardTitle>
                </CardHeader>
                <CardContent class="text-3xl font-semibold tracking-tight">—</CardContent>
              </Card>
            </div>
            <Card class="min-h-80">
              <CardHeader><CardTitle>Service activity</CardTitle></CardHeader>
              <CardContent class="text-sm text-muted-foreground">Dashboard metrics will appear here as services report activity.</CardContent>
            </Card>
          </main>
        </SidebarInset>
      </div>
    </SidebarProvider>
  </div>
</template>
