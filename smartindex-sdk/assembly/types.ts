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
