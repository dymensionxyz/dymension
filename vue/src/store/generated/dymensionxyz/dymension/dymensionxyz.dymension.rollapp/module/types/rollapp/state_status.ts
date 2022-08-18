/* eslint-disable */
export const protobufPackage = "dymensionxyz.dymension.rollapp";

export enum StateStatus {
  /** STATE_STATUS_UNSPECIFIED - zero-value for status ordering */
  STATE_STATUS_UNSPECIFIED = 0,
  /** STATE_STATUS_RECEIVED - STATE_STATUS_RECEIVED defines a rollapp state where the state update transaction was published on dYmension chain */
  STATE_STATUS_RECEIVED = 1,
  /** STATE_STATUS_FINALIZED - STATE_STATUS_FINALIZED defines a rollapp state where the the "Dispute Period" has ended and this state is considered final */
  STATE_STATUS_FINALIZED = 2,
  UNRECOGNIZED = -1,
}

export function stateStatusFromJSON(object: any): StateStatus {
  switch (object) {
    case 0:
    case "STATE_STATUS_UNSPECIFIED":
      return StateStatus.STATE_STATUS_UNSPECIFIED;
    case 1:
    case "STATE_STATUS_RECEIVED":
      return StateStatus.STATE_STATUS_RECEIVED;
    case 2:
    case "STATE_STATUS_FINALIZED":
      return StateStatus.STATE_STATUS_FINALIZED;
    case -1:
    case "UNRECOGNIZED":
    default:
      return StateStatus.UNRECOGNIZED;
  }
}

export function stateStatusToJSON(object: StateStatus): string {
  switch (object) {
    case StateStatus.STATE_STATUS_UNSPECIFIED:
      return "STATE_STATUS_UNSPECIFIED";
    case StateStatus.STATE_STATUS_RECEIVED:
      return "STATE_STATUS_RECEIVED";
    case StateStatus.STATE_STATUS_FINALIZED:
      return "STATE_STATUS_FINALIZED";
    default:
      return "UNKNOWN";
  }
}
