import { Suspense } from "react"

import Loader from "@/share/components/module/loader/Loader"
import { createBrowserRouter } from "react-router"

import { ChangeLangToggle } from "../components/module/languge-selector/ChangeLangToggle"

// const DashboardPage = lazy(() => import("@/pages/dashboard/DashboardPage"))

const router = createBrowserRouter([
  {
    path: "/login",
    element: (
      <Suspense
        fallback={
          <div className="flex h-screen items-center justify-center">
            <Loader />
          </div>
        }
      >
        <div>
          <p>Login Form</p>
          <ChangeLangToggle />
        </div>
      </Suspense>
    ),
  },
])

export default router
