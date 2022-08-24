/* eslint-disable */
import { Reader, util, configure, Writer } from "protobufjs/minimal";
import * as Long from "long";
import { Params } from "../rollapp/params";
import { Rollapp } from "../rollapp/rollapp";
import {
  PageRequest,
  PageResponse,
} from "../cosmos/base/query/v1beta1/pagination";
import { StateInfo } from "../rollapp/state_info";

export const protobufPackage = "dymensionxyz.dymension.rollapp";

/** QueryParamsRequest is request type for the Query/Params RPC method. */
export interface QueryParamsRequest {}

/** QueryParamsResponse is response type for the Query/Params RPC method. */
export interface QueryParamsResponse {
  /** params holds all the parameters of this module. */
  params: Params | undefined;
}

export interface QueryGetRollappRequest {
  rollappId: string;
}

export interface QueryGetRollappResponse {
  rollapp: Rollapp | undefined;
}

export interface QueryAllRollappRequest {
  pagination: PageRequest | undefined;
}

export interface QueryAllRollappResponse {
  rollapp: Rollapp[];
  pagination: PageResponse | undefined;
}

export interface QueryGetStateInfoRequest {
  rollappId: string;
  stateIndex: number;
}

export interface QueryGetStateInfoResponse {
  stateInfo: StateInfo | undefined;
}

export interface QueryAllStateInfoRequest {
  pagination: PageRequest | undefined;
}

export interface QueryAllStateInfoResponse {
  stateInfo: StateInfo[];
  pagination: PageResponse | undefined;
}

const baseQueryParamsRequest: object = {};

export const QueryParamsRequest = {
  encode(_: QueryParamsRequest, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryParamsRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryParamsRequest } as QueryParamsRequest;
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

  fromJSON(_: any): QueryParamsRequest {
    const message = { ...baseQueryParamsRequest } as QueryParamsRequest;
    return message;
  },

  toJSON(_: QueryParamsRequest): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<QueryParamsRequest>): QueryParamsRequest {
    const message = { ...baseQueryParamsRequest } as QueryParamsRequest;
    return message;
  },
};

const baseQueryParamsResponse: object = {};

