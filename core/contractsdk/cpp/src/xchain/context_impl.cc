#include "xchain/context_impl.h"
#include "xchain/contract.pb.h"
#include "xchain/syscall.h"
#include <stdio.h>
#include <stdarg.h>

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
    for (int i=0; i<_call_args.args_size(); i++) {
        auto arg_pair = _call_args.args(i);
        _args.insert(std::make_pair(arg_pair.key(), arg_pair.value()));
    }
    _resp.status = 200;
    _account = Account(_call_args.initiator());

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
    return kUnknownKey;
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

    tx->init(rep.tx());

    return true;
}

bool ContextImpl::query_block(const std::string &blockid, Block* block) {
    pb::QueryBlockRequest req;
    pb::QueryBlockResponse rep;

    req.set_blockid(blockid);
    bool ok = syscall("QueryBlock", req, &rep);
    if (!ok) {
        return false;
    }

    block->init(rep.block());

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

std::unique_ptr<Iterator> ContextImpl::new_iterator(const std::string& start, const std::string& limit) {
    return std::unique_ptr<Iterator>(new Iterator(start, limit, ITERATOR_BATCH_SIZE));
}

Account& ContextImpl::sender() { return _account; }

const std::string& ContextImpl::transfer_amount() const {
    return _call_args.transfer_amount();
}

bool ContextImpl::call(const std::string& module, const std::string& contract,
                      const std::string& method,
                      const std::map<std::string, std::string>& args,
                      Response* xresponse) {
    pb::ContractCallRequest request;
    pb::ContractCallResponse response;
    request.set_module(module);
    request.set_contract(contract);
    request.set_method(method);
    for (auto it=args.begin(); it!=args.end(); it++) {
        auto arg = request.add_args();
        arg->set_key(it->first);
        arg->set_value(it->second);
    }
    bool ok = syscall("ContractCall", request, &response);
    if (!ok) {
        return false;
    }
    xresponse->status = response.response().status();
    xresponse->message = response.response().message();
    xresponse->body = response.response().body();
    return true;
}

bool ContextImpl::cross_query(const std::string& uri,
                      const std::map<std::string, std::string>& args,
                      Response* xresponse) {
    pb::CrossContractQueryRequest request;
    pb::CrossContractQueryResponse response;
    request.set_uri(uri);
    for (auto it=args.begin(); it!=args.end(); it++) {
        auto arg = request.add_args();
        arg->set_key(it->first);
        arg->set_value(it->second);
    }
    bool ok = syscall("CrossContractQuery", request, &response);
    if (!ok) {
        return false;
    }
    xresponse->status = response.response().status();
    xresponse->message = response.response().message();
    xresponse->body = response.response().body();
    return true;
}

void ContextImpl::logf(const char* fmt, ...) {
    va_list args;
    va_start(args, fmt);
    vprintf(fmt, args);
    va_end(args);
    putchar('\n');
}

}  // namespace xchain
