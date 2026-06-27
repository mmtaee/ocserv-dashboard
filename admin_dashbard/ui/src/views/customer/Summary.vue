<script setup lang="ts">
import Logo from '@/layouts/full/logo/Logo.vue';
import { useI18n } from 'vue-i18n';
import SummaryForm from '@/components/customer/SummaryForm.vue';
import {
    CustomerModelCustomerTrafficTypeEnum,
    CustomersApi,
    type CustomerCiscoSetupResponse,
    type CustomerSummaryData,
    type CustomerSummaryResponse
} from '@/api';
import { onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import SummaryResult from '@/components/customer/SummaryResult.vue';
import { useSnackbarStore } from '@/stores/snackbar';

const { t } = useI18n();
const router = useRouter();
const loading = ref(false);

const snapshot = ref<CustomerSummaryResponse>({
    ocserv_user: {
        certificate_enabled: false,
        certificate_available: false,
        deactivated_at: '',
        expire_at: '',
        is_locked: false,
        owner: '',
        rx: 0,
        traffic_size: 0,
        traffic_type: CustomerModelCustomerTrafficTypeEnum.FREE,
        tx: 0,
        username: ''
    },
    usage: {
        bandwidths: {
            rx: 0,
            tx: 0
        },
        date_start: '',
        date_end: ''
    }
});

const api = new CustomersApi();

const result = ref<CustomerSummaryResponse>(snapshot.value);
const ciscoSetup = ref<CustomerCiscoSetupResponse | null>(null);

const hasResult = ref(false);

const customerSummaryData = ref<CustomerSummaryData>({ password: '', username: '' });

onMounted(() => {
    const storedResult = sessionStorage.getItem('customerSummaryResult');
    const storedData = sessionStorage.getItem('customerSummaryData');

    if (!storedResult || !storedData) {
        return;
    }

    try {
        result.value = JSON.parse(storedResult) as CustomerSummaryResponse;
        customerSummaryData.value = JSON.parse(storedData) as CustomerSummaryData;
        hasResult.value = true;
    } catch {
        sessionStorage.removeItem('customerSummaryResult');
        sessionStorage.removeItem('customerSummaryData');
    }
});

const getSummary = (data: CustomerSummaryData) => {
    loading.value = true;
    ciscoSetup.value = null;
    sessionStorage.removeItem('customerCiscoSetup');
    Object.assign(customerSummaryData.value, data);

    api.customersSummaryPost({
        request: data
    })
        .then((res) => {
            result.value = res.data;
            hasResult.value = true;
            sessionStorage.setItem('customerSummaryResult', JSON.stringify(res.data));
            sessionStorage.setItem('customerSummaryData', JSON.stringify(customerSummaryData.value));
        })
        .finally(() => {
            loading.value = false;
        });
};

const newSummary = () => {
    Object.assign(result.value, snapshot.value);
    ciscoSetup.value = null;
    sessionStorage.removeItem('customerCiscoSetup');
    sessionStorage.removeItem('customerSummaryResult');
    sessionStorage.removeItem('customerSummaryData');
    hasResult.value = false;
    Object.assign(customerSummaryData.value, { password: '', username: '' });
};

const disconnect = () => {
    api.customersDisconnectSessionsPost({
        request: customerSummaryData.value
    }).then(() => {
        const snackbar = useSnackbarStore();
        snackbar.show({
            id: 1,
            message: t('USER_DISCONNECTED_SUCCESS_SNACK'),
            color: 'success',
            timeout: 3000
        });
    });
};

const downloadCertificate = () => {
    api.customersCertificatePost(
        {
            request: customerSummaryData.value
        },
        {
            responseType: 'blob'
        }
    ).then((res) => {
        const blob = res.data instanceof Blob ? res.data : new Blob([res.data], { type: 'application/x-pkcs12' });

        const url = window.URL.createObjectURL(blob);
        const link = document.createElement('a');

        link.href = url;
        link.setAttribute('download', `${result.value.ocserv_user.username}.p12`);
        document.body.appendChild(link);
        link.click();
        link.remove();

        window.URL.revokeObjectURL(url);
    });
};

const openCiscoSetup = () => {
    api.customersSetupCiscoPost({
        request: customerSummaryData.value
    }).then((res) => {
        ciscoSetup.value = res.data;
        sessionStorage.setItem('customerCiscoSetup', JSON.stringify(res.data));
        router.push({ name: 'CustomerCiscoSetup' });
    });
};
</script>

<template>
    <div class="authentication customer-summary-page">
        <v-container class="pa-3 pa-sm-6" fluid>
            <v-row class="customer-summary-row d-flex justify-center align-start align-sm-center">
                <v-col class="d-flex align-center" cols="12" sm="8" md="6" lg="4" xl="3" v-if="!hasResult">
                    <v-card class="px-sm-1 px-0 mx-auto w-100" elevation="10" max-width="500" rounded="md">
                        <v-card-item class="pa-5 pa-sm-8">
                            <div class="d-flex justify-center py-4">
                                <Logo />
                            </div>
                            <div class="text-body-1 text-muted text-center mb-5 text-capitalize">
                                {{ t('SUMMARY_GET_TEXT') }}
                            </div>
                            <SummaryForm @getSummary="getSummary" :loading="loading" />
                        </v-card-item>
                    </v-card>
                </v-col>

                <SummaryResult
                    :result="result"
                    v-if="hasResult"
                    @newSummary="newSummary"
                    @disconnect="disconnect"
                    @downloadCertificate="downloadCertificate"
                    @openCiscoSetup="openCiscoSetup"
                />
            </v-row>
        </v-container>
    </div>
</template>

<style scoped>
.customer-summary-page {
    min-height: 100svh;
}

.customer-summary-row {
    min-height: calc(100svh - 24px);
    padding-top: 16px;
    padding-bottom: 16px;
}

@media (min-width: 600px) {
    .customer-summary-row {
        min-height: calc(100svh - 48px);
    }
}
</style>
