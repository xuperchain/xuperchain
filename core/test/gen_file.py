#!/bin/env python
# coding=utf-8
import sys
import json

body1 = {
    "module": "tdpos",
    "method": "nominate_candidate",
    "args": {
        "candidate": "提名address",
        "neturl": "提名neturl"
    }
}

body2 = {
    "module": "tdpos",
    "method": "vote",
    "args": {
        "candidates": ["提名过的address"]
    }
}

body3 = {
    "module": "proposal",
    "method": "Thaw",
    "args": {
        "txid": "提名或者投票addresss时返回的txid"
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

body5 = {
    "version": "1",
    "predistribution": [
        {
            "address": "dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN",
            "quota": "100000000000000000000"
        }
    ],
    "maxblocksize": "128",
    "award": "1000000",
    "decimals": "8",
    "award_decay": {
        "height_gap": 31536000,
        "ratio": 1
    },
    "gas_price": {
        "cpu_rate": 1000,
        "mem_rate": 1000000,
        "disk_rate": 1,
        "xfee_rate": 1
    },
    "new_account_resource_amount": 1000,
    "genesis_consensus": {
        "name": "tdpos",
        "config": {
            "timestamp": "1559021720000000000",
            "proposer_num": "1",
            "period": "3000",
            "alternate_interval": "3000",
            "term_interval": "6000",
            "block_num": "20",
            "vote_unit_price": "1",
            "init_proposer": {
                "1": ["dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN", "nodeaddress"]
            },
            "init_proposer_neturl": {
                "1": ["/ip4/127.0.0.1/tcp/47101/p2p/QmVxeNubpg1ZQjQT8W5yZC9fD7ZB1ViArwvyGUB53sqf8e", "nodeneturl"]
            }
        }
    }
}

body_dict = {'./relate_file/nominate.json': body1, './relate_file/vote.json': body2,
             './relate_file/revoke.json': body3, './relate_file/account.json': body4, './relate_file/xuper.json': body5}
for key in body_dict:
    body_file = json.dumps(body_dict[key], indent=4)
    f = open(key, 'w')
    f.writelines(body_file)
    f.close()
