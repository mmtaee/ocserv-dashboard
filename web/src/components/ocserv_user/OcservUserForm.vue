<script lang="ts" setup>
import {
    type ModelsOcservUser,
    type ModelsOcservUserConfig,
    ModelsOcservUserTrafficTypeEnum,
    type OcservUserCreateOcservUserData,
    type OcservUserUpdateOcservUserData
} from '@/api';
import { useI18n } from 'vue-i18n';
import { computed, reactive, ref, watch } from 'vue';
import { requiredRule } from '@/utils/rules';
import { formatDate } from '@/utils/convertors';
import { getFormFields } from '@/components/ocserv_user/items';

const props = defineProps({
    btnText: {
        type: String,
        default: 'create'
    },
    btnColor: {
        type: String,
        default: 'primary'
    },
    initData: {
        type: Object as () => ModelsOcservUser,
        required: false
    },
    loading: {
        type: Boolean,
        default: false
    },
    groups: {
        type: Array as () => String[],
        default: ['defaults']
    }
});

const emit = defineEmits(['createUser', 'updateUser']);

const { t } = useI18n();
const valid = ref(true);
const isUpdate = ref(false);
const rules = {
    required: (v: string) => requiredRule(v, t)
};

const showDateMenu = ref(false);
const showPassword = ref(false);
const fieldItems = getFormFields();
const chipInputs = reactive<Record<string, string>>({
    dns: '',
    route: '',
    'no-route': '',
    'split-dns': ''
});

const trafficTypes = ref([
    {
        label: t('FREE'),
        value: ModelsOcservUserTrafficTypeEnum.FREE
    },
    {
        label: t('MONTHLY_TRANSMIT'),
        value: ModelsOcservUserTrafficTypeEnum.MONTHLY_TRANSMIT
    },
    {
        label: t('MONTHLY_RECEIVE'),
        value: ModelsOcservUserTrafficTypeEnum.MONTHLY_RECEIVE
    },
    {
        label: t('TOTALLY_RECEIVE'),
        value: ModelsOcservUserTrafficTypeEnum.TOTALLY_RECEIVE
    },
    {
        label: t('TOTALLY_TRANSMIT'),
        value: ModelsOcservUserTrafficTypeEnum.TOTALLY_TRANSMIT
    }
]);

const createData = reactive<OcservUserCreateOcservUserData>({
    config: {} as ModelsOcservUserConfig,
    description: '',
    expire_at: undefined,
    group: 'defaults',
    password: '',
    traffic_size: 0,
    traffic_type: ModelsOcservUserTrafficTypeEnum.FREE,
    username: ''
});
const updateData = reactive<OcservUserUpdateOcservUserData>({});

const addRoutes = (key: string) => {
    const typedKey = key as keyof ModelsOcservUserConfig;
    const input = chipInputs[typedKey];

    if (input) {
        if (!createData.config) createData.config = {};
        if (!Array.isArray(createData.config[typedKey])) {
            createData.config[typedKey] = [] as any;
        }

        const arr = createData.config[typedKey] as string[];

        if (!arr.includes(input)) {
            arr.push(input);
            chipInputs[typedKey] = '';
        }
    }
};

const removeRoute = (key: string, value: string) => {
    if (!createData.config) createData.config = {};
    const typedKey = key as keyof ModelsOcservUserConfig;
    const arr = createData.config[typedKey] as string[];

    let index = arr.findIndex((i) => i == value);
    if (index !== -1) {
        arr.splice(index, 1);
    }
};

const createUser = () => {
    emit('createUser', createData);
};

const updateUser = () => {
    Object.assign(updateData, createData);
    emit('updateUser', props.initData?.uid, updateData);
};

const toggleUnlimited = (val: boolean) => {
    if (val) {
        createData.expire_at = undefined;
    }
};

const expireAtDate = computed<Date>({
    get: () => {
        return createData.expire_at ? new Date(createData.expire_at) : new Date();
    },
    set: (val: Date) => {
        createData.expire_at = formatDate(val);
    }
});

watch(
    () => props.initData,
    () => {
        if (props.initData !== undefined) {
            Object.assign(createData, props.initData);
            isUpdate.value = true;
            if (createData.expire_at == undefined) {
                createData.unlimited = true;
            }
        }
    },
    { immediate: false }
);
</script>

