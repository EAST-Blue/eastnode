@external("env", "consoleLog")
export declare function consoleLog(str: string): void;

@external("env", "valueReturn")
export declare function valueReturn(str: string): void;

@external("env", "panic")
export declare function panic(str: string): void;

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

// [{
//     "id": 1,
//     "hash": "3a6d490a7cf40819cdd826729d921ad5ab4b8347dfbec81179dd09aba0d25b37",
//     "block_hash": "000000009a940db389f3a7cbb8405f4ec14342bed36073b60ee63ed7e117f193",
//     "block_id": 189,
//     "lock_time": 0,
//     "version": 1,
//     "safe": 0
// }]

@external("env", "getTransactionsByBlockHash")
export declare function getTransactionsByBlockHash(block_hash: string): i32;


// [{
//     "id": 1,
//     "value": 5000000000,
//     "spending_tx_id": 0,
//     "spending_tx_hash": ""
//     "spending_tx_index": 0,
//     "sequence": 0,
//     "funding_tx_id": 194,
//     "funding_tx_hash": "3a6d490a7cf40819cdd826729d921ad5ab4b8347dfbec81179dd09aba0d25b37",
//     "funding_tx_index": 0,
//     "signature_script": "",
//     "pk_script": "410449fff9665bfda43017a27b3d32e986378befdd6fa5d4eb097626701ace807a2b3a43e74375dce4ed9028b3b62ba8485358cd48967e854a857a38ecdbfe5b62f8ac",
//     "witness": "",
//     "spender": "",
//     "type": "nonstandard",
// }]

@external("env", "getOutpointsByTransactionHash")
export declare function getOutpointsByTransactionHash(tx_hash: string): i32;