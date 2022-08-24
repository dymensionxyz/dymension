/* eslint-disable */
import {
  StateStatus,
  stateStatusFromJSON,
  stateStatusToJSON,
} from "../rollapp/state_status";
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";
import { BlockDescriptors } from "../rollapp/block_descriptor";

export const protobufPackage = "dymensionxyz.dymension.rollapp";

/** StateInfo defines a rollapps' state. */
export interface StateInfo {
  /**
   * rollappId is the rollapp that the sequencer belongs to and asking to update
   * The rollappId follows the same standard as cosmos chain_id
   */
  rollappId: string;
  /**
   * stateIndex is a sequential increasing number, updating on each
   * state update used for indexing to a specific state info
   */
  stateIndex: number;
  /** sequencer is the bech32-encoded address of the sequencer sent the update */
  sequencer: string;
  /** startHeight is the block height of the first block in the batch */
  startHeight: number;
  /** numBlocks is the number of blocks included in this batch update */
  numBlocks: number;
  /** DAPath is the description of the location on the DA layer */
  DAPath: string;
  /** version is the version of the rollapp */
  version: number;
  /** creationHeight is the height at which the UpdateState took place */
  creationHeight: number;
  /** status is the status of the state update */
  status: StateStatus;
  /**
   * BDs is a list of block description objects (one per block)
   * the list must be ordered by height, starting from startHeight to startHeight+numBlocks-1
   */
  BDs: BlockDescriptors | undefined;
}

const baseStateInfo: object = {
  rollappId: "",
  stateIndex: 0,
  sequencer: "",
  startHeight: 0,
  numBlocks: 0,
  DAPath: "",
  version: 0,
  creationHeight: 0,
  status: 0,
};

