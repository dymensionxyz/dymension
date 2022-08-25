/* eslint-disable */
import {
  OperatingStatus,
  operatingStatusFromJSON,
  operatingStatusToJSON,
} from "../sequencer/operating_status";
import { Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "dymensionxyz.dymension.sequencer";

export interface Scheduler {
  sequencerAddress: string;
  status: OperatingStatus;
}

const baseScheduler: object = { sequencerAddress: "", status: 0 };

export const Scheduler = {
  encode(message: Scheduler, writer: Writer = Writer.create()): Writer {
    if (message.sequencerAddress !== "") {
      writer.uint32(10).string(message.sequencerAddress);
    }
    if (message.status !== 0) {
      writer.uint32(16).int32(message.status);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): Scheduler {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseScheduler } as Scheduler;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.sequencerAddress = reader.string();
          break;
        case 2:
          message.status = reader.int32() as any;
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): Scheduler {
    const message = { ...baseScheduler } as Scheduler;
    if (
      object.sequencerAddress !== undefined &&
      object.sequencerAddress !== null
    ) {
      message.sequencerAddress = String(object.sequencerAddress);
    } else {
      message.sequencerAddress = "";
    }
    if (object.status !== undefined && object.status !== null) {
      message.status = operatingStatusFromJSON(object.status);
    } else {
      message.status = 0;
    }
    return message;
  },

  toJSON(message: Scheduler): unknown {
    const obj: any = {};
    message.sequencerAddress !== undefined &&
      (obj.sequencerAddress = message.sequencerAddress);
    message.status !== undefined &&
      (obj.status = operatingStatusToJSON(message.status));
    return obj;
  },

  fromPartial(object: DeepPartial<Scheduler>): Scheduler {
    const message = { ...baseScheduler } as Scheduler;
    if (
      object.sequencerAddress !== undefined &&
      object.sequencerAddress !== null
    ) {
      message.sequencerAddress = object.sequencerAddress;
    } else {
      message.sequencerAddress = "";
    }
    if (object.status !== undefined && object.status !== null) {
      message.status = object.status;
    } else {
      message.status = 0;
    }
    return message;
  },
};

type Builtin = Date | Function | Uint8Array | string | number | undefined;
export type DeepPartial<T> = T extends Builtin
  ? T
  : T extends Array<infer U>
  ? Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U>
  ? ReadonlyArray<DeepPartial<U>>
  : T extends {}
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>;
