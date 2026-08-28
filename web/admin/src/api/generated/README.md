## @ocserv-dashboard/admin-api@1.0.0

This generator creates TypeScript/JavaScript client that utilizes [axios](https://github.com/axios/axios). The generated Node module can be used in the following environments:

Environment
* Node.js
* Webpack
* Browserify

Language level
* ES5 - you must have a Promises/A+ library installed
* ES6

Module system
* CommonJS
* ES6 module system

It can be used in both TypeScript and JavaScript. In TypeScript, the definition will be automatically resolved via `package.json`. ([Reference](https://www.typescriptlang.org/docs/handbook/declaration-files/consumption.html))

### Building

To build and compile the typescript sources to javascript use:
```
npm install
npm run build
```

### Publishing

First build the package then run `npm publish`

### Consuming

navigate to the folder of your consuming project and run one of the following commands.

_published:_

```
npm install @ocserv-dashboard/admin-api@1.0.0 --save
```

_unPublished (not recommended):_

```
npm install PATH_TO_GENERATED_PACKAGE --save
```

### Documentation for API Endpoints

All URIs are relative to *http://localhost*

Class | Method | HTTP request | Description
------------ | ------------- | ------------- | -------------
*AuthApi* | [**authLogoutPost**](docs/AuthApi.md#authlogoutpost) | **POST** /auth/logout | Logout current session
*AuthApi* | [**authSessionsGet**](docs/AuthApi.md#authsessionsget) | **GET** /auth/sessions | List active user sessions
*AuthApi* | [**authSessionsIdDelete**](docs/AuthApi.md#authsessionsiddelete) | **DELETE** /auth/sessions/{id} | Revoke user session
*HomeApi* | [**homeContainerStatsGet**](docs/HomeApi.md#homecontainerstatsget) | **GET** /home/container-stats | Content of docker system usage stats
*HomeApi* | [**homeGet**](docs/HomeApi.md#homeget) | **GET** /home | Content of home
*HomeApi* | [**homeOcservStatsGet**](docs/HomeApi.md#homeocservstatsget) | **GET** /home/ocserv-stats | Content of ocserv server stats
*HomeApi* | [**homeSystemStatsGet**](docs/HomeApi.md#homesystemstatsget) | **GET** /home/system-stats | Content of os system usage stats
*OCCTLApi* | [**occtlCommandsGet**](docs/OCCTLApi.md#occtlcommandsget) | **GET** /occtl/commands | Occtl Commands
*OCCTLApi* | [**occtlServerInfoGet**](docs/OCCTLApi.md#occtlserverinfoget) | **GET** /occtl/server_info | Server information
*OcservAgentsApi* | [**ocservAgentsGet**](docs/OcservAgentsApi.md#ocservagentsget) | **GET** /ocserv/agents | List Ocserv agents
*OcservAgentsApi* | [**ocservAgentsIdDelete**](docs/OcservAgentsApi.md#ocservagentsiddelete) | **DELETE** /ocserv/agents/{id} | Delete Ocserv agent
*OcservAgentsApi* | [**ocservAgentsIdGet**](docs/OcservAgentsApi.md#ocservagentsidget) | **GET** /ocserv/agents/{id} | Get Ocserv agent
*OcservAgentsApi* | [**ocservAgentsIdPatch**](docs/OcservAgentsApi.md#ocservagentsidpatch) | **PATCH** /ocserv/agents/{id} | Update Ocserv agent
*OcservAgentsApi* | [**ocservAgentsPost**](docs/OcservAgentsApi.md#ocservagentspost) | **POST** /ocserv/agents | Create Ocserv agent
*OcservGroupsApi* | [**ocservGroupsDefaultsGet**](docs/OcservGroupsApi.md#ocservgroupsdefaultsget) | **GET** /ocserv/groups/defaults | Ocserv Defaults Group config
*OcservGroupsApi* | [**ocservGroupsDefaultsPatch**](docs/OcservGroupsApi.md#ocservgroupsdefaultspatch) | **PATCH** /ocserv/groups/defaults | Update Ocserv Defaults Group
*OcservGroupsApi* | [**ocservGroupsGet**](docs/OcservGroupsApi.md#ocservgroupsget) | **GET** /ocserv/groups | List of Ocserv groups
*OcservGroupsApi* | [**ocservGroupsIdDelete**](docs/OcservGroupsApi.md#ocservgroupsiddelete) | **DELETE** /ocserv/groups/{id} | Ocserv Group delete
*OcservGroupsApi* | [**ocservGroupsIdGet**](docs/OcservGroupsApi.md#ocservgroupsidget) | **GET** /ocserv/groups/{id} | Ocserv group detail
*OcservGroupsApi* | [**ocservGroupsIdPatch**](docs/OcservGroupsApi.md#ocservgroupsidpatch) | **PATCH** /ocserv/groups/{id} | Ocserv Group update
*OcservGroupsApi* | [**ocservGroupsLookupGet**](docs/OcservGroupsApi.md#ocservgroupslookupget) | **GET** /ocserv/groups/lookup | List of Ocserv group names
*OcservGroupsApi* | [**ocservGroupsPost**](docs/OcservGroupsApi.md#ocservgroupspost) | **POST** /ocserv/groups | Ocserv Group creation
*OcservOcpasswdApi* | [**ocservUsersOcpasswdGet**](docs/OcservOcpasswdApi.md#ocservusersocpasswdget) | **GET** /ocserv/users/ocpasswd | 
*OcservOcpasswdApi* | [**ocservUsersOcpasswdSyncPost**](docs/OcservOcpasswdApi.md#ocservusersocpasswdsyncpost) | **POST** /ocserv/users/ocpasswd/sync | 
*OcservUnsyncedGroupApi* | [**ocservGroupsSyncPost**](docs/OcservUnsyncedGroupApi.md#ocservgroupssyncpost) | **POST** /ocserv/groups/sync | Ocserv Groups from file
*OcservUnsyncedGroupApi* | [**ocservGroupsUnsyncedGet**](docs/OcservUnsyncedGroupApi.md#ocservgroupsunsyncedget) | **GET** /ocserv/groups/unsynced | list of Unsynced Groups
*OcservUsersApi* | [**ocservUsersBulkDelete**](docs/OcservUsersApi.md#ocservusersbulkdelete) | **DELETE** /ocserv/users/bulk | Bulk delete Ocserv users
*OcservUsersApi* | [**ocservUsersBulkGroupPatch**](docs/OcservUsersApi.md#ocservusersbulkgrouppatch) | **PATCH** /ocserv/users/bulk/group | Bulk assign or remove an Ocserv group
*OcservUsersApi* | [**ocservUsersBulkPatch**](docs/OcservUsersApi.md#ocservusersbulkpatch) | **PATCH** /ocserv/users/bulk | Bulk update Ocserv users
*OcservUsersApi* | [**ocservUsersBulkStatusPatch**](docs/OcservUsersApi.md#ocservusersbulkstatuspatch) | **PATCH** /ocserv/users/bulk/status | Bulk enable or disable Ocserv users
*OcservUsersApi* | [**ocservUsersGet**](docs/OcservUsersApi.md#ocservusersget) | **GET** /ocserv/users | 
*OcservUsersApi* | [**ocservUsersIdActivatePost**](docs/OcservUsersApi.md#ocservusersidactivatepost) | **POST** /ocserv/users/{id}/activate | 
*OcservUsersApi* | [**ocservUsersIdCertificateGet**](docs/OcservUsersApi.md#ocservusersidcertificateget) | **GET** /ocserv/users/{id}/certificate | 
*OcservUsersApi* | [**ocservUsersIdCertificatePost**](docs/OcservUsersApi.md#ocservusersidcertificatepost) | **POST** /ocserv/users/{id}/certificate | 
*OcservUsersApi* | [**ocservUsersIdDelete**](docs/OcservUsersApi.md#ocservusersiddelete) | **DELETE** /ocserv/users/{id} | 
*OcservUsersApi* | [**ocservUsersIdDisconnectByIdPost**](docs/OcservUsersApi.md#ocservusersiddisconnectbyidpost) | **POST** /ocserv/users/{id}/disconnect_by_id | 
*OcservUsersApi* | [**ocservUsersIdGet**](docs/OcservUsersApi.md#ocservusersidget) | **GET** /ocserv/users/{id} | 
*OcservUsersApi* | [**ocservUsersIdLockPost**](docs/OcservUsersApi.md#ocservusersidlockpost) | **POST** /ocserv/users/{id}/lock | 
*OcservUsersApi* | [**ocservUsersIdPatch**](docs/OcservUsersApi.md#ocservusersidpatch) | **PATCH** /ocserv/users/{id} | 
*OcservUsersApi* | [**ocservUsersIdResetUsagePost**](docs/OcservUsersApi.md#ocservusersidresetusagepost) | **POST** /ocserv/users/{id}/reset-usage | Reset Ocserv user usage
*OcservUsersApi* | [**ocservUsersIdSessionLogsGet**](docs/OcservUsersApi.md#ocservusersidsessionlogsget) | **GET** /ocserv/users/{id}/session_logs | 
*OcservUsersApi* | [**ocservUsersIdStatisticsGet**](docs/OcservUsersApi.md#ocservusersidstatisticsget) | **GET** /ocserv/users/{id}/statistics | 
*OcservUsersApi* | [**ocservUsersIdTerminateByIdPost**](docs/OcservUsersApi.md#ocservusersidterminatebyidpost) | **POST** /ocserv/users/{id}/terminate_by_id | 
*OcservUsersApi* | [**ocservUsersIdUnlockPost**](docs/OcservUsersApi.md#ocservusersidunlockpost) | **POST** /ocserv/users/{id}/unlock | 
*OcservUsersApi* | [**ocservUsersPost**](docs/OcservUsersApi.md#ocservuserspost) | **POST** /ocserv/users | 
*OcservUsersApi* | [**ocservUsersUsernameDisconnectPost**](docs/OcservUsersApi.md#ocservusersusernamedisconnectpost) | **POST** /ocserv/users/{username}/disconnect | 
*OcservUsersApi* | [**ocservUsersUsernameTerminatePost**](docs/OcservUsersApi.md#ocservusersusernameterminatepost) | **POST** /ocserv/users/{username}/terminate | 
*ReportApi* | [**reportsSessionLogsGet**](docs/ReportApi.md#reportssessionlogsget) | **GET** /reports/session_logs | Ocserv session logs
*ReportApi* | [**reportsStatisticsGet**](docs/ReportApi.md#reportsstatisticsget) | **GET** /reports/statistics | Ocserv Users Statistics
*ReportApi* | [**reportsTotalBandwidthGet**](docs/ReportApi.md#reportstotalbandwidthget) | **GET** /reports/total-bandwidth | Ocserv Users TotalBandwidth calculating
*ReportApi* | [**reportsUsersGet**](docs/ReportApi.md#reportsusersget) | **GET** /reports/users | Result of all user reports
*SystemApi* | [**systemGet**](docs/SystemApi.md#systemget) | **GET** /system | Get panel System Config
*SystemApi* | [**systemInitGet**](docs/SystemApi.md#systeminitget) | **GET** /system/init | Get panel System init Config
*SystemApi* | [**systemOcservConfigGet**](docs/SystemApi.md#systemocservconfigget) | **GET** /system/ocserv-config | Get structured Ocserv configuration
*SystemApi* | [**systemOcservConfigPatch**](docs/SystemApi.md#systemocservconfigpatch) | **PATCH** /system/ocserv-config | Update structured Ocserv configuration
*SystemApi* | [**systemPatch**](docs/SystemApi.md#systempatch) | **PATCH** /system | Update panel System Config
*SystemApi* | [**systemReleaseGet**](docs/SystemApi.md#systemreleaseget) | **GET** /system/release | Get Dashboard the current and latest release
*SystemApi* | [**systemdDisablePost**](docs/SystemApi.md#systemddisablepost) | **POST** /systemd/disable | Disable Ocserv runtime
*SystemApi* | [**systemdEnablePost**](docs/SystemApi.md#systemdenablepost) | **POST** /systemd/enable | Enable Ocserv runtime
*SystemApi* | [**systemdRestartPost**](docs/SystemApi.md#systemdrestartpost) | **POST** /systemd/restart | Restart Ocserv runtime
*SystemApi* | [**systemdStatusGet**](docs/SystemApi.md#systemdstatusget) | **GET** /systemd/status | Ocserv runtime status
*SystemBackupApi* | [**backupOcservGroupsGet**](docs/SystemBackupApi.md#backupocservgroupsget) | **GET** /backup/ocserv_groups | Backup ocserv groups
*SystemBackupApi* | [**backupOcservUsersGet**](docs/SystemBackupApi.md#backupocservusersget) | **GET** /backup/ocserv_users | Backup ocserv users
*SystemRestoreApi* | [**backupOcservGroupsPost**](docs/SystemRestoreApi.md#backupocservgroupspost) | **POST** /backup/ocserv_groups | Restore ocserv groups
*SystemRestoreApi* | [**backupOcservUsersPost**](docs/SystemRestoreApi.md#backupocservuserspost) | **POST** /backup/ocserv_users | Restore ocserv users
*SystemUserApi* | [**systemUserResetPasswordPost**](docs/SystemUserApi.md#systemuserresetpasswordpost) | **POST** /system/user/reset-password | Reset admin password by secret key
*SystemUsersApi* | [**systemUsersGet**](docs/SystemUsersApi.md#systemusersget) | **GET** /system/users | List of Admin or simple users
*SystemUsersApi* | [**systemUsersIdDelete**](docs/SystemUsersApi.md#systemusersiddelete) | **DELETE** /system/users/{id} | Delete simple user
*SystemUsersApi* | [**systemUsersIdPasswordPost**](docs/SystemUsersApi.md#systemusersidpasswordpost) | **POST** /system/users/{id}/password | Change user password by admin
*SystemUsersApi* | [**systemUsersLoginPost**](docs/SystemUsersApi.md#systemusersloginpost) | **POST** /system/users/login | Admin users login
*SystemUsersApi* | [**systemUsersLookupGet**](docs/SystemUsersApi.md#systemuserslookupget) | **GET** /system/users/lookup | List of Users Lookup
*SystemUsersApi* | [**systemUsersPasswordPost**](docs/SystemUsersApi.md#systemuserspasswordpost) | **POST** /system/users/password | Change user password by self
*SystemUsersApi* | [**systemUsersPost**](docs/SystemUsersApi.md#systemuserspost) | **POST** /system/users | Create user
*SystemUsersApi* | [**systemUsersProfileGet**](docs/SystemUsersApi.md#systemusersprofileget) | **GET** /system/users/profile | Get User Profile


### Documentation For Models

 - [AuthSession](docs/AuthSession.md)
 - [BackupRestoreResponse](docs/BackupRestoreResponse.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiAuthSessionsResponse](docs/GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiAuthSessionsResponse.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiDashboardDockerService](docs/GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiDashboardDockerService.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiDashboardGetHomeResponse](docs/GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiDashboardGetHomeResponse.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiDashboardOcservStatusResponse](docs/GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiDashboardOcservStatusResponse.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiDashboardServerStatusResponse](docs/GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiDashboardServerStatusResponse.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemChangeUserPassword](docs/GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemChangeUserPassword.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemChangeUserPasswordBySelf](docs/GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemChangeUserPasswordBySelf.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemCreateUserData](docs/GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemCreateUserData.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemDashboardRelease](docs/GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemDashboardRelease.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemGetSystemInitResponse](docs/GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemGetSystemInitResponse.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemGetSystemResponse](docs/GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemGetSystemResponse.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemLoginData](docs/GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemLoginData.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemPatchSystemUpdateData](docs/GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemPatchSystemUpdateData.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemResetAdminPassword](docs/GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemResetAdminPassword.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemResetPasswordResponse](docs/GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemResetPasswordResponse.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemUserLoginResponse](docs/GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemUserLoginResponse.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemUsersResponse](docs/GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemUsersResponse.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalUsecaseAdminApiDashboardCPU](docs/GithubComMmtaeeOcservDashboardBackendInternalUsecaseAdminApiDashboardCPU.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalUsecaseAdminApiDashboardCurrentStats](docs/GithubComMmtaeeOcservDashboardBackendInternalUsecaseAdminApiDashboardCurrentStats.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalUsecaseAdminApiDashboardDisk](docs/GithubComMmtaeeOcservDashboardBackendInternalUsecaseAdminApiDashboardDisk.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalUsecaseAdminApiDashboardDockerStats](docs/GithubComMmtaeeOcservDashboardBackendInternalUsecaseAdminApiDashboardDockerStats.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalUsecaseAdminApiDashboardGeneralInfo](docs/GithubComMmtaeeOcservDashboardBackendInternalUsecaseAdminApiDashboardGeneralInfo.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalUsecaseAdminApiDashboardGetHomeUser](docs/GithubComMmtaeeOcservDashboardBackendInternalUsecaseAdminApiDashboardGetHomeUser.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalUsecaseAdminApiDashboardRAM](docs/GithubComMmtaeeOcservDashboardBackendInternalUsecaseAdminApiDashboardRAM.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalUsecaseAdminApiDashboardSwap](docs/GithubComMmtaeeOcservDashboardBackendInternalUsecaseAdminApiDashboardSwap.md)
 - [GithubComMmtaeeOcservDashboardBackendInternalUsecaseAdminApiDashboardTelegramServiceStatus](docs/GithubComMmtaeeOcservDashboardBackendInternalUsecaseAdminApiDashboardTelegramServiceStatus.md)
 - [GroupUnsyncedGroup](docs/GroupUnsyncedGroup.md)
 - [MiddlewaresPermissionDenied](docs/MiddlewaresPermissionDenied.md)
 - [MiddlewaresUnauthorized](docs/MiddlewaresUnauthorized.md)
 - [ModelsAgentAddressType](docs/ModelsAgentAddressType.md)
 - [ModelsDailyTraffic](docs/ModelsDailyTraffic.md)
 - [ModelsExpiryMode](docs/ModelsExpiryMode.md)
 - [ModelsIPBanPoints](docs/ModelsIPBanPoints.md)
 - [ModelsOcservAgent](docs/ModelsOcservAgent.md)
 - [ModelsOcservGroup](docs/ModelsOcservGroup.md)
 - [ModelsOcservGroupConfig](docs/ModelsOcservGroupConfig.md)
 - [ModelsOcservInfo](docs/ModelsOcservInfo.md)
 - [ModelsOcservUser](docs/ModelsOcservUser.md)
 - [ModelsOcservUserCertificateBackup](docs/ModelsOcservUserCertificateBackup.md)
 - [ModelsOcservUserCertificateStatus](docs/ModelsOcservUserCertificateStatus.md)
 - [ModelsOcservUserConfig](docs/ModelsOcservUserConfig.md)
 - [ModelsOcservUserSessionEvent](docs/ModelsOcservUserSessionEvent.md)
 - [ModelsOcservUserSessionLog](docs/ModelsOcservUserSessionLog.md)
 - [ModelsOnlineUserSession](docs/ModelsOnlineUserSession.md)
 - [ModelsServerVersion](docs/ModelsServerVersion.md)
 - [ModelsTrafficType](docs/ModelsTrafficType.md)
 - [ModelsUser](docs/ModelsUser.md)
 - [ModelsUsersLookup](docs/ModelsUsersLookup.md)
 - [OcservAgentCreateInput](docs/OcservAgentCreateInput.md)
 - [OcservAgentUpdateInput](docs/OcservAgentUpdateInput.md)
 - [OcservGroupCreateOcservGroupData](docs/OcservGroupCreateOcservGroupData.md)
 - [OcservGroupOcservGroupsResponse](docs/OcservGroupOcservGroupsResponse.md)
 - [OcservGroupSyncGroupRequest](docs/OcservGroupSyncGroupRequest.md)
 - [OcservGroupUpdateOcservGroupData](docs/OcservGroupUpdateOcservGroupData.md)
 - [OcservUserActivateUserData](docs/OcservUserActivateUserData.md)
 - [OcservUserBulkDeleteResponse](docs/OcservUserBulkDeleteResponse.md)
 - [OcservUserBulkGroupRequest](docs/OcservUserBulkGroupRequest.md)
 - [OcservUserBulkIDsRequest](docs/OcservUserBulkIDsRequest.md)
 - [OcservUserBulkStatusRequest](docs/OcservUserBulkStatusRequest.md)
 - [OcservUserBulkUpdateRequest](docs/OcservUserBulkUpdateRequest.md)
 - [OcservUserBulkUsersResponse](docs/OcservUserBulkUsersResponse.md)
 - [OcservUserCreateOcservUserData](docs/OcservUserCreateOcservUserData.md)
 - [OcservUserOcservUsersResponse](docs/OcservUserOcservUsersResponse.md)
 - [OcservUserUpdateOcservUserData](docs/OcservUserUpdateOcservUserData.md)
 - [OcservuserBulkUpdateItem](docs/OcservuserBulkUpdateItem.md)
 - [OcservuserUpdateOcservUserData](docs/OcservuserUpdateOcservUserData.md)
 - [ReportsOcservUserReportResponse](docs/ReportsOcservUserReportResponse.md)
 - [ReportsSessionLogsResponse](docs/ReportsSessionLogsResponse.md)
 - [RepositoryTopBandwidthUsers](docs/RepositoryTopBandwidthUsers.md)
 - [RepositoryTotalBandwidths](docs/RepositoryTotalBandwidths.md)
 - [RequestErrorResponse](docs/RequestErrorResponse.md)
 - [RequestMeta](docs/RequestMeta.md)
 - [RuntimeActionResponse](docs/RuntimeActionResponse.md)
 - [RuntimeOcservConfig](docs/RuntimeOcservConfig.md)
 - [RuntimeStatusResponse](docs/RuntimeStatusResponse.md)
 - [SystemRekeyMethod](docs/SystemRekeyMethod.md)


<a id="documentation-for-authorization"></a>
## Documentation For Authorization

Endpoints do not require authorization.

