<script setup lang="ts">
import { Pencil, Plus, RefreshCw, Save, Trash2 } from "@lucide/vue";
import { computed, onMounted, reactive, shallowRef } from "vue";
import { storeToRefs } from "pinia";
import { useI18n } from "vue-i18n";
import { normalizeApiError } from "@/api/http";
import { MASTER_SERVER, useServerStore } from "@/stores/server";
import {
  ModelsAgentAddressType,
  type RuntimeOcservConfig,
} from "@/api/generated";
import {
  createOcservAgent,
  deleteOcservAgent,
  getOcservAgent,
  getOcservConfig,
  getSystemConfig,
  getSystemRelease,
  updateOcservAgent,
  updateOcservConfig,
  updateSystemConfig,
  type OcservAgent,
  type OcservAgentCreate,
  type SystemConfig,
  type SystemUpdateData,
} from "@/api/services/system";
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
import { Checkbox } from "@/components/ui/checkbox";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from "@/components/ui/empty";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";

const { t } = useI18n({ useScope: "global" });
const serverStore = useServerStore();
const { agents, selectedServer } = storeToRefs(serverStore);
const loading = shallowRef(true),
  saving = shallowRef<"system" | "ocserv" | "agent" | "delete" | null>(null),
  error = shallowRef(""),
  success = shallowRef("");
const release = shallowRef<{ current: string; latest: string } | null>(null),
  ocservOriginal = shallowRef<RuntimeOcservConfig>({}),
  allowOcservUpdate = shallowRef(false),
  ocservError = shallowRef(""),
  activeTab = shallowRef<"release" | "panel" | "ocserv" | "agents">("release"),
  agentDialogOpen = shallowRef(false),
  editingAgentId = shallowRef<number | null>(null),
  deleteAgent = shallowRef<OcservAgent | null>(null);
