package com.baidu.xuper;

import java.util.ArrayList;
import java.util.List;

public class PrefixRange {
    public static byte[] generateLimit(byte[] prefix) {
        int len = prefix.length;
        List<Byte> limitList = new ArrayList<>();
        for (int i = len - 1; i >= 0; i--) {
            byte c = prefix[i];
            if (c < 0x7f) {
                for (int j = 0; j < i; j++) {
                    limitList.add(prefix[j]);
                }
                limitList.add((byte) (c + 1));
                break;
            }
        }
        byte[] limit = new byte[limitList.size()];
        for (int i = 0; i < limitList.size(); i++) {
            limit[i] = limitList.get(i);
        }
        return limit;
    }
}
