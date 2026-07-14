// Package agent maintains the registry of TEE agents and their action log.
// Agents register with an attestation policy, may rotate it after a timelock,
// and append actions verified against the policy currently in force. A
// governance-gated denylist prevents byte-identical policy fingerprints from
// registering new agents or appending attested actions. Unrevoking is fully
// reversible since agent state is never mutated by revocation.
package agent
