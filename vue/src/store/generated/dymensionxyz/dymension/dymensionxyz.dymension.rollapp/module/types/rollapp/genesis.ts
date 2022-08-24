/* eslint-disable */
import { Params } from "../rollapp/params";
import { Rollapp } from "../rollapp/rollapp";
import { StateInfo } from "../rollapp/state_info";
import { Writer, Reader } from "protobufjs/minimal";

export const protobufPackage = "dymensionxyz.dymension.rollapp";

/** GenesisState defines the rollapp module's genesis state. */
export interface GenesisState {
  params: Params | undefined;
  rollappList: Rollapp[];
  /** this line is used by starport scaffolding # genesis/proto/state */
  stateInfoList: StateInfo[];
}

const baseGenesisState: object = {};

export const GenesisState = {
  encode(message: GenesisState, writer: Writer = Writer.create()): Writer {
    if (message.params !== undefined) {
      Params.encode(message.params, writer.uint32(10).fork()).ldelim();
    }
    for (const v of message.rollappList) {
      Rollapp.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    for (const v of message.stateInfoList) {
      StateInfo.encode(v!, writer.uint32(26).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): GenesisState {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseGenesisState } as GenesisState;
    message.rollappList = [];
    message.stateInfoList = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.params = Params.decode(reader, reader.uint32());
          break;
        case 2:
          message.rollappList.push(Rollapp.decode(reader, reader.uint32()));
          break;
        case 3:
          message.stateInfoList.push(StateInfo.decode(reader, reader.uint32()));
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
    message.rollappList = [];
    message.stateInfoList = [];
    if (object.params !== undefined && object.params !== null) {
      message.params = Params.fromJSON(object.params);
    } else {
      message.params = undefined;
    }
    if (object.rollappList !== undefined && object.rollappList !== null) {
      for (const e of object.rollappList) {
        message.rollappList.push(Rollapp.fromJSON(e));
      }
    }
    if (object.stateInfoList !== undefined && object.stateInfoList !== null) {
      for (const e of object.stateInfoList) {
        message.stateInfoList.push(StateInfo.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: GenesisState): unknown {
    const obj: any = {};
    message.params !== undefined &&
      (obj.params = message.params ? Params.toJSON(message.params) : undefined);
    if (message.rollappList) {
      obj.rollappList = message.rollappList.map((e) =>
        e ? Rollapp.toJSON(e) : undefined
      );
    } else {
      obj.rollappList = [];
    }
    if (message.stateInfoList) {
      obj.stateInfoList = message.stateInfoList.map((e) =>
        e ? StateInfo.toJSON(e) : undefined
      );
    } else {
      obj.stateInfoList = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<GenesisState>): GenesisState {
    const message = { ...baseGenesisState } as GenesisState;
    message.rollappList = [];
    message.stateInfoList = [];
    if (object.params !== undefined && object.params !== null) {
      message.params = Params.fromPartial(object.params);
    } else {
      message.params = undefined;
    }
    if (object.rollappList !== undefined && object.rollappList !== null) {
      for (const e of object.rollappList) {
        message.rollappList.push(Rollapp.fromPartial(e));
      }
    }
    if (object.stateInfoList !== undefined && object.stateInfoList !== null) {
      for (const e of object.stateInfoList) {
        message.stateInfoList.push(StateInfo.fromPartial(e));
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
