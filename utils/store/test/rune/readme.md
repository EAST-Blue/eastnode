# INDEXER 1

## ETCHING
tx: 58962c96f1b981a5fafbd64b3d44a8ac0922dee330a8211e73395681cf3588b2

## MINT

### MINT_1: DEFAULT POINTER
tx: 7ba48104b4e7e36d010497f402648c576b29b66f42e5792c2a400add970c33ed
100 -> index 0
```json
{
  "NASI•GORENG•PEDAS": {
    "7ba48104b4e7e36d010497f402648c576b29b66f42e5792c2a400add970c33ed:0": 100
  }
}
```

### MINT_2: DEFAULT POINTER, OP_RETURN ON INDEX 0
tx: acd77c4dbd8d43d508799f0a7e7cdad5cded13da68b7ccb1616a5e7ac20ab09a
100 -> index 1 (default index non op_return)
```json
{
  "NASI•GORENG•PEDAS": {
    "acd77c4dbd8d43d508799f0a7e7cdad5cded13da68b7ccb1616a5e7ac20ab09a:1": 100,
    "7ba48104b4e7e36d010497f402648c576b29b66f42e5792c2a400add970c33ed:0": 100
  }

```

### MINT_3: POINTER INDEX 3
tx: 73ec8431e09fa6f6375f4acb4ac7323b094e92396f847973f2c82fbcfc8aec78
100 -> index 3
```json
{
  "NASI•GORENG•PEDAS": {
    "73ec8431e09fa6f6375f4acb4ac7323b094e92396f847973f2c82fbcfc8aec78:3": 100,
    "acd77c4dbd8d43d508799f0a7e7cdad5cded13da68b7ccb1616a5e7ac20ab09a:1": 100,
    "7ba48104b4e7e36d010497f402648c576b29b66f42e5792c2a400add970c33ed:0": 100
  }
}

```
### MINT_4: NO ELIGIBLE OUTPUT, ONLY OP_RETURN
tx: 9b57374c1e9eae1fd58b9c8b0b678cc8601c122c8616d335d164cb1eb72c7aa5
amount will be transfered into change address hehe
100 -> index 1
```json
{
  "NASI•GORENG•PEDAS": {
    "73ec8431e09fa6f6375f4acb4ac7323b094e92396f847973f2c82fbcfc8aec78:3": 100,
    "acd77c4dbd8d43d508799f0a7e7cdad5cded13da68b7ccb1616a5e7ac20ab09a:1": 100,
    "9b57374c1e9eae1fd58b9c8b0b678cc8601c122c8616d335d164cb1eb72c7aa5:1": 100,
    "7ba48104b4e7e36d010497f402648c576b29b66f42e5792c2a400add970c33ed:0": 100
  }
}

```
### MINT_5: INVALID POINTER, INDEX > VOUTS.LENGTH
tx: 1e84565ebe16c21344530529d5fea15d84549ca0fcbd56bc80168405c8b2c847

```json
{
  "NASI•GORENG•PEDAS": {
    "73ec8431e09fa6f6375f4acb4ac7323b094e92396f847973f2c82fbcfc8aec78:3": 100,
    "acd77c4dbd8d43d508799f0a7e7cdad5cded13da68b7ccb1616a5e7ac20ab09a:1": 100,
    "9b57374c1e9eae1fd58b9c8b0b678cc8601c122c8616d335d164cb1eb72c7aa5:1": 100,
    "7ba48104b4e7e36d010497f402648c576b29b66f42e5792c2a400add970c33ed:0": 100
  }
}

```

### MINT_6: INVALID POINTER, INDEX  === VOUTS.LENGTH
Unfortunately, the 9b57374c1e9eae1fd58b9c8b0b678cc8601c122c8616d335d164cb1eb72c7aa5:1 has been destroyed because the outpoint included in this transaction.
tx: 58f46c471849e72610c369e5263a0ade230b3f0a2d664215b046c94893d597c5

