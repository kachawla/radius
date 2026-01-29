# Automatic Application Discovery & App Definition Generation

* **Author**: Reshrahim team (ported from Reshrahim/radius spec `001-auto-app-discovery`)
* **Created**: 2026-01-28
* **Status**: Draft

## Topic Summary

Radius should make it trivial to adopt Radius for an existing application **without** first authoring Resource Types or Recipes by hand.

This feature adds a multi-step workflow:

1. **Discover**: analyze a local codebase and produce a human-readable discovery report
2. **Generate**: produce a complete, editable `app.bicep` application definition (services + dependencies + recipe bindings)
3. **Deploy**: deploy the generated app definition to an environment

A core design goal is **skills-first architecture**: the same underlying “skills” power the CLI (`rad`), AI coding agents (via MCP), and a programmatic API.

## Top level goals

- Zero/low-friction onboarding for **existing applications**.
- Detect common infrastructure dependencies (datastores, caches, queues, storage) and deployable services.
- Generate a valid `./radius/app.bicep` that developers can review/edit before deployment.
- Build capabilities as **composable skills** that can be invoked independently (CLI, MCP, API).
- Prefer deterministic outcomes where possible (same input → same output).

## Non-goals (out of scope)

- Remote repository cloning / cloud-based analysis of the codebase (analysis is local-only).
- Executing user code (static analysis only).
- Fully automated “one command from repo → production deploy” with zero user choices (interactive recipe selection remains part of the v1 UX).
- A complete enterprise-grade solution for parsing arbitrary wiki content with high reliability (see open questions).

## User profile and challenges

### Primary user
- **Developer** with an existing application who wants to deploy using Radius quickly.

### Secondary user
- **Platform engineer** who wants Radius to apply team standards (naming, tags, dev/prod policies) and prefer trusted/internal recipes.

### Challenges today
- Users must understand Radius concepts (Resource Types, Recipes, app.bicep wiring) before they can deploy.
- Teams often have existing IaC and standards, but adopting Radius can mean duplicating work.
- Recipe discovery and selection is hard without a guided workflow.

### Positive user outcome
- A developer can point Radius at a directory and receive:
  - services detected (what will run)
  - dependencies detected (what infra is needed)
  - team practices detected (how to name/tag/size resources)
  - a generated `app.bicep` that deploys successfully after review

## Key scenarios

### Scenario 1: Discover an existing application (P1)
A developer runs `rad app discover .` and Radius produces `./radius/discovery.md` listing detected services, dependencies, confidence levels, and evidence.

### Scenario 2: Generate an application definition (P1)
After discovery, a developer runs `rad app generate` and Radius creates `./radius/app.bicep`, prompting for recipe choices when multiple options exist.

### Scenario 3: Deploy to an environment (P1)
A developer runs `rad deploy ./radius/app.bicep -e dev` and gets a working deployment.

### Scenario 4: AI agent experience via MCP (P1)
An AI agent can invoke the same discovery/generation skills via MCP, present results conversationally, and write the same artifacts.

### Scenario 5: Apply team infrastructure practices (P2)
A platform engineer provides or Radius detects team standards (tags, naming patterns, environment-specific defaults). Generation applies these practices.

## Desired user experience outcome (end-to-end)

### The scenario
Node.js e-commerce app with:
- Services: `api-server` (Express, :3000), `worker`
- Dependencies: PostgreSQL, Redis, Azure Blob Storage

Goal: existing codebase → deployable Radius app with **no manual Resource Type/Recipe authoring**.

### Workflow overview

```
     USER                                RADIUS                           EXTERNAL SOURCES
       |                                   |                                     |
       |  1. DISCOVER                      |                                     |
       |  "Analyze my codebase"            |                                     |
       |---------------------------------->|                                     |
       |                                   |  Analyze codebase...                |
       |                                   |  - Detect dependencies              |
       |                                   |  - Find services                    |
       |                                   |  - Extract team practices           |
       |                                   |                                     |
       |  ./radius/discovery.md            |                                     |
       |<----------------------------------|                                     |
       |                                   |                                     |
       |  2. GENERATE                      |                                     |
       |  "Create my app definition"       |                                     |
       |---------------------------------->|                                     |
       |                                   |  Generate Resource Types...         |
       |                                   |  (applying team practices)          |
       |                                   |                                     |
       |                                   |  Search for Recipes...              |
       |                                   |------------------------------------>|
       |                                   |                    AVM, Terraform,  |
       |                                   |<------------------- Bicep repos     |
       |                                   |                                     |
       |  Recipe Options:                  |                                     |
       |  PostgreSQL -> [1] Azure (AVM)    |                                     |
       |               [2] Container       |                                     |
       |<----------------------------------|                                     |
       |                                   |                                     |
       |  Select: 1                        |                                     |
       |---------------------------------->|                                     |
       |                                   |  Generate app.bicep...              |
       |                                   |  Validate...                        |
       |                                   |                                     |
       |  ./radius/app.bicep               |                                     |
       |<----------------------------------|                                     |
       |                                   |                                     |
       |  3. DEPLOY                        |                                     |
       |  "Deploy to production"           |                                     |
       |---------------------------------->|                                     |
       |                                   |  Provision infrastructure...        |
       |                                   |------------------------------------>|
       |                                   |                    Azure PostgreSQL |
       |                                   |<-------------------Azure Redis      |
       |                                   |                    Storage Account  |
       |                                   |  Deploy containers...               |
       |                                   |                                     |
       |  Deployment Complete!             |                                     |
       |  https://my-app.azurecontainer... |                                     |
       |<----------------------------------|                                     |
       v                                   v                                     v
```