export const QueryParamsResponse = {
  encode(
    message: QueryParamsResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.params !== undefined) {
      Params.encode(message.params, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryParamsResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryParamsResponse } as QueryParamsResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.params = Params.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryParamsResponse {
    const message = { ...baseQueryParamsResponse } as QueryParamsResponse;
    if (object.params !== undefined && object.params !== null) {
      message.params = Params.fromJSON(object.params);
    } else {
      message.params = undefined;
    }
    return message;
  },

  toJSON(message: QueryParamsResponse): unknown {
    const obj: any = {};
    message.params !== undefined &&
      (obj.params = message.params ? Params.toJSON(message.params) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<QueryParamsResponse>): QueryParamsResponse {
    const message = { ...baseQueryParamsResponse } as QueryParamsResponse;
    if (object.params !== undefined && object.params !== null) {
      message.params = Params.fromPartial(object.params);
    } else {
      message.params = undefined;
    }
    return message;
  },
};

const baseQueryGetRollappRequest: object = { rollappId: "" };

export const QueryGetRollappRequest = {
  encode(
    message: QueryGetRollappRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.rollappId !== "") {
      writer.uint32(10).string(message.rollappId);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryGetRollappRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryGetRollappRequest } as QueryGetRollappRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.rollappId = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetRollappRequest {
    const message = { ...baseQueryGetRollappRequest } as QueryGetRollappRequest;
    if (object.rollappId !== undefined && object.rollappId !== null) {
      message.rollappId = String(object.rollappId);
    } else {
      message.rollappId = "";
    }
    return message;
  },

  toJSON(message: QueryGetRollappRequest): unknown {
    const obj: any = {};
    message.rollappId !== undefined && (obj.rollappId = message.rollappId);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetRollappRequest>
  ): QueryGetRollappRequest {
    const message = { ...baseQueryGetRollappRequest } as QueryGetRollappRequest;
    if (object.rollappId !== undefined && object.rollappId !== null) {
      message.rollappId = object.rollappId;
    } else {
      message.rollappId = "";
    }
    return message;
  },
};

const baseQueryGetRollappResponse: object = {};

export const QueryGetRollappResponse = {
  encode(
    message: QueryGetRollappResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.rollapp !== undefined) {
      Rollapp.encode(message.rollapp, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryGetRollappResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetRollappResponse,
    } as QueryGetRollappResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.rollapp = Rollapp.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetRollappResponse {
    const message = {
      ...baseQueryGetRollappResponse,
    } as QueryGetRollappResponse;
    if (object.rollapp !== undefined && object.rollapp !== null) {
      message.rollapp = Rollapp.fromJSON(object.rollapp);
    } else {
      message.rollapp = undefined;
    }
    return message;
  },

  toJSON(message: QueryGetRollappResponse): unknown {
    const obj: any = {};
    message.rollapp !== undefined &&
      (obj.rollapp = message.rollapp
        ? Rollapp.toJSON(message.rollapp)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetRollappResponse>
  ): QueryGetRollappResponse {
    const message = {
      ...baseQueryGetRollappResponse,
    } as QueryGetRollappResponse;
    if (object.rollapp !== undefined && object.rollapp !== null) {
      message.rollapp = Rollapp.fromPartial(object.rollapp);
    } else {
      message.rollapp = undefined;
    }
    return message;
  },
};

const baseQueryAllRollappRequest: object = {};

export const QueryAllRollappRequest = {
  encode(
    message: QueryAllRollappRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.pagination !== undefined) {
      PageRequest.encode(message.pagination, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryAllRollappRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryAllRollappRequest } as QueryAllRollappRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.pagination = PageRequest.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryAllRollappRequest {
    const message = { ...baseQueryAllRollappRequest } as QueryAllRollappRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryAllRollappRequest): unknown {
    const obj: any = {};
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageRequest.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryAllRollappRequest>
  ): QueryAllRollappRequest {
    const message = { ...baseQueryAllRollappRequest } as QueryAllRollappRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

const baseQueryAllRollappResponse: object = {};

export const QueryAllRollappResponse = {
  encode(
    message: QueryAllRollappResponse,
    writer: Writer = Writer.create()
  ): Writer {
    for (const v of message.rollapp) {
      Rollapp.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    if (message.pagination !== undefined) {
      PageResponse.encode(
        message.pagination,
        writer.uint32(18).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryAllRollappResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryAllRollappResponse,
    } as QueryAllRollappResponse;
    message.rollapp = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.rollapp.push(Rollapp.decode(reader, reader.uint32()));
          break;
        case 2:
          message.pagination = PageResponse.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryAllRollappResponse {
    const message = {
      ...baseQueryAllRollappResponse,
    } as QueryAllRollappResponse;
    message.rollapp = [];
    if (object.rollapp !== undefined && object.rollapp !== null) {
      for (const e of object.rollapp) {
        message.rollapp.push(Rollapp.fromJSON(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryAllRollappResponse): unknown {
    const obj: any = {};
    if (message.rollapp) {
      obj.rollapp = message.rollapp.map((e) =>
        e ? Rollapp.toJSON(e) : undefined
      );
    } else {
      obj.rollapp = [];
    }
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageResponse.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryAllRollappResponse>
  ): QueryAllRollappResponse {
    const message = {
      ...baseQueryAllRollappResponse,
    } as QueryAllRollappResponse;
    message.rollapp = [];
    if (object.rollapp !== undefined && object.rollapp !== null) {
      for (const e of object.rollapp) {
        message.rollapp.push(Rollapp.fromPartial(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

const baseQueryGetStateInfoRequest: object = { rollappId: "", stateIndex: 0 };

export const QueryGetStateInfoRequest = {
  encode(
    message: QueryGetStateInfoRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.rollappId !== "") {
      writer.uint32(10).string(message.rollappId);
    }
    if (message.stateIndex !== 0) {
      writer.uint32(16).uint64(message.stateIndex);
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetStateInfoRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetStateInfoRequest,
    } as QueryGetStateInfoRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.rollappId = reader.string();
          break;
        case 2:
          message.stateIndex = longToNumber(reader.uint64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetStateInfoRequest {
    const message = {
      ...baseQueryGetStateInfoRequest,
    } as QueryGetStateInfoRequest;
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
    return message;
  },

  toJSON(message: QueryGetStateInfoRequest): unknown {
    const obj: any = {};
    message.rollappId !== undefined && (obj.rollappId = message.rollappId);
    message.stateIndex !== undefined && (obj.stateIndex = message.stateIndex);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetStateInfoRequest>
  ): QueryGetStateInfoRequest {
    const message = {
      ...baseQueryGetStateInfoRequest,
    } as QueryGetStateInfoRequest;
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
    return message;
  },
};

const baseQueryGetStateInfoResponse: object = {};

export const QueryGetStateInfoResponse = {
  encode(
    message: QueryGetStateInfoResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.stateInfo !== undefined) {
      StateInfo.encode(message.stateInfo, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetStateInfoResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetStateInfoResponse,
    } as QueryGetStateInfoResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.stateInfo = StateInfo.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetStateInfoResponse {
    const message = {
      ...baseQueryGetStateInfoResponse,
    } as QueryGetStateInfoResponse;
    if (object.stateInfo !== undefined && object.stateInfo !== null) {
      message.stateInfo = StateInfo.fromJSON(object.stateInfo);
    } else {
      message.stateInfo = undefined;
    }
    return message;
  },

  toJSON(message: QueryGetStateInfoResponse): unknown {
    const obj: any = {};
    message.stateInfo !== undefined &&
      (obj.stateInfo = message.stateInfo
        ? StateInfo.toJSON(message.stateInfo)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetStateInfoResponse>
  ): QueryGetStateInfoResponse {
    const message = {
      ...baseQueryGetStateInfoResponse,
    } as QueryGetStateInfoResponse;
    if (object.stateInfo !== undefined && object.stateInfo !== null) {
      message.stateInfo = StateInfo.fromPartial(object.stateInfo);
    } else {
      message.stateInfo = undefined;
    }
    return message;
  },
};

const baseQueryAllStateInfoRequest: object = {};

export const QueryAllStateInfoRequest = {
  encode(
    message: QueryAllStateInfoRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.pagination !== undefined) {
      PageRequest.encode(message.pagination, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryAllStateInfoRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryAllStateInfoRequest,
    } as QueryAllStateInfoRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.pagination = PageRequest.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryAllStateInfoRequest {
    const message = {
      ...baseQueryAllStateInfoRequest,
    } as QueryAllStateInfoRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryAllStateInfoRequest): unknown {
    const obj: any = {};
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageRequest.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryAllStateInfoRequest>
  ): QueryAllStateInfoRequest {
    const message = {
      ...baseQueryAllStateInfoRequest,
    } as QueryAllStateInfoRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

const baseQueryAllStateInfoResponse: object = {};

export const QueryAllStateInfoResponse = {
  encode(
    message: QueryAllStateInfoResponse,
    writer: Writer = Writer.create()
  ): Writer {
    for (const v of message.stateInfo) {
      StateInfo.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    if (message.pagination !== undefined) {
      PageResponse.encode(
        message.pagination,
        writer.uint32(18).fork()
      ).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryAllStateInfoResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryAllStateInfoResponse,
    } as QueryAllStateInfoResponse;
    message.stateInfo = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.stateInfo.push(StateInfo.decode(reader, reader.uint32()));
          break;
        case 2:
          message.pagination = PageResponse.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryAllStateInfoResponse {
    const message = {
      ...baseQueryAllStateInfoResponse,
    } as QueryAllStateInfoResponse;
    message.stateInfo = [];
    if (object.stateInfo !== undefined && object.stateInfo !== null) {
      for (const e of object.stateInfo) {
        message.stateInfo.push(StateInfo.fromJSON(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryAllStateInfoResponse): unknown {
    const obj: any = {};
    if (message.stateInfo) {
      obj.stateInfo = message.stateInfo.map((e) =>
        e ? StateInfo.toJSON(e) : undefined
      );
    } else {
      obj.stateInfo = [];
    }
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageResponse.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryAllStateInfoResponse>
  ): QueryAllStateInfoResponse {
    const message = {
      ...baseQueryAllStateInfoResponse,
    } as QueryAllStateInfoResponse;
    message.stateInfo = [];
    if (object.stateInfo !== undefined && object.stateInfo !== null) {
      for (const e of object.stateInfo) {
        message.stateInfo.push(StateInfo.fromPartial(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

/** Query defines the gRPC querier service. */
export interface Query {
  /** Parameters queries the parameters of the module. */
  Params(request: QueryParamsRequest): Promise<QueryParamsResponse>;
  /** Queries a Rollapp by index. */
  Rollapp(request: QueryGetRollappRequest): Promise<QueryGetRollappResponse>;
  /** Queries a list of Rollapp items. */
  RollappAll(request: QueryAllRollappRequest): Promise<QueryAllRollappResponse>;
  /** Queries a StateInfo by index. */
  StateInfo(
    request: QueryGetStateInfoRequest
  ): Promise<QueryGetStateInfoResponse>;
  /** Queries a list of StateInfo items. */
  StateInfoAll(
    request: QueryAllStateInfoRequest
  ): Promise<QueryAllStateInfoResponse>;
}

export class QueryClientImpl implements Query {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  Params(request: QueryParamsRequest): Promise<QueryParamsResponse> {
    const data = QueryParamsRequest.encode(request).finish();
    const promise = this.rpc.request(
      "dymensionxyz.dymension.rollapp.Query",
      "Params",
      data
    );
    return promise.then((data) => QueryParamsResponse.decode(new Reader(data)));
  }

  Rollapp(request: QueryGetRollappRequest): Promise<QueryGetRollappResponse> {
    const data = QueryGetRollappRequest.encode(request).finish();
    const promise = this.rpc.request(
      "dymensionxyz.dymension.rollapp.Query",
      "Rollapp",
      data
    );
    return promise.then((data) =>
      QueryGetRollappResponse.decode(new Reader(data))
    );
  }

  RollappAll(
    request: QueryAllRollappRequest
  ): Promise<QueryAllRollappResponse> {
    const data = QueryAllRollappRequest.encode(request).finish();
    const promise = this.rpc.request(
      "dymensionxyz.dymension.rollapp.Query",
      "RollappAll",
      data
    );
    return promise.then((data) =>
      QueryAllRollappResponse.decode(new Reader(data))
    );
  }

  StateInfo(
    request: QueryGetStateInfoRequest
  ): Promise<QueryGetStateInfoResponse> {
    const data = QueryGetStateInfoRequest.encode(request).finish();
    const promise = this.rpc.request(
      "dymensionxyz.dymension.rollapp.Query",
      "StateInfo",
      data
    );
    return promise.then((data) =>
      QueryGetStateInfoResponse.decode(new Reader(data))
    );
  }

  StateInfoAll(
    request: QueryAllStateInfoRequest
  ): Promise<QueryAllStateInfoResponse> {
    const data = QueryAllStateInfoRequest.encode(request).finish();
    const promise = this.rpc.request(
      "dymensionxyz.dymension.rollapp.Query",
      "StateInfoAll",
      data
    );
    return promise.then((data) =>
      QueryAllStateInfoResponse.decode(new Reader(data))
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
