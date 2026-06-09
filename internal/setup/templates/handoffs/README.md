# Kumite Handoffs

This directory stores durable project-local mirrors of the current Kumite chain handoffs.

The pi-subagents runtime may store step outputs in a temporary chain-run directory. Kumite agents mirror the latest scout, planner, implementer, reviewer, and rework summaries here so future sessions and curator updates can inspect the last workflow without depending on runtime temp paths.
