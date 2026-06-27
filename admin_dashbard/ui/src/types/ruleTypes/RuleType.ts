export type Validator = (value: string, t: (key: string) => string) => true | string;
