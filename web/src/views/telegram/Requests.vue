<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import {
    TelegramAPI,
    type TelegramRequestModel,
    type TelegramPackage
} from '@/api/telegram';
import { useSnackbarStore } from '@/stores/snackbar';
import UiParentCard from '@/components/shared/UiParentCard.vue';

const { t } = useI18n();
const snackbar = useSnackbarStore();

const loading = ref(false);
const tab = ref<'pending' | 'awaiting_payment' | 'payment_uploaded' | 'history'>('pending');
const items = ref<TelegramRequestModel[]>([]);
const total = ref(0);
const page = ref(1);
const size = 20;

const detailDialog = ref(false);
const selected = ref<TelegramRequestModel | null>(null);
const adminNote = ref('');
const overrideUsername = ref('');
const overridePassword = ref('');
const owner = ref('');
const groupName = ref('');
const receiptObjectUrl = ref<string | null>(null);
const packages = ref<TelegramPackage[]>([]);

const STATUS_BY_TAB: Record<string, string> = {
    pending: 'pending',
    awaiting_payment: 'awaiting_payment',
    payment_uploaded: 'payment_uploaded',
    history: ''
};

const load = async () => {
    loading.value = true;
    try {
        const params: any = { page: page.value, size };
        const status = STATUS_BY_TAB[tab.value];
        if (status) params.status = status;
        const res = await TelegramAPI.listRequests(params);
        items.value = res.data.result || [];
        total.value = res.data.meta.total_records;
    } finally {
        loading.value = false;
    }
};

const loadPackages = async () => {
    const res = await TelegramAPI.listPackages(true);
    packages.value = res.data;
};

const openDetails = async (req: TelegramRequestModel) => {
    selected.value = req;
    adminNote.value = req.admin_note || '';
    overrideUsername.value = req.desired_username || '';
    overridePassword.value = '';
    owner.value = '';
    groupName.value = '';
    detailDialog.value = true;
    if (req.receipt_file_path) {
        try {
            const blob = await TelegramAPI.fetchReceiptBlob(req.id);
            receiptObjectUrl.value = URL.createObjectURL(blob);
        } catch (e) {
            receiptObjectUrl.value = null;
        }
    } else {
        receiptObjectUrl.value = null;
    }
};

const closeDetails = () => {
    detailDialog.value = false;
    if (receiptObjectUrl.value) {
        URL.revokeObjectURL(receiptObjectUrl.value);
        receiptObjectUrl.value = null;
    }
};

const approve = async () => {
    if (!selected.value) return;
    loading.value = true;
    try {
        await TelegramAPI.approve(selected.value.id, adminNote.value);
        snackbar.show({ id: 1, message: t('TELEGRAM_REQUEST_APPROVED'), color: 'success', timeout: 3000 });
        closeDetails();
        await load();
    } finally {
        loading.value = false;
    }
};

const reject = async () => {
    if (!selected.value) return;
    if (!confirm(t('CONFIRM_REJECT'))) return;
    loading.value = true;
    try {
        await TelegramAPI.reject(selected.value.id, adminNote.value);
        snackbar.show({ id: 1, message: t('TELEGRAM_REQUEST_REJECTED'), color: 'warning', timeout: 3000 });
        closeDetails();
        await load();
    } finally {
        loading.value = false;
    }
};

const confirmPayment = async () => {
    if (!selected.value) return;
    loading.value = true;
    try {
        await TelegramAPI.confirmPayment(selected.value.id, {
            override_username: overrideUsername.value || undefined,
            override_password: overridePassword.value || undefined,
            owner: owner.value || undefined,
            group: groupName.value || undefined,
            admin_note: adminNote.value || undefined
        });
        snackbar.show({ id: 1, message: t('TELEGRAM_REQUEST_DELIVERED'), color: 'success', timeout: 3000 });
        closeDetails();
        await load();
    } finally {
        loading.value = false;
    }
};

const findPackageTitle = (id?: number): string => {
    if (!id) return '—';
    const p = packages.value.find((p) => p.id === id);
    return p ? p.title : `#${id}`;
};

watch(tab, () => {
    page.value = 1;
    load();
});

onMounted(async () => {
    await loadPackages();
    await load();
});

onBeforeUnmount(() => {
    if (receiptObjectUrl.value) URL.revokeObjectURL(receiptObjectUrl.value);
});
</script>

