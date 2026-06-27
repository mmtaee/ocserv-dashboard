import type { OcservUsersGetSortEnum } from '@/api';

interface Meta {
    page: number;
    size: number;
    order?: string;
    sort: OcservUsersGetSortEnum;
    total_records: number;
}

export type { Meta };
