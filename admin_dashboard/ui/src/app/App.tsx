import QueryProvider from "@/share/config/QueryProvider"
import router from "@/share/routes/router"
import { RouterProvider } from "react-router"

function App() {
  return (
    <QueryProvider>
      <RouterProvider router={router} />
    </QueryProvider>
  )
}

export default App
