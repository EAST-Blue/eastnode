@external("env", "consoleLog")
export declare function consoleLog(str: string): void;

@external("env", "valueReturn")
export declare function valueReturn(str: string): void;

@external("env", "panic")
export declare function panic(str: string): void;

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

@external("env", "selectNative")
export declare function selectNative(statement: string, args: string): i32;

@external("env", "getBlockByHeight")
export declare function getBlockByHeight(height: u64): i32;

@external("env", "getTransactionsByBlockHash")
export declare function getTransactionsByBlockHash(block_hash: string): i32;

@external("env", "getOutpointsByTransactionHash")
export declare function getOutpointsByTransactionHash(tx_hash: string): i32;

@external("env", "contractAddress")
export declare function contractAddress(): i32;

// TODD: better naming for this function
@external("env", "getTransactionV1sByBlockHeight")
export declare function envGetTransactionV1sByBlockHeight(height: u64): i32;
