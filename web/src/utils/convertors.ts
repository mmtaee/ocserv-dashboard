import { ModelsOcservUserTrafficTypeEnum } from '@/api';
import { useI18n } from 'vue-i18n';

const numberToFixer = (n: number, fixture: number = 4): string => {
    if (n === 0) return '0';

    const threshold = 1 / 10 ** fixture;
    if (Math.abs(n) < threshold) return '0';

    return n.toFixed(fixture);
};

const bytesToGB = (bytes: number, fixture: number = 6): string => {
    if (bytes === 0) return '0';

    const result = bytes / 1024 ** 3;
    return numberToFixer(result, fixture);
};

const formatDateTime = (dateString: string | undefined, message: string | undefined): string => {
    if (!dateString) {
        return message || '';
    }
    const date = new Date(dateString);
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');

    return `${year}-${month}-${day} ${hours}:${minutes}`;
};

const formatDate = (date: Date | string | null | undefined): string => {
    if (!date) return '';

    // If a string is passed, convert to Date
    const d = typeof date === 'string' ? new Date(date) : date;

    if (isNaN(d.getTime())) return ''; // invalid date

    const year = d.getFullYear();
    const month = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');

    return `${year}-${month}-${day}`;
};

const formatDateTimeWithRelative = (dateString: string | undefined, message: string | undefined): string => {
    if (!dateString) {
        return message || '';
    }

    const { t } = useI18n();
    const formatted = formatDateTime(dateString, message);
    const date = new Date(dateString);
    const now = new Date();

    // Calculate difference in milliseconds
    const diffTime = now.getTime() - date.getTime();

    // Helper to get full year/month/day difference
    const diffYears = now.getFullYear() - date.getFullYear();
    const diffMonths = (now.getFullYear() - date.getFullYear()) * 12 + (now.getMonth() - date.getMonth());
    const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24));

    let relative = '';

    if (diffDays === 0) {
        relative = t('TODAY');
    } else if (diffDays === 1) {
        relative = t('YESTERDAY');
    } else if (diffDays === -1) {
        relative = t('TOMORROW');
    } else if (Math.abs(diffYears) >= 1) {
        if (diffYears > 0) {
            relative = `${diffYears} year${diffYears > 1 ? 's' : ''} ago`;
        } else {
            relative = `in ${Math.abs(diffYears)} year${Math.abs(diffYears) > 1 ? 's' : ''}`;
        }
    } else if (Math.abs(diffMonths) >= 1) {
        if (diffMonths > 0) {
            relative = `${diffMonths} month${diffMonths > 1 ? 's' : ''} ago`;
        } else {
            relative = `in ${Math.abs(diffMonths)} month${Math.abs(diffMonths) > 1 ? 's' : ''}`;
        }
    } else {
        if (diffDays > 1) {
            relative = `${diffDays} days ago`;
        } else if (diffDays < -1) {
            relative = `in ${Math.abs(diffDays)} days`;
        }
    }

    return `${formatted} (${relative})`;
};

const formatDateWithRelative = (dateString: string | undefined, message: string | undefined): string => {
    if (!dateString) {
        return message || '';
    }

    const { t } = useI18n();
    const date = new Date(dateString);
    const now = new Date();

    // Strip time to compare only dates
    const dateOnly = new Date(date.getFullYear(), date.getMonth(), date.getDate());
    const nowOnly = new Date(now.getFullYear(), now.getMonth(), now.getDate());

    const diffTime = nowOnly.getTime() - dateOnly.getTime();
    const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24));

    let relative = '';

    if (diffDays === 0) {
        relative = t('TODAY');
    } else if (diffDays === 1) {
        relative = t('YESTERDAY');
    } else if (diffDays === -1) {
        relative = t('TOMORROW');
    } else if (diffDays > 1) {
        relative = `${diffDays} ${t('DAYS_AGO')}`;
    } else if (diffDays < -1) {
        relative = `in ${Math.abs(diffDays)} ${t('DAYS')}`;
    }

    // Format date only (e.g., YYYY-MM-DD)
    const formatted = formatDate(dateString);

    return `${formatted} (${relative})`;
};

const trafficTypesTransformer = (item: ModelsOcservUserTrafficTypeEnum): string => {
    const { t } = useI18n();

    switch (item) {
        case ModelsOcservUserTrafficTypeEnum.FREE:
            return t('FREE');
        case ModelsOcservUserTrafficTypeEnum.MONTHLY_TRANSMIT:
            return t('MONTHLY_TRANSMIT');
	case ModelsOcservUserTrafficTypeEnum.MONTHLY_RECEIVE:
            return t('MONTHLY_RECEIVE');
        case ModelsOcservUserTrafficTypeEnum.MONTHLY_RX_TX:
            return t('MONTHLY_RX_TX');
        case ModelsOcservUserTrafficTypeEnum.TOTALLY_RECEIVE:
            return t('TOTALLY_RECEIVE');
        case ModelsOcservUserTrafficTypeEnum.TOTALLY_TRANSMIT:
            return t('TOTALLY_TRANSMIT');
        case ModelsOcservUserTrafficTypeEnum.TOTALLY_RX_TX:
            return t('TOTALLY_RX_TX');
        default:
            return item;
    }
};

const toISODateString = (date: Date): string => {
    date.setHours(0, 0, 0, 0); // reset to midnight
    return date.toISOString().split('T')[0]; // keep only YYYY-MM-DD
};

export {
    bytesToGB,
    formatDateTime,
    formatDate,
    formatDateTimeWithRelative,
    formatDateWithRelative,
    trafficTypesTransformer,
    numberToFixer,
    toISODateString
};