```json
{
  "NASI•GORENG•PEDAS": {
    "73ec8431e09fa6f6375f4acb4ac7323b094e92396f847973f2c82fbcfc8aec78:3": 100,
    "acd77c4dbd8d43d508799f0a7e7cdad5cded13da68b7ccb1616a5e7ac20ab09a:1": 100,
    "7ba48104b4e7e36d010497f402648c576b29b66f42e5792c2a400add970c33ed:0": 100
  }
}
```
## EDICT 
### EDICT_1: TF 10 to index 0, remaning default index 0
Transfer 1 inputs total 100. Set the amount to 10 and output to index 0. The remaining 90 balance will be transferred to the first non-OP_RETURN output (index 0)
tx: 3c0995475cb14be31328ab7631dc3f7e4559c891459b29cf9caafd9b3d696485
```json
{
  "NASI•GORENG•PEDAS": {
    "73ec8431e09fa6f6375f4acb4ac7323b094e92396f847973f2c82fbcfc8aec78:3": 100,
    "3c0995475cb14be31328ab7631dc3f7e4559c891459b29cf9caafd9b3d696485:0": 100, // replaced the output
    "acd77c4dbd8d43d508799f0a7e7cdad5cded13da68b7ccb1616a5e7ac20ab09a:1": 100
  }
}
```
### EDICT_2: TF 10 to index 1, remaning default index 0
Transfer 1 inputs total 100. Set the amount to 10 and output to index 1. The remaining 90 balance will be transferred to the first non-OP_RETURN output (index 0)
tx: c3775672b925b7173dea8d9da5fc6db5b214658091de31818fee38701f6819a3

```json
{
  "NASI•GORENG•PEDAS": {
    "73ec8431e09fa6f6375f4acb4ac7323b094e92396f847973f2c82fbcfc8aec78:3": 100,
    "acd77c4dbd8d43d508799f0a7e7cdad5cded13da68b7ccb1616a5e7ac20ab09a:1": 100,
    "c3775672b925b7173dea8d9da5fc6db5b214658091de31818fee38701f6819a3:0": 90,
    "c3775672b925b7173dea8d9da5fc6db5b214658091de31818fee38701f6819a3:1": 10
  }
}
```

### EDICT_3: Multiple TFs, 28, 83, 100 of total inputs 190 into index 0,1,2
Because the total inputs are 190, the last TF amount become 79 instead of 100
tx: ff419a53bc5acf9d12b2a175c88bfef651793065cd56b007458e43b9eb0e1185

```json
{
  "NASI•GORENG•PEDAS": {
    "ff419a53bc5acf9d12b2a175c88bfef651793065cd56b007458e43b9eb0e1185:0": 28,
    "ff419a53bc5acf9d12b2a175c88bfef651793065cd56b007458e43b9eb0e1185:1": 83,
    "ff419a53bc5acf9d12b2a175c88bfef651793065cd56b007458e43b9eb0e1185:2": 79,
    "acd77c4dbd8d43d508799f0a7e7cdad5cded13da68b7ccb1616a5e7ac20ab09a:1": 100,
    "c3775672b925b7173dea8d9da5fc6db5b214658091de31818fee38701f6819a3:1": 10
  }
}

```

