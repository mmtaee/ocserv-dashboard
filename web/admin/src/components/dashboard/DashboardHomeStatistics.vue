<script setup lang="ts">
import type { DeepReadonly } from "vue";

import type { DashboardOverview } from "@/api/services/dashboard";
import BandwidthChart from "@/components/dashboard/BandwidthChart.vue";
import IpBansOverview from "@/components/dashboard/IpBansOverview.vue";
import OnlineSessionsOverview from "@/components/dashboard/OnlineSessionsOverview.vue";
import OnlineUsersOverview from "@/components/dashboard/OnlineUsersOverview.vue";
import TopBandwidthUsers from "@/components/dashboard/TopBandwidthUsers.vue";
import TrafficChart from "@/components/dashboard/TrafficChart.vue";

const props = defineProps<{
  overview: DeepReadonly<DashboardOverview> | null;
  loading: boolean;
}>();
</script>

<template>
  <section class="flex flex-col gap-6">
    <div class="grid gap-6 xl:grid-cols-3">
      <OnlineUsersOverview
        :users="props.overview?.users ?? null"
        :loading="loading"
      />
      <IpBansOverview
        class="xl:col-span-2"
        :bans="props.overview?.ip_bans ?? []"
        :loading="loading"
      />
    </div>
    <OnlineSessionsOverview
      :sessions="props.overview?.users?.online_users_session ?? []"
      :loading="loading"
    />
    <TopBandwidthUsers
      :users="props.overview?.top_bandwidth_user ?? null"
      :loading="loading"
    />
    <div class="grid gap-6 xl:grid-cols-3">
      <TrafficChart
        class="xl:col-span-2"
        :data="props.overview?.statistics ?? []"
        :loading="loading"
      />
      <BandwidthChart
        :total="props.overview?.total_bandwidth ?? null"
        :loading="loading"
      />
    </div>
  </section>
</template>
