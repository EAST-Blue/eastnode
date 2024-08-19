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
} from "./env";
import { Value } from "assemblyscript-json/assembly/JSON";
import { TransactionV1, VinV1, VoutV1 } from "./types";

export class Table {
  public name: string;
  public schema: Column[];

  constructor(name: string, schema: Column[]) {
    this.name = name;
    this.schema = schema;
  }

  public init(primaryKey: string): void {
    create(this.name, primaryKey, this.schema);
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
  primaryKey: string,
  tableSchema: TableSchema
): void {
  createTable(tableName, primaryKey, toStringSchema(tableSchema));
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
      UTXOs.push(
        new UTXO(
          getResultFromJson(jsonObj as JSON.Obj, "id", "int64"),
          getResultFromJson(jsonObj as JSON.Obj, "value", "int64"),
          getResultFromJson(jsonObj as JSON.Obj, "spending_tx_hash", "string"),
          getResultFromJson(jsonObj as JSON.Obj, "spending_tx_index", "int64"),
          getResultFromJson(
            jsonObj as JSON.Obj,
            "spending_block_hash",
            "string"
          ),
          getResultFromJson(
            jsonObj as JSON.Obj,
            "spending_block_height",
            "int64"
          ),
          getResultFromJson(
            jsonObj as JSON.Obj,
            "spending_block_tx_index",
            "int64"
          ),
          getResultFromJson(jsonObj as JSON.Obj, "sequence", "int64"),
          getResultFromJson(jsonObj as JSON.Obj, "funding_tx_hash", "string"),
          getResultFromJson(jsonObj as JSON.Obj, "funding_tx_index", "int64"),
          getResultFromJson(
            jsonObj as JSON.Obj,
            "funding_block_hash",
            "string"
          ),
          getResultFromJson(
            jsonObj as JSON.Obj,
            "funding_block_height",
            "int64"
          ),
          getResultFromJson(
            jsonObj as JSON.Obj,
            "funding_block_tx_index",
            "int64"
          ),
          getResultFromJson(jsonObj as JSON.Obj, "signature_script", "string"),
          getResultFromJson(jsonObj as JSON.Obj, "pk_script", "string"),
          getResultFromJson(jsonObj as JSON.Obj, "witness", "string"),
          getResultFromJson(jsonObj as JSON.Obj, "spender", "string"),
          getResultFromJson(jsonObj as JSON.Obj, "type", "string"),
          getResultFromJson(jsonObj as JSON.Obj, "p2sh_asm_scripts", "array"),
          getResultFromJson(jsonObj as JSON.Obj, "pk_asm_scripts", "array"),
          getResultFromJson(jsonObj as JSON.Obj, "witness_asm_scripts", "array")
        )
      );
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
            <u64>parseInt(getResultFromJson(vin as JSON.Obj, "value", "int64"))
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

// // TODO: add network