const panel = reactive<SystemUpdateData>({
  auto_delete_inactive_users: false,
  client_profile_connection_name: "",
  client_profile_server_address: "",
  client_profile_server_port: 0,
  google_captcha_secret_key: "",
  google_captcha_site_key: "",
  keep_inactive_user_days: 0,
});
const agent = reactive<OcservAgentCreate>({
  name: "",
  address: "",
  address_type: ModelsAgentAddressType.AgentAddressTypeDomain,
  port: 443,
  token: "",
});
const ocserv = reactive<RuntimeOcservConfig>({});
const numberFields: (keyof RuntimeOcservConfig)[] = [
  "auth_timeout",
  "ban_reset_time",
  "cookie_timeout",
  "dpd",
  "keepalive",
  "log_level",
  "max_ban_score",
  "max_clients",
  "max_same_clients",
  "min_reauth_time",
  "mobile_dpd",
  "mtu",
  "rate_limit_ms",
  "rekey_time",
  "switch_to_tcp_timeout",
  "tcp_port",
  "udp_port",
];
const booleanFields: (keyof RuntimeOcservConfig)[] = [
  "cisco_client_compat",
  "deny_roaming",
  "dtls_legacy",
  "ping_leases",
  "predictable_ips",
  "try_mtu_discovery",
  "tunnel_all_dns",
];
const textFields: (keyof RuntimeOcservConfig)[] = [
  "banner",
  "ipv4_network",
  "pre_login_banner",
];
const updateAvailable = computed(() =>
  Boolean(release.value && release.value.current !== release.value.latest),
);
const agentValid = computed(() =>
  Boolean(
    agent.name.trim() &&
      agent.address.trim() &&
      agent.token.trim() &&
      (agent.port == null ||
        (Number.isInteger(agent.port) && agent.port >= 1 && agent.port <= 65535)),
  ),
);
function setPanel(value: SystemConfig): boolean {
  const keys = Object.keys(panel) as (keyof SystemUpdateData)[];
  if (keys.some((key) => typeof value[key] !== typeof panel[key])) return false;
  Object.assign(panel, value);
  return true;
}
function resetAgent(): void {
  editingAgentId.value = null;
  Object.assign(agent, {
    name: "",
    address: "",
    address_type: ModelsAgentAddressType.AgentAddressTypeDomain,
    port: 443,
    token: "",
  });
}
function openCreateAgent(): void {
  resetAgent();
  agentDialogOpen.value = true;
}
function textValue(key: keyof RuntimeOcservConfig): string {
  return (ocserv[key] as string | undefined) ?? "";
}
function numberValue(key: keyof RuntimeOcservConfig): number | undefined {
  return ocserv[key] as number | undefined;
}
function booleanValue(key: keyof RuntimeOcservConfig): boolean | undefined {
  return ocserv[key] as boolean | undefined;
}
function updateText(key: keyof RuntimeOcservConfig, value: string): void {
  (ocserv as Record<string, unknown>)[key] = value || undefined;
}
function updateBoolean(key: keyof RuntimeOcservConfig, value: boolean): void {
  (ocserv as Record<string, unknown>)[key] = value;
}
function setOcserv(value: RuntimeOcservConfig): void {
  Object.keys(ocserv).forEach(
    (key) => delete ocserv[key as keyof RuntimeOcservConfig],
  );
  Object.assign(ocserv, value);
}
function updateNumber(key: keyof RuntimeOcservConfig, value: string): void {
  (ocserv as Record<string, unknown>)[key] =
    value === "" ? undefined : Number(value);
}
async function refresh(): Promise<void> {
  loading.value = true;
  error.value = "";
  try {
    const [system, config, nextRelease] = await Promise.all([
      getSystemConfig(),
      getOcservConfig(),
      getSystemRelease(),
      serverStore.refreshAgents(),
    ]);
    if (!setPanel(system))
      throw new Error(t("systemSettings.invalidSystemConfig"));
    allowOcservUpdate.value = config.allow_update;
    const { allow_update: _allowUpdate, ...ocservConfig } = config;
    ocservOriginal.value = ocservConfig;
    setOcserv(ocservConfig);
    release.value = nextRelease;
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    loading.value = false;
  }
}
async function savePanel(): Promise<void> {
  saving.value = "system";
  error.value = "";
  try {
    setPanel(await updateSystemConfig({ ...panel }));
    success.value = t("systemSettings.panelSaved");
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    saving.value = null;
  }
}
async function saveOcserv(): Promise<void> {
  if (!allowOcservUpdate.value) {
    error.value = t("systemSettings.updatesUnavailable");
    return;
  }
  const changed = Object.fromEntries(
    Object.entries(ocserv).filter(
      ([key, value]) =>
        JSON.stringify(value) !==
        JSON.stringify(ocservOriginal.value[key as keyof RuntimeOcservConfig]),
    ),
  ) as RuntimeOcservConfig;
  saving.value = "ocserv";
  error.value = "";
  ocservError.value = "";
  try {
    await updateOcservConfig(changed);
    await refresh();
    success.value = t("systemSettings.ocservSaved");
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    saving.value = null;
  }
}
async function editAgent(item: OcservAgent): Promise<void> {
  if (item.id == null) return;
  saving.value = "agent";
  try {
    Object.assign(agent, await getOcservAgent(item.id));
    editingAgentId.value = item.id;
    agentDialogOpen.value = true;
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    saving.value = null;
  }
}
async function saveAgent(): Promise<void> {
  if (!agentValid.value) {
    error.value = t("systemSettings.agentRequired");
    return;
  }
  saving.value = "agent";
  error.value = "";
  try {
    if (editingAgentId.value == null) await createOcservAgent({ ...agent });
    else await updateOcservAgent(editingAgentId.value, { ...agent });
    await serverStore.refreshAgents();
    resetAgent();
    agentDialogOpen.value = false;
    success.value = t("systemSettings.agentSaved");
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    saving.value = null;
  }
}
async function removeAgent(): Promise<void> {
  if (deleteAgent.value?.id == null) return;
  saving.value = "delete";
  try {
    await deleteOcservAgent(deleteAgent.value.id);
    deleteAgent.value = null;
    await serverStore.refreshAgents();
    success.value = t("systemSettings.agentDeleted");
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    saving.value = null;
  }
}
onMounted(refresh);
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <h1 class="text-2xl font-semibold tracking-tight">
          {{ t("navigation.settings") }}
        </h1>
        <p class="text-sm text-muted-foreground">
          {{ t("systemSettings.description") }}
        </p>
      </div>
      <Button
        type="button"
        variant="outline"
        :disabled="loading || saving !== null"
        @click="refresh"
        ><Spinner v-if="loading" data-icon="inline-start" /><RefreshCw
          v-else
          data-icon="inline-start"
        />{{ t("systemSettings.refresh") }}</Button
      >
    </div>
    <Alert v-if="error" variant="error"
      ><AlertTitle>{{ t("systemSettings.requestFailed") }}</AlertTitle
      ><AlertDescription>{{ error }}</AlertDescription></Alert
    ><Alert v-if="success" variant="success"
      ><AlertTitle>{{ t("systemSettings.success") }}</AlertTitle
      ><AlertDescription>{{ success }}</AlertDescription></Alert
    >
    <div class="flex flex-wrap gap-2" role="tablist">
      <Button
        v-for="tab in ['release', 'panel', 'ocserv', 'agents'] as const"
        v-show="tab !== 'agents' || selectedServer === MASTER_SERVER"
        :key="tab"
        type="button"
        :variant="activeTab === tab ? 'default' : 'outline'"
        role="tab"
        :aria-selected="activeTab === tab"
        @click="activeTab = tab"
        >{{ t(`systemSettings.tabs.${tab}`) }}</Button
      >
    </div>
    <Card v-if="activeTab === 'release'"
      ><CardHeader
        ><CardTitle>{{ t("systemSettings.release") }}</CardTitle
        ><CardDescription>{{
          t("systemSettings.releaseDescription")
        }}</CardDescription></CardHeader
      ><CardContent v-if="release" class="grid gap-3 sm:grid-cols-2"
        ><div>
          <p class="text-sm text-muted-foreground">
            {{ t("systemSettings.currentVersion") }}
          </p>
          <p>{{ release.current }}</p>
        </div>
        <div>
          <p class="text-sm text-muted-foreground">
            {{ t("systemSettings.latestVersion") }}
          </p>
          <p>{{ release.latest }}</p>
        </div>
        <Alert
          class="sm:col-span-2"
          :variant="updateAvailable ? 'warning' : 'success'"
        >
          <AlertDescription>{{
            t(
              updateAvailable
                ? "systemSettings.updateAvailable"
                : "systemSettings.upToDate",
            )
          }}</AlertDescription>
        </Alert></CardContent
      ></Card
    >
    <Card v-if="activeTab === 'panel'"
      ><CardHeader
        ><CardTitle>{{ t("systemSettings.panel") }}</CardTitle
        ><CardDescription>{{
          t("systemSettings.panelDescription")
        }}</CardDescription></CardHeader
      ><CardContent
        ><FieldGroup
          ><Field
            ><FieldLabel for="connection-name">{{
              t("systemSettings.connectionName")
            }}</FieldLabel
            ><Input
              id="connection-name"
              v-model="panel.client_profile_connection_name"
              :disabled="saving !== null" /></Field
          ><Field
            ><FieldLabel for="server-address">{{
              t("systemSettings.serverAddress")
            }}</FieldLabel
            ><Input
              id="server-address"
              v-model="panel.client_profile_server_address"
              :disabled="saving !== null" /></Field
          ><Field
            ><FieldLabel for="server-port">{{
              t("systemSettings.serverPort")
            }}</FieldLabel
            ><Input
              id="server-port"
              v-model.number="panel.client_profile_server_port"
              min="1"
              type="number"
              :disabled="saving !== null" /></Field
          ><Field
            ><FieldLabel for="inactive-days">{{
              t("systemSettings.inactiveDays")
            }}</FieldLabel
            ><Input
              id="inactive-days"
              v-model.number="panel.keep_inactive_user_days"
              min="0"
              type="number"
              :disabled="saving !== null" /></Field
          ><Field
            ><div class="flex items-center gap-2">
              <Checkbox
                id="auto-delete"
                v-model="panel.auto_delete_inactive_users"
                :disabled="saving !== null"
              /><FieldLabel for="auto-delete">{{
                t("systemSettings.autoDelete")
              }}</FieldLabel>
            </div></Field
          ><Field
            ><FieldLabel for="captcha-site">{{
              t("systemSettings.captchaSite")
            }}</FieldLabel
            ><Input
              id="captcha-site"
              v-model="panel.google_captcha_site_key"
              :disabled="saving !== null" /></Field
          ><Field
            ><FieldLabel for="captcha-secret">{{
              t("systemSettings.captchaSecret")
            }}</FieldLabel
            ><Input
              id="captcha-secret"
              v-model="panel.google_captcha_secret_key"
              type="password"
              :disabled="saving !== null" /></Field></FieldGroup></CardContent
      ><CardFooter
        ><Button type="button" :disabled="saving !== null" @click="savePanel"
          ><Spinner v-if="saving === 'system'" data-icon="inline-start" /><Save
            v-else
            data-icon="inline-start"
          />{{ t("systemSettings.save") }}</Button
        ></CardFooter
      ></Card
    >
    <Card v-if="activeTab === 'ocserv'"
      ><CardHeader
        ><CardTitle>{{ t("systemSettings.ocserv") }}</CardTitle
        ><CardDescription>{{
          t("systemSettings.ocservDescription")
        }}</CardDescription></CardHeader
      ><CardContent
        ><FieldGroup
          ><Field v-for="key in textFields" :key="key"
            ><FieldLabel :for="`ocserv-${key}`">{{
              t("systemSettings.configField", { field: key })
            }}</FieldLabel
            ><Input
              :id="`ocserv-${key}`"
              :model-value="textValue(key)"
              :disabled="saving !== null || !allowOcservUpdate"
              @update:model-value="updateText(key, String($event))" /></Field
          ><Field v-for="key in numberFields" :key="key"
            ><FieldLabel :for="`ocserv-${key}`">{{
              t("systemSettings.configField", { field: key })
            }}</FieldLabel
            ><Input
              :id="`ocserv-${key}`"
              :model-value="numberValue(key)"
              type="number"
              :disabled="saving !== null || !allowOcservUpdate"
              @update:model-value="updateNumber(key, String($event))" /></Field
          ><Field v-for="key in booleanFields" :key="key"
            ><div class="flex items-center gap-2">
              <Checkbox
                :id="`ocserv-${key}`"
                :model-value="booleanValue(key)"
                :disabled="saving !== null || !allowOcservUpdate"
                @update:model-value="updateBoolean(key, Boolean($event))"
              /><FieldLabel :for="`ocserv-${key}`">{{
                t("systemSettings.configField", { field: key })
              }}</FieldLabel>
            </div></Field
          ><Field
            ><FieldLabel for="ocserv-dns">{{
              t("systemSettings.configField", { field: "dns" })
            }}</FieldLabel
            ><Input
              id="ocserv-dns"
              :model-value="ocserv.dns?.join(', ') ?? ''"
              :disabled="saving !== null || !allowOcservUpdate"
              @update:model-value="
                ocserv.dns = String($event)
                  .split(',')
                  .map((item) => item.trim())
                  .filter(Boolean)
              " /></Field
          ><Field
            ><FieldLabel for="ocserv-rekey">{{
              t("systemSettings.configField", { field: "rekey_method" })
            }}</FieldLabel
            ><select
              id="ocserv-rekey"
              v-model="ocserv.rekey_method"
              class="h-9 rounded-md border border-input bg-transparent px-3 text-sm"
              :disabled="saving !== null || !allowOcservUpdate"
            >
              <option value="ssl">ssl</option>
              <option value="new-tunnel">new-tunnel</option>
            </select></Field
          ></FieldGroup
        ><FieldDescription v-if="allowOcservUpdate">{{
          t("systemSettings.ocservHelp")
        }}</FieldDescription
        ><FieldDescription v-else>{{
          t("systemSettings.readOnly")
        }}</FieldDescription
        ><FieldError v-if="ocservError">{{
          ocservError
        }}</FieldError> </CardContent
      ><CardFooter
        ><Button
          type="button"
          :disabled="saving !== null || !allowOcservUpdate"
          @click="saveOcserv"
          ><Spinner v-if="saving === 'ocserv'" data-icon="inline-start" /><Save
            v-else
            data-icon="inline-start"
          />{{ t("systemSettings.save") }}</Button
        ></CardFooter
      ></Card
    >
    <Card
      v-if="activeTab === 'agents' && selectedServer === MASTER_SERVER"
      ><CardHeader
        ><CardTitle>{{ t("systemSettings.agents") }}</CardTitle
        ><CardDescription>{{
          t("systemSettings.agentsDescription")
        }}</CardDescription
        ><CardAction
          ><Button
            type="button"
            variant="outline"
            :disabled="saving !== null"
            @click="openCreateAgent"
            ><Plus data-icon="inline-start" />{{
              t("systemSettings.newAgent")
            }}</Button
          ></CardAction
        ></CardHeader
      ><CardContent class="flex flex-col gap-6"
        ><Empty v-if="!loading && !agents.length"
          ><EmptyHeader
            ><EmptyTitle>{{ t("systemSettings.noAgents") }}</EmptyTitle
            ><EmptyDescription>{{
              t("systemSettings.noAgentsDescription")
            }}</EmptyDescription></EmptyHeader
          ></Empty
        >
        <div v-else class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead class="text-muted-foreground">
              <tr>
                <th class="p-2 text-start">
                  {{ t("systemSettings.agentName") }}
                </th>
                <th class="p-2 text-start">
                  {{ t("systemSettings.agentAddress") }}
                </th>
                <th class="p-2 text-start">
                  {{ t("systemSettings.addressType") }}
                </th>
                <th class="p-2 text-start">
                  {{ t("systemSettings.serverPort") }}
                </th>
                <th class="p-2 text-start">
                  {{ t("systemSettings.actions") }}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in agents" :key="item.id">
                <td class="p-2">{{ item.name }}</td>
                <td class="p-2">{{ item.address }}</td>
                <td class="p-2">
                  {{
                    t(
                      item.address_type === "ip"
                        ? "systemSettings.ip"
                        : "systemSettings.domain",
                    )
                  }}
                </td>
                <td class="p-2">{{ item.port ?? 443 }}</td>
                <td class="flex gap-2 p-2">
                  <Button
                    size="sm"
                    type="button"
                    variant="outline"
                    :disabled="saving !== null"
                    @click="editAgent(item)"
                    ><Pencil data-icon="inline-start" />{{
                      t("systemSettings.edit")
                    }}</Button
                  ><Button
                    size="sm"
                    type="button"
                    variant="destructive"
                    :disabled="saving !== null"
                    @click="deleteAgent = item"
                    ><Trash2 data-icon="inline-start" />{{
                      t("systemSettings.delete")
                    }}</Button
                  >
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        </CardContent>
      </Card>
    >
    <Sheet v-model:open="agentDialogOpen"
      ><SheetContent class="overflow-y-auto"
        ><form class="flex flex-col gap-6" @submit.prevent="saveAgent">
          <SheetHeader
            ><SheetTitle>{{
              t(
                editingAgentId === null
                  ? "systemSettings.createAgent"
                  : "systemSettings.updateAgent",
              )
            }}</SheetTitle
            ><SheetDescription>{{
              t("systemSettings.agentsDescription")
            }}</SheetDescription></SheetHeader
          ><FieldGroup class="p-4">
            <Field
              ><FieldLabel for="sheet-agent-name">{{
                t("systemSettings.agentName")
              }}</FieldLabel
              ><Input
                id="sheet-agent-name"
                v-model="agent.name"
                required
                :disabled="saving !== null" /></Field
            ><Field
              ><FieldLabel for="sheet-agent-address">{{
                t("systemSettings.agentAddress")
              }}</FieldLabel
              ><Input
                id="sheet-agent-address"
                v-model="agent.address"
                required
                :disabled="saving !== null" /></Field
            ><Field
              ><FieldLabel for="sheet-agent-type">{{
                t("systemSettings.addressType")
              }}</FieldLabel
              ><select
                id="sheet-agent-type"
                v-model="agent.address_type"
                class="h-9 rounded-md border border-input bg-transparent px-3 text-sm"
                :disabled="saving !== null"
              >
                <option :value="ModelsAgentAddressType.AgentAddressTypeDomain">
                  {{ t("systemSettings.domain") }}
                </option>
                <option :value="ModelsAgentAddressType.AgentAddressTypeIP">
                  {{ t("systemSettings.ip") }}
                </option>
              </select></Field
            ><Field
              ><FieldLabel for="sheet-agent-port">{{
                t("systemSettings.serverPort")
              }}</FieldLabel
              ><Input
                id="sheet-agent-port"
                v-model.number="agent.port"
                type="number"
                min="1"
                max="65535"
                :disabled="saving !== null" /></Field
            ><Field
              ><FieldLabel for="sheet-agent-token">{{
                t("systemSettings.agentToken")
              }}</FieldLabel
              ><Input
                id="sheet-agent-token"
                v-model="agent.token"
                required
                type="password"
                :disabled="saving !== null" /></Field></FieldGroup
          ><SheetFooter
            ><Button
              type="button"
              variant="outline"
              :disabled="saving !== null"
              @click="agentDialogOpen = false"
              >{{ t("systemSettings.cancel") }}</Button
            ><Button type="submit" :disabled="saving !== null"
              ><Spinner
                v-if="saving === 'agent'"
                data-icon="inline-start"
              /><Save v-else data-icon="inline-start" />{{
                t(
                  editingAgentId === null
                    ? "systemSettings.createAgent"
                    : "systemSettings.updateAgent",
                )
              }}</Button
            ></SheetFooter
          >
        </form></SheetContent
      ></Sheet
    >
    <AlertDialog :open="deleteAgent !== null"
      ><AlertDialogContent
        ><AlertDialogHeader
          ><AlertDialogTitle>{{
            t("systemSettings.deleteAgentTitle")
          }}</AlertDialogTitle
          ><AlertDialogDescription>{{
            t("systemSettings.deleteAgentDescription", {
              name: deleteAgent?.name ?? "",
            })
          }}</AlertDialogDescription></AlertDialogHeader
        ><AlertDialogFooter
          ><AlertDialogCancel
            :disabled="saving === 'delete'"
            @click="deleteAgent = null"
            >{{ t("systemSettings.cancel") }}</AlertDialogCancel
          ><AlertDialogAction
            :disabled="saving === 'delete'"
            @click.prevent="removeAgent"
            ><Spinner
              v-if="saving === 'delete'"
              data-icon="inline-start"
            /><Trash2 v-else data-icon="inline-start" />{{
              t("systemSettings.delete")
            }}</AlertDialogAction
          ></AlertDialogFooter
        ></AlertDialogContent
      ></AlertDialog
    >
  </div>
</template>
