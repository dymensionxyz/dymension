// Package agent maintains the registry of TEE agents, their append-only
// attested action log, reputation, and per-agent escrows.
// Agents register with an attestation policy, may rotate it after a timelock,
// append actions verified against the policy currently in force, and pay
// attested transfers under an owner-set, per-window spend budget. A transfer
// may carry an enclave-attested invoice_ref, which settles at most once per
// (agent, recipient) and leaves a permanent, queryable settlement receipt. A
// governance-gated denylist prevents byte-identical policy fingerprints from
// registering new agents, appending attested actions, or fulfilling eIBC orders
// through agent-bound on-demand LPs. Unrevoking is fully reversible since agent
// state is never mutated by revocation.
package agent
