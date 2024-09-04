export class VinV1 {
  txHash: string;
  index: u32;
  value: u64;
  witness: string[];

  constructor(txHash: string, index: u32, value: u64, witness: string[]) {
    this.txHash = txHash;
    this.index = index;
    this.value = value;
    this.witness = witness;
  }

  toJson(): string {
    return (
      '{ "txHash": "' +
      this.txHash +
      '", "index": ' +
      this.index.toString() +
      ', "value": ' +
      this.value.toString() +
      ', "witness": [' +
      this.witness.map<string>((w: string) => '"' + w + '"').join(",") +
      "] }"
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
