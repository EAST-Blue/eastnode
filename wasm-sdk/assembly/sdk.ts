import { JSON } from "assemblyscript-json/assembly";
import {
  createTable,
  deleteItem,
  getBlockByHeight,
  getOutpointsByTransactionHash,
  getTransactionsByBlockHash,
  insertItem,
  selectItems,
  updateItem,
} from "./env";

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

  public select(whereCondition: TableSchema): string {
    return selectRow(this.name, whereCondition);
  }

  public insert(values: TableSchema): void {
    insertRow(this.name, values);
  }

  public update(whereCondition: TableSchema, values: TableSchema): void {
    updateRows(this.name, whereCondition, values);
  }

  public delete(whereCondition: TableSchema): void {
    deleteRows(this.name, whereCondition)
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

export class UTXO {
  id: u64;
  value: u64;
  spendingTxId: string;
  spendingTxHash: string;
  spendingTxIndex: string;
  sequence: u64;
  fundingTxId: string;
  fundingTxHash: string;
  fundingTxIndex: string;
  signatureScript: string;
  pkScript: string;
  witness: string;
  spender: string;
  type: string;

  constructor(
    id: string,
    value: string,
    spendingTxId: string,
    spendingTxHash: string,
    spendingTxIndex: string,
    sequence: string,
    fundingTxId: string,
    fundingTxHash: string,
    fundingTxIndex: string,
    signatureScript: string,
    pkScript: string,
    witness: string,
    spender: string,
    type: string
  ) {
    this.id = u64(parseInt(id));
    this.value = u64(parseInt(value));
    this.spendingTxId = spendingTxId;
    this.spendingTxHash = spendingTxHash;
    this.spendingTxIndex = spendingTxIndex;
    this.sequence = u64(parseInt(sequence));
    this.fundingTxId = fundingTxId;
    this.fundingTxHash = fundingTxHash;
    this.fundingTxIndex = fundingTxIndex;
    this.signatureScript = signatureScript;
    this.pkScript = pkScript;
    this.witness = witness;
    this.spender = spender;
    this.type = type;
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
  } else {
    let valueOrNull: JSON.Str | null = jsonObj.getString(fieldName);
    if (valueOrNull != null) {
      let value: string = valueOrNull.valueOf();
      return value;
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
): string {
  const ptr = selectItems(tableName, toStringSchema(whereCondition));
  const result = ptrToString(ptr);

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

export function getTxUTXOByBlockHeight(block_height: u64): UTXO[] {
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
      const txHash = getResultFromJson(
        jsonObj as JSON.Obj,
        "block_hash",
        "string"
      );
      txHashes.push(txHash);
    }
  }

  // Get UTXOs
  const UTXOs: UTXO[] = [];
  for (let i = 0; i < txHashes.length; i++) {
    const utxosPtr = getOutpointsByTransactionHash(txHashes[i]);
    const utxosStr = ptrToString(utxosPtr);
    const jsonResultUtxos = toJsonArray(utxosStr);

    for (let i = 0; i < jsonResultUtxos.valueOf().length; i++) {
      const jsonObj = jsonResultUtxos.valueOf()[i];

      if (jsonObj.isObj) {
        UTXOs.push(
          new UTXO(
            getResultFromJson(jsonObj as JSON.Obj, "id", "int64"),
            getResultFromJson(jsonObj as JSON.Obj, "value", "int64"),
            getResultFromJson(jsonObj as JSON.Obj, "spending_tx_id", "string"),
            getResultFromJson(
              jsonObj as JSON.Obj,
              "spending_tx_hash",
              "string"
            ),
            getResultFromJson(
              jsonObj as JSON.Obj,
              "spending_tx_index",
              "string"
            ),
            getResultFromJson(jsonObj as JSON.Obj, "sequence", "int64"),
            getResultFromJson(jsonObj as JSON.Obj, "funding_tx_id", "string"),
            getResultFromJson(jsonObj as JSON.Obj, "funding_tx_hash", "string"),
            getResultFromJson(
              jsonObj as JSON.Obj,
              "funding_tx_index",
              "string"
            ),
            getResultFromJson(
              jsonObj as JSON.Obj,
              "signature_script",
              "string"
            ),
            getResultFromJson(jsonObj as JSON.Obj, "pk_script", "string"),
            getResultFromJson(jsonObj as JSON.Obj, "witness", "string"),
            getResultFromJson(jsonObj as JSON.Obj, "spender", "string"),
            getResultFromJson(jsonObj as JSON.Obj, "type", "string")
          )
        );
      }
    }
  }
  return UTXOs;
}
