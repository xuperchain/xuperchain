#!/bin/env python
#coding=utf-8
import sys
import json

body1 = {
    "module": "tdpos",
    "method": "nominate_candidate",
    "args": {
        "candidate": "提名address"
         }
}

body2 = {
    "module": "tdpos",
    "method": "vote",
    "args" : {
        "candidates": ["提名过的address"]
        }
}

body3 = {
    "module":"proposal",
    "method": "Thaw",
    "args" : {
        "txid":"提名或者投票addresss时返回的txid"
        }
}

body4 = {
    "module_name": "xkernel",
    "method_name": "NewAccount",
    "args": {
        "account_name": "1559479167343170",
        "acl": "{\"pm\": {\"rule\": 1,\"acceptValue\": 0.6},\"aksWeight\": {\"dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN\": 0.5,\"TH\": 0.5}}"
    }
}

body_dict = {'./relate_file/nominate.json':body1, './relate_file/vote.json':body2, './relate_file/revoke.json':body3, './relate_file/account.json':body4}
for key in body_dict:
    body_file = json.dumps(body_dict[key], indent=4)
    f = open(key,'w')
    f.writelines(body_file)
    f.close()