<template>
    <v-row>
        <v-col cols="12">
            <UiParentCard :title="t('TELEGRAM_REQUESTS')">
                <v-tabs v-model="tab" color="primary" align-tabs="start" class="mb-3">
                    <v-tab value="pending">{{ t('TELEGRAM_TAB_PENDING') }}</v-tab>
                    <v-tab value="awaiting_payment">{{ t('TELEGRAM_TAB_AWAITING') }}</v-tab>
                    <v-tab value="payment_uploaded">{{ t('TELEGRAM_TAB_UPLOADED') }}</v-tab>
                    <v-tab value="history">{{ t('TELEGRAM_TAB_HISTORY') }}</v-tab>
                </v-tabs>

                <v-table density="comfortable">
                    <thead>
                        <tr class="text-capitalize bg-lightprimary">
                            <th>#</th>
                            <th>{{ t('CHAT_ID') }}</th>
                            <th>{{ t('TYPE') }}</th>
                            <th>{{ t('PACKAGE') }}</th>
                            <th>{{ t('STATUS') }}</th>
                            <th>{{ t('CREATED_AT') }}</th>
                            <th>{{ t('ACTION') }}</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="r in items" :key="r.id">
                            <td>{{ r.id }}</td>
                            <td>{{ r.chat_id }}</td>
                            <td>{{ r.type }}</td>
                            <td>{{ findPackageTitle(r.package_id) }}</td>
                            <td>
                                <v-chip size="small" variant="tonal">{{ r.status }}</v-chip>
                            </td>
                            <td>{{ new Date(r.created_at).toLocaleString() }}</td>
                            <td>
                                <v-btn size="small" variant="text" @click="openDetails(r)">
                                    {{ t('VIEW') }}
                                </v-btn>
                            </td>
                        </tr>
                        <tr v-if="!items.length">
                            <td colspan="7" class="text-center text-grey">
                                {{ t('NO_DATA') }}
                            </td>
                        </tr>
                    </tbody>
                </v-table>
            </UiParentCard>
        </v-col>

        <v-dialog v-model="detailDialog" max-width="800" @update:modelValue="(v) => !v && closeDetails()">
            <v-card v-if="selected">
                <v-card-title>
                    {{ t('TELEGRAM_REQUEST_DETAILS') }} #{{ selected.id }}
                </v-card-title>
                <v-card-text>
                    <v-row>
                        <v-col cols="12" md="6">
                            <div><strong>{{ t('TYPE') }}:</strong> {{ selected.type }}</div>
                            <div><strong>{{ t('STATUS') }}:</strong> {{ selected.status }}</div>
                            <div><strong>{{ t('CHAT_ID') }}:</strong> {{ selected.chat_id }}</div>
                            <div v-if="selected.telegram_username">
                                <strong>{{ t('TELEGRAM_USERNAME') }}:</strong> @{{ selected.telegram_username }}
                            </div>
                            <div>
                                <strong>{{ t('PACKAGE') }}:</strong> {{ findPackageTitle(selected.package_id) }}
                            </div>
                            <div v-if="selected.target_ocserv_id">
                                <strong>{{ t('TARGET_OCSERV_USER') }}:</strong> #{{ selected.target_ocserv_id }}
                            </div>
                            <div v-if="selected.desired_username">
                                <strong>{{ t('DESIRED_USERNAME') }}:</strong> {{ selected.desired_username }}
                            </div>
                            <div v-if="selected.user_message">
                                <strong>{{ t('USER_MESSAGE') }}:</strong> {{ selected.user_message }}
                            </div>
                        </v-col>
                        <v-col cols="12" md="6">
                            <div v-if="receiptObjectUrl">
                                <strong>{{ t('RECEIPT') }}:</strong>
                                <a :href="receiptObjectUrl" target="_blank">
                                    <img :src="receiptObjectUrl" style="max-width: 100%; max-height: 280px; margin-top: 8px;" alt="receipt" />
                                </a>
                            </div>
                            <div v-else class="text-grey">
                                {{ t('NO_RECEIPT') }}
                            </div>
                        </v-col>
                    </v-row>

                    <v-divider class="my-3" />

                    <v-textarea
                        v-model="adminNote"
                        :label="t('ADMIN_NOTE')"
                        rows="2"
                        variant="outlined"
                        density="comfortable"
                    />

                    <template v-if="selected.status === 'payment_uploaded' && selected.type === 'new'">
                        <v-row class="mt-1">
                            <v-col cols="12" md="6">
                                <v-text-field
                                    v-model="overrideUsername"
                                    :label="t('OVERRIDE_USERNAME')"
                                    variant="outlined"
                                    density="comfortable"
                                />
                            </v-col>
                            <v-col cols="12" md="6">
                                <v-text-field
                                    v-model="overridePassword"
                                    :label="t('OVERRIDE_PASSWORD')"
                                    variant="outlined"
                                    density="comfortable"
                                />
                            </v-col>
                            <v-col cols="12" md="6">
                                <v-text-field
                                    v-model="owner"
                                    :label="t('OWNER')"
                                    placeholder="telegram"
                                    variant="outlined"
                                    density="comfortable"
                                />
                            </v-col>
                            <v-col cols="12" md="6">
                                <v-text-field
                                    v-model="groupName"
                                    :label="t('GROUP')"
                                    placeholder="defaults"
                                    variant="outlined"
                                    density="comfortable"
                                />
                            </v-col>
                        </v-row>
                    </template>
                </v-card-text>

                <v-card-actions class="px-4 pb-4">
                    <v-btn variant="text" @click="closeDetails">{{ t('CLOSE') }}</v-btn>
                    <v-spacer />
                    <v-btn
                        v-if="selected.status === 'pending'"
                        color="primary"
                        :loading="loading"
                        @click="approve"
                    >
                        {{ t('APPROVE') }}
                    </v-btn>
                    <v-btn
                        v-if="selected.status === 'payment_uploaded'"
                        color="success"
                        :loading="loading"
                        @click="confirmPayment"
                    >
                        {{ t('TELEGRAM_CONFIRM_PAYMENT') }}
                    </v-btn>
                    <v-btn
                        v-if="selected.status !== 'delivered' && selected.status !== 'rejected'"
                        color="error"
                        variant="outlined"
                        :loading="loading"
                        @click="reject"
                    >
                        {{ t('REJECT') }}
                    </v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>
    </v-row>
</template>
