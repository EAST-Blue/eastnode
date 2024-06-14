import secp256k1 from 'secp256k1'
import {createHash} from 'crypto'
import * as borsh from 'borsh';
import * as secp from '@noble/secp256k1';

(async () => {
    // const method = "Mutation/Query"

    // type = "call", "view", "deploy", "genesis", transfer

    // {
    //     type: "transfer",
    //     method_name: "",
    //     args: []string
    // }

    // {
    //     type: "deploy",
    //     method_name: "",
    //     args: []byte
    // }

    // {
    //     type: "function",
    //     method_name: "",
    //     args: ["hello", "world"]
    // }

    const privKey = Buffer.from("7c67c815e1c4a25fe70d95aad9440b682bdcbe6e2baf34d460966e605705ea8e", "hex")

    const pubKey = Buffer.from(secp256k1.publicKeyCreate(privKey), "hex") 

    const mainStruct = {
        struct: {
            id: 'string',
            signature: 'string',
            transaction: 'string'
        }
    }

    const transactionStruct = {
        struct: {
            signer: 'string',
            receiver: 'string',
            actions: 'string'
        }
    }

    const actionStruct = {
        struct : {
            kind: 'string',
            method_name: 'string',
            args: {
                array: {
                    type: 'string'
                }
            }
        }
    } 

    const actionsStruct =  {
        array: {
            type: actionStruct
        }
    }

    const params = {
        id: 'sha256(signature)',
        signature: 'secp256k1.ecdsaSign(transaction, privKey)',
        transaction: {
            signer: pubKey.toString("hex"),
            receiver: "abc",
            actions: [{
                kind: "call",
                method_name: "asd",
                args: ["hello", "world"]
            }]
        }
    }

    const actionsPacked = Buffer.from(borsh.serialize(actionsStruct, params.transaction.actions), "hex").toString("hex")
    // console.log(actionsPacked)

    const txPackedByte = Buffer.from(borsh.serialize(transactionStruct, {
        signer: pubKey.toString("hex"),
        receiver: 'bc1pkskdm7qk0z4gr8cgy38ysa00gyftj364gmf3uruse80c6gzunf6s0ywcsh',
        actions: actionsPacked
    }))
    // console.log(txPacked) 
    console.log(txPackedByte)

    // const signatureNoble = await (await secp.signAsync(Buffer.from("hello").toString("hex"), privKey)).toCompactHex()
    // console.log(signatureNoble)

    const hashedMsg = createHash('sha256').update(txPackedByte).digest()
    console.log(hashedMsg)

    const sigObj = secp256k1.ecdsaSign(hashedMsg, privKey)
    const signature = Buffer.from(sigObj.signature, "hex").toString("hex")
    console.log(signature)

    const txID = createHash('sha256').update(signature).digest("hex")
    console.log(txID)

    const mainPacked = Buffer.from(borsh.serialize(mainStruct, {
        id: txID,
        signature: signature,
        transaction: txPackedByte.toString("hex")
    })).toString('hex')

    console.log(mainPacked)

    // const isValid = secp.verify("a8a15ec2716b8c41c63368d1b93c562204e835414a7a764d82da70657f8164e6dd8aa567bdb8219fa0d43f5157fa189ed8435818d94a87fa11e08b4df7ffa75d", hashedMsg, pubKey.toString('hex'));
    // console.log(isValid)
    
    // console.log(Buffer.from(sigObj.signature, "hex").toString("hex"))

    // // verify the signature
    // console.log(secp256k1.ecdsaVerify(sigObj.signature, hashedMsg, pubKey))
    // console.log()
})();