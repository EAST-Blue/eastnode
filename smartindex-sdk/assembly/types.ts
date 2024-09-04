import { JSON } from "assemblyscript-json/assembly";
import { getResultFromJson } from "./sdk";
export class VinV1 {
  txHash: string;
  index: u32;
  value: u64;

  constructor(txHash: string, index: u32, value: u64) {
    this.txHash = txHash;
    this.index = index;
    this.value = value;
  }

  toJson(): string {
    return (
      '{ "txHash": "' +
      this.txHash +
      '", "index": ' +
      this.index.toString() +
      ', "value": ' +
      this.value.toString() +
      " }"
    );
  }
}

export class VoutV1 {
  txHash: string;
  index: u32;
  address: string;
  pkScript: string;
  value: u64;

  constructor(
    txHash: string,
    index: u32,
    address: string,
    pkScript: string,
    value: u64
  ) {
    this.txHash = txHash;
    this.index = index;
    this.address = address;
    this.pkScript = pkScript;
    this.value = value;
  }

  toJson(): string {
    return (
      '{ "txHash": "' +
      this.txHash +
      '", "index": ' +
      this.index.toString() +
      ', "address": "' +
      this.address +
      '", "pkScript": "' +
      this.pkScript +
      '", "value": ' +
      this.value.toString() +
      " }"
    );
  }
}

export class TransactionV1 {
  hash: string;
  lockTime: u32;
  version: u32;

  vins: VinV1[];
  vouts: VoutV1[];

  constructor(
    hash: string,
    lockTime: u32,
    version: u32,
    vins: VinV1[],
    vouts: VoutV1[]
  ) {
    this.hash = hash;
    this.lockTime = lockTime;
    this.version = version;
    this.vins = vins;
    this.vouts = vouts;
  }

  toJson(): string {
    return (
      '{ "hash": "' +
      this.hash +
      '", "lockTime": ' +
      this.lockTime.toString() +
      ', "version": ' +
      this.version.toString() +
      ', "vins": [' +
      this.vins.map<string>((vin: VinV1) => vin.toJson()).join(",") +
      '], "vouts": [' +
      this.vouts.map<string>((vout: VoutV1) => vout.toJson()).join(",") +
      "] }"
    );
  }
}

export class VinV2 {
  txHash: string;
  txIndex: string;
  blockHash: string;
  blockHeight: string;
  blockTxIndex: string;
  sequence: u64;
  signatureScript: string;
  witness: string;

  fundingTxHash: string;
  fundingTxIndex: string;


  pkScript: string;
  value: u64;
  spender: string;
  type: string;

  p2shAsmScripts: string[];
  pkAsmScripts: string[];
  witnessAsmScripts: string[];

  constructor(
    txHash: string,
    txIndex: string,
    blockHash: string,
    blockHeight: string,
    blockTxIndex: string,
    sequence: u64,
    signatureScript: string,
    witness: string,
    fundingTxHash: string,
    fundingTxIndex: string,
    pkScript: string,
    value: u64,
    spender: string,
    type: string,
    p2shAsmScripts: string[],
    pkAsmScripts: string[],
    witnessAsmScripts: string[]
  ) {
    this.txHash = txHash;
    this.txIndex = txIndex;
    this.blockHash = blockHash;
    this.blockHeight = blockHeight;
    this.blockTxIndex = blockTxIndex;
    this.sequence = sequence;
    this.signatureScript = signatureScript;
    this.witness = witness;
    this.fundingTxHash = fundingTxHash;
    this.fundingTxIndex = fundingTxIndex;
    this.pkScript = pkScript;
    this.value = value;
    this.spender = spender;
    this.type = type;
    this.p2shAsmScripts = p2shAsmScripts;
    this.pkAsmScripts = pkAsmScripts;
    this.witnessAsmScripts = witnessAsmScripts;
  }

