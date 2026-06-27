<script setup lang="ts">
import { onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { TelegramAPI, type TelegramRequestModel, type TelegramPackage } from '@/api/telegram';
import { OcservGroupsApi, OcservUsersGetSortEnum } from '@/api';
import { getAuthorization } from '@/utils/request';
import { useSnackbarStore } from '@/stores/snackbar';
import UiParentCard from '@/components/shared/UiParentCard.vue';
import Pagination from '@/components/shared/Pagination.vue';
import type { Meta } from '@/types/metaTypes/MetaType';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const snackbar = useSnackbarStore();

const loading = ref(false);
const tab = ref<'pending' | 'awaiting_payment' | 'payment_uploaded' | 'history'>('pending');
const items = ref<TelegramRequestModel[]>([]);
const meta = reactive<Meta>({
    page: 1,
    size: 10,
    sort: OcservUsersGetSortEnum.ASC,
    total_records: 0
});

const detailDialog = ref(false);
const selected = ref<TelegramRequestModel | null>(null);
const adminNote = ref('');
const approveCardNumber = ref('');
const approveCardHolder = ref('');
const replyToUser = ref('');
const overrideUsername = ref('');
const overridePassword = ref('');
const owner = ref('');
const groupName = ref('');
const receiptObjectUrl = ref<string | null>(null);
const packages = ref<TelegramPackage[]>([]);
const groups = ref<string[]>(['defaults']);

const STATUS_BY_TAB: Record<string, string> = {
    pending: 'pending',
    awaiting_payment: 'awaiting_payment',
    payment_uploaded: 'payment_uploaded',
    history: ''
};

const load = async () => {
    loading.value = true;
    try {
        const page = Number(route.query.page) || 1;
        const size = Math.max(Number(route.query.size) || 10, 1);
        const sort = route.query.sort === 'DESC' ? 'DESC' : 'ASC';
        const status = STATUS_BY_TAB[tab.value];
        const res = await TelegramAPI.listRequests({
            page,
            size,
            sort,
            order: 'created_at',
            ...(status ? { status } : {})
        });
        items.value = res.data.result || [];
        Object.assign(meta, res.data.meta);
        meta.sort = route.query.sort === 'DESC' ? OcservUsersGetSortEnum.DESC : OcservUsersGetSortEnum.ASC;
    } finally {
        loading.value = false;
    }
};

const updateMeta = (newMeta: Meta) => {
    Object.assign(meta, newMeta);
    load();
};

const loadPackages = async () => {
    const res = await TelegramAPI.listPackages(true);
    packages.value = res.data;
};

const loadGroups = async () => {
    try {
        const api = new OcservGroupsApi();
        const res = await api.ocservGroupsLookupGet({ ...getAuthorization() });
        const list = (res.data || []) as string[];
        groups.value = list.length ? list : ['defaults'];
    } catch {
        groups.value = ['defaults'];
    }
};

const openDetails = async (req: TelegramRequestModel) => {
    selected.value = req;
    adminNote.value = req.admin_note || '';
    replyToUser.value = '';
    approveCardNumber.value = '';
    approveCardHolder.value = '';
    if (req.status === 'pending') {
        try {
            const s = await TelegramAPI.getSettings();
            approveCardNumber.value = s.data.card_number || '';
            approveCardHolder.value = s.data.card_holder || '';
        } catch {
            /* ignore */
        }
    }
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
        await TelegramAPI.approve(selected.value.id, {
            admin_note: adminNote.value || undefined,
            card_number: approveCardNumber.value || undefined,
            card_holder: approveCardHolder.value || undefined,
            reply_to_user: replyToUser.value || undefined
        });
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

const displayContact = (r: TelegramRequestModel): string => {
    const u = (r.telegram_username || '').trim();
    if (u) return `@${u}`;
    return String(r.chat_id);
};

const removeRequest = async (r: TelegramRequestModel) => {
    if (!confirm(t('CONFIRM_DELETE'))) return;
    loading.value = true;
    try {
        await TelegramAPI.deleteRequest(r.id);
        snackbar.show({ id: 1, message: t('TELEGRAM_REQUEST_DELETED'), color: 'success', timeout: 3000 });
        const p = Number(route.query.page) || 1;
        await load();
        if (!items.value.length && p > 1) {
            await router.replace({ query: { ...route.query, page: String(p - 1) } });
            await load();
        }
    } catch {
        snackbar.show({ id: 1, message: t('TELEGRAM_REQUEST_DELETE_FAILED'), color: 'error', timeout: 5000 });
    } finally {
        loading.value = false;
    }
};

const isRequestDeletable = (r: TelegramRequestModel) =>
    !['pending', 'awaiting_payment', 'payment_uploaded'].includes(r.status);

watch(tab, async () => {
    await router.replace({ query: { ...route.query, page: '1' } });
    await load();
});

onMounted(async () => {
    await Promise.all([loadPackages(), loadGroups()]);
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
                <v-progress-linear :active="loading" indeterminate />

                <div v-if="!loading">
                    <v-tabs v-model="tab" color="primary" align-tabs="start" class="mb-3 px-md-15">
                        <v-tab value="pending">{{ t('TELEGRAM_TAB_PENDING') }}</v-tab>
                        <v-tab value="awaiting_payment">{{ t('TELEGRAM_TAB_AWAITING') }}</v-tab>
                        <v-tab value="payment_uploaded">{{ t('TELEGRAM_TAB_UPLOADED') }}</v-tab>
                        <v-tab value="history">{{ t('TELEGRAM_TAB_HISTORY') }}</v-tab>
                    </v-tabs>

                    <v-table v-if="items.length > 0" density="comfortable" class="px-md-15">
                        <thead>
                            <tr class="text-capitalize bg-lightprimary">
                                <th class="text-left">#</th>
                                <th class="text-left">{{ t('TELEGRAM_REQUEST_CONTACT') }}</th>
                                <th class="text-left">{{ t('TYPE') }}</th>
                                <th class="text-left">{{ t('PACKAGE') }}</th>
                                <th class="text-left">{{ t('STATUS') }}</th>
                                <th class="text-left">{{ t('CREATED_AT') }}</th>
                                <th class="text-left">{{ t('ACTION') }}</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="r in items" :key="r.id">
                                <td>{{ r.id }}</td>
                                <td>{{ displayContact(r) }}</td>
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
                                    <v-btn
                                        v-if="tab === 'history' && isRequestDeletable(r)"
                                        icon="mdi-delete"
                                        size="small"
                                        variant="text"
                                        color="error"
                                        :title="t('DELETE')"
                                        @click="removeRequest(r)"
                                    />
                                </td>
                            </tr>
                        </tbody>
                    </v-table>
                </div>

                <div v-if="loading || items.length == 0" class="ms-md-5 mb-md-5 text-capitalize">
                    {{ t('NO_DATA') }}
                </div>

                <Pagination :totalRecords="meta.total_records" @update="updateMeta" />
            </UiParentCard>
        </v-col>

        <v-dialog v-model="detailDialog" max-width="800" @update:modelValue="(v) => !v && closeDetails()">
            <v-card v-if="selected">
                <v-card-title> {{ t('TELEGRAM_REQUEST_DETAILS') }} #{{ selected.id }} </v-card-title>
                <v-card-text>
                    <v-row>
                        <v-col cols="12" md="6">
                            <div>
                                <strong>{{ t('TYPE') }}:</strong> {{ selected.type }}
                            </div>
                            <div>
                                <strong>{{ t('STATUS') }}:</strong> {{ selected.status }}
                            </div>
                            <div>
                                <strong>{{ t('CHAT_ID') }}:</strong> {{ selected.chat_id }}
                            </div>
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
                                    <img
                                        :src="receiptObjectUrl"
                                        style="max-width: 100%; max-height: 280px; margin-top: 8px"
                                        alt="receipt"
                                    />
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

                    <template v-if="selected.status === 'pending'">
                        <v-row dense class="mt-2">
                            <v-col cols="12" md="6">
                                <v-text-field
                                    v-model="approveCardNumber"
                                    :label="t('TELEGRAM_APPROVE_CARD_NUMBER')"
                                    variant="outlined"
                                    density="comfortable"
                                    :hint="t('TELEGRAM_APPROVE_CARD_HINT')"
                                    persistent-hint
                                />
                            </v-col>
                            <v-col cols="12" md="6">
                                <v-text-field
                                    v-model="approveCardHolder"
                                    :label="t('TELEGRAM_APPROVE_CARD_HOLDER')"
                                    variant="outlined"
                                    density="comfortable"
                                    :hint="t('TELEGRAM_APPROVE_CARD_HOLDER_HINT')"
                                    persistent-hint
                                />
                            </v-col>
                            <v-col cols="12">
                                <v-textarea
                                    v-model="replyToUser"
                                    :label="t('TELEGRAM_REPLY_TO_USER')"
                                    rows="3"
                                    variant="outlined"
                                    density="comfortable"
                                    :hint="t('TELEGRAM_REPLY_TO_USER_HINT')"
                                    persistent-hint
                                />
                            </v-col>
                        </v-row>
                    </template>

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
                                <v-select
                                    v-model="groupName"
                                    :items="groups"
                                    :label="t('GROUP')"
                                    variant="outlined"
                                    density="comfortable"
                                    clearable
                                />
                            </v-col>
                        </v-row>
                    </template>
                </v-card-text>

                <v-card-actions class="px-4 pb-4">
                    <v-btn variant="text" @click="closeDetails">{{ t('CLOSE') }}</v-btn>
                    <v-spacer />
                    <v-btn v-if="selected.status === 'pending'" color="primary" :loading="loading" @click="approve">
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
