---
title: Architecture
last-update: unset
owner: curator
status: seed
---

# Architecture

This live document describes the overall architecture of the project and indexes deeper architecture memory by project domain. Keep this file high-level enough that agents can quickly route themselves to the right domain-specific document.

## System Overview

Document the runtime shape, major modules, data flow, external services, and command boundaries.

## Architecture Documentation Index

Add domain-specific architecture docs here when a domain has stable architecture details worth retrieving later.

Recommended path shape:

- `.kumite/memory/architecture/<domain>/<domain>-architecture.md`

Current domain docs:

- None yet.

## Architecture Rules

Record constraints future changes must follow. Include package boundaries, layering rules, allowed dependencies, and forbidden shortcuts.

## Important Flows

Describe the flows most likely to affect implementation work. Prefer concise, named flows with links to code paths.

## Curation Guidance

- Keep broad architectural rules here.
- Move deep domain details into domain-specific files and link them from `Architecture Documentation Index`.
- Add or update architecture memory only after implementation is complete, reviewed, and the user has classified the work as ready.
- Do not duplicate business/product rules here; link to `.kumite/memory/business-rules.md` when behavior depends on domain policy.

## Change Log

- Seeded by kumite setup. Replace this entry when the first real architecture review is completed.
