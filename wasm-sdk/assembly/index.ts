import { u128 } from "as-bignum/assembly";
import { Column, toJson, toSchema } from "./config";
import {
  createTable,
  deleteItem,
  insertItem,
  updateItem,
  selectItems,
  consoleLog,
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

  let idOrNull: JSON.Str | null = jsonResult.getString("id"); 
  if (idOrNull != null) {
    let id: string = idOrNull.valueOf();
    consoleLog("id " + id)
  }
  let addressOrNull: JSON.Str | null = jsonResult.getString("address"); 
  if (addressOrNull != null) {
    let address: string = addressOrNull.valueOf();
    consoleLog("address " + address)
  }
  let valueOrNull: JSON.Str | null = jsonResult.getString("value"); 
  if (valueOrNull != null) {
    let value: string = valueOrNull.valueOf();
    consoleLog("value " + value)
  }
}

export function index(params: string[]): void {
  const a = u128.fromString(params[0]);
}