<template>
    <v-form v-model="valid">
        <v-row align="center" justify="start">
            <v-col cols="12">
                <h3 class="text-capitalize">{{ t('MAIN') }}</h3>
            </v-col>
            <v-col cols="12" lg="4" md="6">
                <v-label class="font-weight-bold mb-1 text-capitalize">{{ t('USERNAME') }}</v-label>
                <v-text-field
                    v-model="createData.username"
                    :readonly="isUpdate"
                    :rules="isUpdate ? [] : [rules.required]"
                    autocomplete="new-password"
                    color="primary"
                    hide-details
                    variant="outlined"
                />
            </v-col>
            <v-col cols="12" lg="4" md="6">
                <v-label class="font-weight-bold mb-1 text-capitalize">{{ t('PASSWORD') }}</v-label>
                <v-text-field
                    v-model="createData.password"
                    :append-inner-icon="showPassword ? 'mdi-eye-off-outline' : 'mdi-eye-outline'"
                    :rules="[rules.required]"
                    :type="showPassword ? 'text' : 'password'"
                    autocomplete="new-password"
                    color="primary"
                    hide-details
                    variant="outlined"
                    @click:append-inner="showPassword = !showPassword"
                />
            </v-col>
            <v-col cols="12" lg="4" md="6">
                <v-label class="font-weight-bold mb-1 text-capitalize">{{ t('GROUP') }}</v-label>
                <v-select
                    v-model="createData.group"
                    :items="groups"
                    :rules="[rules.required]"
                    color="primary"
                    hide-details
                    variant="outlined"
                />
            </v-col>
            <v-col cols="12" lg="4" md="6">
                <v-label class="font-weight-bold mb-1 text-capitalize">{{ t('TRAFFIC_TYPE') }}</v-label>
                <v-select
                    v-model="createData.traffic_type"
                    :items="trafficTypes"
                    :rules="[rules.required]"
                    color="primary"
                    hide-details
                    item-title="label"
                    item-value="value"
                    variant="outlined"
                    @update:modelValue="
                        (v) => (v == ModelsOcservUserTrafficTypeEnum.FREE ? (createData.traffic_size = 0) : false)
                    "
                />
            </v-col>
            <v-col cols="12" lg="4" md="6">
                <v-label class="font-weight-bold mb-1 text-capitalize">{{ t('TRAFFIC_SIZE') }}</v-label>
                <v-text-field
                    v-model.number="createData.traffic_size"
                    :disabled="createData.traffic_type == ModelsOcservUserTrafficTypeEnum.FREE"
                    :rules="createData.traffic_type == ModelsOcservUserTrafficTypeEnum.FREE ? [] : [rules.required]"
                    color="primary"
                    hide-details
                    suffix="GB"
                    type="number"
                    variant="outlined"
                    @update:modelValue="
                        (val: any) => {
                            createData.traffic_size = Boolean(val) ? (Number(val) as any) : null;
                        }
                    "
                />
            </v-col>
            <v-col cols="12" lg="4" md="6">
                <v-row align="center" justify="start">
                    <v-col cols="12" md="8">
                        <v-menu v-model="showDateMenu" :close-on-content-click="false" transition="scale-transition">
                            <template #activator="{ props }">
                                <v-label class="font-weight-bold mb-1 text-capitalize">
                                    {{ t('EXPIRE_AT') }}
                                </v-label>
                                <v-text-field
                                    :disabled="createData.unlimited"
                                    :model-value="createData.expire_at ? formatDate(createData.expire_at) : ''"
                                    color="primary"
                                    hide-details
                                    readonly
                                    v-bind="props"
                                    variant="outlined"
                                    clearable
                                    @click:clear="createData.expire_at = ''"
                                />
                            </template>
                            <v-date-picker
                                :disabled="createData.unlimited"
                                v-model="expireAtDate"
                                :header="t('EXPIRE_AT')"
                                elevation="24"
                                title=""
                                @update:model-value="() => (showDateMenu = false)"
                            />
                        </v-menu>
                    </v-col>
                    <v-col cols="12" md="auto" class="mt-md-6 ma-sm-0 pa-sm-0">
                        <v-checkbox
                            v-model="createData.unlimited"
                            :label="t('UNLIMITED')"
                            class="text-capitalize text-subtitle-2"
                            color="primary"
                            hide-details
                            @update:model-value="(val: unknown) => toggleUnlimited(val as boolean)"
                        />
                    </v-col>
                </v-row>
            </v-col>

            <v-col cols="12" md="11">
                <h3 class="text-capitalize">{{ t('NETWORK_CONFIGURATION') }}</h3>
            </v-col>
            <template v-for="field in fieldItems.fields.filter((f) => f.type === 'text')" :key="field.key">
                <v-col cols="12" lg="4" md="6">
                    <v-label class="font-weight-bold mb-1 text-capitalize">{{ field.label }}</v-label>
                    <v-text-field
                        v-model="createData.config[field.key as keyof ModelsOcservUserConfig]"
                        :hint="field.hint"
                        :placeholder="field.example"
                        :rules="field.rules"
                        color="primary"
                        variant="outlined"
                    />
                </v-col>
            </template>

            <v-col cols="12" md="11">
                <h3 class="text-capitalize">{{ t('PERFORMANCE_AND_SESSION_SETTINGS') }}</h3>
            </v-col>
            <template v-for="field in fieldItems.fields.filter((f) => f.type === 'number')" :key="field.key">
                <v-col cols="12" lg="4" md="6">
                    <v-label class="font-weight-bold mb-1 text-capitalize">{{ field.label }}</v-label>
                    <v-text-field
                        v-model.number="createData.config[field.key as keyof ModelsOcservUserConfig]"
                        :hint="field.hint"
                        color="primary"
                        min="0"
                        type="number"
                        variant="outlined"
                        @update:modelValue="
                            (val: any) => {
                                createData.config[field.key as keyof ModelsOcservUserConfig] = Boolean(val)
                                    ? (Number(val) as any)
                                    : null;
                            }
                        "
                    />
                </v-col>
            </template>

            <v-col cols="12" md="11">
                <h3 class="text-capitalize">{{ t('ACCESS_AND_FEATURE_CONTROLS') }}</h3>
            </v-col>
            <template v-for="field in fieldItems.fields.filter((f) => f.type === 'switch')" :key="field.key">
                <v-col cols="12" md="3">
                    <v-row align="center" justify="center">
                        <v-col cols="6" md="12">
                            <v-label class="font-weight-bold mb-1 text-capitalize">{{ field.label }}</v-label>
                            <v-switch
                                v-model="createData.config[field.key as keyof ModelsOcservUserConfig]"
                                :hint="field.hint"
                                color="primary"
                                variant="outlined"
                            />
                        </v-col>
                    </v-row>
                </v-col>
            </template>

            <v-col cols="12" md="11">
                <h3 class="text-capitalize">{{ t('ROUTES') }}</h3>
            </v-col>
            <template v-for="field in fieldItems.textFields" :key="field.key">
                <v-col cols="12">
                    <v-col cols="12" md="3" sm="12">
                        <v-label class="font-weight-bold mb-1 text-capitalize">{{ field.label }}</v-label>
                        <v-text-field
                            v-model="chipInputs[field.key]"
                            :hint="field.hint"
                            :placeholder="field.example"
                            :rules="field.rules"
                            append-inner-icon="mdi-plus-circle-outline"
                            color="primary"
                            variant="outlined"
                            @keydown.enter="addRoutes(field.key)"
                            @click:append-inner="addRoutes(field.key)"
                        />
                    </v-col>
                    <v-col class="pa-0 px-3 ma-0">
                        <v-card class="overflow-y-auto" height="180">
                            <v-card-title class="text-subtitle-2 pa-3"> {{ field.label }}:</v-card-title>
                            <v-card-text>
                                <v-chip
                                    v-for="chip in createData.config[field.key as keyof ModelsOcservUserConfig]"
                                    :key="`${field.key}-${chip}`"
                                    class="me-2 my-1"
                                    color="primary"
                                >
                                    {{ chip }}
                                    <v-icon color="error" end @click="removeRoute(field.key, chip)">mdi-delete</v-icon>
                                </v-chip>
                            </v-card-text>
                        </v-card>
                    </v-col>
                </v-col>
            </template>
        </v-row>
    </v-form>

    <v-row align="center" class="me-0 mt-5" justify="end">
        <v-col cols="auto">
            <v-btn
                :color="btnColor"
                :disabled="!valid"
                :loading="loading"
                @click="isUpdate ? updateUser() : createUser()"
            >
                {{ btnText }}
            </v-btn>
        </v-col>
    </v-row>
</template>
