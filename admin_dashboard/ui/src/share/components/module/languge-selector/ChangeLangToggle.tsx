import { Button } from "@/share/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from "@/share/components/ui/dropdown-menu"
import { useI18n } from "@/share/config/i18n"

export function ChangeLangToggle() {
  const {
    t,
    locale,
    setLocale,
    availableLocales,
    localeConfig,
    localeConfigs,
  } = useI18n()

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline">{localeConfig.name}</Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-32">
        <DropdownMenuGroup>
          <DropdownMenuLabel>{t.languageSelector.title}</DropdownMenuLabel>
          <DropdownMenuRadioGroup value={locale} onValueChange={setLocale}>
            {availableLocales.map((loc) => (
              <DropdownMenuRadioItem key={loc} value={loc}>
                {localeConfigs[loc].name}
              </DropdownMenuRadioItem>
            ))}
          </DropdownMenuRadioGroup>
        </DropdownMenuGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
