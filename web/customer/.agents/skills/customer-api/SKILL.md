---
name: customer-api
description: Maintain frontend integration with the backend Customers Swagger tag.
---

Treat the latest Customers Swagger tag as authoritative. Keep endpoint calls in src/api/services/customers.ts, schema types in src/api/generated, and bearer handling in src/api/http.ts. Never put URLs in components. Login is public; every other customer request is protected. Match validation, query names, pagination, file response types, and response shapes exactly.
