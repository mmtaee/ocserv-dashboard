<script lang="ts" setup>
import { reactive, ref, watch } from 'vue';
import { type ModelsOcservGroup, type ModelsOcservGroupConfig, type OcservGroupCreateOcservGroupData } from '@/api';
import { useI18n } from 'vue-i18n';
import { requiredRule } from '@/utils/rules';
import { getFormFields } from '@/components/ocserv_group/items';

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
        type: Object as () => ModelsOcservGroup,
        required: false
    },
    loading: {
        type: Boolean,
        default: false
    }
});

const emit = defineEmits(['createGroup', 'updateGroup']);

const { t } = useI18n();
const valid = ref(true);
const isUpdate = ref(false);
const rules = {
    required: (v: string) => requiredRule(v, t)
};

const createData = reactive<OcservGroupCreateOcservGroupData>({ config: {}, name: '' });

const fieldItems = getFormFields();
const chipInputs = reactive<Record<string, string>>({
    dns: '',
    route: '',
    'no-route': '',
    'split-dns': ''
});

const createGroup = () => {
    emit('createGroup', createData);
};

const updateGroup = () => {
    emit('updateGroup', props.initData?.id, createData.config);
};

const addRoutes = (key: string) => {
    const typedKey = key as keyof ModelsOcservGroupConfig;
    const input = chipInputs[typedKey];

    if (input) {
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
    const typedKey = key as keyof ModelsOcservGroupConfig;
    const arr = createData.config[typedKey] as string[];

    let index = arr.findIndex((i) => i == value);
    if (index !== -1) {
        arr.splice(index, 1);
    }
};

watch(
    () => props.initData,
    () => {
        console.log('props.initData: ', props.initData);
        if (props.initData !== undefined) {
            Object.assign(createData, props.initData);
            isUpdate.value = true;
        }
    },
    { immediate: false }
);
</script>

<template>
    <v-form v-model="valid">
        <v-row align="center" justify="start">
            <v-col cols="12" v-if="createData.name">
                <h3 class="text-capitalize">{{ t('MAIN') }}</h3>
            </v-col>
            <v-col cols="12" lg="4" md="6" v-if="createData.name">
                <v-label class="font-weight-bold mb-1 text-capitalize">{{ t('GROUP_NAME') }}</v-label>
                <v-text-field
                    v-model="createData.name"
                    :readonly="isUpdate"
                    :rules="isUpdate ? [] : [rules.required]"
                    color="primary"
                    hide-details
                    variant="outlined"
                />
            </v-col>
            <v-col cols="12" md="11">
                <h3 class="text-capitalize">{{ t('NETWORK_CONFIGURATION') }}</h3>
            </v-col>
            <template v-for="field in fieldItems.fields.filter((f) => f.type === 'text')" :key="field.key">
                <v-col cols="12" lg="4" md="6">
                    <v-label class="font-weight-bold mb-1 text-capitalize">{{ field.label }}</v-label>
                    <v-text-field
                        v-model="createData.config[field.key as keyof ModelsOcservGroupConfig]"
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
                        v-model.number="createData.config[field.key as keyof ModelsOcservGroupConfig]"
                        :hint="field.hint"
                        color="primary"
                        min="0"
                        type="number"
                        variant="outlined"
                        @update:modelValue="
                            (val: any) => {
                                createData.config[field.key as keyof ModelsOcservGroupConfig] = Boolean(val)
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
                                v-model="createData.config[field.key as keyof ModelsOcservGroupConfig]"
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
                                    v-for="chip in createData.config[field.key as keyof ModelsOcservGroupConfig]"
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
                @click="isUpdate ? updateGroup() : createGroup()"
            >
                {{ btnText }}
            </v-btn>
        </v-col>
    </v-row>
</template>
