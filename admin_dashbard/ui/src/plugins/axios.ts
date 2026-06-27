import axios, {
    type AxiosInstance,
    type AxiosRequestConfig,
    type AxiosResponse,
    type InternalAxiosRequestConfig
} from 'axios';
import { useI18n } from 'vue-i18n';
import { useSnackbarStore } from '@/stores/snackbar';
import type { SnackbarItem } from '@/types/storeTypes/StoreSnackbarType';
import { useLoadingStore } from '@/stores/loading';

const host = window.location.host; // e.g. "example.com:3000"
const protocol = window.location.protocol; // "http:" or "https:"
export const ApiUrl = import.meta.env.VITE_API_URL ?? `${protocol}//${host}/api`;

const config: AxiosRequestConfig = { baseURL: ApiUrl };

const api: AxiosInstance = axios.create(config);

api.interceptors.request.use(
    (config: InternalAxiosRequestConfig): InternalAxiosRequestConfig => {
        const loadingStore = useLoadingStore();

        const skipLoading = config.headers?.['x-skip-loading'] === 'true';

        if (!skipLoading) {
            loadingStore.show();
        } else {
            delete config.headers?.['x-skip-loading'];
        }

        return config;
    },
    (error) => {
        const loadingStore = useLoadingStore();
        loadingStore.hide();
        return Promise.reject(error);
    }
);

// Response interceptor: Handle 401/400
api.interceptors.response.use(
    (response: AxiosResponse): AxiosResponse => {
        const loadingStore = useLoadingStore();
        loadingStore.hide();
        return response;
    },
    (error) => {
        const loadingStore = useLoadingStore();
        loadingStore.hide();
        const { response } = error;
        if (response) {
            if (response.status === 401) {
                localStorage.removeItem('token');
                window.location.href = '/login';
            }
            if (response.status === 400) {
                if (
                    !window.location.pathname.startsWith('/summary') &&
                    response.data.hasOwnProperty('error') &&
                    response.data['error'] == 'record not found'
                ) {
                    window.location.href = '/not_found';
                }
                const snackbar = useSnackbarStore();
                let errorList = response.data.error;
                if (Array.isArray(errorList)) {
                    const messages: SnackbarItem[] = errorList.map((message: string, index: number) => ({
                        id: index + 1,
                        message,
                        color: 'error',
                        timeout: (index + 1) * 1000 + 4000
                    }));
                    snackbar.show(messages);
                } else {
                    const { t } = useI18n();
                    snackbar.show({
                        id: 1,
                        message: typeof errorList === 'string' ? errorList : t('UNKNOWN_ERROR'),
                        color: 'error',
                        timeout: 4000
                    });
                }
            }
            if (response.status == 404) {
                window.location.href = '/not_found';
            }
        }
        return Promise.reject(error);
    }
);

export default api;
