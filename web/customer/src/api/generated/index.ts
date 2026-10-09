export type ExpiryMode = "unlimited" | "fixed" | "first_connection";
export type TrafficType =
  | "Free"
  | "MonthlyTransmit"
  | "MonthlyReceive"
  | "MonthlyRxTx"
  | "TotallyTransmit"
  | "TotallyReceive"
  | "TotallyRxTx";
export type SessionEvent =
  "user-agent" | "handshake" | "periodic-stats" | "disconnect";

export interface Customer {
  certificate_available?: boolean;
  certificate_enabled?: boolean;
  deactivated_at?: string;
  expire_at?: string;
  expire_days_after_first_connection?: number;
  expiry_mode?: ExpiryMode;
  first_connected_at?: string;
  is_locked?: boolean;
  owner?: string;
  running_rx?: number;
  running_tx?: number;
  traffic_size?: number;
  traffic_type?: TrafficType;
  username?: string;
}
export interface Bandwidth {
  rx: number;
  tx: number;
}
export interface Usage {
  bandwidths?: Bandwidth;
  date_end?: string;
  date_start?: string;
}
export interface SummaryResponse {
  ocserv_user?: Customer;
  usage?: Usage;
}
export interface LoginData {
  username: string;
  password: string;
}
export interface LoginResponse {
  expires_at: string;
  token: string;
  user: Customer;
}
export interface ChangePasswordInput {
  password: string;
}
export interface RequestMeta {
  page: number;
  size: number;
  total_records: number;
}
export interface SessionLog {
  created_at: string;
  event: SessionEvent;
  ip?: string;
  message: string;
  username: string;
}
export interface ActivitiesResponse {
  meta: RequestMeta;
  result?: SessionLog[];
}
export interface DailyTraffic {
  date?: string;
  rx?: number;
  tx?: number;
}
export interface OnlineUserSession {
  "Average RX"?: string;
  "Average TX"?: string;
  Device: string;
  Groupname?: string;
  ID: number;
  IPv4: string;
  "Session started at": string;
  Username?: string;
  "_Last connected at"?: string;
  vhost: string;
}
export interface CiscoSetup {
  certificate_import_uri?: string;
  certificate_password?: string;
  connection_create_uri?: string;
  connection_name?: string;
  server_address?: string;
  server_port?: number;
}
export interface DateRangeQuery {
  date_start?: string;
  date_end?: string;
}
export interface ActivitiesQuery extends DateRangeQuery {
  page?: number;
  size?: number;
}
