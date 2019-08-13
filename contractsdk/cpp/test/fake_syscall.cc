#include <assert.h>
#include <iostream>
#include <fstream>
#include <vector>
#include <cstdlib>
#include <pthread.h>
#include <ctime>
#include <sstream>
#include "xchain/xchain.h"
#include "table/types.h"
#include "test/fake_syscall.h"

namespace xchain { namespace cdt {

static FakeContract kFakeContract;
static pthread_mutex_t kMutex;

std::string StrToHex(const std::string& data)
{
    const std::string hex = "0123456789ABCDEF";
    std::stringstream ss;

    for (std::string::size_type i = 0; i < data.size(); ++i)
        ss << hex[(unsigned char)data[i] >> 4] << hex[(unsigned char)data[i] & 0xf];
    return ss.str();
}

std::string HexToStr(const std::string& str)
{
    std::string result;
    for (size_t i = 0; i < str.length(); i += 2)
    {
        std::string byte = str.substr(i, 2);
        char chr = (char)(int)strtol(byte.c_str(), NULL, 16);
        result.push_back(chr);
    }
    return result;
}

void load_rwset(const std::string& filename, MapType* rwset) {
    std::ifstream fi(filename, std::ios::binary);
    if (!fi) {
        std::cout << "open file error when load: " << filename << std::endl;
        return;
    }
    char buf[1024];
    while(fi.getline(buf, sizeof(buf))) {
        char *np = std::strtok(buf, ",");
        np = std::strtok(NULL, ",");
        (*rwset)[buf] = HexToStr(np);
    }
    fi.close();
}

void store_rwset(const std::string& filename, const MapType& rwset) {
    std::ofstream fo(filename, std::ios::trunc & std::ios::binary);
    if (!fo) {
        std::cout << "open file error" << std::endl;
        return;
    }
    for (auto &v : rwset) {
        fo << v.first  + "," + StrToHex(v.second) << std::endl;
    }
    fo.close();
}

FakeContract::~FakeContract() {
}

FakeContract::FakeContract() {
    _syscall = {
        {"GetCallArgs" , [this](std::string args) -> std::string {
            pb::CallArgs req;
            //初始化
            req.set_method(_method);
            for (auto &v : _args) {
                auto marg = req.add_args();
                marg->set_key(v.first);
                marg->set_value(v.second);
            }
            req.set_initiator(_initiator);
            for (auto &v : _auth_require) {
                req.add_auth_require(v);
            }

            std::string value;
            bool ok = req.SerializeToString(&value);
            return value;
        }},
        {"GetObject" , [this](std::string args) -> std::string {
            pb::GetRequest req;
            assert(req.ParseFromString(args));
            pb::GetResponse resp;
            if (this->_rwset.find(req.key()) != this->_rwset.end()) {
                resp.set_value(this->_rwset[req.key()]);
            } else {
                _is_error = true;
            }
            std::string value;
            resp.SerializeToString(&value);
            return value;
        }},

        {"PutObject" , [this](std::string args) -> std::string {
            pb::PutRequest req;
            assert(req.ParseFromString(args));
            this->_rwset[req.key()] = req.value();
            pb::PutResponse resp;
            std::string value;
            resp.SerializeToString(&value);
            return value;
        }},
        {"DeleteObject" , [this](std::string args) -> std::string {
            pb::DeleteRequest req;
            assert(req.ParseFromString(args));
            this->_rwset.erase(req.key());
            pb::DeleteResponse resp;
            std::string value;
            resp.SerializeToString(&value);
            return value;
        }},
        {"NewIterator" , [this](std::string args) -> std::string {
            pb::IteratorRequest req;
            assert(req.ParseFromString(args));
            int cnt = 0;
            pb::IteratorResponse resp;
            for (auto &v : this->_rwset) {
                if (v.first.compare(req.start()) < 0) {
                    continue;
                }
                if (v.first.compare(req.limit()) > 0) {
                    continue;
                }
                auto item = resp.add_items();
                item->set_key(v.first);
                item->set_value(v.second);
                cnt += 1;

                if (cnt >= req.cap()) {
                    break;
                }
            }
            std::string value;
            resp.SerializeToString(&value);
            return value;
        }},
        {"SetOutput" , [this](std::string args) -> std::string {
            pb::SetOutputRequest req;
            assert(req.ParseFromString(args));
            this->resp = req.response();
            pb::SetOutputResponse resp;
            std::string value;
            resp.SerializeToString(&value);
            store_rwset(kFileName, _rwset);
            return value;
        }},
        {"QueryTx", [this](std::string args) -> std::string {
            pb::QueryTxRequest req;
            assert(req.ParseFromString(args));
            pb::QueryTxResponse resp;
            pb::Transaction *tx = new pb::Transaction();
            tx->set_txid("c9d3390118c509b094c6e2cf4b369d849ce2dd50f2254a54e9a9b5626d7d9422");
            tx->set_blockid("5a9266b17608dce11f84bddd9a3eae37cf36d3a4f33fd95b53e25077e6e16757");
            resp.set_allocated_tx(tx);
            std::string value;
            resp.SerializeToString(&value);
            return value;        
        }},
        {"QueryBlock", [this](std::string args) -> std::string {
            pb::QueryBlockRequest req;
            assert(req.ParseFromString(args));
            pb::QueryBlockResponse resp;
            pb::Block *block = new pb::Block();
            block->set_blockid("5a9266b17608dce11f84bddd9a3eae37cf36d3a4f33fd95b53e25077e6e16757");
            block->set_proposer("alice");
            resp.set_allocated_block(block);
            std::string value;
            resp.SerializeToString(&value);
            return value;        
        }},
    };
}

uint32_t FakeContract::call_method(const std::string& m, const std::string& r) {
    resp.set_status(0);
    _is_error = false;
    assert(_syscall.end() != _syscall.find(m));
    _buf = _syscall[m](r);
    if (_is_error) {
        resp.set_status(500);
        resp.set_message("call error");
        resp.SerializeToString(&_buf);
    }
    return _buf.size();
}
uint32_t FakeContract::fetch_response(char* res, uint32_t len) {
    std::memcpy(res, _buf.data(), len);
    return _is_error? 0 : 1;
}

void FakeContract::init(MapType rwset, std::string initiator, std::vector<std::string> auth,
        std::string m, MapType args) {
    load_rwset(kFileName, &_rwset);
    for (auto &v : rwset) {
        _rwset[v.first] = v.second;
    }
    _initiator = initiator;
    _auth_require.assign(auth.begin(), auth.end());
    _args = args;
    _method = m;
}

void ctx_init(MapType rwset, std::string initiator, std::vector<std::string> auth,
        std::string m, MapType args) {
    kFakeContract.init(rwset, initiator, auth, m, args);
}

void ctx_lock() {
    pthread_mutex_lock(&kMutex);
}

void ctx_unlock() {
    pthread_mutex_unlock(&kMutex);
}

bool ctx_assert(int status) {
    return status == kFakeContract.resp.status();
}

bool ctx_assert(int status, std::string message, std::string body) {
    return status == kFakeContract.resp.status() &&
        (message == kFakeContract.resp.message() && body == kFakeContract.resp.body());
}

extern "C" uint32_t call_method(const char* method, uint32_t method_len,
        const char* request, uint32_t request_len) {
    std::string m(method, method_len);
    std::string a(request, request_len );
    return kFakeContract.call_method(m, a);
}

extern "C" uint32_t fetch_response(char* response, uint32_t response_len) {
    auto a = kFakeContract.fetch_response(response, response_len);
    return a;
}
}}
