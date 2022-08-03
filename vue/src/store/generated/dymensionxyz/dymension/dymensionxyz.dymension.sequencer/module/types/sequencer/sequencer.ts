/* eslint-disable */
import { Any } from "../google/protobuf/any";
import { Description } from "../sequencer/description";
import { Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "dymensionxyz.dymension.sequencer";

/**
 * Sequencer defines a sequencer identified by its' address (sequencerAddress).
 * The sequencer could be attached to only one rollapp (rollappId).
 */
export interface Sequencer {
  /** sequencerAddress is the bech32-encoded address of the sequencer account. */
  sequencerAddress: string;
  /** creator is the bech32-encoded address of the account sent the transaction (sequencer creator) */
  creator: string;
  /** pubkey is the public key of the sequencer, as a Protobuf Any. */
  pubkey: Any | undefined;
  /** rollappId defines the rollapp to which the sequencer belongs. */
  rollappId: string;
  /** description defines the descriptive terms for the sequencer. */
  description: Description | undefined;
}

const baseSequencer: object = {
  sequencerAddress: "",
  creator: "",
  rollappId: "",
};

export const Sequencer = {
  encode(message: Sequencer, writer: Writer = Writer.create()): Writer {
    if (message.sequencerAddress !== "") {
      writer.uint32(10).string(message.sequencerAddress);
    }
    if (message.creator !== "") {
      writer.uint32(18).string(message.creator);
    }
    if (message.pubkey !== undefined) {
      Any.encode(message.pubkey, writer.uint32(26).fork()).ldelim();
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

  decode(input: Reader | Uint8Array, length?: number): Sequencer {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseSequencer } as Sequencer;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.sequencerAddress = reader.string();
          break;
        case 2:
          message.creator = reader.string();
          break;
        case 3:
          message.pubkey = Any.decode(reader, reader.uint32());
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

  fromJSON(object: any): Sequencer {
    const message = { ...baseSequencer } as Sequencer;
    if (
      object.sequencerAddress !== undefined &&
      object.sequencerAddress !== null
    ) {
      message.sequencerAddress = String(object.sequencerAddress);
    } else {
      message.sequencerAddress = "";
    }
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = String(object.creator);
    } else {
      message.creator = "";
    }
    if (object.pubkey !== undefined && object.pubkey !== null) {
      message.pubkey = Any.fromJSON(object.pubkey);
    } else {
      message.pubkey = undefined;
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

  toJSON(message: Sequencer): unknown {
    const obj: any = {};
    message.sequencerAddress !== undefined &&
      (obj.sequencerAddress = message.sequencerAddress);
    message.creator !== undefined && (obj.creator = message.creator);
    message.pubkey !== undefined &&
      (obj.pubkey = message.pubkey ? Any.toJSON(message.pubkey) : undefined);
    message.rollappId !== undefined && (obj.rollappId = message.rollappId);
    message.description !== undefined &&
      (obj.description = message.description
        ? Description.toJSON(message.description)
        : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<Sequencer>): Sequencer {
    const message = { ...baseSequencer } as Sequencer;
    if (
      object.sequencerAddress !== undefined &&
      object.sequencerAddress !== null
    ) {
      message.sequencerAddress = object.sequencerAddress;
    } else {
      message.sequencerAddress = "";
    }
    if (object.creator !== undefined && object.creator !== null) {
      message.creator = object.creator;
    } else {
      message.creator = "";
    }
    if (object.pubkey !== undefined && object.pubkey !== null) {
      message.pubkey = Any.fromPartial(object.pubkey);
    } else {
      message.pubkey = undefined;
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
