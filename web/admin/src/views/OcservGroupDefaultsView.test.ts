import { flushPromises, mount } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";

import { i18n } from "@/locales";
import { mockOcservDefaultGroupConfig } from "@/mocks";
import OcservGroupDefaultsView from "@/views/OcservGroupDefaultsView.vue";

const service = vi.hoisted(() => ({
  get: vi.fn(),
  update: vi.fn(),
}));

vi.mock("@/api/services/ocserv-groups", () => ({
  getOcservDefaultGroupConfig: service.get,
  updateOcservDefaultGroupConfig: service.update,
}));

function mountView() {
  return mount(OcservGroupDefaultsView, {
    global: {
      plugins: [i18n],
      stubs: {
        DashboardLayout: { template: "<div><slot /></div>" },
      },
    },
  });
}

afterEach(() => {
  vi.clearAllMocks();
});

describe("OcservGroupDefaultsView", () => {
  it("shows loading state while fetching the configuration", () => {
    service.get.mockReturnValue(new Promise(() => undefined));

    const wrapper = mountView();

    expect(wrapper.find('[aria-busy="true"]').exists()).toBe(true);
  });

  it("populates the form from the fetched configuration", async () => {
    service.get.mockResolvedValue(
      structuredClone(mockOcservDefaultGroupConfig),
    );

    const wrapper = mountView();
    await flushPromises();

    expect(wrapper.get<HTMLInputElement>("#ipv4-network").element.value).toBe(
      "192.168.1.0/24",
    );
    expect(wrapper.get<HTMLInputElement>("#dns").element.value).toBe(
      "8.8.8.8, 1.1.1.1",
    );
    expect(
      wrapper.get<HTMLInputElement>("#max-same-clients").element.value,
    ).toBe("2");
  });

  it("submits edited values with the generated PATCH request shape", async () => {
    service.get.mockResolvedValue(
      structuredClone(mockOcservDefaultGroupConfig),
    );
    service.update.mockResolvedValue(undefined);
    const wrapper = mountView();
    await flushPromises();

    await wrapper.get("#ipv4-network").setValue("10.20.0.0/16");
    await wrapper.get("form").trigger("submit");
    await flushPromises();

    expect(service.update).toHaveBeenCalledWith({
      config: expect.objectContaining({
        dns: ["8.8.8.8", "1.1.1.1"],
        "ipv4-network": "10.20.0.0/16",
      }),
    });
    expect(wrapper.text()).toContain("Default group configuration saved.");
  });

  it("marks non-integer numeric values as invalid", async () => {
    service.get.mockResolvedValue(
      structuredClone(mockOcservDefaultGroupConfig),
    );
    const wrapper = mountView();
    await flushPromises();

    const input = wrapper.get<HTMLInputElement>("#max-same-clients");
    await input.setValue("1.5");

    expect(input.element.validity.stepMismatch).toBe(true);
  });

  it("renders a fetch error and retries the request", async () => {
    service.get
      .mockRejectedValueOnce(new Error("connection failed"))
      .mockResolvedValueOnce(structuredClone(mockOcservDefaultGroupConfig));
    const wrapper = mountView();
    await flushPromises();

    expect(wrapper.text()).toContain("connection failed");
    await wrapper.get("button").trigger("click");
    await flushPromises();

    expect(service.get).toHaveBeenCalledTimes(2);
    expect(wrapper.find("#ipv4-network").exists()).toBe(true);
  });

  it("preserves the form and reports an update failure", async () => {
    service.get.mockResolvedValue(
      structuredClone(mockOcservDefaultGroupConfig),
    );
    service.update.mockRejectedValue(new Error("save failed"));
    const wrapper = mountView();
    await flushPromises();

    await wrapper.get("form").trigger("submit");
    await flushPromises();

    expect(wrapper.text()).toContain("save failed");
    expect(wrapper.find("#ipv4-network").exists()).toBe(true);
  });
});
