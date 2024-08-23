// SDK imports
import {
  consoleLog,
  valueReturn,
  Column,
  toJson,
  ptrToString,
  getResultFromJson,
  create,
  selectRow,
  insertRow,
  updateRows,
  deleteRows,
  selectNative,
  getTxUTXOByBlockHeight,
  Table,
  toJsonArray,
  JSON,
  getTxsByBlockHeight,
  getContractAddress
} from "@east-bitcoin-lib/smartindex-sdk/assembly";
export { allocate } from "@east-bitcoin-lib/smartindex-sdk/assembly/external";

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
  return ordinalsTable.select([new Column("id", id)]).toString();
}

// Testing functions
export function insertItemTest(): void {
  ordinalsTable.insert([
    new Column("id", "0"),
    new Column("address", "bc1q0d4836j3ekmm9cz7v3kcf0sdsxtmzg4ttpu7dm"),
    new Column("value", "1000"),
  ]);
  ordinalsTable.insert([
    new Column("id", "1"),
    new Column(
      "address",
      "bc1pkskdm7qk0z4gr8cgy38ysa00gyftj364gmf3uruse80c6gzunf6s0ywcsh"
    ),
    new Column("value", "100"),
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

  consoleLog(getResultFromJson(result, "id", "string"));
  consoleLog(getResultFromJson(result, "address", "string"));
  consoleLog(getResultFromJson(result, "value", "string"));
}

// Unit test purpose
export function index(block_height_ptr: i32): void {
  const block_height: u64 = u64(parseInt(ptrToString(block_height_ptr)));
  const utxos = getTxUTXOByBlockHeight(block_height);

  for (let i = 0; i < utxos.length; i++) {
    consoleLog("spendingTxHash: " + utxos[i].spendingTxHash);
    consoleLog("fundingTxHash: " + utxos[i].fundingTxHash);
    consoleLog("p2shAsmScripts: " + utxos[i].p2shAsmScripts.join(" "));
    consoleLog("pkAsmScripts: " + utxos[i].pkAsmScripts.join(" "));
    consoleLog("witnessAsmScripts: " + utxos[i].witnessAsmScripts.join(" "));
  }
}

export function processString(str_ptr: i32): void {
  const input = ptrToString(str_ptr);
  const output = "output for " + input;

  valueReturn(output);
}

export function selectNativeTest(): void {
  const ptr = selectNative("SELECT * from temp_ordinals", "[]");
  const result = toJsonArray(ptrToString(ptr));
  for (let i = 0; i < result.valueOf().length; i++) {
    const jsonObj = result.valueOf()[i];
    if (jsonObj.isObj) {
      consoleLog(getResultFromJson(jsonObj as JSON.Obj, "id", "string"));
      consoleLog(getResultFromJson(jsonObj as JSON.Obj, "address", "string"));
      consoleLog(getResultFromJson(jsonObj as JSON.Obj, "value", "string"));
    }
  }
}

export function testGetContractAddress(): void {
  const contractAddress = getContractAddress();
  valueReturn("contract address: " + contractAddress);
}
