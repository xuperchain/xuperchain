#include "xchain/context_impl.h"
#include <stdio.h>
#include "xchain/contract.pb.h"
#include "xchain/xchain.pb.h"
#include "xchain/syscall.h"

namespace xchain {

ContextImpl::ContextImpl() {}

ContextImpl::~ContextImpl() {}

const std::string& ContextImpl::method() { return _call_args.method(); }

bool ContextImpl::init() {
    pb::GetCallArgsRequest req;
    bool ok = syscall("GetCallArgs", req, &_call_args);
    if (!ok) {
        return false;
    }
    _args.insert(_call_args.args().begin(), _call_args.args().end());
    _resp.status = 200;
    return true;
}

const std::map<std::string, std::string>& ContextImpl::args() const {
    return _args;
}

const std::string& ContextImpl::arg(const std::string& name) const {
    auto it = _args.find(name);
    if (it != _args.end()) {
        return it->second;
    }
    return std::move(std::string(""));
}


const std::string& ContextImpl::initiator() const {
    return _call_args.initiator();
}

int ContextImpl::auth_require_size() const {
    return _call_args.auth_require_size();
}

const std::string& ContextImpl::auth_require(int idx) const {
    return _call_args.auth_require(idx);
}

bool ContextImpl::get_object(const std::string& key, std::string* value) {
    pb::GetRequest req;
    pb::GetResponse rep;
    req.set_key(key);
    bool ok = syscall("GetObject", req, &rep);
    if (!ok) {
        return false;
    }
    *value = rep.value();
    return true;
}

bool ContextImpl::put_object(const std::string& key, const std::string& value) {
    pb::PutRequest req;
    pb::PutResponse rep;
    req.set_key(key);
    req.set_value(value);
    bool ok = syscall("PutObject", req, &rep);
    if (!ok) {
        return false;
    }
    return true;
}

bool ContextImpl::delete_object(const std::string& key) {
    pb::DeleteRequest req;
    pb::DeleteResponse rep;
    req.set_key(key);
    bool ok = syscall("DeleteObject", req, &rep);
    if (!ok) {
        return false;
    }
    return true;
}

bool ContextImpl::query_tx(const std::string &txid, Transaction* tx) {
    pb::QueryTxRequest req;
    pb::QueryTxResponse rep;
    req.set_txid(txid);
    bool ok = syscall("QueryTx", req, &rep);
    if (!ok) {
        return false;
    }

    pb::Transaction* pbtx = new pb::Transaction();
    if (!pbtx->ParseFromString(rep.tx())) {
        return  false;
    }

    tx->init(pbtx);

    return true;
}

void ContextImpl::ok(const std::string& body) {
    _resp.status = 200;
    _resp.body = body;
}

void ContextImpl::error(const std::string& body) {
    _resp.status = 500;
    _resp.message = body;
}

Response* ContextImpl::mutable_response() { return &_resp; }

const Response& ContextImpl::get_response() { return _resp; }
}  // namespace xchain
