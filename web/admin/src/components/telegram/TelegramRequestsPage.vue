<script setup lang="ts">
import { Download, Eye, RefreshCw, Trash2 } from "@lucide/vue";
import { computed, onMounted, reactive, shallowRef } from "vue";
import { useI18n } from "vue-i18n";
import { normalizeApiError } from "@/api/http";
import {
  approveTelegramRequest,
  confirmTelegramPayment,
  deleteTelegramAccount,
  deleteTelegramRequest,
  downloadTelegramReceipt,
  getTelegramAccounts,
  getTelegramRequest,
  getTelegramRequests,
  rejectTelegramRequest,
  telegramRequestStatuses,
  telegramRequestTypes,
  type TelegramAccount,
  type TelegramRequest,
  type TelegramRequestStatus,
  type TelegramRequestType,
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
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from "@/components/ui/empty";
import { Field, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination";
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
import { Textarea } from "@/components/ui/textarea";

const { locale, t } = useI18n({ useScope: "global" });
const requests = shallowRef<TelegramRequest[]>([]);
const selected = shallowRef<TelegramRequest | null>(null);
const accounts = shallowRef<TelegramAccount[]>([]);
const loading = shallowRef(true);
const detailLoading = shallowRef(false);
const mutating = shallowRef(false);
const error = shallowRef("");
const success = shallowRef("");
const page = shallowRef(1);
const size = shallowRef(25);
const total = shallowRef(0);
const status = shallowRef("all");
const type = shallowRef("all");
const order = shallowRef("created_at");
const sort = shallowRef<"ASC" | "DESC">("DESC");
const sheetOpen = shallowRef(false);
const action = shallowRef<"approve" | "reject" | "confirm" | null>(null);
const confirmOpen = shallowRef(false);
const deleteTarget = shallowRef<{
  kind: "request" | "account";
  id: number;
} | null>(null);
const actionForm = reactive({
  admin_note: "",
  card_number: "",
  card_holder: "",
  reply_to_user: "",
  override_username: "",
  override_password: "",
  group: "",
});
const totalPages = computed(() =>
  Math.max(1, Math.ceil(total.value / size.value)),
);
const canApprove = computed(() => selected.value?.status === "pending");
const canConfirm = computed(
  () => selected.value?.status === "payment_uploaded",
);
const canReject = computed(
  () =>
    selected.value?.status !== "delivered" &&
    selected.value?.status !== "rejected",
);
const canDelete = computed(
  () =>
    selected.value?.status === "delivered" ||
    selected.value?.status === "rejected",
);
async function load(nextPage = page.value): Promise<void> {
  loading.value = true;
  error.value = "";
  try {
    const response = await getTelegramRequests({
      page: nextPage,
      size: size.value,
      order: order.value,
      sort: sort.value,
      status:
        status.value === "all"
          ? undefined
          : (status.value as TelegramRequestStatus),
      type:
        type.value === "all" ? undefined : (type.value as TelegramRequestType),
    });
    requests.value = response.result ?? [];
    page.value = response.meta?.page ?? nextPage;
    size.value = response.meta?.size ?? size.value;
    total.value = response.meta?.total_records ?? requests.value.length;
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    loading.value = false;
  }
}
async function openDetails(item: TelegramRequest): Promise<void> {
  if (item.id == null) return;
  selected.value = item;
  accounts.value = [];
  action.value = null;
  sheetOpen.value = true;
  detailLoading.value = true;
  error.value = "";
  try {
    selected.value = await getTelegramRequest(item.id);
    if (selected.value.target_ocserv_user_id)
      accounts.value = await getTelegramAccounts(
        selected.value.target_ocserv_user_id,
      );
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
    sheetOpen.value = false;
  } finally {
    detailLoading.value = false;
  }
}
function beginAction(next: "approve" | "reject" | "confirm"): void {
  Object.assign(actionForm, {
    admin_note: "",
    card_number: "",
    card_holder: "",
    reply_to_user: "",
    override_username: "",
    override_password: "",
    group: "",
  });
  action.value = next;
}
async function submitAction(): Promise<void> {
  if (!selected.value?.id || !action.value) return;
  mutating.value = true;
  error.value = "";
  try {
    if (action.value === "approve")
      await approveTelegramRequest(selected.value.id, {
        admin_note: actionForm.admin_note,
        card_number: actionForm.card_number,
        card_holder: actionForm.card_holder,
        reply_to_user: actionForm.reply_to_user,
      });
    else if (action.value === "reject")
      await rejectTelegramRequest(selected.value.id, {
        admin_note: actionForm.admin_note,
      });
    else
      await confirmTelegramPayment(selected.value.id, {
        admin_note: actionForm.admin_note,
        group: actionForm.group,
        override_username: actionForm.override_username,
        override_password: actionForm.override_password,
      });
    success.value = t(`telegram.${action.value}Success`);
    action.value = null;
    selected.value = await getTelegramRequest(selected.value.id);
    await load();
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    mutating.value = false;
  }
}
function askDelete(kind: "request" | "account", id?: number): void {
  if (id == null) return;
  deleteTarget.value = { kind, id };
  confirmOpen.value = true;
}
async function remove(): Promise<void> {
  if (!deleteTarget.value) return;
  mutating.value = true;
  error.value = "";
  try {
    if (deleteTarget.value.kind === "request") {
      await deleteTelegramRequest(deleteTarget.value.id);
      sheetOpen.value = false;
      success.value = t("telegram.requestDeleted");
      await load();
      if (requests.value.length === 0 && page.value > 1)
        await load(page.value - 1);
    } else {
      await deleteTelegramAccount(deleteTarget.value.id);
      accounts.value = accounts.value.filter(
        (item) => item.id !== deleteTarget.value?.id,
      );
      success.value = t("telegram.accountDeleted");
    }
    confirmOpen.value = false;
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    mutating.value = false;
  }
}
async function downloadReceipt(): Promise<void> {
  if (!selected.value?.id) return;
  mutating.value = true;
  error.value = "";
  try {
    const blob = await downloadTelegramReceipt(selected.value.id);
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `telegram-receipt-${selected.value.id}`;
    link.click();
    URL.revokeObjectURL(url);
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    mutating.value = false;
  }
}
function formatDate(value?: string): string {
  if (!value) return "—";
  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? value
    : new Intl.DateTimeFormat(locale.value, {
        dateStyle: "medium",
        timeStyle: "short",
      }).format(date);
}
onMounted(() => load());
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
        <CardTitle>{{ t("navigation.telegramRequests") }}</CardTitle>
        <CardDescription>
          {{ t("telegram.requestsDescription") }}
        </CardDescription>
        <CardAction>
          <Button variant="outline" :disabled="loading" @click="load()">
            <RefreshCw data-icon="inline-start" />
            {{ t("telegram.refresh") }}
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent class="flex flex-col gap-4">
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <Field>
            <FieldLabel>{{ t("telegram.status") }}</FieldLabel>
            <Select v-model="status" @update:model-value="load(1)">
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="all">{{ t("telegram.all") }}</SelectItem>
                  <SelectItem
                    v-for="value in telegramRequestStatuses"
                    :key="value"
                    :value="value"
                  >
                    {{ t(`telegram.statuses.${value}`) }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </Field>
          <Field>
            <FieldLabel>{{ t("telegram.type") }}</FieldLabel>
            <Select v-model="type" @update:model-value="load(1)">
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="all">{{ t("telegram.all") }}</SelectItem>
                  <SelectItem
                    v-for="value in telegramRequestTypes"
                    :key="value"
                    :value="value"
                  >
                    {{ t(`telegram.types.${value}`) }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </Field>
          <Field>
            <FieldLabel>{{ t("telegram.order") }}</FieldLabel>
            <Select v-model="order" @update:model-value="load(1)">
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="created_at">
                    {{ t("telegram.createdAt") }}
                  </SelectItem>
                  <SelectItem value="updated_at">
                    {{ t("telegram.updatedAt") }}
                  </SelectItem>
                  <SelectItem value="status">
                    {{ t("telegram.status") }}
                  </SelectItem>
                  <SelectItem value="type">{{ t("telegram.type") }}</SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </Field>
          <Field>
            <FieldLabel>{{ t("telegram.sort") }}</FieldLabel>
            <Select v-model="sort" @update:model-value="load(1)">
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="DESC">
                    {{ t("telegram.descending") }}
                  </SelectItem>
                  <SelectItem value="ASC">
                    {{ t("telegram.ascending") }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </Field>
        </div>
        <div
          v-if="loading && requests.length === 0"
          class="flex flex-col gap-3"
        >
          <Skeleton v-for="n in 6" :key="n" class="h-12 w-full" />
        </div>
        <Empty v-else-if="requests.length === 0">
          <EmptyHeader>
            <EmptyTitle>{{ t("telegram.noRequests") }}</EmptyTitle>
            <EmptyDescription>
              {{ t("telegram.noRequestsDescription") }}
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
        <Table v-else>
          <TableHeader>
            <TableRow>
              <TableHead>{{ t("telegram.id") }}</TableHead>
              <TableHead>{{ t("telegram.telegramUser") }}</TableHead>
              <TableHead>{{ t("telegram.type") }}</TableHead>
              <TableHead>{{ t("telegram.status") }}</TableHead>
              <TableHead>{{ t("telegram.createdAt") }}</TableHead>
              <TableHead class="text-end">
                {{ t("telegram.actions") }}
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="item in requests" :key="item.id">
              <TableCell>{{ item.id }}</TableCell>
              <TableCell>{{ item.telegram_username || "—" }}</TableCell>
              <TableCell>{{ t(`telegram.types.${item.type}`) }}</TableCell>
              <TableCell>
                <Badge variant="secondary">
                  {{ t(`telegram.statuses.${item.status}`) }}
                </Badge>
              </TableCell>
              <TableCell>{{ formatDate(item.created_at) }}</TableCell>
              <TableCell class="text-end">
                <Button
                  size="icon-sm"
                  variant="ghost"
                  :aria-label="t('telegram.view')"
                  @click="openDetails(item)"
                >
                  <Eye />
                </Button>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
      <CardFooter
        v-if="total > 0"
        class="grid items-center gap-3 sm:grid-cols-[1fr_auto_1fr]"
      >
        <span class="text-sm text-muted-foreground">
          {{ t("telegram.pageStatus", { page, pages: totalPages, total }) }}
        </span>
        <Pagination
          v-slot="{ page: currentPage }"
          class="mx-auto w-auto"
          :items-per-page="size"
          :page="page"
          :total="total"
          :disabled="loading"
          show-edges
          @update:page="load"
        >
          <PaginationContent v-slot="{ items }">
            <PaginationPrevious :label="t('telegram.previous')" />
            <template
              v-for="(item, index) in items"
              :key="`${item.type}-${index}`"
            >
              <PaginationItem
                v-if="item.type === 'page'"
                :value="item.value"
                :is-active="item.value === currentPage"
              >
                {{ item.value }}
              </PaginationItem>
              <PaginationEllipsis
                v-else
                :index="index"
                :label="t('telegram.morePages')"
              />
            </template>
            <PaginationNext :label="t('telegram.next')" />
          </PaginationContent>
        </Pagination>
      </CardFooter>
    </Card>
    <Sheet v-model:open="sheetOpen">
      <SheetContent class="overflow-y-auto sm:max-w-xl">
        <SheetHeader>
          <SheetTitle>
            {{ t("telegram.requestDetails", { id: selected?.id }) }}
          </SheetTitle>
          <SheetDescription>
            {{ t("telegram.requestDetailsDescription") }}
          </SheetDescription>
        </SheetHeader>
        <div v-if="detailLoading" class="px-4">
          <Skeleton class="h-80 w-full" />
        </div>
        <div v-else-if="selected" class="flex flex-col gap-5 px-4">
          <dl class="grid grid-cols-2 gap-3 text-sm">
            <dt class="text-muted-foreground">{{ t("telegram.status") }}</dt>
            <dd>{{ t(`telegram.statuses.${selected.status}`) }}</dd>
            <dt class="text-muted-foreground">{{ t("telegram.type") }}</dt>
            <dd>{{ t(`telegram.types.${selected.type}`) }}</dd>
            <dt class="text-muted-foreground">
              {{ t("telegram.telegramUser") }}
            </dt>
            <dd>{{ selected.telegram_username || "—" }}</dd>
            <dt class="text-muted-foreground">
              {{ t("telegram.desiredUsername") }}
            </dt>
            <dd>{{ selected.desired_username || "—" }}</dd>
            <dt class="text-muted-foreground">{{ t("telegram.packageId") }}</dt>
            <dd>{{ selected.package_id || "—" }}</dd>
            <dt class="text-muted-foreground">{{ t("telegram.createdAt") }}</dt>
            <dd>{{ formatDate(selected.created_at) }}</dd>
            <dt class="text-muted-foreground">
              {{ t("telegram.userMessage") }}
            </dt>
            <dd>{{ selected.user_message || "—" }}</dd>
            <dt class="text-muted-foreground">{{ t("telegram.adminNote") }}</dt>
            <dd>{{ selected.admin_note || "—" }}</dd>
          </dl>
          <div v-if="accounts.length" class="flex flex-col gap-2">
            <h3 class="font-medium">{{ t("telegram.linkedAccounts") }}</h3>
            <div
              v-for="account in accounts"
              :key="account.id"
              class="flex items-center justify-between rounded-md border p-3"
            >
              <span>
                {{ account.telegram_username || account.chat_id }} ·
                {{ t(`telegram.languages.${account.language}`) }}
              </span>
              <Button
                size="icon-sm"
                variant="ghost"
                :aria-label="t('telegram.unlink')"
                @click="askDelete('account', account.id)"
              >
                <Trash2 />
              </Button>
            </div>
          </div>
          <form v-if="action" @submit.prevent="submitAction">
            <FieldGroup>
              <Field>
                <FieldLabel for="admin-note">
                  {{ t("telegram.adminNote") }}
                </FieldLabel>
                <Textarea
                  id="admin-note"
                  v-model="actionForm.admin_note"
                  maxlength="1024"
                />
              </Field>
              <template v-if="action === 'approve'">
                <Field>
                  <FieldLabel for="card-number-override">
                    {{ t("telegram.cardNumber") }}
                  </FieldLabel>
                  <Input
                    id="card-number-override"
                    v-model="actionForm.card_number"
                    maxlength="64"
                  />
                </Field>
                <Field>
                  <FieldLabel for="card-holder-override">
                    {{ t("telegram.cardHolder") }}
                  </FieldLabel>
                  <Input
                    id="card-holder-override"
                    v-model="actionForm.card_holder"
                    maxlength="128"
                  />
                </Field>
                <Field>
                  <FieldLabel for="reply">
                    {{ t("telegram.replyToUser") }}
                  </FieldLabel>
                  <Textarea
                    id="reply"
                    v-model="actionForm.reply_to_user"
                    maxlength="1024"
                  />
                </Field>
              </template>
              <template v-if="action === 'confirm'">
                <Field>
                  <FieldLabel for="override-username">
                    {{ t("telegram.overrideUsername") }}
                  </FieldLabel>
                  <Input
                    id="override-username"
                    v-model="actionForm.override_username"
                    minlength="3"
                    maxlength="64"
                  />
                </Field>
                <Field>
                  <FieldLabel for="override-password">
                    {{ t("telegram.overridePassword") }}
                  </FieldLabel>
                  <Input
                    id="override-password"
                    v-model="actionForm.override_password"
                    type="password"
                    minlength="4"
                    maxlength="64"
                  />
                </Field>
                <Field>
                  <FieldLabel for="group">{{ t("telegram.group") }}</FieldLabel>
                  <Input id="group" v-model="actionForm.group" maxlength="16" />
                </Field>
              </template>
              <Button type="submit" :disabled="mutating">
                <Spinner v-if="mutating" data-icon="inline-start" />
                {{ t(`telegram.${action}`) }}
              </Button>
            </FieldGroup>
          </form>
        </div>
        <SheetFooter v-if="selected && !action" class="flex-row flex-wrap">
          <Button
            v-if="selected.receipt_file_path"
            variant="outline"
            :disabled="mutating"
            @click="downloadReceipt"
          >
            <Download data-icon="inline-start" />
            {{ t("telegram.receipt") }}
          </Button>
          <Button v-if="canApprove" @click="beginAction('approve')">
            {{ t("telegram.approve") }}
          </Button>
          <Button v-if="canConfirm" @click="beginAction('confirm')">
            {{ t("telegram.confirm") }}
          </Button>
          <Button
            v-if="canReject"
            variant="outline"
            @click="beginAction('reject')"
          >
            {{ t("telegram.reject") }}
          </Button>
          <Button
            v-if="canDelete"
            variant="destructive"
            @click="askDelete('request', selected.id)"
          >
            {{ t("telegram.delete") }}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
    <AlertDialog v-model:open="confirmOpen">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>
            {{
              t(
                deleteTarget?.kind === "account"
                  ? "telegram.unlinkAccount"
                  : "telegram.deleteRequest",
              )
            }}
          </AlertDialogTitle>
          <AlertDialogDescription>
            {{
              t(
                deleteTarget?.kind === "account"
                  ? "telegram.unlinkAccountConfirm"
                  : "telegram.deleteRequestConfirm",
              )
            }}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>{{ t("telegram.cancel") }}</AlertDialogCancel>
          <AlertDialogAction :disabled="mutating" @click="remove">
            {{
              t(
                deleteTarget?.kind === "account"
                  ? "telegram.unlink"
                  : "telegram.delete",
              )
            }}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
</template>
