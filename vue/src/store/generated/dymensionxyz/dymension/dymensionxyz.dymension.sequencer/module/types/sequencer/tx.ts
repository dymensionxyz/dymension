/* eslint-disable */
import { Reader, Writer } from "protobufjs/minimal";
import { Description } from "../sequencer/description";

export const protobufPackage = "dymensionxyz.dymension.sequencer";

export interface MsgCreateSequencer {
  creator: string;
  sequencerAddress: string;
  pubkey: string;
  rollappId: string;
  description: Description | undefined;
}

export interface MsgCreateSequencerResponse {}

const baseMsgCreateSequencer: object = {
  creator: "",
  sequencerAddress: "",
  pubkey: "",
  rollappId: "",
};

export const MsgCreateSequencer = {
  encode(
    message: MsgCreateSequencer,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.creator !== "") {
      writer.uint32(10).string(message.creator);
    }
    if (message.sequencerAddress !== "") {
      writer.uint32(18).string(message.sequencerAddress);
    }
    if (message.pubkey !== "") {
      writer.uint32(26).string(message.pubkey);
    }
    if (message.rollappId !== "") {
      writer.uint32(34).string(message.rollappId);
    }
    if (message.description !== undefined) {
      Description.encode(
        message.description,
        writer.uint32(42).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgCreateSequencer {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgCreateSequencer } as MsgCreateSequencer;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.creator = reader.string();
          break;
        case 2:
          message.sequencerAddress = reader.string();
          break;
        case 3:
          message.pubkey = reader.string();
          break;
        case 4:
          message.rollappId = reader.string();
          break;
        case 5:
          message.description = Description.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgCreateSequencer {
    const message = { ...baseMsgCreateSequencer } as MsgCreateSequencer;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (
      object.sequencerAddress !== undefined &&
      object.sequencerAddress !== null
    ) {
      message.sequencerAddress = String(object.sequencerAddress);
    } else {
      message.sequencerAddress = "";
    }
    if (object.pubkey !== undefined && object.pubkey !== null) {
      message.pubkey = String(object.pubkey);
    } else {
      message.pubkey = "";
    }
    if (object.rollappId !== undefined && object.rollappId !== null) {
      message.rollappId = String(object.rollappId);
    } else {
      message.rollappId = "";
    }
    if (object.description !== undefined && object.description !== null) {
      message.description = Description.fromJSON(object.description);
    } else {
      message.description = undefined;
    }
    return message;
  },

  toJSON(message: MsgCreateSequencer): unknown {
    const obj: any = {};
    message.creator !== undefined && (obj.creator = message.creator);
    message.sequencerAddress !== undefined &&
      (obj.sequencerAddress = message.sequencerAddress);
    message.pubkey !== undefined && (obj.pubkey = message.pubkey);
    message.rollappId !== undefined && (obj.rollappId = message.rollappId);
    message.description !== undefined &&
      (obj.description = message.description
        ? Description.toJSON(message.description)
        : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgCreateSequencer>): MsgCreateSequencer {
    const message = { ...baseMsgCreateSequencer } as MsgCreateSequencer;
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (
      object.sequencerAddress !== undefined &&
      object.sequencerAddress !== null
    ) {
      message.sequencerAddress = object.sequencerAddress;
    } else {
      message.sequencerAddress = "";
    }
    if (object.pubkey !== undefined && object.pubkey !== null) {
      message.pubkey = object.pubkey;
    } else {
      message.pubkey = "";
    }
    if (object.rollappId !== undefined && object.rollappId !== null) {
      message.rollappId = object.rollappId;
    } else {
      message.rollappId = "";
    }
    if (object.description !== undefined && object.description !== null) {
      message.description = Description.fromPartial(object.description);
    } else {
      message.description = undefined;
    }
    return message;
  },
};

const baseMsgCreateSequencerResponse: object = {};

export const MsgCreateSequencerResponse = {
  encode(
    _: MsgCreateSequencerResponse,
    writer: Writer = Writer.create()
  ): Writer {
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): MsgCreateSequencerResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgCreateSequencerResponse,
    } as MsgCreateSequencerResponse;
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

  fromJSON(_: any): MsgCreateSequencerResponse {
    const message = {
      ...baseMsgCreateSequencerResponse,
    } as MsgCreateSequencerResponse;
    return message;
  },

  toJSON(_: MsgCreateSequencerResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<MsgCreateSequencerResponse>
  ): MsgCreateSequencerResponse {
    const message = {
      ...baseMsgCreateSequencerResponse,
    } as MsgCreateSequencerResponse;
    return message;
  },
};

/** Msg defines the Msg service. */
export interface Msg {
  /** this line is used by starport scaffolding # proto/tx/rpc */
  CreateSequencer(
    request: MsgCreateSequencer
  ): Promise<MsgCreateSequencerResponse>;
}

export class MsgClientImpl implements Msg {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  CreateSequencer(
    request: MsgCreateSequencer
  ): Promise<MsgCreateSequencerResponse> {
    const data = MsgCreateSequencer.encode(request).finish();
    const promise = this.rpc.request(
      "dymensionxyz.dymension.sequencer.Msg",
      "CreateSequencer",
      data
    );
    return promise.then((data) =>
      MsgCreateSequencerResponse.decode(new Reader(data))
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