  static fromJson(jsonObj: JSON.Obj): VinV2 {
    return new VinV2(
      getResultFromJson(jsonObj, "tx_hash", "string"),
      getResultFromJson(jsonObj, "tx_index", "int64"),
      getResultFromJson(jsonObj, "block_hash", "string"),
      getResultFromJson(jsonObj, "block_height", "string"),
      getResultFromJson(jsonObj, "block_tx_index", "int64"),
      u64(parseInt(getResultFromJson(jsonObj, "sequence", "int64"))),
      getResultFromJson(jsonObj, "signature_script", "string"),
      getResultFromJson(jsonObj, "witness", "string"),
      getResultFromJson(jsonObj, "funding_tx_hash", "string"),
      getResultFromJson(jsonObj, "funding_tx_index", "int64"),
      getResultFromJson(jsonObj, "pk_script", "string"),
      u64(parseInt(getResultFromJson(jsonObj, "value", "int64"))),
      getResultFromJson(jsonObj, "spender", "string"),
      getResultFromJson(jsonObj, "type", "string"),
      getResultFromJson(jsonObj, "p2sh_asm_scripts", "array").split(";"),
      getResultFromJson(jsonObj, "pk_asm_scripts", "array").split(";"),
      getResultFromJson(jsonObj, "witness_asm_scripts", "array").split(";")
    );
  }
}

export class VoutV2 {
  txHash: string;
  txIndex: string;
  blockHash: string;
  blockHeight: string;
  blockTxIndex: string;

  pkScript: string;
  value: u64;
  spender: string;
  type: string;

  p2shAsmScripts: string[];
  pkAsmScripts: string[];

  constructor(
    txHash: string,
    txIndex: string,
    blockHash: string,
    blockHeight: string,
    blockTxIndex: string,
    pkScript: string,
    value: u64,
    spender: string,
    type: string,
    p2shAsmScripts: string[],
    pkAsmScripts: string[]
  ) {
    this.txHash = txHash;
    this.txIndex = txIndex;
    this.blockHash = blockHash;
    this.blockHeight = blockHeight;
    this.blockTxIndex = blockTxIndex;
    this.pkScript = pkScript;
    this.value = value;
    this.spender = spender;
    this.type = type;
    this.p2shAsmScripts = p2shAsmScripts;
    this.pkAsmScripts = pkAsmScripts;
  }

  static fromJson(jsonObj: JSON.Obj): VoutV2 {
    return new VoutV2(
      getResultFromJson(jsonObj, "tx_hash", "string"),
      getResultFromJson(jsonObj, "tx_index", "int64"),
      getResultFromJson(jsonObj, "block_hash", "string"),
      getResultFromJson(jsonObj, "block_height", "string"),
      getResultFromJson(jsonObj, "block_tx_index", "int64"),
      getResultFromJson(jsonObj, "pk_script", "string"),
      u64(parseInt(getResultFromJson(jsonObj, "value", "int64"))),
      getResultFromJson(jsonObj, "spender", "string"),
      getResultFromJson(jsonObj, "type", "string"),
      getResultFromJson(jsonObj, "p2sh_asm_scripts", "array").split(";"),
      getResultFromJson(jsonObj, "pk_asm_scripts", "array").split(";")
    );
  }
}

export class TransactionV2 {
  hash: string;
  lockTime: u32;
  version: u32;

  vins: VinV2[];
  vouts: VoutV2[];

  constructor(
    hash: string,
    lockTime: u32,
    version: u32,
    vins: VinV2[],
    vouts: VoutV2[]
  ) {
    this.hash = hash;
    this.lockTime = lockTime;
    this.version = version;
    this.vins = vins;
    this.vouts = vouts;
  }
}

export class TransactionV3 {
  hash: string;
  lockTime: u32;
  version: u32;
  safe: bool;
  blockId: u32;
  blockHash: string;
  blockHeight: u64;
  blockIndex: u32;

  constructor(
    hash: string,
    lockTime: u32,
    version: u32,
    safe: bool,
    blockId: u32,
    blockHash: string,
    blockHeight: u64,
    blockIndex: u32
  ) {
    this.hash = hash;
    this.lockTime = lockTime;
    this.version = version;
    this.safe = safe;
    this.blockId = blockId;
    this.blockHash = blockHash;
    this.blockHeight = blockHeight;
    this.blockIndex = blockIndex;
  }

  toJson(): string {
    return (
      '{ "hash": "' +
      this.hash +
      '", "lockTime": ' +
      this.lockTime.toString() +
      ', "version": ' +
      this.version.toString() +
      ', "safe": ' +
      this.safe.toString() +
      ', "blockId": ' +
      this.blockId.toString() +
      ', "blockHash": "' +
      this.blockHash +
      '", "blockHeight": ' +
      this.blockHeight.toString() +
      ', "blockIndex": ' +
      this.blockIndex.toString() +
      " }"
    );
  }
}
