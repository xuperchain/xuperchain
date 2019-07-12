#ifndef XCHAIN_TRANSACTION_H
#define XCHAIN_TRANSACTION_H

#include "xchain/contract.pb.h"

namespace pb = xchain::contract::sdk;

namespace xchain {

struct TxInput {
    std::string ref_txid;
    int32_t ref_offset;
    std::string from_addr;
    std::string amount;

    TxInput(std::string reftxid, int32_t refoffset, std::string fromaddr, std::string amou)  
        : ref_txid(std::move(reftxid)), ref_offset(refoffset), 
          from_addr(std::move(fromaddr)), amount(std::move(amou))
    {    
    }

    TxInput(const TxInput& other)
        : ref_txid(std::move(other.ref_txid)), ref_offset(other.ref_offset), 
          from_addr(std::move(other.from_addr)), amount(std::move(other.amount))
    {
    }

    TxInput(TxInput&& other)
        : ref_txid(std::move(other.ref_txid)), ref_offset(other.ref_offset), 
          from_addr(std::move(other.from_addr)), amount(std::move(other.amount))
    {
    } 

    TxInput& operator=(const TxInput& other);  
};

struct TxOutput {
    std::string amount;
    std::string to_addr;

    TxOutput(std::string amou, std::string toaddr)  
        : amount(std::move(amou)), to_addr(std::move(toaddr)) 
    {    
    }

    TxOutput(const TxOutput& other)
        : amount(std::move(other.amount)), to_addr(std::move(other.to_addr))
    {
    }

    TxOutput(TxOutput&& other)
        : amount(std::move(other.amount)), to_addr(std::move(other.to_addr))
    {
    } 

    TxOutput& operator=(const TxOutput& other);  

};

class Transaction {

public:
    Transaction();
    virtual ~Transaction();
    void init(pb::Transaction pbtx);

public:
    std::string txid;
    std::string blockid;
    std::string desc;
    std::string initiator;
    std::vector<std::string> auth_require;
    std::vector<TxInput> tx_inputs;
    std::vector<TxOutput> tx_outputs;
};

}  // namespace xchain

#endif
