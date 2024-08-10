NASIPADANGUGUGAK
Etching: Normal etching, no fancy setup
Transaction 1: http://localhost:3000/tx/4c512b8c33edba01fdd6f354d5fb964fd5c569d4df0913bf2845f12bd10bcd60

Mint 1: Mint 100 amount to index 0
http://localhost:3000/tx/dcb5926315930115b69b2de7c4c4d8957088edfdc2fa72d070e71e768d5d3ac0
Mint 2: http://localhost:3000/tx/3b1eafa7c3953f36ef8439ff76b8a68f001b452cf7a386b8b8eed24debb7936a
Mint 3: Without index 0, auto assign to non OP_RETURN output index 1
http://localhost:3000/tx/875321bae9ed3fa9621c829f9ecf2f64414936ade77f64b0467689171b1c8fa5
Mint 4: Specify the index using pointer to index 2
http://localhost:3000/tx/3ec1f90504263001ae3efe4ad9fc591a82f4c0d6b4bd7e15f92ef86dcc23113c

Edict 1: Transfer 2 inputs, each with 100 balance. Set the amount to 10 and output to index 0. The remaining 190 balance will be transferred to the first non-OP_RETURN output.
http://localhost:3000/tx/8274b9857286ee8a9a9a8c23ce730adfa19f36a4ace94437d0113be0c9c219a1
Edict 2: Transfer 1 input of 200 RUNE, set the amount to 10 and output to index 1. The balance is split into 2 UTXOs on the same address.
http://localhost:3000/tx/5b5d09109c429c1dee77e6d44a5be70f95fa91066c9b2ac642f7ac772e511c1f
Edict 3: Transfer 1 input of 190 RUNE, set the amount to 69 and output to index 1. The balance is split into 2 UTXOs on a different address.
http://localhost:3000/tx/8b604ebb0426d24dd947f4bd46bed9d592e62d84aa25e0fa4c4f988edb32e45c
Edict 4: Transfer 1 input of 121 RUNE, set the amounts to 90, 20, and 11 for the 3 outputs.
http://localhost:3000/tx/c6377d43b83d51ec3b1b8d295a1a5deab5f93e886e12b076a4388af924573b27

```
{
  "entry": {
    "block": 118,
    "burned": 0,
    "divisibility": 0,
    "etching": "4c512b8c33edba01fdd6f354d5fb964fd5c569d4df0913bf2845f12bd10bcd60",
    "mints": 4,
    "number": 0,
    "premine": 0,
    "spaced_rune": "NASIPADANGUGUGAK",
    "symbol": "💸",
    "terms": {
      "amount": 100,
      "cap": 20,
      "height": [
        100,
        200
      ],
      "offset": [
        null,
        null
      ]
    },
    "timestamp": 1723100347,
    "turbo": false
  },
  "id": "118:1",
  "mintable": true,
  "parent": null
}
```

```
{
  "NASIPADANGUGUGAK": {
    "5b5d09109c429c1dee77e6d44a5be70f95fa91066c9b2ac642f7ac772e511c1f:1": 10,
    "c6377d43b83d51ec3b1b8d295a1a5deab5f93e886e12b076a4388af924573b27:0": 90,
    "c6377d43b83d51ec3b1b8d295a1a5deab5f93e886e12b076a4388af924573b27:1": 20,
    "c6377d43b83d51ec3b1b8d295a1a5deab5f93e886e12b076a4388af924573b27:2": 11,
    "8b604ebb0426d24dd947f4bd46bed9d592e62d84aa25e0fa4c4f988edb32e45c:1": 69,
    "875321bae9ed3fa9621c829f9ecf2f64414936ade77f64b0467689171b1c8fa5:1": 100,
    "dcb5926315930115b69b2de7c4c4d8957088edfdc2fa72d070e71e768d5d3ac0:0": 100
  }
}
```
