import { JSON } from "assemblyscript-json/assembly";
import {
  contractAddress,
  createTable,
  deleteItem,
  getBlockByHeight,
  getOutpointsByTransactionHash,
  getTransactionsByBlockHash,
  insertItem,
  selectItems,
  updateItem,
  envGetTransactionV1sByBlockHeight,
  envGetTransactionV2sByBlockHeight,
  envGetNetwork,
  consoleLog,
  envGetTransactionByHash,
  envGetLastHeight,
} from "./env";
import { Value } from "assemblyscript-json/assembly/JSON";
import { TransactionV1, TransactionV3, VinV1, VinV2, VoutV1, VoutV2 } from "./types";
import { Network } from "./constants";

export class TableOption {
  primaryKey: string;
  // Default indexes are using btree, TODO: add more options
  indexes: string[];
  // MediumTexts are using text type mediumtext for the column
  mediumTexts: string[];

  constructor(primaryKey: string, indexes: string[], mediumTexts: string[]) {
    this.primaryKey = primaryKey;
    this.indexes = indexes;
    this.mediumTexts = mediumTexts;
  }

  toJson(): string {
    let obj = "{";
    obj += `"primaryKey": "${this.primaryKey}",`;
    obj += `"indexes": [`;
    for (let i = 0; i < this.indexes.length; i++) {
      obj += `"${this.indexes[i]}"`;
      if (i < this.indexes.length - 1) {
        obj += ",";
      }
    }
    obj += "]";
    obj += ",";
    obj += `"mediumTexts": [`;
    for (let i = 0; i < this.mediumTexts.length; i++) {
      obj += `"${this.mediumTexts[i]}"`;
      if (i < this.mediumTexts.length - 1) {
        obj += ",";
      }
    }
    obj += "]";
    obj += "}";

    return obj;
  }
}

export class Table {
  public name: string;
  public schema: Column[];

  constructor(name: string, schema: Column[]) {
    this.name = name;
    this.schema = schema;
  }

  public init(option: TableOption): void {
    create(this.name, this.schema, option);
  }

  public select(whereCondition: TableSchema): JSON.Obj {
    return selectRow(this.name, whereCondition);
  }

  public insert(values: TableSchema): void {
    insertRow(this.name, values);
  }

  public update(whereCondition: TableSchema, values: TableSchema): void {
    updateRows(this.name, whereCondition, values);
  }

  public delete(whereCondition: TableSchema): void {
    deleteRows(this.name, whereCondition);
  }
}

export class Column {
  name: string;
  type: string;

  constructor(name: string, type: string) {
    this.name = name;
    this.type = type;
  }
}

export class Transaction {
  hash: string;
  utxos: UTXO[];

  constructor(hash: string, utxo: UTXO[]) {
    this.hash = hash;
    this.utxos = utxo;
  }
}

export class UTXO {
  id: u64;
  value: u64;
  spendingTxHash: string;
  spendingTxIndex: string;
  spendingBlockHash: string;
  spendingBlockHeight: string;
  spendingBlockTxIndex: string;
  sequence: u64;
  fundingTxHash: string;
  fundingTxIndex: string;
  fundingBlockHash: string;
  fundingBlockHeight: string;
  fundingBlockTxIndex: string;
  signatureScript: string;
  pkScript: string;
  witness: string;
  spender: string;
  type: string;
  p2shAsmScripts: string[];
  pkAsmScripts: string[];
  witnessAsmScripts: string[];

  constructor(
    id: string,
    value: string,
    spendingTxHash: string,
    spendingTxIndex: string,
    spendingBlockHash: string,
    spendingBlockHeight: string,
    spendingBlockTxIndex: string,
    sequence: string,
    fundingTxHash: string,
    fundingTxIndex: string,
    fundingBlockHash: string,
    fundingBlockHeight: string,
    fundingBlockTxIndex: string,
    signatureScript: string,
    pkScript: string,
    witness: string,
    spender: string,
    type: string,
    p2shAsmScripts: string,
    pkAsmScripts: string,
    witnessAsmScripts: string
  ) {
    this.id = u64(parseInt(id));
    this.value = u64(parseInt(value));
    this.spendingTxHash = spendingTxHash;
    this.spendingTxIndex = spendingTxIndex;
    this.spendingBlockHash = spendingBlockHash;
    this.spendingBlockHeight = spendingBlockHeight;
    this.spendingBlockTxIndex = spendingBlockTxIndex;
    this.sequence = u64(parseInt(sequence));
    this.fundingTxHash = fundingTxHash;
    this.fundingTxIndex = fundingTxIndex;
    this.fundingBlockHash = fundingBlockHash;
    this.fundingBlockHeight = fundingBlockHeight;
    this.fundingBlockTxIndex = fundingBlockTxIndex;
    this.signatureScript = signatureScript;
    this.pkScript = pkScript;
    this.witness = witness;
    this.spender = spender;
    this.type = type;
    this.p2shAsmScripts = p2shAsmScripts.split(";");
    this.pkAsmScripts = pkAsmScripts.split(";");
    this.witnessAsmScripts = witnessAsmScripts.split(";");
  }