## Architecture (skills-first)

All interfaces invoke the same skills layer:

```
+--------------------------------------------------------------------------+
|                          RADIUS DISCOVERY                                |
+--------------------------------------------------------------------------+
|                                                                          |
|    +------------------+   +------------------+   +------------------+    |
|    |    AI Agents     |   |     rad CLI      |   |  Programmatic    |    |
|    |   (via MCP)      |   |                  |   |      API         |    |
|    +--------+---------+   +--------+---------+   +--------+---------+    |
|             |                      |                      |              |
|             +----------------------+----------------------+              |
|                                    |                                     |
|                                    v                                     |
|    +----------------------------------------------------------------+    |
|    |                        SKILLS LAYER                            |    |
|    +----------------------------------------------------------------+    |
|    |  discover_dependencies | discover_services | discover_team_    |    |
|    |  discover_recipes | generate_resource_types | generate_app_    |    |
|    |  validate_app_definition                                       |    |
|    +----------------------------------------------------------------+    |
|                                    |                                     |
|                                    v                                     |
|    +----------------------------------------------------------------+    |
|    |                        CORE ENGINE                             |    |
|    +----------------------------------------------------------------+    |
|    |  Language Analyzers | Team Practices Analyzer | Bicep Generator|    |
|    +----------------------------------------------------------------+    |
|                                                                          |
+--------------------------------------------------------------------------+
```

### Skills (proposed contracts)

| Skill | Phase | Purpose | Output (shape) |
|------|------|---------|----------------|
| `discover_dependencies` | Discover | Detect infra dependencies from code | `{dependencies: [{type, technology, confidence, evidence}]}` |
| `discover_services` | Discover | Detect deployable services/entrypoints | `{services: [{name, type, port, entrypoint}]}` |
| `discover_team_practices` | Discover | Detect team conventions from IaC/docs/config | `{practices: [{category, convention, source, environment}]}` |
| `generate_resource_types` | Generate | Produce resource type schemas based on detected deps + practices | `{resourceTypes: [{name, schema, outputs}]}` |
| `discover_recipes` | Generate | Find matching recipes from configured sources | `{recipes: [{resourceType, name, source, iacType}]}` |
| `generate_app_definition` | Generate | Generate `app.bicep` | `{path, content}` |
| `validate_app_definition` | Generate | Validate generated output | `{valid, errors: []}` |

## Key dependencies and risks

- Dependency: recipe source availability (AVM/internal repos). Must degrade gracefully.
- Risk: determinism vs “smart” generation. Requires a clear stance on catalogs vs on-the-fly generation.
- Risk: documentation parsing for team practices is hard to make reliable/deterministic.
- Risk: container image strategy (Dockerfiles vs placeholders) impacts “deploy immediately” success rate.

## Current state

In Radius today (upstream), users generally author `app.bicep` manually or start from templates/samples. This proposal adds a new “discovery → generate” workflow and a skills layer that can be reused by CLI and MCP.

> Note: Whether upstream already has primitives for “skills via MCP” and “app discovery” needs validation against `radius-project/radius` (see “Backend alignment” below).

## Key investments (what we will build)

### Investment 1: Discovery outputs (`rad app discover`)
- Implement discovery as local static analysis.
- Output `./radius/discovery.md` with:
  - dependencies + confidence tiers + evidence
  - services + ports + entrypoints
  - team practices + sources
  - warnings for partial failures

### Investment 2: Generation (`rad app generate`)
- Generate Resource Types (catalog or hybrid; see OQ-1).
- Discover and rank recipes from configured sources.
- Interactive selection when multiple options exist (plus `--accept-defaults`).
- Generate `./radius/app.bicep` with wiring between services and resources.
- Validate Bicep syntax and references.

### Investment 3: MCP + programmatic skills
- Expose each skill with JSON I/O schemas.
- Provide `rad mcp serve` with stdio + HTTP transports.
- Ensure CLI and MCP share the same implementation.

### Investment 4: Team practices
- Read `.radius/team-practices.yaml` if present.
- Detect practices from existing IaC (Terraform/Bicep/ARM).
- (Optional) ingest structured docs sources, with clear precedence rules.

## Open questions to resolve (summary)

- **OQ-1**: Pre-defined catalog vs generated Resource Types (determinism vs flexibility)
- **OQ-2**: Resource Type schema design strategy (derived vs standardized vs minimal+extensions)
- **OQ-3**: Container image strategy in generated app.bicep
- **OQ-4**: MCP server deployment model
- **OQ-5**: Recipe source configuration UX
- **OQ-6/OQ-7**: Team practices sources + parsing strategy

## Appendix: Requirements & acceptance criteria

(Ported from original spec; consider moving into separate appendix file once stabilized.)
- Functional requirements (FR-01..)
- Non-functional requirements (NFR-01..)
- Success criteria (SC-001..)
- Edge cases