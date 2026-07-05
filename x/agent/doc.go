// Package agent maintains the registry of TEE agents and their action log.
// Agents register with an immutable attestation policy and append attested
// actions verified against it. A governance-gated denylist of policy
// fingerprints revokes compromised TEE images fleet-wide: a revoked policy can
// neither register new agents nor append attested actions, and unrevoking is
// fully reversible since agent state is never mutated by revocation.
package agent