  static fromJson(jsonObj: JSON.Obj): UTXO {
    return new UTXO(
      getResultFromJson(jsonObj, "id", "int64"),
      getResultFromJson(jsonObj, "value", "int64"),
      getResultFromJson(jsonObj, "spending_tx_hash", "string"),
      getResultFromJson(jsonObj, "spending_tx_index", "int64"),
      getResultFromJson(jsonObj, "spending_block_hash", "string"),
      getResultFromJson(jsonObj, "spending_block_height", "int64"),
      getResultFromJson(jsonObj, "spending_block_tx_index", "int64"),
      getResultFromJson(jsonObj, "sequence", "int64"),
      getResultFromJson(jsonObj, "funding_tx_hash", "string"),
      getResultFromJson(jsonObj, "funding_tx_index", "int64"),
      getResultFromJson(jsonObj, "funding_block_hash", "string"),
      getResultFromJson(jsonObj, "funding_block_height", "int64"),
      getResultFromJson(jsonObj, "funding_block_tx_index", "int64"),
      getResultFromJson(jsonObj, "signature_script", "string"),
      getResultFromJson(jsonObj, "pk_script", "string"),
      getResultFromJson(jsonObj, "witness", "string"),
      getResultFromJson(jsonObj, "spender", "string"),
      getResultFromJson(jsonObj, "type", "string"),
      getResultFromJson(jsonObj, "p2sh_asm_scripts", "array"),
      getResultFromJson(jsonObj, "pk_asm_scripts", "array"),
      getResultFromJson(jsonObj, "witness_asm_scripts", "array")
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

export type TableSchema = Column[];

export function toStringSchema(tableDefinition: TableSchema): string {
  const obj = JSON.Value.Object();
  for (let i = 0; i < tableDefinition.length; i += 1) {
    obj.set(tableDefinition[i].name, tableDefinition[i].type);
  }
  return obj.toString();
}

export function toJson(jsonString: string): JSON.Obj {
  let jsonObj: JSON.Obj = <JSON.Obj>JSON.parse(jsonString);

  return jsonObj;
}

export function toJsonArray(jsonString: string): JSON.Arr {
  let jsonObj: JSON.Arr = <JSON.Arr>JSON.parse(jsonString);

  return jsonObj;
}

export function getResultFromJson(
  jsonObj: JSON.Obj,
  fieldName: string,
  type: string
): string {
  if (type === "int64") {
    let valueOrNull: JSON.Integer | null = jsonObj.getInteger(fieldName);
    if (valueOrNull != null) {
      let value: i64 = valueOrNull.valueOf();
      return value.toString();
    }
  } else if (type === "string") {
    let valueOrNull: JSON.Str | null = jsonObj.getString(fieldName);
    if (valueOrNull != null) {
      let value: string = valueOrNull.valueOf();
      return value;
    }
  } else if (type === "array") {
    let valueOrNull: JSON.Arr | null = jsonObj.getArr(fieldName);
    if (valueOrNull != null) {
      let value: Value[] = valueOrNull.valueOf();
      return value.join(";");
    }
  }

  return "";
}

export function ptrToString(ptr: i64): string {
  // get length
  let len = load<u32>(usize(ptr - 4));
  return String.UTF16.decodeUnsafe(<usize>ptr, <usize>len);
}

// Wrapped functions
export function create(
  tableName: string,
  tableSchema: TableSchema,
  option: TableOption
): void {
  createTable(tableName, toStringSchema(tableSchema), option.toJson());
}

export function selectRow(
  tableName: string,
  whereCondition: TableSchema
): JSON.Obj {
  const ptr = selectItems(tableName, toStringSchema(whereCondition));
  const result = toJson(ptrToString(ptr));

  return result;
}

export function insertRow(tableName: string, values: TableSchema): void {
  insertItem(tableName, toStringSchema(values));
}

export function updateRows(
  tableName: string,
  whereCondition: TableSchema,
  values: TableSchema
): void {
  updateItem(tableName, toStringSchema(whereCondition), toStringSchema(values));
}

export function deleteRows(
  tableName: string,
  whereCondition: TableSchema
): void {
  deleteItem(tableName, toStringSchema(whereCondition));
}

export function getUTXOByTransactionHash(hash: string): UTXO[] {
  const utxosPtr = getOutpointsByTransactionHash(hash);
  const utxosStr = ptrToString(utxosPtr);
  const jsonUTXOs = toJsonArray(utxosStr);
  const UTXOs: UTXO[] = [];

  for (let i = 0; i < jsonUTXOs.valueOf().length; i++) {
    const jsonObj = jsonUTXOs.valueOf()[i];

    if (jsonObj.isObj) {
      UTXOs.push(UTXO.fromJson(jsonObj as JSON.Obj));
    }
  }

  return UTXOs;
}

export function getTxHashesByBlockHeight(block_height: u64): string[] {
  // Get Block
  const ptr = getBlockByHeight(block_height);
  const result = ptrToString(ptr);

  const jsonResult = toJson(result);

  // Get Txs
  const hash = getResultFromJson(jsonResult, "hash", "string");
  const txHashesPtr = getTransactionsByBlockHash(hash);
  const resultTxHashes = ptrToString(txHashesPtr);
  const jsonResultTxHashes = toJsonArray(resultTxHashes);

  const txHashes: string[] = [];
  for (let i = 0; i < jsonResultTxHashes.valueOf().length; i++) {
    const jsonObj = jsonResultTxHashes.valueOf()[i];

    if (jsonObj.isObj) {
      const txHash = getResultFromJson(jsonObj as JSON.Obj, "hash", "string");
      txHashes.push(txHash);
    }
  }
  return txHashes;
}

export function getTxUTXOByBlockHeight(block_height: u64): UTXO[] {
  const txHashes = getTxHashesByBlockHeight(block_height);
  // Get UTXOs
  let UTXOs: UTXO[] = [];
  for (let i = 0; i < txHashes.length; i++) {
    const jsonResultUtxos = getUTXOByTransactionHash(txHashes[i]);
    UTXOs = UTXOs.concat(jsonResultUtxos);
  }
  return UTXOs;
}

export function getTxsByBlockHeight(block_height: u64): Transaction[] {
  const txHashes = getTxHashesByBlockHeight(block_height);

  // Get UTXOs
  let txs: Transaction[] = [];
  for (let i = 0; i < txHashes.length; i++) {
    const jsonResultUtxos = getUTXOByTransactionHash(txHashes[i]);
    let UTXOs: UTXO[] = jsonResultUtxos;
    txs.push(new Transaction(txHashes[i], UTXOs));
  }
  return txs;
}

export function getContractAddress(): string {
  return ptrToString(contractAddress());
}

export function getTransactionV1sByBlockHeight(height: u64): TransactionV1[] {
  const transactions: TransactionV1[] = [];

  const ptr = envGetTransactionV1sByBlockHeight(height);
  const trxsJson = toJsonArray(ptrToString(ptr));

  for (let i = 0; i < trxsJson.valueOf().length; i++) {
    const trxJson = trxsJson.valueOf()[i];

    const hash = getResultFromJson(trxJson as JSON.Obj, "hash", "string");
    const lockTime = <u32>(
      parseInt(getResultFromJson(trxJson as JSON.Obj, "lock_time", "int64"))
    );
    const version = <u32>(
      parseInt(getResultFromJson(trxJson as JSON.Obj, "version", "int64"))
    );

    const vinsJson = (trxJson as JSON.Obj).getArr("vins");
    const vins: VinV1[] = [];
    if (vinsJson) {
      for (let j = 0; j < vinsJson.valueOf().length; j++) {
        const vin = vinsJson.valueOf()[j];
        vins.push(
          new VinV1(
            getResultFromJson(vin as JSON.Obj, "tx_hash", "string"),
            <u32>parseInt(getResultFromJson(vin as JSON.Obj, "index", "int64")),
            <u64>parseInt(getResultFromJson(vin as JSON.Obj, "value", "int64")),
            getResultFromJson(vin as JSON.Obj, "witness", "string").split(",")
          )
        );
      }
    }

    const voutsJson = (trxJson as JSON.Obj).getArr("vouts");
    const vouts: VoutV1[] = [];
    if (voutsJson) {
      for (let j = 0; j < voutsJson.valueOf().length; j++) {
        const vout = voutsJson.valueOf()[j];
        vouts.push(
          new VoutV1(
            getResultFromJson(vout as JSON.Obj, "tx_hash", "string"),
            <u32>(
              parseInt(getResultFromJson(vout as JSON.Obj, "index", "int64"))
            ),
            getResultFromJson(vout as JSON.Obj, "address", "string"),
            getResultFromJson(vout as JSON.Obj, "pk_script", "string"),
            <u64>parseInt(getResultFromJson(vout as JSON.Obj, "value", "int64"))
          )
        );
      }
    }

    transactions.push(
      new TransactionV1(hash, <u32>lockTime, <u32>version, vins, vouts)
    );
  }

  return transactions;
}

export function getTransactionV2sByBlockHeight(height: u64): TransactionV2[] {
  const transactions: TransactionV2[] = [];

  const ptr = envGetTransactionV2sByBlockHeight(height);
  const trxsJson = toJsonArray(ptrToString(ptr));

  for (let i = 0; i < trxsJson.valueOf().length; i++) {
    const trxJson = trxsJson.valueOf()[i];

    const hash = getResultFromJson(trxJson as JSON.Obj, "hash", "string");
    const lockTime = <u32>(
      parseInt(getResultFromJson(trxJson as JSON.Obj, "lock_time", "int64"))
    );
    const version = <u32>(
      parseInt(getResultFromJson(trxJson as JSON.Obj, "version", "int64"))
    );

    const vinsJson = (trxJson as JSON.Obj).getArr("vins");
    const vins: VinV2[] = [];
    if (vinsJson) {
      for (let j = 0; j < vinsJson.valueOf().length; j++) {
        const vin = vinsJson.valueOf()[j];
        vins.push(VinV2.fromJson(vin as JSON.Obj));
      }
    }

    const voutsJson = (trxJson as JSON.Obj).getArr("vouts");
    const vouts: VoutV2[] = [];
    if (voutsJson) {
      for (let j = 0; j < voutsJson.valueOf().length; j++) {
        const vout = voutsJson.valueOf()[j];
        vouts.push(VoutV2.fromJson(vout as JSON.Obj));
      }
    }

    transactions.push(
      new TransactionV2(hash, <u32>lockTime, <u32>version, vins, vouts)
    );
  }

  return transactions;
}

export function getNetwork(): Network {
  const networkStr = ptrToString(envGetNetwork());

  if (networkStr === "mainnet") {
    return Network.Mainnet;
  } else if (networkStr === "testnet") {
    return Network.Testnet;
  } else if (networkStr === "signet") {
    return Network.Signet;
  } else {
    return Network.Regtest;
  }
}

export function getLastHeight(): u32 {
  const lastHeightStr = ptrToString(envGetLastHeight());
  consoleLog("LAST HEIGHT:");
  consoleLog(lastHeightStr);
  return u32(parseInt(lastHeightStr));
}

export function getTransactionByHash(_hash: string): TransactionV3 | null {
  const transactionPtr = envGetTransactionByHash(_hash);
  const str = ptrToString(transactionPtr)
  if(str === "null") {
    return null
  }

  const jsonTransaction = toJson(str);

  const hash = getResultFromJson(jsonTransaction, "hash", "string");
  const lockTime = <u32>(
    parseInt(getResultFromJson(jsonTransaction, "lock_time", "int64"))
  );
  const version = <u32>(
    parseInt(getResultFromJson(jsonTransaction, "version", "int64"))
  );
  const safe = getResultFromJson(jsonTransaction, "safe", "string") === "true";
  const blockId = <u32>(
    parseInt(getResultFromJson(jsonTransaction, "block_id", "int64"))
  );
  const blockHash = getResultFromJson(jsonTransaction, "block_hash", "string");
  const blockHeight = <u64>(
    parseInt(getResultFromJson(jsonTransaction, "block_height", "int64"))
  );
  const blockIndex = <u32>(
    parseInt(getResultFromJson(jsonTransaction, "block_index", "int64"))
  );

  return new TransactionV3(
    hash,
    lockTime,
    version,
    safe,
    blockId,
    blockHash,
    blockHeight,
    blockIndex
  );
}
