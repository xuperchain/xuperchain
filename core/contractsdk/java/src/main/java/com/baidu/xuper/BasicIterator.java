package com.baidu.xuper;

import com.baidu.xuper.contractpb.Contract;
import com.baidu.xuper.contractpb.SyscallGrpc;
import com.google.protobuf.ByteString;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.Iterator;

public class BasicIterator implements Iterator {
    static final int CAP = 100;

    private final SyscallGrpc.SyscallBlockingStub client;
    private final Contract.IteratorRequest.Builder builder;
    private byte[] start;
    private byte[] limit;
    private ArrayList<Contract.IteratorItem> items;
    private int it;

    public static BasicIterator newIterator(SyscallGrpc.SyscallBlockingStub client,
                                            Contract.SyscallHeader header, byte[] start, byte[] limit) {
        BasicIterator iter = new BasicIterator(client, header, start, limit);
        if (!iter.load()) {
            iter.it = -1;
        } else {
            iter.it = 0;
        }
        return iter;
    }

    private BasicIterator(SyscallGrpc.SyscallBlockingStub client, Contract.SyscallHeader header,
                          byte[] start, byte[] limit) {
        this.client = client;
        this.builder = Contract.IteratorRequest.newBuilder().setHeader(header);
        this.start = start;
        this.limit = limit;
        this.items = new ArrayList<>();
    }

    @Override
    public boolean hasNext() {
        boolean ret = end();
        if (ret) {
            return false;
        }
        return true;
    }

    @Override
    public Object next() {
        boolean ret = end();
        if (ret) {
            return null;
        }
        Contract.IteratorItem curItem = this.items.get(this.it);
        this.it++;
        if (end()) {
            if (!load()) {
                this.it = -1;
            } else {
                this.it = 0;
            }
        }

        return curItem;
    }

    @Override
    public void remove() {
    }

    private boolean load() {
        this.items.clear();
        if (Arrays.equals(this.start, this.limit)) {
            return false;
        }
        Contract.IteratorResponse iteratorResponse = rangeQuery(this.start, this.limit, CAP + 1);
        for (int i = 0; i < iteratorResponse.getItemsCount(); i++) {
            this.items.add(iteratorResponse.getItems(i));
        }
        if (!this.items.isEmpty()) {
            int len = this.items.size();
            if (len == CAP + 1) {
                this.start = this.items.get(len - 1).getKey().toByteArray();
                this.items.remove(len - 1);
            } else {
                this.start = this.limit;
            }
        }
        return true;
    }

    private boolean end() {
        return this.it >= this.items.size() || this.it < 0;
    }

    private Contract.IteratorResponse rangeQuery(byte[] start, byte[] limit, int cap) {
        Contract.IteratorRequest request = this.builder
                .setStart(ByteString.copyFrom(start)).setLimit(ByteString.copyFrom(limit)).setCap(cap).build();

        return this.client.newIterator(request);
    }

}
