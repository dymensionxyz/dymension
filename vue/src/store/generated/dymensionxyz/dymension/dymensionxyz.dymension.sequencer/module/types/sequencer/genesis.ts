/* eslint-disable */
import { Params } from "../sequencer/params";
import { Sequencer } from "../sequencer/sequencer";
import { SequencersByRollapp } from "../sequencer/sequencers_by_rollapp";
import { Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "dymensionxyz.dymension.sequencer";

/** GenesisState defines the sequencer module's genesis state. */
export interface GenesisState {
  params: Params | undefined;
  sequencerList: Sequencer[];
  /** this line is used by starport scaffolding # genesis/proto/state */
  sequencersByRollappList: SequencersByRollapp[];
}

const baseGenesisState: object = {};

export const GenesisState = {
  encode(message: GenesisState, writer: Writer = Writer.create()): Writer {
    if (message.params !== undefined) {
      Params.encode(message.params, writer.uint32(10).fork()).ldelim();
    }
    for (const v of message.sequencerList) {
      Sequencer.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    for (const v of message.sequencersByRollappList) {
      SequencersByRollapp.encode(v!, writer.uint32(26).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): GenesisState {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseGenesisState } as GenesisState;
    message.sequencerList = [];
    message.sequencersByRollappList = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.params = Params.decode(reader, reader.uint32());
          break;
        case 2:
          message.sequencerList.push(Sequencer.decode(reader, reader.uint32()));
          break;
        case 3:
          message.sequencersByRollappList.push(
            SequencersByRollapp.decode(reader, reader.uint32())
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GenesisState {
    const message = { ...baseGenesisState } as GenesisState;
    message.sequencerList = [];
    message.sequencersByRollappList = [];
    if (object.params !== undefined && object.params !== null) {
      message.params = Params.fromJSON(object.params);
    } else {
      message.params = undefined;
    }
    if (object.sequencerList !== undefined && object.sequencerList !== null) {
      for (const e of object.sequencerList) {
        message.sequencerList.push(Sequencer.fromJSON(e));
      }
    }
    if (
      object.sequencersByRollappList !== undefined &&
      object.sequencersByRollappList !== null
    ) {
      for (const e of object.sequencersByRollappList) {
        message.sequencersByRollappList.push(SequencersByRollapp.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: GenesisState): unknown {
    const obj: any = {};
    message.params !== undefined &&
      (obj.params = message.params ? Params.toJSON(message.params) : undefined);
    if (message.sequencerList) {
      obj.sequencerList = message.sequencerList.map((e) =>
        e ? Sequencer.toJSON(e) : undefined
      );
    } else {
      obj.sequencerList = [];
    }
    if (message.sequencersByRollappList) {
      obj.sequencersByRollappList = message.sequencersByRollappList.map((e) =>
        e ? SequencersByRollapp.toJSON(e) : undefined
      );
    } else {
      obj.sequencersByRollappList = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<GenesisState>): GenesisState {
    const message = { ...baseGenesisState } as GenesisState;
    message.sequencerList = [];
    message.sequencersByRollappList = [];
    if (object.params !== undefined && object.params !== null) {
      message.params = Params.fromPartial(object.params);
    } else {
      message.params = undefined;
    }
    if (object.sequencerList !== undefined && object.sequencerList !== null) {
      for (const e of object.sequencerList) {
        message.sequencerList.push(Sequencer.fromPartial(e));
      }
    }
    if (
      object.sequencersByRollappList !== undefined &&
      object.sequencersByRollappList !== null
    ) {
      for (const e of object.sequencersByRollappList) {
        message.sequencersByRollappList.push(
          SequencersByRollapp.fromPartial(e)
        );
      }
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
