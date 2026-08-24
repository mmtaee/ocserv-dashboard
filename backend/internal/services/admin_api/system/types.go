package system

import systemusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/system"

import runtimesystemusecase "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/system"

type GetSystemInitResponse = systemusecase.GetSystemInitResponse
type GetSystemResponse = systemusecase.GetSystemResponse
type PatchSystemUpdateData = systemusecase.PatchSystemUpdateData
type LoginData = systemusecase.LoginData
type UserLoginResponse = systemusecase.UserLoginResponse
type CreateUserData = systemusecase.CreateUserData
type UsersResponse = systemusecase.UsersResponse
type ChangeUserPassword = systemusecase.ChangeUserPassword
type ChangeUserPasswordBySelf = systemusecase.ChangeUserPasswordBySelf
type ResetPasswordResponse = systemusecase.ResetPasswordResponse
type DashboardRelease = systemusecase.DashboardRelease
type ResetAdminPassword = systemusecase.ResetAdminPassword
type StatusResponse = runtimesystemusecase.Status
type ActionResponse = runtimesystemusecase.ActionResult
type OcservConfig = runtimesystemusecase.OcservConfig