export const StateInfo = {
  encode(message: StateInfo, writer: Writer = Writer.create()): Writer {
    if (message.rollappId !== "") {
      writer.uint32(10).string(message.rollappId);
    }
    if (message.stateIndex !== 0) {
      writer.uint32(16).uint64(message.stateIndex);
    }
    if (message.sequencer !== "") {
      writer.uint32(26).string(message.sequencer);
    }
    if (message.startHeight !== 0) {
      writer.uint32(32).uint64(message.startHeight);
    }
    if (message.numBlocks !== 0) {
      writer.uint32(40).uint64(message.numBlocks);
    }
    if (message.DAPath !== "") {
      writer.uint32(50).string(message.DAPath);
    }
    if (message.version !== 0) {
      writer.uint32(56).uint64(message.version);
    }
    if (message.creationHeight !== 0) {
      writer.uint32(64).uint64(message.creationHeight);
    }
    if (message.status !== 0) {
      writer.uint32(72).int32(message.status);
    }
    if (message.BDs !== undefined) {
      BlockDescriptors.encode(message.BDs, writer.uint32(82).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): StateInfo {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseStateInfo } as StateInfo;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.rollappId = reader.string();
          break;
        case 2:
          message.stateIndex = longToNumber(reader.uint64() as Long);
          break;
        case 3:
          message.sequencer = reader.string();
          break;
        case 4:
          message.startHeight = longToNumber(reader.uint64() as Long);
          break;
        case 5:
          message.numBlocks = longToNumber(reader.uint64() as Long);
          break;
        case 6:
          message.DAPath = reader.string();
          break;
        case 7:
          message.version = longToNumber(reader.uint64() as Long);
          break;
        case 8:
          message.creationHeight = longToNumber(reader.uint64() as Long);
          break;
        case 9:
          message.status = reader.int32() as any;
          break;
        case 10:
          message.BDs = BlockDescriptors.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): StateInfo {
    const message = { ...baseStateInfo } as StateInfo;
    if (object.rollappId !== undefined && object.rollappId !== null) {
      message.rollappId = String(object.rollappId);
    } else {
      message.rollappId = "";
    }
    if (object.stateIndex !== undefined && object.stateIndex !== null) {
      message.stateIndex = Number(object.stateIndex);
    } else {
      message.stateIndex = 0;
    }
    if (object.sequencer !== undefined && object.sequencer !== null) {
      message.sequencer = String(object.sequencer);
    } else {
      message.sequencer = "";
    }
    if (object.startHeight !== undefined && object.startHeight !== null) {
      message.startHeight = Number(object.startHeight);
    } else {
      message.startHeight = 0;
    }
    if (object.numBlocks !== undefined && object.numBlocks !== null) {
      message.numBlocks = Number(object.numBlocks);
    } else {
      message.numBlocks = 0;
    }
    if (object.DAPath !== undefined && object.DAPath !== null) {
      message.DAPath = String(object.DAPath);
    } else {
      message.DAPath = "";
    }
    if (object.version !== undefined && object.version !== null) {
      message.version = Number(object.version);
    } else {
      message.version = 0;
    }
    if (object.creationHeight !== undefined && object.creationHeight !== null) {
      message.creationHeight = Number(object.creationHeight);
    } else {
      message.creationHeight = 0;
    }
    if (object.status !== undefined && object.status !== null) {
      message.status = stateStatusFromJSON(object.status);
    } else {
      message.status = 0;
    }
    if (object.BDs !== undefined && object.BDs !== null) {
      message.BDs = BlockDescriptors.fromJSON(object.BDs);
    } else {
      message.BDs = undefined;
    }
    return message;
  },

  toJSON(message: StateInfo): unknown {
    const obj: any = {};
    message.rollappId !== undefined && (obj.rollappId = message.rollappId);
    message.stateIndex !== undefined && (obj.stateIndex = message.stateIndex);
    message.sequencer !== undefined && (obj.sequencer = message.sequencer);
    message.startHeight !== undefined &&
      (obj.startHeight = message.startHeight);
    message.numBlocks !== undefined && (obj.numBlocks = message.numBlocks);
    message.DAPath !== undefined && (obj.DAPath = message.DAPath);
    message.version !== undefined && (obj.version = message.version);
    message.creationHeight !== undefined &&
      (obj.creationHeight = message.creationHeight);
    message.status !== undefined &&
      (obj.status = stateStatusToJSON(message.status));
    message.BDs !== undefined &&
      (obj.BDs = message.BDs
        ? BlockDescriptors.toJSON(message.BDs)
        : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<StateInfo>): StateInfo {
    const message = { ...baseStateInfo } as StateInfo;
    if (object.rollappId !== undefined && object.rollappId !== null) {
      message.rollappId = object.rollappId;
    } else {
      message.rollappId = "";
    }
    if (object.stateIndex !== undefined && object.stateIndex !== null) {
      message.stateIndex = object.stateIndex;
    } else {
      message.stateIndex = 0;
    }
    if (object.sequencer !== undefined && object.sequencer !== null) {
      message.sequencer = object.sequencer;
    } else {
      message.sequencer = "";
    }
    if (object.startHeight !== undefined && object.startHeight !== null) {
      message.startHeight = object.startHeight;
    } else {
      message.startHeight = 0;
    }
    if (object.numBlocks !== undefined && object.numBlocks !== null) {
      message.numBlocks = object.numBlocks;
    } else {
      message.numBlocks = 0;
    }
    if (object.DAPath !== undefined && object.DAPath !== null) {
      message.DAPath = object.DAPath;
    } else {
      message.DAPath = "";
    }
    if (object.version !== undefined && object.version !== null) {
      message.version = object.version;
    } else {
      message.version = 0;
    }
    if (object.creationHeight !== undefined && object.creationHeight !== null) {
      message.creationHeight = object.creationHeight;
    } else {
      message.creationHeight = 0;
    }
    if (object.status !== undefined && object.status !== null) {
      message.status = object.status;
    } else {
      message.status = 0;
    }
    if (object.BDs !== undefined && object.BDs !== null) {
      message.BDs = BlockDescriptors.fromPartial(object.BDs);
    } else {
      message.BDs = undefined;
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
