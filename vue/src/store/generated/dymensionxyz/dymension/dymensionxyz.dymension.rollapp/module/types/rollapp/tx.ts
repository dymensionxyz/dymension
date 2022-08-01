/* eslint-disable */
import { Reader, util, configure, Writer } from "protobufjs/minimal";
import * as Long from "long";
import { Sequencers } from "../sequencer/sequencers";

export const protobufPackage = "dymensionxyz.dymension.rollapp";

export interface MsgCreateRollapp {
  /** creator is the bech32-encoded address of the rollapp creator. */
  creator: string;
  /**
   * rollappId is the unique identifier of the rollapp chain.
   * The rollappId follows the same standard as cosmos chain_id.
   */
  rollappId: string;
  /** genesisPath is the description of the genesis file location on the DA. */
  codeStamp: string;
  /** genesisPath is the description of the genesis file location on the DA. */
  genesisPath: string;
  /**
   * maxWithholdingBlocks is the maximum number of blocks for
   * an active sequencer to send a state update (MsgUpdateState).
   */
  maxWithholdingBlocks: number;
  /** maxSequencers is the maximum number of sequencers. */
  maxSequencers: number;
  /**
   * permissionedAddresses is a bech32-encoded address list of the
   * sequencers that are allowed to serve this rollappId.
   * In the case of an empty list, the rollapp is considered permissionless.
   */
  permissionedAddresses: Sequencers | undefined;
}

export interface MsgCreateRollappResponse {}

const baseMsgCreateRollapp: object = {
  creator: "",
  rollappId: "",
  codeStamp: "",
  genesisPath: "",
  maxWithholdingBlocks: 0,
  maxSequencers: 0,
};

export const MsgCreateRollapp = {
  encode(message: MsgCreateRollapp, writer: Writer = Writer.create()): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.rollappId !== "") {
      writer.uint32(18).string(message.rollappId);
    }
    if (message.codeStamp !== "") {
      writer.uint32(26).string(message.codeStamp);
    }
    if (message.genesisPath !== "") {
      writer.uint32(34).string(message.genesisPath);
    }
    if (message.maxWithholdingBlocks !== 0) {
      writer.uint32(40).uint64(message.maxWithholdingBlocks);
    }
    if (message.maxSequencers !== 0) {
      writer.uint32(48).uint64(message.maxSequencers);
    }
    if (message.permissionedAddresses !== undefined) {
      Sequencers.encode(
        message.permissionedAddresses,
        writer.uint32(58).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgCreateRollapp {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgCreateRollapp } as MsgCreateRollapp;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.rollappId = reader.string();
          break;
        case 3:
          message.codeStamp = reader.string();
          break;
        case 4:
          message.genesisPath = reader.string();
          break;
        case 5:
          message.maxWithholdingBlocks = longToNumber(reader.uint64() as Long);
          break;
        case 6:
          message.maxSequencers = longToNumber(reader.uint64() as Long);
          break;
        case 7:
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

  fromJSON(object: any): MsgCreateRollapp {
    const message = { ...baseMsgCreateRollapp } as MsgCreateRollapp;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.rollappId !== undefined && object.rollappId !== null) {
      message.rollappId = String(object.rollappId);
    } else {
      message.rollappId = "";
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

  toJSON(message: MsgCreateRollapp): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.rollappId !== undefined && (obj.rollappId = message.rollappId);
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

  fromPartial(object: DeepPartial<MsgCreateRollapp>): MsgCreateRollapp {
    const message = { ...baseMsgCreateRollapp } as MsgCreateRollapp;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.rollappId !== undefined && object.rollappId !== null) {
      message.rollappId = object.rollappId;
    } else {
      message.rollappId = "";
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

const baseMsgCreateRollappResponse: object = {};

export const MsgCreateRollappResponse = {
  encode(
    _: MsgCreateRollappResponse,
    writer: Writer = Writer.create()
  ): Writer {
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): MsgCreateRollappResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgCreateRollappResponse,
    } as MsgCreateRollappResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): MsgCreateRollappResponse {
    const message = {
      ...baseMsgCreateRollappResponse,
    } as MsgCreateRollappResponse;
    return message;
  },

  toJSON(_: MsgCreateRollappResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<MsgCreateRollappResponse>
  ): MsgCreateRollappResponse {
    const message = {
      ...baseMsgCreateRollappResponse,
    } as MsgCreateRollappResponse;
    return message;
  },
};

/** Msg defines the Msg service. */
export interface Msg {
  /** this line is used by starport scaffolding # proto/tx/rpc */
  CreateRollapp(request: MsgCreateRollapp): Promise<MsgCreateRollappResponse>;
}

export class MsgClientImpl implements Msg {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  CreateRollapp(request: MsgCreateRollapp): Promise<MsgCreateRollappResponse> {
    const data = MsgCreateRollapp.encode(request).finish();
    const promise = this.rpc.request(
      "dymensionxyz.dymension.rollapp.Msg",
      "CreateRollapp",
      data
    );
    return promise.then((data) =>
      MsgCreateRollappResponse.decode(new Reader(data))
    );
  }
}

interface Rpc {
  request(
    service: string,
    method: string,
    data: Uint8Array
  ): Promise<Uint8Array>;
}

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
