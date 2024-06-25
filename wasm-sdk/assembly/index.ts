import { u128 } from "as-bignum/assembly";
import { Column, toJson, toSchema } from "./config";
import {
  createTable,
  deleteItem,
  insertItem,
  updateItem,
  selectItems,
  consoleLog,
  getBlockByHeight,
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

  getResultFromJson(jsonResult, "id", "string")
  getResultFromJson(jsonResult, "address", "string")
  getResultFromJson(jsonResult, "value", "string")
}

export function index(block_height: u64): void {
  const ptr = getBlockByHeight(block_height);
  const result = toString(ptr);

  const jsonResult = toJson(result);

  getResultFromJson(jsonResult, "id", "int64")
  getResultFromJson(jsonResult, "version", "int64")
  getResultFromJson(jsonResult, "height", "string")
  getResultFromJson(jsonResult, "previous_block", "string")
  getResultFromJson(jsonResult, "merkle_root", "string")
  getResultFromJson(jsonResult, "hash", "string")
  getResultFromJson(jsonResult, "time", "int64")
  getResultFromJson(jsonResult, "nonce", "int64")
  getResultFromJson(jsonResult, "bits", "int64")

}

function getResultFromJson(jsonObj: JSON.Obj, fieldName: string, type: string): void {
  if (type === "int64") {
    let valueOrNull: JSON.Integer | null = jsonObj.getInteger(fieldName);
    if (valueOrNull != null) {
      let value: i64 = valueOrNull.valueOf();
      consoleLog(fieldName + ": " + value.toString());
    }
  } else {
    let valueOrNull: JSON.Str | null = jsonObj.getString(fieldName);
    if (valueOrNull != null) {
      let value: string = valueOrNull.valueOf();
      consoleLog(fieldName + ": " + value);
    }
  }
}
