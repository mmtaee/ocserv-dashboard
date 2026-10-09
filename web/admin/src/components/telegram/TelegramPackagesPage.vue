<script setup lang="ts">
import { Pencil, Plus, RefreshCw, Trash2 } from "@lucide/vue";
import { onMounted, reactive, shallowRef } from "vue";
import { useI18n } from "vue-i18n";
import { normalizeApiError } from "@/api/http";
import {
  createTelegramPackage,
  deleteTelegramPackage,
  getTelegramPackages,
  telegramTrafficTypes,
  updateTelegramPackage,
  type TelegramPackage,
  type TelegramPackageCreate,
} from "@/api/services/telegram";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from "@/components/ui/empty";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Skeleton } from "@/components/ui/skeleton";
import { Spinner } from "@/components/ui/spinner";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

const { t } = useI18n({ useScope: "global" });
const packages = shallowRef<TelegramPackage[]>([]);
const loading = shallowRef(true);
const mutating = shallowRef(false);
const error = shallowRef("");
const success = shallowRef("");
const editorOpen = shallowRef(false);
const deleteOpen = shallowRef(false);
const selected = shallowRef<TelegramPackage | null>(null);
const validation = shallowRef("");
const includeInactive = shallowRef(true);
const form = reactive({
  title: "",
  days: 30,
  traffic_size_gb: 0,
  traffic_type: "Free",
  price_text: "",
  is_active: true,
});
async function load(): Promise<void> {
  loading.value = true;
  error.value = "";
  try {
    packages.value = await getTelegramPackages(includeInactive.value);
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    loading.value = false;
  }
}
function openEditor(item?: TelegramPackage): void {
  selected.value = item ?? null;
  Object.assign(form, {
    title: item?.title ?? "",
    days: item?.days ?? 30,
    traffic_size_gb: item?.traffic_size_gb ?? 0,
    traffic_type: item?.traffic_type ?? "Free",
    price_text: item?.price_text ?? "",
    is_active: item?.is_active ?? true,
  });
  validation.value = "";
  editorOpen.value = true;
}
async function save(): Promise<void> {
  if (
    form.title.trim().length < 2 ||
    form.title.length > 128 ||
    form.days < 1 ||
    form.days > 3650 ||
    form.traffic_size_gb < 0 ||
    form.traffic_size_gb > 100000 ||
    form.price_text.length > 64
  ) {
    validation.value = t("telegram.validation.package");
    return;
  }
  const request = {
    ...form,
    title: form.title.trim(),
    traffic_type: form.traffic_type as TelegramPackageCreate["traffic_type"],
  };
  mutating.value = true;
  error.value = "";
  try {
    selected.value?.id == null
      ? await createTelegramPackage(request)
      : await updateTelegramPackage(selected.value.id, request);
    editorOpen.value = false;
    success.value = t("telegram.packageSaved");
    await load();
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    mutating.value = false;
  }
}
async function remove(): Promise<void> {
  if (selected.value?.id == null) return;
  mutating.value = true;
  error.value = "";
  try {
    await deleteTelegramPackage(selected.value.id);
    deleteOpen.value = false;
    success.value = t("telegram.packageDeleted");
    await load();
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    mutating.value = false;
  }
}
onMounted(load);
</script>

