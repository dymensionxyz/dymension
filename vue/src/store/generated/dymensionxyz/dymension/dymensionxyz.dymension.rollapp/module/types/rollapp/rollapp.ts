/* eslint-disable */
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";
import { Sequencers } from "../rollapp/sequencers";

export const protobufPackage = "dymensionxyz.dymension.rollapp";

export interface Rollapp {
  rollappId: string;
  creator: string;
  version: number;
  codeStamp: string;
  genesisPath: string;
  maxWithholdingBlocks: number;
  maxSequencers: number;
  permissionedAddresses: Sequencers | undefined;
}

const baseRollapp: object = {
  rollappId: "",
  creator: "",
  version: 0,
  codeStamp: "",
  genesisPath: "",
  maxWithholdingBlocks: 0,
  maxSequencers: 0,
};

export const Rollapp = {
  encode(message: Rollapp, writer: Writer = Writer.create()): Writer {
    if (message.rollappId !== "") {
      writer.uint32(10).string(message.rollappId);
    }
    if (message.creator !== "") {
      writer.uint32(18).string(message.creator);
    }
    if (message.version !== 0) {
      writer.uint32(24).uint64(message.version);
    }
    if (message.codeStamp !== "") {
      writer.uint32(34).string(message.codeStamp);
    }
    if (message.genesisPath !== "") {
      writer.uint32(42).string(message.genesisPath);
    }
    if (message.maxWithholdingBlocks !== 0) {
      writer.uint32(48).uint64(message.maxWithholdingBlocks);
    }
    if (message.maxSequencers !== 0) {
      writer.uint32(56).uint64(message.maxSequencers);
    }
    if (message.permissionedAddresses !== undefined) {
      Sequencers.encode(
        message.permissionedAddresses,
        writer.uint32(66).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): Rollapp {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseRollapp } as Rollapp;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.rollappId = reader.string();
          break;
        case 2:
          message.creator = reader.string();
          break;
        case 3:
          message.version = longToNumber(reader.uint64() as Long);
          break;
        case 4:
          message.codeStamp = reader.string();
          break;
        case 5:
          message.genesisPath = reader.string();
          break;
        case 6:
          message.maxWithholdingBlocks = longToNumber(reader.uint64() as Long);
          break;
        case 7:
          message.maxSequencers = longToNumber(reader.uint64() as Long);
          break;
        case 8:
          message.permissionedAddresses = Sequencers.decode(
            reader,
            reader.uint32()
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): Rollapp {
    const message = { ...baseRollapp } as Rollapp;
    if (object.rollappId !== undefined && object.rollappId !== null) {
      message.rollappId = String(object.rollappId);
    } else {
      message.rollappId = "";
    }
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.version !== undefined && object.version !== null) {
      message.version = Number(object.version);
    } else {
      message.version = 0;
    }
    if (object.codeStamp !== undefined && object.codeStamp !== null) {
      message.codeStamp = String(object.codeStamp);
    } else {
      message.codeStamp = "";
    }
    if (object.genesisPath !== undefined && object.genesisPath !== null) {
      message.genesisPath = String(object.genesisPath);
    } else {
      message.genesisPath = "";
    }
    if (
      object.maxWithholdingBlocks !== undefined &&
      object.maxWithholdingBlocks !== null
    ) {
      message.maxWithholdingBlocks = Number(object.maxWithholdingBlocks);
    } else {
      message.maxWithholdingBlocks = 0;
    }
    if (object.maxSequencers !== undefined && object.maxSequencers !== null) {
      message.maxSequencers = Number(object.maxSequencers);
    } else {
      message.maxSequencers = 0;
    }
    if (
      object.permissionedAddresses !== undefined &&
      object.permissionedAddresses !== null
    ) {
      message.permissionedAddresses = Sequencers.fromJSON(
        object.permissionedAddresses
      );
    } else {
      message.permissionedAddresses = undefined;
    }
    return message;
  },

  toJSON(message: Rollapp): unknown {
    const obj: any = {};
    message.rollappId !== undefined && (obj.rollappId = message.rollappId);
    message.creator !== undefined && (obj.creator = message.creator);
    message.version !== undefined && (obj.version = message.version);
    message.codeStamp !== undefined && (obj.codeStamp = message.codeStamp);
    message.genesisPath !== undefined &&
      (obj.genesisPath = message.genesisPath);
    message.maxWithholdingBlocks !== undefined &&
      (obj.maxWithholdingBlocks = message.maxWithholdingBlocks);
    message.maxSequencers !== undefined &&
      (obj.maxSequencers = message.maxSequencers);
    message.permissionedAddresses !== undefined &&
      (obj.permissionedAddresses = message.permissionedAddresses
        ? Sequencers.toJSON(message.permissionedAddresses)
        : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<Rollapp>): Rollapp {
    const message = { ...baseRollapp } as Rollapp;
    if (object.rollappId !== undefined && object.rollappId !== null) {
      message.rollappId = object.rollappId;
    } else {
      message.rollappId = "";
    }
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.version !== undefined && object.version !== null) {
      message.version = object.version;
    } else {
      message.version = 0;
    }
    if (object.codeStamp !== undefined && object.codeStamp !== null) {
      message.codeStamp = object.codeStamp;
    } else {
      message.codeStamp = "";
    }
    if (object.genesisPath !== undefined && object.genesisPath !== null) {
      message.genesisPath = object.genesisPath;
    } else {
      message.genesisPath = "";
    }
    if (
      object.maxWithholdingBlocks !== undefined &&
      object.maxWithholdingBlocks !== null
    ) {
      message.maxWithholdingBlocks = object.maxWithholdingBlocks;
    } else {
      message.maxWithholdingBlocks = 0;
    }
    if (object.maxSequencers !== undefined && object.maxSequencers !== null) {
      message.maxSequencers = object.maxSequencers;
    } else {
      message.maxSequencers = 0;
    }
    if (
      object.permissionedAddresses !== undefined &&
      object.permissionedAddresses !== null
    ) {
      message.permissionedAddresses = Sequencers.fromPartial(
        object.permissionedAddresses
      );
    } else {
      message.permissionedAddresses = undefined;
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
