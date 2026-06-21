
import Cookies from "js-cookie"
import { AUTH_COOKIE_NAME } from "../constant/global"

export const setToken = async (token: string) => {


  Cookies.set(AUTH_COOKIE_NAME, token, {
    expires: 120, // 120 days
    // path: "/",
    // sameSite: "lax",
    // secure,
  })
}

export const getToken = async (): Promise<string | undefined> => {
  return Cookies.get(AUTH_COOKIE_NAME)
}

export const removeToken = async () => {
  Cookies.remove(AUTH_COOKIE_NAME)
}
