// config/orval.config.ts
import { BASE_API_URL } from "@/share/constant/global"
import { defineConfig } from "orval"

export default defineConfig({
  adminDashboard: {
    input: "http://185.226.117.170:8080/openapi.json",
    output: {
      baseUrl: BASE_API_URL,
      mode: "tags-split",
      target: "../../api/endpoints",
      schemas: "../../api/models",
      client: "react-query",
      mock: false,
      httpClient: "fetch",
      formatter: "prettier",
      override: {
        mutator: {
          path: "../ovral/customInstance.ts",
          name: "customInstance",
        },
      },
    },
  },
})
