import api from '@/plugins/axios';

export interface TelegramSettings {
    enabled: boolean;
    bot_token: string;
    bot_username: string;
    admin_chat_id: number;
    low_quota_threshold_mb: number;
    default_language: string;
    ocserv_host: string;
    card_number: string;
    card_holder: string;
    support_username: string;
}

export interface TelegramSettingsPatch {
    enabled?: boolean;
    bot_token?: string;
    admin_chat_id?: number;
    low_quota_threshold_mb?: number;
    default_language?: string;
    ocserv_host?: string;
    card_number?: string;
    card_holder?: string;
    support_username?: string;
}

export interface TelegramPackage {
    id: number;
    title: string;
    days: number;
    traffic_size_gb: number;
    traffic_type: string;
    price_text: string;
    is_active: boolean;
    created_at?: string;
    updated_at?: string;
}

export interface TelegramRequestModel {
    id: number;
    chat_id: number;
    telegram_username: string;
    type: 'new' | 'renew';
    package_id?: number;
    target_ocserv_id?: number;
    desired_username: string;
    status:
        | 'pending'
        | 'awaiting_payment'
        | 'payment_uploaded'
        | 'approved'
        | 'rejected'
        | 'delivered';
    receipt_file_path: string;
    user_message: string;
    admin_note: string;
    delivered_at?: string;
    awaiting_payment_message_id?: number;
    created_at: string;
    updated_at: string;
}

export interface TelegramRequestsResponse {
    meta: { page: number; size: number; total_records: number };
    result: TelegramRequestModel[];
}

export interface TelegramAccount {
    id: number;
    chat_id: number;
    telegram_username: string;
    language: string;
    ocserv_user_id: number;
    created_at: string;
    last_low_quota_notified_at?: string;
}

const auth = () => ({
    headers: { Authorization: `Bearer ${localStorage.getItem('token') ?? ''}` }
});

export const TelegramAPI = {
    // Settings
    getSettings: () => api.get<TelegramSettings>('/telegram/settings', auth()),
    updateSettings: (payload: TelegramSettingsPatch) =>
        api.patch<TelegramSettings>('/telegram/settings', payload, auth()),
    test: (message?: string) =>
        api.post<{ status: string }>('/telegram/test', { message: message ?? '' }, auth()),

    // Packages
    listPackages: (includeInactive = true) =>
        api.get<TelegramPackage[]>(`/telegram/packages?include_inactive=${includeInactive}`, auth()),
    createPackage: (payload: Partial<TelegramPackage>) =>
        api.post<TelegramPackage>('/telegram/packages', payload, auth()),
    updatePackage: (id: number, payload: Partial<TelegramPackage>) =>
        api.patch<TelegramPackage>(`/telegram/packages/${id}`, payload, auth()),
    deletePackage: (id: number) => api.delete(`/telegram/packages/${id}`, auth()),

    // Requests
    listRequests: (
        params: {
            status?: string;
            type?: string;
            page?: number;
            size?: number;
            sort?: string;
            order?: string;
        } = {}
    ) => {
        const query = new URLSearchParams();
        if (params.status) query.set('status', params.status);
        if (params.type) query.set('type', params.type);
        if (params.page) query.set('page', String(params.page));
        if (params.size) query.set('size', String(params.size));
        if (params.sort) query.set('sort', params.sort);
        if (params.order) query.set('order', params.order);
        const qs = query.toString();
        return api.get<TelegramRequestsResponse>(`/telegram/requests${qs ? `?${qs}` : ''}`, auth());
    },
    getRequest: (id: number) =>
        api.get<TelegramRequestModel>(`/telegram/requests/${id}`, auth()),
    receiptUrl: (id: number) => {
        const base = (api.defaults.baseURL || '').replace(/\/$/, '');
        const token = localStorage.getItem('token') ?? '';
        return `${base}/telegram/requests/${id}/receipt?_=${encodeURIComponent(token)}`;
    },
    fetchReceiptBlob: async (id: number) => {
        const res = await api.get(`/telegram/requests/${id}/receipt`, {
            ...auth(),
            responseType: 'blob'
        });
        return res.data as Blob;
    },
    approve: (
        id: number,
        payload?: {
            admin_note?: string;
            card_number?: string;
            card_holder?: string;
            reply_to_user?: string;
        }
    ) =>
        api.post<TelegramRequestModel>(
            `/telegram/requests/${id}/approve`,
            payload ?? {},
            auth()
        ),
    reject: (id: number, adminNote?: string) =>
        api.post<TelegramRequestModel>(
            `/telegram/requests/${id}/reject`,
            { admin_note: adminNote ?? '' },
            auth()
        ),
    confirmPayment: (
        id: number,
        payload: {
            override_username?: string;
            override_password?: string;
            owner?: string;
            group?: string;
            admin_note?: string;
        } = {}
    ) =>
        api.post<{ status: string; username: string }>(
            `/telegram/requests/${id}/confirm-payment`,
            payload,
            auth()
        ),
    deleteRequest: (id: number) => api.delete(`/telegram/requests/${id}`, auth()),

    // Linked accounts
    accountsForUser: (ocservUserUid: string) =>
        api.get<TelegramAccount[]>(
            `/telegram/accounts?ocserv_user_uid=${encodeURIComponent(ocservUserUid)}`,
            auth()
        ),
    deleteAccount: (id: number) => api.delete(`/telegram/accounts/${id}`, auth())
};
