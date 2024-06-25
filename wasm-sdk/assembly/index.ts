import { u128 } from "as-bignum/assembly";
import { Column, toJson, toJsonArray, toSchema } from "./config";
import {
  createTable,
  deleteItem,
  insertItem,
  updateItem,
  selectItems,
  consoleLog,
  getBlockByHeight,
  getTransactionsByBlockHash,
  getOutpointsByTransactionHash,
} from "./env";
import { toString } from "./utils";
import { JSON } from "assemblyscript-json/assembly";

const ORDINALS_TABLE_NAME = "ordinals";
const ORDINALS_PRIMARY_KEY = "id";
const ORDINALS_TABLE_SCHEMA = [
  new Column("id", "int64"),
  new Column("address", "string"),
  new Column("value", "int64"),
];

export function init(): void {
  const generatedSchema = toSchema(ORDINALS_TABLE_SCHEMA);
  createTable(ORDINALS_TABLE_NAME, ORDINALS_PRIMARY_KEY, generatedSchema);
}

export function getOrdinal(id: string): string {
  const whereCondition = toSchema([new Column("id", id)]);
  const ptr = selectItems(ORDINALS_TABLE_NAME, whereCondition);
  const result = toString(ptr);

  return result;
}

export function insertItemTest(): void {
  // will be casted
  const values = toSchema([
    new Column("id", "0"),
    new Column("address", "bc1q0d4836j3ekmm9cz7v3kcf0sdsxtmzg4ttpu7dm"),
    new Column("value", "1000"),
  ]);
  insertItem(ORDINALS_TABLE_NAME, values);
}

export function updateItemTest(): void {
  // will be casted
  const whereCondition = toSchema([new Column("id", "0")]);
  const values = toSchema([
    new Column("address", "bc1qjr4gcelycyck4yxcnx5xt3w26u28veyu7meley"),
  ]);
  updateItem(ORDINALS_TABLE_NAME, whereCondition, values);
}

export function deleteItemTest(): void {
  // will be casted
  const whereCondition = toSchema([new Column("id", "0")]);
  deleteItem(ORDINALS_TABLE_NAME, whereCondition);
}

export function selectItemTest(): void {
  // will be casted
  const whereCondition = toSchema([new Column("id", "0")]);
  const ptr = selectItems(ORDINALS_TABLE_NAME, whereCondition);
  const result = toString(ptr);

  const jsonResult = toJson(result);

  getResultFromJson(jsonResult, "id", "string");
  getResultFromJson(jsonResult, "address", "string");
  getResultFromJson(jsonResult, "value", "string");
}

export function index(block_height: u64): void {
  // Get Block
  const ptr = getBlockByHeight(block_height);
  const result = toString(ptr);

  const jsonResult = toJson(result);

  getResultFromJson(jsonResult, "id", "int64");
  getResultFromJson(jsonResult, "version", "int64");
  getResultFromJson(jsonResult, "height", "string");
  getResultFromJson(jsonResult, "previous_block", "string");
  getResultFromJson(jsonResult, "merkle_root", "string");
  getResultFromJson(jsonResult, "time", "int64");
  getResultFromJson(jsonResult, "nonce", "int64");
  getResultFromJson(jsonResult, "bits", "int64");

  // Get Txs
  const hash = getResultFromJson(jsonResult, "hash", "string");
  const txHashesPtr = getTransactionsByBlockHash(hash);
  const resultTxHashes = toString(txHashesPtr);
  const jsonResultTxHashes = toJsonArray(resultTxHashes);

  const txHashes: string[] = []
  for (let i = 0; i < jsonResultTxHashes.valueOf().length; i++) {
    const jsonObj = jsonResultTxHashes.valueOf()[i];

    if (jsonObj.isObj) {
      getResultFromJson(jsonObj as JSON.Obj, "id", "int64");
      getResultFromJson(jsonObj as JSON.Obj, "hash", "string");
      getResultFromJson(jsonObj as JSON.Obj, "block_id", "int64");
      getResultFromJson(jsonObj as JSON.Obj, "lock_time", "int64");
      getResultFromJson(jsonObj as JSON.Obj, "version", "int64");
      getResultFromJson(jsonObj as JSON.Obj, "safe", "int64");
      const txHash = getResultFromJson(jsonObj as JSON.Obj, "block_hash", "string");
      txHashes.push(txHash)
    }
  }

  // Get UTXOs

  for (let i = 0; i < txHashes.length; i++) {
      const utxosPtr = getOutpointsByTransactionHash(txHashes[i])
      const utxosStr = toString(utxosPtr);
      const jsonResultUtxos = toJsonArray(utxosStr);

      for (let i = 0; i < jsonResultUtxos.valueOf().length; i++) {
        const jsonObj = jsonResultUtxos.valueOf()[i];
    
        if (jsonObj.isObj) {
          getResultFromJson(jsonObj as JSON.Obj, "id", "int64");
          getResultFromJson(jsonObj as JSON.Obj, "value", "int64");
          getResultFromJson(jsonObj as JSON.Obj, "spending_tx_id", "string");
          getResultFromJson(jsonObj as JSON.Obj, "spending_tx_hash", "string");
          getResultFromJson(jsonObj as JSON.Obj, "spending_tx_index", "string");
          getResultFromJson(jsonObj as JSON.Obj, "sequence", "int64");
          getResultFromJson(jsonObj as JSON.Obj, "funding_tx_id", "string");
          getResultFromJson(jsonObj as JSON.Obj, "funding_tx_hash", "string");
          getResultFromJson(jsonObj as JSON.Obj, "funding_tx_index", "string");
          getResultFromJson(jsonObj as JSON.Obj, "signature_script", "string");
          getResultFromJson(jsonObj as JSON.Obj, "pk_script", "string");
          getResultFromJson(jsonObj as JSON.Obj, "witness", "string");
          getResultFromJson(jsonObj as JSON.Obj, "spender", "string");
          getResultFromJson(jsonObj as JSON.Obj, "type", "string");
        }
      }



  }

}

function getResultFromJson(
  jsonObj: JSON.Obj,
  fieldName: string,
  type: string
): string {
  if (type === "int64") {
    let valueOrNull: JSON.Integer | null = jsonObj.getInteger(fieldName);
    if (valueOrNull != null) {
      let value: i64 = valueOrNull.valueOf();
      consoleLog(fieldName + ": " + value.toString());
      return value.toString();
    }
  } else {
    let valueOrNull: JSON.Str | null = jsonObj.getString(fieldName);
    if (valueOrNull != null) {
      let value: string = valueOrNull.valueOf();
      consoleLog(fieldName + ": " + value);
      return value;
    }
  }

  return "";
}
