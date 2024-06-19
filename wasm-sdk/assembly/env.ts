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
// export declare function getBlock(id: u64): ArrayBuffer;
// export declare function getTransaction(id: u64): ArrayBuffer;