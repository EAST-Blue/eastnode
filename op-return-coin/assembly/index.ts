import {
  valueReturn,
  Column,
  ptrToString,
  getResultFromJson,
  getTxUTXOByBlockHeight,
  Table,
  getUTXOByTransactionHash,
} from "east-wasm-sdk/assembly";
export { allocate } from "east-wasm-sdk/assembly/external";
import { decode } from "as-hex/assembly";

const stateTable = new Table("state", [
  new Column("id", "int64"),
  new Column("indexed_block_height", "int64"),
]);

const ledgerTable = new Table("ledger", [
  new Column("address", "string"),
  new Column("coin", "string"),
  new Column("balance", "int64"),
]);

const transactionsTable = new Table("transactions", [
  new Column("coin", "string"),
  new Column("from", "string"),
  new Column("to", "string"),
  new Column("amount", "int64"),
]);

export function init(): void {
  stateTable.init("indexed_block_height");
  stateTable.insert([
    new Column("id", "0"),
    new Column("indexed_block_height", "0"),
  ]);

  ledgerTable.init("");
  transactionsTable.init("");
}

export function getIndexedBlock(): void {
  const latestBlock: string = getResultFromJson(
    stateTable.select([new Column("id", "0")]),
    "indexed_block_height",
    "string"
  );
  valueReturn(latestBlock);
}

export function getBalance(address_str: i32): void {
  const address = ptrToString(address_str);
  const balance = _getBalance(address);
  valueReturn(balance.toString());
}

function _getBalance(address: string): i64 {
  let balance: i64 = 0;
  const selectResult = ledgerTable.select([new Column("address", address)]);
  if (getResultFromJson(selectResult, "error", "string") != "not-found") {
    balance = i64(
      parseInt(getResultFromJson(selectResult, "balance", "string"))
    );
  } else {
    return -1;
  }
  return balance;
}

export function index(number_of_blocks_ptr: i32): void {
  const number_of_block: i64 = i64(parseInt(ptrToString(number_of_blocks_ptr)));

  // get latest state
  const latestBlock: i64 = i64(
    parseInt(
      getResultFromJson(
        stateTable.select([new Column("id", "0")]),
        "indexed_block_height",
        "string"
      )
    )
  );

  for (let i = 1; i < number_of_block + 1; i++) {
    const utxos = getTxUTXOByBlockHeight(latestBlock + i);

    for (let i = 0; i < utxos.length; i++) {
      if (utxos[i].pkAsmScripts.length == 2) {
        if (utxos[i].pkAsmScripts[0] === "OP_RETURN") {
          const msg = decode(utxos[i].pkAsmScripts[1]);
          if (msg.substr(0, 4) === "EAST") {
            const command = msg.charAt(4);
            const args = msg.substr(5).split("_");

            // get all utxo in the same tx
            const sameTxUTXOs = getUTXOByTransactionHash(
              utxos[i].fundingTxHash
            );

            let from = "";
            let to = "";

            for (let j = 0; j < sameTxUTXOs.length; j++) {
              if (sameTxUTXOs[j].spendingTxHash === utxos[i].fundingTxHash) {
                from = sameTxUTXOs[j].spender;
              }
            }

            for (let j = 0; j < sameTxUTXOs.length; j++) {
              if (sameTxUTXOs[j].fundingTxHash === utxos[i].fundingTxHash) {
                if (
                  sameTxUTXOs[j].spender != from &&
                  sameTxUTXOs[j].spender.length > 1
                ) {
                  to = sameTxUTXOs[j].spender;
                }
              }
            }

            if (command === "i") {
              const coinName = args[0];
              const amount = args[1];

              ledgerTable.insert([
                new Column("address", from),
                new Column("coin", coinName),
                new Column("balance", amount),
              ]);
              transactionsTable.insert([
                new Column("coin", coinName),
                new Column("from", ""),
                new Column("to", from),
                new Column("amount", amount),
              ]);
            } else if (command === "t") {
              const coinName = args[0];
              const amount = args[1];

              // decrease from owner
              let fromBalance = _getBalance(from);
              let toBalance = _getBalance(to);

              if (fromBalance >= i64(parseInt(amount))) {
                fromBalance -= i64(parseInt(amount));

                ledgerTable.update(
                  [new Column("address", from)],
                  [
                    new Column("address", from),
                    new Column("coin", coinName),
                    new Column("balance", fromBalance.toString()),
                  ]
                );

                if (toBalance == -1) {
                  ledgerTable.insert([
                    new Column("address", to),
                    new Column("coin", coinName),
                    new Column("balance", amount),
                  ]);
                } else {
                  toBalance += i64(parseInt(amount));
                  ledgerTable.update(
                    [new Column("address", to)],
                    [
                      new Column("address", to),
                      new Column("coin", coinName),
                      new Column("balance", toBalance.toString()),
                    ]
                  );
                }

                transactionsTable.insert([
                  new Column("coin", coinName),
                  new Column("from", from),
                  new Column("to", to),
                  new Column("amount", amount),
                ]);
              }
            }
          }
        }
      }
    }

    stateTable.update(
      [new Column("id", "0")],
      [new Column("indexed_block_height", (latestBlock + i).toString())]
    );
  }
}
