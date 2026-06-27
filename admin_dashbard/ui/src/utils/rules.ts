import type { Validator } from '@/types/ruleTypes/RuleType';

const ipOrRangeRule: Validator = (v, t) => {
    const ipFormat =
        /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(?:\/(?:3[0-2]|[1-2]?[0-9]))?$|^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.)?(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(?:\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){2}$/;
    return v && !v.match(ipFormat) ? t('IP_FORMAT_WITH_RANGE_REQUIRED') : true;
};

const requiredRule: Validator = (v, t) => {
    return !!v || t('FIELD_REQUIRED');
};

const numberRule: Validator = (v, t) => {
    return v && isNaN(Number(v)) ? t('FIELD_NUMBER_REQUIRED') : true;
};

const ipWithNetmaskRule: Validator = (v, t) => {
    const ipv4Segment = '(25[0-5]|2[0-4]\\d|1\\d{2}|[1-9]?\\d)';
    const ipv4 = `(${ipv4Segment}\\.){3}${ipv4Segment}`;
    const pattern = new RegExp(`^${ipv4}/${ipv4}$`);

    return v && !pattern.test(v) ? t('IP_WITH_NETMASK_FORMAT_REQUIRED') : true;
};

const ipRule: Validator = (v, t) => {
    const ipFormat =
        /^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
    return v && !v.match(ipFormat) ? t('IP_FORMAT_REQUIRED') : true;
};

const domainRule: Validator = (v, t) => {
    const domainRegex = /^(?!:\/\/)([a-zA-Z0-9-]{1,63}\.)+[a-zA-Z]{2,}$/;
    return v && !domainRegex.test(v) ? t('DOMAIN_FORMAT_REQUIRED') : true;
};

const ipWithRangeRule: Validator = (v, t) => {
    if (!v) return true; // Allow empty if not required

    const ipSegment = '(25[0-5]|2[0-4]\\d|1\\d{2}|[1-9]?\\d)';
    const ipPattern = `(${ipSegment}\\.){3}${ipSegment}`;

    const cidrRegex = new RegExp(`^${ipPattern}/([0-9]|[1-2][0-9]|3[0-2])$`);
    const netmaskRegex = new RegExp(`^${ipPattern}/${ipPattern}$`);

    if (cidrRegex.test(v) || netmaskRegex.test(v)) {
        return true;
    }

    return t('(IP/SUBNET)_or_(IP/RANGE)_FORMAT_REQUIRED');
};

export { ipOrRangeRule, requiredRule, numberRule, ipRule, ipWithNetmaskRule, domainRule, ipWithRangeRule };
