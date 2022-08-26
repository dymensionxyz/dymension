/* eslint-disable */
export const protobufPackage = "dymensionxyz.dymension.sequencer";

/** OperatingStatus defines the operating status of a sequencer */
export enum OperatingStatus {
  /** OPERATING_STATUS_UNSPECIFIED - OPERATING_STATUS_UNSPECIFIED defines zero-value for status ordering */
  OPERATING_STATUS_UNSPECIFIED = 0,
  /** OPERATING_STATUS_PROPOSER - OPERATING_STATUS_PROPOSER defines a sequencer that is active and can propose state updates */
  OPERATING_STATUS_PROPOSER = 1,
  /** OPERATING_STATUS_INACTIVE - OPERATING_STATUS_INACTIVE defines a sequencer that is not active and won't be scheduled */
  OPERATING_STATUS_INACTIVE = 2,
  UNRECOGNIZED = -1,
}

export function operatingStatusFromJSON(object: any): OperatingStatus {
  switch (object) {
    case 0:
    case "OPERATING_STATUS_UNSPECIFIED":
      return OperatingStatus.OPERATING_STATUS_UNSPECIFIED;
    case 1:
    case "OPERATING_STATUS_PROPOSER":
      return OperatingStatus.OPERATING_STATUS_PROPOSER;
    case 2:
    case "OPERATING_STATUS_INACTIVE":
      return OperatingStatus.OPERATING_STATUS_INACTIVE;
    case -1:
    case "UNRECOGNIZED":
    default:
      return OperatingStatus.UNRECOGNIZED;
  }
}

export function operatingStatusToJSON(object: OperatingStatus): string {
  switch (object) {
    case OperatingStatus.OPERATING_STATUS_UNSPECIFIED:
      return "OPERATING_STATUS_UNSPECIFIED";
    case OperatingStatus.OPERATING_STATUS_PROPOSER:
      return "OPERATING_STATUS_PROPOSER";
    case OperatingStatus.OPERATING_STATUS_INACTIVE:
      return "OPERATING_STATUS_INACTIVE";
    default:
      return "UNKNOWN";
  }
}
