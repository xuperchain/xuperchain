package com.baidu.xuper;

import com.baidu.xuper.contractpb.Contract;

public class ContractIteratorItem {
    private com.baidu.xuper.contractpb.Contract.IteratorItem item;

    public ContractIteratorItem(Contract.IteratorItem item) {
        this.item = item;
    }

    public byte[] getKey() {
        return this.item.getKey().toByteArray();
    }

    public byte[] getValue() {
        return this.item.getValue().toByteArray();
    }
}