package binance

// https://testnet-dex.binance.org/api/v1/tx/A8FCD14430FDA557C5744ECC18AA9C9704B739E31FA6FA8328FDD8206F2F47EF?format=json
const binanceTxNewOrder = `
{
 "jsonrpc": "2.0",
 "id": "",
 "result": {
   "txs": [
     {
       "hash": "A8FCD14430FDA557C5744ECC18AA9C9704B739E31FA6FA8328FDD8206F2F47EF",
       "height": "35559022",
       "index": 0,
       "tx_result": {
         "data": "eyJvcmRlcl9pZCI6IkU5M0FGQTI1MUY1QTFFRUJCOEUzNzNEREM1MDM5NDMwMjY5NEEyMjAtMTA5MzU1In0=",
         "log": "Msg 0: ",
         "tags": [
           {
             "key": "YWN0aW9u",
             "value": "b3JkZXJOZXc="
           }
         ]
       },
       "tx": "4QHwYl3uCmfObcBDChTpOvolH1oe67jjc93FA5QwJpSiIBIvRTkzQUZBMjUxRjVBMUVFQkI4RTM3M0REQzUwMzk0MzAyNjk0QTIyMC0xMDkzNTUaC1pDQi1GMDBfQk5CIAIoATCBRTiAqNa5B0ABEnIKJuta6YchAv+Fh/6SCT/nXtoa/c4s1p3cvs0B6EH24jX94NocvGBdEkAkCKvNMD184fFCCK7HQ+BaRQ5NBmW6c7x3Ur2UL6MNswIqN/X9+ZTvRms151aF9speNnyNYZNDrmOrUoyIj8cAGPanKiCq1gY=",
       "proof": {
         "RootHash": "8E6F9DC69873E9F12AC1D84B7D25ED27039924663430042138B3CAA91584E9F6",
         "Data": "4QHwYl3uCmfObcBDChTpOvolH1oe67jjc93FA5QwJpSiIBIvRTkzQUZBMjUxRjVBMUVFQkI4RTM3M0REQzUwMzk0MzAyNjk0QTIyMC0xMDkzNTUaC1pDQi1GMDBfQk5CIAIoATCBRTiAqNa5B0ABEnIKJuta6YchAv+Fh/6SCT/nXtoa/c4s1p3cvs0B6EH24jX94NocvGBdEkAkCKvNMD184fFCCK7HQ+BaRQ5NBmW6c7x3Ur2UL6MNswIqN/X9+ZTvRms151aF9speNnyNYZNDrmOrUoyIj8cAGPanKiCq1gY=",
         "Proof": {
           "total": "1",
           "index": "0",
           "leaf_hash": "jm+dxphz6fEqwdhLfSXtJwOZJGY0MAQhOLPKqRWE6fY=",
           "aunts": []
         }
       }
     }
   ],
   "total_count": "1"
 }
}`

// https://testnet-dex.binance.org/api/v1/tx/10C4E872A5DC842BE72AC8DE9C6A13F97DF6D345336F01B87EBA998F5A3BC36D?format=json
const binanceTxTransferWithdraw = `
{
 "jsonrpc": "2.0",
 "id": "",
 "result": {
   "txs": [
     {
       "hash": "10C4E872A5DC842BE72AC8DE9C6A13F97DF6D345336F01B87EBA998F5A3BC36D",
       "height": "35345060",
       "index": 0,
       "tx_result": {
         "log": "Msg 0: ",
         "tags": [
           {
             "key": "c2VuZGVy",
             "value": "dGJuYjFnZ2RjeWhrOHJjN2ZnenA4d2Eyc3UyMjBhY2xjZ2djc2Q5NHllNQ=="
           },
           {
             "key": "cmVjaXBpZW50",
             "value": "dGJuYjF5eWNuNG1oNmZmd3BqZjU4NHQ4bHBwN2MyN2dodTAzZ3B2cWtmag=="
           },
           {
             "key": "YWN0aW9u",
             "value": "c2VuZA=="
           }
         ]
       },
       "tx": "3gHwYl3uClYqLIf6CicKFEIbgl7HHjyUCCd3VQ4pT+4/hCMQEg8KCFJVTkUtQTFGEIDC1y8SJwoUITE67vpKXBkmh6rP8IfYV5F+PigSDwoIUlVORS1BMUYQgMLXLxJwCibrWumHIQOki6+6K5zhbjAndqURWmVv5ZVY+ePXfi/DxUTzcenLWhJAUr5kAtjMfsb+IO+7ligNJRXhpL8WZLkH0IIWeQ2Cb4xEcN8ANIVgKjzU6IQYOKnNYpoCpMWQJTYXFg+Q95ztCBiSsyogFRoMd2l0aGRyYXc6Qk5CIAE=",
       "proof": {
         "RootHash": "A06D7798436C26BAF00177873C901C8A2337F8B0C18A75AAA9D86D615BE24938",
         "Data": "3gHwYl3uClYqLIf6CicKFEIbgl7HHjyUCCd3VQ4pT+4/hCMQEg8KCFJVTkUtQTFGEIDC1y8SJwoUITE67vpKXBkmh6rP8IfYV5F+PigSDwoIUlVORS1BMUYQgMLXLxJwCibrWumHIQOki6+6K5zhbjAndqURWmVv5ZVY+ePXfi/DxUTzcenLWhJAUr5kAtjMfsb+IO+7ligNJRXhpL8WZLkH0IIWeQ2Cb4xEcN8ANIVgKjzU6IQYOKnNYpoCpMWQJTYXFg+Q95ztCBiSsyogFRoMd2l0aGRyYXc6Qk5CIAE=",
         "Proof": {
           "total": "1",
           "index": "0",
           "leaf_hash": "oG13mENsJrrwAXeHPJAciiM3+LDBinWqqdhtYVviSTg=",
           "aunts": []
         }
       }
     }
   ],
   "total_count": "1"
 }
}`

