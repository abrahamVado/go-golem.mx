# Multi-tenancy

Always scope tenant-owned reads and writes by `company_id`. Never trust client-provided company IDs for private routes.
