<script setup lang="ts">
import { shallowRef } from "vue";
import { useI18n } from "vue-i18n";

import OcservSyncGroups from "@/components/ocserv-sync/OcservSyncGroups.vue";
import OcservSyncUsers from "@/components/ocserv-sync/OcservSyncUsers.vue";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

const { t } = useI18n({ useScope: "global" });
const tab = shallowRef<"users" | "groups">("users");
</script>

<template>
  <div class="flex flex-col gap-6 p-4 md:p-6">
    <Card>
      <CardHeader>
        <CardTitle>{{ t("ocservSync.title") }}</CardTitle>
        <CardDescription>{{ t("ocservSync.description") }}</CardDescription>
      </CardHeader>
      <CardContent class="flex flex-col gap-6">
        <div
          :aria-label="t('ocservSync.tabList')"
          class="flex gap-2"
          role="tablist"
        >
          <Button
            :aria-selected="tab === 'users'"
            role="tab"
            :variant="tab === 'users' ? 'default' : 'outline'"
            @click="tab = 'users'"
          >
            {{ t("ocservSync.users") }}
          </Button>
          <Button
            :aria-selected="tab === 'groups'"
            role="tab"
            :variant="tab === 'groups' ? 'default' : 'outline'"
            @click="tab = 'groups'"
          >
            {{ t("ocservSync.groups") }}
          </Button>
        </div>
        <OcservSyncUsers v-if="tab === 'users'" role="tabpanel" />
        <OcservSyncGroups v-else role="tabpanel" />
      </CardContent>
    </Card>
  </div>
</template>
