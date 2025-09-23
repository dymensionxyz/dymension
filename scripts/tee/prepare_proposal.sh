# Check required environment variables

: ${PROPOSAL_PATH:?Error: PROPOSAL_PATH is not set}
: ${POLICY_VALUES_PATH:?Error: POLICY_VALUES_PATH is not set}
: ${POLICY_QUERY_PATH:?Error: POLICY_QUERY_PATH is not set}
: ${POLICY_STRUCTURE_PATH:?Error: POLICY_STRUCTURE_PATH is not set}
: ${GCP_ROOT_CERT_PATH:?Error: GCP_ROOT_CERT_PATH is not set}

POLICY_VALUES=$(cat "${POLICY_VALUES_PATH}")
POLICY_STRUCTURE=$(cat "${POLICY_STRUCTURE_PATH}")
POLICY_QUERY=$(cat "${POLICY_QUERY_PATH}")
GCP_ROOT_CERT=$(cat "${GCP_ROOT_CERT_PATH}")

jq --arg pv "$POLICY_VALUES" \
   --arg ps "$POLICY_STRUCTURE" \
   --arg pq "$POLICY_QUERY" \
   --arg gc "$GCP_ROOT_CERT" \
   '.messages[0].params.tee_config.policy_values = $pv |
    .messages[0].params.tee_config.policy_structure = $ps |
    .messages[0].params.tee_config.policy_query = $pq |
    .messages[0].params.tee_config.gcp_root_cert_pem = $gc' \
   "${PROPOSAL_PATH}" > "${PROPOSAL_PATH}.populated.json"
