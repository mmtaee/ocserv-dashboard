import App from "@/app/App"

import { DirectionProvider } from "../components/ui/direction"
import { useI18n } from "./i18n"

const RootProvider = () => {
  const { localeConfig } = useI18n()
  return (
    <DirectionProvider dir={localeConfig.direction}>
      <App />
    </DirectionProvider>
  )
}

export default RootProvider
