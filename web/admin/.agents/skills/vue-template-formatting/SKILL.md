---
name: vue-template-formatting
description: Enforce this project's Vue template interpolation layout when creating, editing, formatting, or reviewing `.vue` files.
---

# Vue Template Formatting

Keep multiline element content vertically aligned. When an element's interpolation
is placed on multiple lines, put the opening tag, complete interpolation, and
closing tag on separate lines.

Use:

```vue
<AlertDescription>
  {{ t("dashboard.containerUsageError") }}
</AlertDescription>
```

Do not split the interpolation delimiters across the element tags:

```vue
<AlertDescription>{{
  t("dashboard.containerUsageError")
}}</AlertDescription>
```

An interpolation may remain inline when the entire element is intentionally kept
on one readable line:

```vue
<AlertTitle>{{ title }}</AlertTitle>
```

Apply this rule after running formatters, and include it in the final review of
every changed Vue template.
