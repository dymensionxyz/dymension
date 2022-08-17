/* eslint-disable */
import { Sequencers } from "../shared/sequencers";
import { Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "dymensionxyz.dymension.sequencer";

/**
 * SequencersByRollapp defines an map between rollappId to a list of
 * all sequencers that belongs to it.
 */
export interface SequencersByRollapp {
  /**
   * rollappId is the unique identifier of the rollapp chain.
   * The rollappId follows the same standard as cosmos chain_id.
   */
  rollappId: string;
  /** list of sequencers' account address */
  sequencers: Sequencers | undefined;
}

const baseSequencersByRollapp: object = { rollappId: "" };

export const SequencersByRollapp = {
  encode(
    message: SequencersByRollapp,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.rollappId !== "") {
      writer.uint32(10).string(message.rollappId);
    }
    if (message.sequencers !== undefined) {
      Sequencers.encode(message.sequencers, writer.uint32(18).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): SequencersByRollapp {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseSequencersByRollapp } as SequencersByRollapp;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.rollappId = reader.string();
          break;
        case 2:
          message.sequencers = Sequencers.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): SequencersByRollapp {
    const message = { ...baseSequencersByRollapp } as SequencersByRollapp;
    if (object.rollappId !== undefined && object.rollappId !== null) {
      message.rollappId = String(object.rollappId);
    } else {
      message.rollappId = "";
    }
    if (object.sequencers !== undefined && object.sequencers !== null) {
      message.sequencers = Sequencers.fromJSON(object.sequencers);
    } else {
      message.sequencers = undefined;
    }
    return message;
  },

  toJSON(message: SequencersByRollapp): unknown {
    const obj: any = {};
    message.rollappId !== undefined && (obj.rollappId = message.rollappId);
    message.sequencers !== undefined &&
      (obj.sequencers = message.sequencers
        ? Sequencers.toJSON(message.sequencers)
        : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<SequencersByRollapp>): SequencersByRollapp {
    const message = { ...baseSequencersByRollapp } as SequencersByRollapp;
    if (object.rollappId !== undefined && object.rollappId !== null) {
      message.rollappId = object.rollappId;
    } else {
      message.rollappId = "";
    }
    if (object.sequencers !== undefined && object.sequencers !== null) {
      message.sequencers = Sequencers.fromPartial(object.sequencers);
    } else {
      message.sequencers = undefined;
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
