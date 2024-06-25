@external("env", "consoleLog")
export declare function consoleLog(str: string): void;

// Host function to DB
// CreateTable
// Insert
// Select
// Update

@external("env", "createTable")
export declare function createTable(tableName: string, primaryKey: string, tableSchema: string): boolean;

@external("env", "insertItem")
export declare function insertItem(tableName: string, values: string): boolean;

@external("env", "updateItem")
export declare function updateItem(tableName: string, whereCondition: string, values: string): boolean;

@external("env", "deleteItem")
export declare function deleteItem(tableName: string, whereCondition: string): boolean;

@external("env", "selectItem")
export declare function selectItems(tableName: string, whereCondition: string): i32;


// Host function to Bitcoin node
// Get block id
// Get transaction id

// declare function __get(
//     k: ArrayBuffer,
//     v: ArrayBuffer,
//   ): void;
// 


// {
//     "id": 1, // ascending number from dolt,
//     "version": 1, // version number for the block
//     "height": 0,
//     "previous_block": "0000000000000000000000000000000000000000000000000000000000000000",
//     "merkle_root": "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
//     "hash": "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f",
//     "time": 1231006505000,
//     "nonce": 2083236893,
//     "bits": 486604799
// }

@external("env", "getBlockByHeight")
export declare function getBlockByHeight(height: u64): i32;
// export declare function getTransaction(id: u64): ArrayBuffer;