// https://testnet-dex.binance.org/api/v1/tx/523546F263ABA7BDDFFEE82B9A362D0B8BD4F114D58880CF78A77D4D43E7847A?format=json
const binanceTxOutboundFromPool = `
{
 "jsonrpc": "2.0",
 "id": "",
 "result": {
   "txs": [
     {
       "hash": "523546F263ABA7BDDFFEE82B9A362D0B8BD4F114D58880CF78A77D4D43E7847A",
       "height": "35340678",
       "index": 0,
       "tx_result": {
         "log": "Msg 0: ",
         "tags": [
           {
             "key": "c2VuZGVy",
             "value": "dGJuYjF5eWNuNG1oNmZmd3BqZjU4NHQ4bHBwN2MyN2dodTAzZ3B2cWtmag=="
           },
           {
             "key": "cmVjaXBpZW50",
             "value": "dGJuYjFnZ2RjeWhrOHJjN2ZnenA4d2Eyc3UyMjBhY2xjZ2djc2Q5NHllNQ=="
           },
           {
             "key": "YWN0aW9u",
             "value": "c2VuZA=="
           }
         ]
       },
       "tx": "3gHwYl3uClgqLIf6CigKFCExOu76SlwZJoeqz/CH2FeRfj4oEhAKCFJVTkUtQTFGEICMjZ4CEigKFEIbgl7HHjyUCCd3VQ4pT+4/hCMQEhAKCFJVTkUtQTFGEICMjZ4CEnAKJuta6YchA4qGFSnfOnMFBASpOdYfdpTguZhKJaMZxir4RDzHeb6VEkCoKLgBJFSqxAJxwpeLxumNlKfj3Qtc4V+GVnGooRr/rmKCewweZ5Wc7xT3DqSdkB1oo169zcU5tYpVZm5hmwqJGIe5KiAKGgxPVVRCT1VORDo5NDY=",
       "proof": {
         "RootHash": "1EC05BB121F24DB3E4F04A2EC92710896218B614E94629D4443D3B05065ED46C",
         "Data": "3gHwYl3uClgqLIf6CigKFCExOu76SlwZJoeqz/CH2FeRfj4oEhAKCFJVTkUtQTFGEICMjZ4CEigKFEIbgl7HHjyUCCd3VQ4pT+4/hCMQEhAKCFJVTkUtQTFGEICMjZ4CEnAKJuta6YchA4qGFSnfOnMFBASpOdYfdpTguZhKJaMZxir4RDzHeb6VEkCoKLgBJFSqxAJxwpeLxumNlKfj3Qtc4V+GVnGooRr/rmKCewweZ5Wc7xT3DqSdkB1oo169zcU5tYpVZm5hmwqJGIe5KiAKGgxPVVRCT1VORDo5NDY=",
         "Proof": {
           "total": "1",
           "index": "0",
           "leaf_hash": "HsBbsSHyTbPk8EouyScQiWIYthTpRinURD07BQZe1Gw=",
           "aunts": []
         }
       }
     }
   ],
   "total_count": "1"
 }
}`

