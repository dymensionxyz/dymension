/* eslint-disable */
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";
import { StateInfo } from "../rollapp/state_info";

export const protobufPackage = "dymensionxyz.dymension.rollapp";

export interface RollappStateInfo {
  rollappId: string;
  sateInfo: StateInfo | undefined;
  stateIndex: number;
}

const baseRollappStateInfo: object = { rollappId: "", stateIndex: 0 };

export const RollappStateInfo = {
  encode(message: RollappStateInfo, writer: Writer = Writer.create()): Writer {
    if (message.rollappId !== "") {
      writer.uint32(10).string(message.rollappId);
    }
    if (message.sateInfo !== undefined) {
      StateInfo.encode(message.sateInfo, writer.uint32(18).fork()).ldelim();
    }
    if (message.stateIndex !== 0) {
      writer.uint32(24).uint64(message.stateIndex);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): RollappStateInfo {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseRollappStateInfo } as RollappStateInfo;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.rollappId = reader.string();
          break;
        case 2:
          message.sateInfo = StateInfo.decode(reader, reader.uint32());
          break;
        case 3:
          message.stateIndex = longToNumber(reader.uint64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): RollappStateInfo {
    const message = { ...baseRollappStateInfo } as RollappStateInfo;
    if (object.rollappId !== undefined && object.rollappId !== null) {
      message.rollappId = String(object.rollappId);
    } else {
      message.rollappId = "";
    }
    if (object.sateInfo !== undefined && object.sateInfo !== null) {
      message.sateInfo = StateInfo.fromJSON(object.sateInfo);
    } else {
      message.sateInfo = undefined;
    }
    if (object.stateIndex !== undefined && object.stateIndex !== null) {
      message.stateIndex = Number(object.stateIndex);
    } else {
      message.stateIndex = 0;
    }
    return message;
  },

  toJSON(message: RollappStateInfo): unknown {
    const obj: any = {};
    message.rollappId !== undefined && (obj.rollappId = message.rollappId);
    message.sateInfo !== undefined &&
      (obj.sateInfo = message.sateInfo
        ? StateInfo.toJSON(message.sateInfo)
        : undefined);
    message.stateIndex !== undefined && (obj.stateIndex = message.stateIndex);
    return obj;
  },

  fromPartial(object: DeepPartial<RollappStateInfo>): RollappStateInfo {
    const message = { ...baseRollappStateInfo } as RollappStateInfo;
    if (object.rollappId !== undefined && object.rollappId !== null) {
      message.rollappId = object.rollappId;
    } else {
      message.rollappId = "";
    }
    if (object.sateInfo !== undefined && object.sateInfo !== null) {
      message.sateInfo = StateInfo.fromPartial(object.sateInfo);
    } else {
      message.sateInfo = undefined;
    }
    if (object.stateIndex !== undefined && object.stateIndex !== null) {
      message.stateIndex = object.stateIndex;
    } else {
      message.stateIndex = 0;
    }
    return message;
  },
};

declare var self: any | undefined;
declare var window: any | undefined;
var globalThis: any = (() => {
  if (typeof globalThis !== "undefined") return globalThis;
  if (typeof self !== "undefined") return self;
  if (typeof window !== "undefined") return window;
  if (typeof global !== "undefined") return global;
  throw "Unable to locate global object";
})();

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

function longToNumber(long: Long): number {
  if (long.gt(Number.MAX_SAFE_INTEGER)) {
    throw new globalThis.Error("Value is larger than Number.MAX_SAFE_INTEGER");
  }
  return long.toNumber();
}

if (util.Long !== Long) {
  util.Long = Long as any;
  configure();
}
