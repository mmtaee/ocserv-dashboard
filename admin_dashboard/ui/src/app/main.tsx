import { StrictMode } from "react"

import RootProvider from "@/share/config/RootProvider.tsx"
import { I18nProvider } from "@/share/config/i18n"
import { createRoot } from "react-dom/client"

import "./index.css"

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <I18nProvider>
      <RootProvider />
    </I18nProvider>
  </StrictMode>
)