const binanceTxOutboundFromPool1 = `
{
  "jsonrpc": "2.0",
  "id": "",
  "result": {
    "txs": [
      {
        "hash": "2C199678C1C33CF324DD99E373D5DF9437FBD2BA49E43E35EBB5B0F29180D93F",
        "height": "35339328",
        "index": 0,
        "tx_result": {
          "log": "Msg 0: ",
          "tags": [
            {
              "key": "c2VuZGVy",
              "value": "dGJuYjF5eWNuNG1oNmZmd3BqZjU4NHQ4bHBwN2MyN2dodTAzZ3B2cWtmag=="
            },
            {
              "key": "cmVjaXBpZW50",
              "value": "dGJuYjFnZ2RjeWhrOHJjN2ZnenA4d2Eyc3UyMjBhY2xjZ2djc2Q5NHllNQ=="
            },
            {
              "key": "YWN0aW9u",
              "value": "c2VuZA=="
            }
          ]
        },
        "tx": "3gHwYl3uClgqLIf6CigKFCExOu76SlwZJoeqz/CH2FeRfj4oEhAKCFJVTkUtQTFGEICG2PAkEigKFEIbgl7HHjyUCCd3VQ4pT+4/hCMQEhAKCFJVTkUtQTFGEICG2PAkEnAKJuta6YchA4qGFSnfOnMFBASpOdYfdpTguZhKJaMZxir4RDzHeb6VEkBPCTCewhYLFrS4MD90owM8zRfvBiQaR03HvqX2b9pYyQOXgzNnmy0aYhL4BY/IFHd6Zl8FpgI7pEqP8Ybn6FCSGIe5KiAJGgxPVVRCT1VORDo4MjU=",
        "proof": {
          "RootHash": "DE6B52EF102C6F18F5F417C3B2EA3DED324437AA0F85E3508C2DFC2EE0A97927",
          "Data": "3gHwYl3uClgqLIf6CigKFCExOu76SlwZJoeqz/CH2FeRfj4oEhAKCFJVTkUtQTFGEICG2PAkEigKFEIbgl7HHjyUCCd3VQ4pT+4/hCMQEhAKCFJVTkUtQTFGEICG2PAkEnAKJuta6YchA4qGFSnfOnMFBASpOdYfdpTguZhKJaMZxir4RDzHeb6VEkBPCTCewhYLFrS4MD90owM8zRfvBiQaR03HvqX2b9pYyQOXgzNnmy0aYhL4BY/IFHd6Zl8FpgI7pEqP8Ybn6FCSGIe5KiAJGgxPVVRCT1VORDo4MjU=",
          "Proof": {
            "total": "1",
            "index": "0",
            "leaf_hash": "3mtS7xAsbxj19BfDsuo97TJEN6oPheNQjC38LuCpeSc=",
            "aunts": []
          }
        }
      }
    ],
    "total_count": "1"
  }
}`

const binanceTxSwapLOKToBNB = `{
  "jsonrpc": "2.0",
  "id": "",
  "result": {
    "txs": [
      {
        "hash": "126634C328357071366189B779ADC1C085027B7CCC101996E977B6C4D277F781",
        "height": "42677241",
        "index": 0,
        "tx_result": {
          "log": "Msg 0: ",
          "tags": [
            {
              "key": "c2VuZGVy",
              "value": "dGJuYjE5MHRncDV1Y2hubGNwc2s3bjduZmZ5cGt3bHpoY3FnZTI3eGtmaA=="
            },
            {
              "key": "cmVjaXBpZW50",
              "value": "dGJuYjFoendmazZ0M3NxamZ1emxyMHVyOWxqOTIwZ3MzN2dnOTJndGF5OQ=="
            },
            {
              "key": "YWN0aW9u",
              "value": "c2VuZA=="
            }
          ]
        },
        "tx": "2AHwYl3uClQqLIf6CiYKFCvWgNOYvP+Awt6fppSQNnfFfAEZEg4KB0xPSy0zQzAQgMLXLxImChS4nJtpcYAkngvjfwZfyKp6IR8hBRIOCgdMT0stM0MwEIDC1y8Scgom61rphyECjAsBDgI5+O5rBODfcFfk/3BVio3yGn8Ksz/Bz7RXDlwSQMsmMcDp8efpc3yA5njHw8+qgX2KIuYhSQO6F4qgLNAuXlS0Cn+HxtFOGwDE00Kh6TrChWMnumXmqb1mDEUzC+8Y1LYqIJnBARoIU1dBUDpCTkI=",
        "proof": {
          "RootHash": "A6A71574E494AC3294101AFF54006FEBB0BF5EAA34ACDDE891168886292AFED6",
          "Data": "2AHwYl3uClQqLIf6CiYKFCvWgNOYvP+Awt6fppSQNnfFfAEZEg4KB0xPSy0zQzAQgMLXLxImChS4nJtpcYAkngvjfwZfyKp6IR8hBRIOCgdMT0stM0MwEIDC1y8Scgom61rphyECjAsBDgI5+O5rBODfcFfk/3BVio3yGn8Ksz/Bz7RXDlwSQMsmMcDp8efpc3yA5njHw8+qgX2KIuYhSQO6F4qgLNAuXlS0Cn+HxtFOGwDE00Kh6TrChWMnumXmqb1mDEUzC+8Y1LYqIJnBARoIU1dBUDpCTkI=",
          "Proof": {
            "total": "1",
            "index": "0",
            "leaf_hash": "pqcVdOSUrDKUEBr/VABv67C/Xqo0rN3okRaIhikq/tY=",
            "aunts": []
          }
        }
      }
    ],
    "total_count": "1"
  }
}`
