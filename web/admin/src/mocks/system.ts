import type {
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemGetSystemInitResponse,
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemGetSystemResponse,
} from "@/api/generated";

export const mockSystemInit = {
  first_init: true,
  google_captcha_site_key: "",
  telegram_bot_enabled: true,
} satisfies GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemGetSystemInitResponse;

export const mockSystemConfig = {
  auto_delete_inactive_users: true,
  client_profile_connection_name: "Ocserv Test VPN",
  client_profile_server_address: "vpn.test.local",
  client_profile_server_port: 443,
  first_init: true,
  google_captcha_secret_key: "",
  google_captcha_site_key: "",
  keep_inactive_user_days: 30,
} satisfies GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemGetSystemResponse;
