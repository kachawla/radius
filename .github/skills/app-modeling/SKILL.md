## Guardrails

- Do NOT invent property names not in the schema. If unsure, read the relevant reference file.
- Do NOT set readOnly properties — they are output by recipes at deploy time.
- Do NOT generate recipes if one already exists in `resource-types-contrib`.
- Do NOT reference readOnly properties of other resources in Bicep (e.g. `database.properties.host`) — these are not available at compile time. Use connection auto-injection.
- Do NOT use array syntax where the schema specifies object maps (`connections`, `containers`, `ports`, `volumes`, `env` are all object maps).
- Do NOT place `connections` inside `containers` — it is a top-level property under `properties`.
- Do NOT include comments explaining skill rules or why properties are absent. The generated app.bicep must be clean, production-ready Bicep. Comments like "do not set readOnly properties" or "use existing recipe" must NOT appear in output.
- Ask for clarification if the app's architecture is ambiguous.