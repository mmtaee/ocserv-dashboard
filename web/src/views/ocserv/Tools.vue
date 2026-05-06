<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import UiParentCard from '@/components/shared/UiParentCard.vue';
import SystemdStatus from '@/components/ocserv/systemd/Status.vue';
import SystemdActions from '@/components/ocserv/systemd/Actions.vue';
import { type PropType, ref } from 'vue';
import UiChildCard from '@/components/shared/UiChildCard.vue';

type ActionState = 'enabling' | 'disabling' | 'restarting' | null;

const state = ref(null);
const { t } = useI18n();

const childRef = ref<InstanceType<typeof SystemdStatus> | null>(null);
const callGetStatus = () => {
    childRef.value?.getStatus();
};

const currentStatus = ref<'enabling' | 'disabling' | 'restarting' | null>(null);

const handleCurrentStatus = (v: ActionState) => {
    currentStatus.value = v;
};
</script>

<template>
    <UiParentCard variant="flat" :title="t('SYSTEMD_HANDLER_PAGE_TITLE')">
        <UiChildCard :title="t('STATUS')" class="px-5">
            <SystemdStatus @state="(s) => (state = s)" ref="childRef" :currentStatus="currentStatus" />
        </UiChildCard>

        <UiChildCard :title="t('SERVICE_ACTIONS')" class="px-5" :height="200">
            <SystemdActions @getState="callGetStatus" :state="state || 'inactive'" @currentStatus="handleCurrentStatus" />
        </UiChildCard>
    </UiParentCard>
</template>
