import { consoleLog, valueReturn } from "./env";
import {
  Column,
  toJson,
  ptrToString,
  getResultFromJson,
  create,
  selectRow,
  insertRow,
  updateRows,
  deleteRows,
  getTxUTXOByBlockHeight,
  Table,
} from "./sdk";

// TODO: create Table class
const ordinalsTable = new Table("ordinals", [
  new Column("id", "int64"),
  new Column("address", "string"),
  new Column("value", "int64"),
]);

export function init(): void {
  ordinalsTable.init("id");
}

export function getOrdinal(id_ptr: i32): string {
  const id = ptrToString(id_ptr);
  return ordinalsTable.select([new Column("id", id)]);
}

// Testing functions
export function insertItemTest(): void {
  ordinalsTable.insert([
    new Column("id", "0"),
    new Column("address", "bc1q0d4836j3ekmm9cz7v3kcf0sdsxtmzg4ttpu7dm"),
    new Column("value", "1000"),
  ]);
}

export function updateItemTest(): void {
  ordinalsTable.update(
    [new Column("id", "0")],
    [new Column("address", "bc1qjr4gcelycyck4yxcnx5xt3w26u28veyu7meley")]
  );
}

export function deleteItemTest(): void {
  ordinalsTable.delete([new Column("id", "0")]);
}

export function selectItemTest(): void {
  const result = ordinalsTable.select([new Column("id", "0")]);
  const jsonResult = toJson(result);

  consoleLog(getResultFromJson(jsonResult, "id", "string"));
  consoleLog(getResultFromJson(jsonResult, "address", "string"));
  consoleLog(getResultFromJson(jsonResult, "value", "string"));
}

// Unit test purpose
export function index(block_height_ptr: i32): void {
  const block_height: u64 = u64(parseInt(ptrToString(block_height_ptr)));
  const utxos = getTxUTXOByBlockHeight(block_height);

  for (let i = 0; i < utxos.length; i++) {
    consoleLog(utxos[i].fundingTxHash);
  }
}

export function processString(str_ptr: i32): void {
  const input = ptrToString(str_ptr);
  const output = "output for " + input;

  valueReturn(output);
}