### EDICT_4: Divide amounts into all outputs
Because the total inputs are 190, the last TF amount become 79 instead of 100
Divide 100 amounts into all outputs (7 - 1 (op_return))
tx: 428ac779b02094ae4009626d0b5e03727bec57a44207e025fbe8a641a9d9cb82
```json
{
  "NASI•GORENG•PEDAS": {
    "428ac779b02094ae4009626d0b5e03727bec57a44207e025fbe8a641a9d9cb82:0": 35,
    "428ac779b02094ae4009626d0b5e03727bec57a44207e025fbe8a641a9d9cb82:1": 13,
    "428ac779b02094ae4009626d0b5e03727bec57a44207e025fbe8a641a9d9cb82:2": 13,
    "428ac779b02094ae4009626d0b5e03727bec57a44207e025fbe8a641a9d9cb82:3": 13,
    "428ac779b02094ae4009626d0b5e03727bec57a44207e025fbe8a641a9d9cb82:4": 13,
    "428ac779b02094ae4009626d0b5e03727bec57a44207e025fbe8a641a9d9cb82:6": 13,
    "ff419a53bc5acf9d12b2a175c88bfef651793065cd56b007458e43b9eb0e1185:0": 28,
    "ff419a53bc5acf9d12b2a175c88bfef651793065cd56b007458e43b9eb0e1185:1": 83,
    "ff419a53bc5acf9d12b2a175c88bfef651793065cd56b007458e43b9eb0e1185:2": 79,
    "c3775672b925b7173dea8d9da5fc6db5b214658091de31818fee38701f6819a3:1": 10
  }
}

```

# INDEXER 2

## EDICT 
### EDICT_1: Divide amounts into all outputs (edict: 0 amount)
Transfer 3 inputs total 300. Set the amount to 0
tx: 59f5afee2338faeeaab3d4403d8099b511ee826e6f6b508cf767fef26c2aa91f
```json
{
  "NASI•GORENG•PEDAS": {
    "59f5afee2338faeeaab3d4403d8099b511ee826e6f6b508cf767fef26c2aa91f:0": 50,
    "59f5afee2338faeeaab3d4403d8099b511ee826e6f6b508cf767fef26c2aa91f:1": 50,
    "59f5afee2338faeeaab3d4403d8099b511ee826e6f6b508cf767fef26c2aa91f:2": 50,
    "59f5afee2338faeeaab3d4403d8099b511ee826e6f6b508cf767fef26c2aa91f:3": 50,
    "59f5afee2338faeeaab3d4403d8099b511ee826e6f6b508cf767fef26c2aa91f:4": 50,
    "59f5afee2338faeeaab3d4403d8099b511ee826e6f6b508cf767fef26c2aa91f:6": 50,
  }

### EDICT_2: Divide amounts into all outputs (edict: 0 amount)
Transfer 1 inputs total 50. Set the amount to 0
tx: d9889c573ff5eb2befc28f0b45e374fe41a462b8273d5e1bdc419795e93a6152
```json
{
  "NASI•GORENG•PEDAS": {
    "59f5afee2338faeeaab3d4403d8099b511ee826e6f6b508cf767fef26c2aa91f:1": 50,
    "59f5afee2338faeeaab3d4403d8099b511ee826e6f6b508cf767fef26c2aa91f:2": 50,
    "59f5afee2338faeeaab3d4403d8099b511ee826e6f6b508cf767fef26c2aa91f:3": 50,
    "59f5afee2338faeeaab3d4403d8099b511ee826e6f6b508cf767fef26c2aa91f:4": 50,
    "59f5afee2338faeeaab3d4403d8099b511ee826e6f6b508cf767fef26c2aa91f:6": 50,
    "d9889c573ff5eb2befc28f0b45e374fe41a462b8273d5e1bdc419795e93a6152:0": 9,
    "d9889c573ff5eb2befc28f0b45e374fe41a462b8273d5e1bdc419795e93a6152:1": 9,
    "d9889c573ff5eb2befc28f0b45e374fe41a462b8273d5e1bdc419795e93a6152:2": 8,
    "d9889c573ff5eb2befc28f0b45e374fe41a462b8273d5e1bdc419795e93a6152:3": 8,
    "d9889c573ff5eb2befc28f0b45e374fe41a462b8273d5e1bdc419795e93a6152:4": 8,
    "d9889c573ff5eb2befc28f0b45e374fe41a462b8273d5e1bdc419795e93a6152:6": 8
  }
}
```