<template>
  <div class="flex flex-col gap-6">
    <Alert v-if="error" variant="error">
      <AlertTitle>{{ t("telegram.requestFailed") }}</AlertTitle>
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>
    <Alert v-if="success" variant="success">
      <AlertTitle>{{ t("telegram.success") }}</AlertTitle>
      <AlertDescription>{{ success }}</AlertDescription>
    </Alert>
    <Card>
      <CardHeader>
        <CardTitle>{{ t("navigation.telegramPackages") }}</CardTitle>
        <CardDescription>
          {{ t("telegram.packagesDescription") }}
        </CardDescription>
        <CardAction class="flex gap-2">
          <Field orientation="horizontal">
            <Checkbox
              id="include-inactive"
              v-model="includeInactive"
              @update:model-value="load"
            />
            <FieldLabel for="include-inactive">
              {{ t("telegram.includeInactive") }}
            </FieldLabel>
          </Field>
          <Button variant="outline" :disabled="loading" @click="load">
            <RefreshCw data-icon="inline-start" />
            {{ t("telegram.refresh") }}
          </Button>
          <Button @click="openEditor()">
            <Plus data-icon="inline-start" />
            {{ t("telegram.createPackage") }}
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent>
        <div v-if="loading" class="flex flex-col gap-3">
          <Skeleton v-for="n in 4" :key="n" class="h-12 w-full" />
        </div>
        <Empty v-else-if="packages.length === 0">
          <EmptyHeader>
            <EmptyTitle>{{ t("telegram.noPackages") }}</EmptyTitle>
            <EmptyDescription>
              {{ t("telegram.noPackagesDescription") }}
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
        <Table v-else>
          <TableHeader>
            <TableRow>
              <TableHead>{{ t("telegram.title") }}</TableHead>
              <TableHead>{{ t("telegram.duration") }}</TableHead>
              <TableHead>{{ t("telegram.traffic") }}</TableHead>
              <TableHead>{{ t("telegram.price") }}</TableHead>
              <TableHead>{{ t("telegram.status") }}</TableHead>
              <TableHead class="text-end">
                {{ t("telegram.actions") }}
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="item in packages" :key="item.id">
              <TableCell class="font-medium">{{ item.title }}</TableCell>
              <TableCell>
                {{ t("telegram.days", { count: item.days ?? 0 }) }}
              </TableCell>
              <TableCell>
                {{ t(`telegram.trafficTypes.${item.traffic_type}`) }} ·
                {{ item.traffic_size_gb ?? 0 }} GB
              </TableCell>
              <TableCell>{{ item.price_text || "—" }}</TableCell>
              <TableCell>
                <Badge :variant="item.is_active ? 'default' : 'secondary'">
                  {{
                    t(item.is_active ? "telegram.active" : "telegram.inactive")
                  }}
                </Badge>
              </TableCell>
              <TableCell>
                <div class="flex justify-end gap-2">
                  <Button
                    size="icon-sm"
                    variant="ghost"
                    :aria-label="t('telegram.edit')"
                    @click="openEditor(item)"
                  >
                    <Pencil />
                  </Button>
                  <Button
                    size="icon-sm"
                    variant="ghost"
                    :aria-label="t('telegram.delete')"
                    @click="
                      selected = item;
                      deleteOpen = true;
                    "
                  >
                    <Trash2 />
                  </Button>
                </div>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
    <Sheet v-model:open="editorOpen">
      <SheetContent class="overflow-y-auto">
        <SheetHeader>
          <SheetTitle>
            {{
              t(selected ? "telegram.editPackage" : "telegram.createPackage")
            }}
          </SheetTitle>
          <SheetDescription>
            {{ t("telegram.packageFormDescription") }}
          </SheetDescription>
        </SheetHeader>
        <form class="px-4" @submit.prevent="save">
          <FieldGroup>
            <Field>
              <FieldLabel for="package-title">
                {{ t("telegram.title") }}
              </FieldLabel>
              <Input id="package-title" v-model="form.title" maxlength="128" />
            </Field>
            <div class="grid grid-cols-2 gap-4">
              <Field>
                <FieldLabel for="package-days">
                  {{ t("telegram.durationDays") }}
                </FieldLabel>
                <Input
                  id="package-days"
                  v-model="form.days"
                  type="number"
                  min="1"
                  max="3650"
                />
              </Field>
              <Field>
                <FieldLabel for="package-traffic">
                  {{ t("telegram.trafficSize") }}
                </FieldLabel>
                <Input
                  id="package-traffic"
                  v-model="form.traffic_size_gb"
                  type="number"
                  min="0"
                  max="100000"
                />
              </Field>
            </div>
            <Field>
              <FieldLabel>{{ t("telegram.trafficType") }}</FieldLabel>
              <Select v-model="form.traffic_type">
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem
                      v-for="type in telegramTrafficTypes"
                      :key="type"
                      :value="type"
                    >
                      {{ t(`telegram.trafficTypes.${type}`) }}
                    </SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </Field>
            <Field>
              <FieldLabel for="package-price">
                {{ t("telegram.price") }}
              </FieldLabel>
              <Input
                id="package-price"
                v-model="form.price_text"
                maxlength="64"
              />
            </Field>
            <Field orientation="horizontal">
              <Checkbox id="package-active" v-model="form.is_active" />
              <FieldLabel for="package-active">
                {{ t("telegram.active") }}
              </FieldLabel>
            </Field>
            <FieldError v-if="validation">{{ validation }}</FieldError>
          </FieldGroup>
        </form>
        <SheetFooter>
          <Button :disabled="mutating" @click="save">
            <Spinner v-if="mutating" data-icon="inline-start" />
            {{ t("telegram.save") }}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
    <AlertDialog v-model:open="deleteOpen">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{{ t("telegram.deletePackage") }}</AlertDialogTitle>
          <AlertDialogDescription>
            {{ t("telegram.deletePackageConfirm", { title: selected?.title }) }}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>{{ t("telegram.cancel") }}</AlertDialogCancel>
          <AlertDialogAction :disabled="mutating" @click="remove">
            {{ t("telegram.delete") }}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
</template>
