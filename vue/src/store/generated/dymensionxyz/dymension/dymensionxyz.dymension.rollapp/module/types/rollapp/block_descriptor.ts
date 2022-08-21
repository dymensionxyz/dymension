/* eslint-disable */
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "dymensionxyz.dymension.rollapp";

export interface BlockDescriptor {
  height: number;
  stateRoot: string;
  intermediateStatesRoot: string;
}

const baseBlockDescriptor: object = {
  height: 0,
  stateRoot: "",
  intermediateStatesRoot: "",
};

export const BlockDescriptor = {
  encode(message: BlockDescriptor, writer: Writer = Writer.create()): Writer {
    if (message.height !== 0) {
      writer.uint32(8).uint64(message.height);
    }
    if (message.stateRoot !== "") {
      writer.uint32(18).string(message.stateRoot);
    }
    if (message.intermediateStatesRoot !== "") {
      writer.uint32(26).string(message.intermediateStatesRoot);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): BlockDescriptor {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseBlockDescriptor } as BlockDescriptor;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.height = longToNumber(reader.uint64() as Long);
          break;
        case 2:
          message.stateRoot = reader.string();
          break;
        case 3:
          message.intermediateStatesRoot = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): BlockDescriptor {
    const message = { ...baseBlockDescriptor } as BlockDescriptor;
    if (object.height !== undefined && object.height !== null) {
      message.height = Number(object.height);
    } else {
      message.height = 0;
    }
    if (object.stateRoot !== undefined && object.stateRoot !== null) {
      message.stateRoot = String(object.stateRoot);
    } else {
      message.stateRoot = "";
    }
    if (
      object.intermediateStatesRoot !== undefined &&
      object.intermediateStatesRoot !== null
    ) {
      message.intermediateStatesRoot = String(object.intermediateStatesRoot);
    } else {
      message.intermediateStatesRoot = "";
    }
    return message;
  },

  toJSON(message: BlockDescriptor): unknown {
    const obj: any = {};
    message.height !== undefined && (obj.height = message.height);
    message.stateRoot !== undefined && (obj.stateRoot = message.stateRoot);
    message.intermediateStatesRoot !== undefined &&
      (obj.intermediateStatesRoot = message.intermediateStatesRoot);
    return obj;
  },

  fromPartial(object: DeepPartial<BlockDescriptor>): BlockDescriptor {
    const message = { ...baseBlockDescriptor } as BlockDescriptor;
    if (object.height !== undefined && object.height !== null) {
      message.height = object.height;
    } else {
      message.height = 0;
    }
    if (object.stateRoot !== undefined && object.stateRoot !== null) {
      message.stateRoot = object.stateRoot;
    } else {
      message.stateRoot = "";
    }
    if (
      object.intermediateStatesRoot !== undefined &&
      object.intermediateStatesRoot !== null
    ) {
      message.intermediateStatesRoot = object.intermediateStatesRoot;
    } else {
      message.intermediateStatesRoot = "";
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
