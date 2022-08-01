/* eslint-disable */
import { Reader, Writer } from "protobufjs/minimal";
import { Params } from "../sequencer/params";
import { Sequencer } from "../sequencer/sequencer";
import {
  PageRequest,
  PageResponse,
} from "../cosmos/base/query/v1beta1/pagination";

export const protobufPackage = "dymensionxyz.dymension.sequencer";

/** QueryParamsRequest is request type for the Query/Params RPC method. */
export interface QueryParamsRequest {}

/** QueryParamsResponse is response type for the Query/Params RPC method. */
export interface QueryParamsResponse {
  /** params holds all the parameters of this module. */
  params: Params | undefined;
}

export interface QueryGetSequencerRequest {
  sequencerAddress: string;
}

export interface QueryGetSequencerResponse {
  sequencer: Sequencer | undefined;
}

export interface QueryAllSequencerRequest {
  pagination: PageRequest | undefined;
}

export interface QueryAllSequencerResponse {
  sequencer: Sequencer[];
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

const baseQueryGetSequencerRequest: object = { sequencerAddress: "" };

export const QueryGetSequencerRequest = {
  encode(
    message: QueryGetSequencerRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.sequencerAddress !== "") {
      writer.uint32(10).string(message.sequencerAddress);
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetSequencerRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetSequencerRequest,
    } as QueryGetSequencerRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.sequencerAddress = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetSequencerRequest {
    const message = {
      ...baseQueryGetSequencerRequest,
    } as QueryGetSequencerRequest;
    if (
      object.sequencerAddress !== undefined &&
      object.sequencerAddress !== null
    ) {
      message.sequencerAddress = String(object.sequencerAddress);
    } else {
      message.sequencerAddress = "";
    }
    return message;
  },

  toJSON(message: QueryGetSequencerRequest): unknown {
    const obj: any = {};
    message.sequencerAddress !== undefined &&
      (obj.sequencerAddress = message.sequencerAddress);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetSequencerRequest>
  ): QueryGetSequencerRequest {
    const message = {
      ...baseQueryGetSequencerRequest,
    } as QueryGetSequencerRequest;
    if (
      object.sequencerAddress !== undefined &&
      object.sequencerAddress !== null
    ) {
      message.sequencerAddress = object.sequencerAddress;
    } else {
      message.sequencerAddress = "";
    }
    return message;
  },
};

const baseQueryGetSequencerResponse: object = {};

export const QueryGetSequencerResponse = {
  encode(
    message: QueryGetSequencerResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.sequencer !== undefined) {
      Sequencer.encode(message.sequencer, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryGetSequencerResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryGetSequencerResponse,
    } as QueryGetSequencerResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.sequencer = Sequencer.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryGetSequencerResponse {
    const message = {
      ...baseQueryGetSequencerResponse,
    } as QueryGetSequencerResponse;
    if (object.sequencer !== undefined && object.sequencer !== null) {
      message.sequencer = Sequencer.fromJSON(object.sequencer);
    } else {
      message.sequencer = undefined;
    }
    return message;
  },

  toJSON(message: QueryGetSequencerResponse): unknown {
    const obj: any = {};
    message.sequencer !== undefined &&
      (obj.sequencer = message.sequencer
        ? Sequencer.toJSON(message.sequencer)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryGetSequencerResponse>
  ): QueryGetSequencerResponse {
    const message = {
      ...baseQueryGetSequencerResponse,
    } as QueryGetSequencerResponse;
    if (object.sequencer !== undefined && object.sequencer !== null) {
      message.sequencer = Sequencer.fromPartial(object.sequencer);
    } else {
      message.sequencer = undefined;
    }
    return message;
  },
};

const baseQueryAllSequencerRequest: object = {};

export const QueryAllSequencerRequest = {
  encode(
    message: QueryAllSequencerRequest,
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
  ): QueryAllSequencerRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryAllSequencerRequest,
    } as QueryAllSequencerRequest;
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

  fromJSON(object: any): QueryAllSequencerRequest {
    const message = {
      ...baseQueryAllSequencerRequest,
    } as QueryAllSequencerRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryAllSequencerRequest): unknown {
    const obj: any = {};
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageRequest.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryAllSequencerRequest>
  ): QueryAllSequencerRequest {
    const message = {
      ...baseQueryAllSequencerRequest,
    } as QueryAllSequencerRequest;
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageRequest.fromPartial(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },
};

const baseQueryAllSequencerResponse: object = {};

export const QueryAllSequencerResponse = {
  encode(
    message: QueryAllSequencerResponse,
    writer: Writer = Writer.create()
  ): Writer {
    for (const v of message.sequencer) {
      Sequencer.encode(v!, writer.uint32(10).fork()).ldelim();
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
  ): QueryAllSequencerResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryAllSequencerResponse,
    } as QueryAllSequencerResponse;
    message.sequencer = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.sequencer.push(Sequencer.decode(reader, reader.uint32()));
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

  fromJSON(object: any): QueryAllSequencerResponse {
    const message = {
      ...baseQueryAllSequencerResponse,
    } as QueryAllSequencerResponse;
    message.sequencer = [];
    if (object.sequencer !== undefined && object.sequencer !== null) {
      for (const e of object.sequencer) {
        message.sequencer.push(Sequencer.fromJSON(e));
      }
    }
    if (object.pagination !== undefined && object.pagination !== null) {
      message.pagination = PageResponse.fromJSON(object.pagination);
    } else {
      message.pagination = undefined;
    }
    return message;
  },

  toJSON(message: QueryAllSequencerResponse): unknown {
    const obj: any = {};
    if (message.sequencer) {
      obj.sequencer = message.sequencer.map((e) =>
        e ? Sequencer.toJSON(e) : undefined
      );
    } else {
      obj.sequencer = [];
    }
    message.pagination !== undefined &&
      (obj.pagination = message.pagination
        ? PageResponse.toJSON(message.pagination)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryAllSequencerResponse>
  ): QueryAllSequencerResponse {
    const message = {
      ...baseQueryAllSequencerResponse,
    } as QueryAllSequencerResponse;
    message.sequencer = [];
    if (object.sequencer !== undefined && object.sequencer !== null) {
      for (const e of object.sequencer) {
        message.sequencer.push(Sequencer.fromPartial(e));
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
  /** Queries a Sequencer by index. */
  Sequencer(
    request: QueryGetSequencerRequest
  ): Promise<QueryGetSequencerResponse>;
  /** Queries a list of Sequencer items. */
  SequencerAll(
    request: QueryAllSequencerRequest
  ): Promise<QueryAllSequencerResponse>;
}

export class QueryClientImpl implements Query {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  Params(request: QueryParamsRequest): Promise<QueryParamsResponse> {
    const data = QueryParamsRequest.encode(request).finish();
    const promise = this.rpc.request(
      "dymensionxyz.dymension.sequencer.Query",
      "Params",
      data
    );
    return promise.then((data) => QueryParamsResponse.decode(new Reader(data)));
  }

  Sequencer(
    request: QueryGetSequencerRequest
  ): Promise<QueryGetSequencerResponse> {
    const data = QueryGetSequencerRequest.encode(request).finish();
    const promise = this.rpc.request(
      "dymensionxyz.dymension.sequencer.Query",
      "Sequencer",
      data
    );
    return promise.then((data) =>
      QueryGetSequencerResponse.decode(new Reader(data))
    );
  }

  SequencerAll(
    request: QueryAllSequencerRequest
  ): Promise<QueryAllSequencerResponse> {
    const data = QueryAllSequencerRequest.encode(request).finish();
    const promise = this.rpc.request(
      "dymensionxyz.dymension.sequencer.Query",
      "SequencerAll",
      data
    );
    return promise.then((data) =>
      QueryAllSequencerResponse.decode(new Reader(data))
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
