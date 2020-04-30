#include "xchain/trust_operators/trust_operators.h"
#include "xchain/trust_operators/tf.pb.h"

extern "C" uint32_t xvm_tfcall(const char *inputptr, size_t inputlen,
                               char **outputptr, size_t *outputlen);

bool tfcall(const TrustFunctionCallRequest &request,
            TrustFunctionCallResponse *response) {
  std::string req;
  char *out = nullptr;
  size_t out_size = 0;
  request.SerializeToString(&req);
  auto ok = xvm_tfcall(req.data(), req.size(), &out, &out_size);
  // 注意是返回1表示失败
  if (ok) {
    return false;
  }
  auto ret = response->ParseFromString(std::string(out, out_size));
  free(out);
  return ret;
}

TrustOperators::TrustOperators(xchain::Context *ctx, const uint32_t svn)
    : _ctx(ctx), _svn(svn) {}

// map_to_string convert a <string, string>map to a string
std::string map_to_string(std::map<std::string, std::string> str_map) {
  std::map<std::string, std::string>::iterator it;
  std::string str = "{";
  for (it = str_map.begin(); it != str_map.end(); ++it) {
    if (it != str_map.begin()) {
      str = str + ",";
    }
    str = str + '"' + it->first + '"';
    str = str + ":";
    str = str + '"' + it->second + '"';
  }
  str = str + "}";
  return str;
}

/*
    ops supports encrypted data operations;
    op is one of {add, sub, mul};
    left_op is left operand(cipher1 | commitment1), right_op is right
    operand(cipher2 | commitment2).
*/
bool TrustOperators::binary_ops(const std::string op, const operand &left_op,
                                const operand &right_op, std::string *result) {
  TrustFunctionCallRequest req;
  req.set_method(op);
  std::map<std::string, std::string> args_map;
  args_map = {{"l", left_op.cipher},
              {"r", right_op.cipher},
              {"o", "key"},
              {"commitment", left_op.commitment},
              {"commitment2", right_op.commitment}};
  req.set_args(map_to_string(args_map));

  req.set_svn(_svn);
  req.set_address(_ctx->initiator());
  TrustFunctionCallResponse resp;
  if (!tfcall(req, &resp)) {
    return false;
  }
  assert(resp.has_kvs());
  // tfcall only returns one kv pair {"key": encrypted_result}
  if (resp.kvs().kv(0).key() == "key") {
    *result = resp.kvs().kv(0).value();
    return true;
  }
  return false;
}

bool TrustOperators::add(const operand &left_op, const operand &right_op,
                         std::string *result) {
  return binary_ops("add", left_op, right_op, result);
}

bool TrustOperators::sub(const operand &left_op, const operand &right_op,
                         std::string *result) {
  return binary_ops("sub", left_op, right_op, result);
}

bool TrustOperators::mul(const operand &left_op, const operand &right_op,
                         std::string *result) {
  return binary_ops("mul", left_op, right_op, result);
}

// kind = "commitment" -> authorize a user to use data, return a commitment.
// kind = "ownership"  -> share data to a user, return re-encrypted data.
bool TrustOperators::authorize(const auth_info &auth, std::string *result) {
  TrustFunctionCallRequest req;
  req.set_method("authorize");
  std::map<std::string, std::string> args_map;
  args_map = {{"ciphertext", auth.data}, {"to", auth.to}, {"kind", auth.kind}};

  req.set_args(map_to_string(args_map));
  req.set_svn(_svn);
  req.set_address(_ctx->initiator());
  req.set_publickey(auth.pubkey);
  req.set_signature(auth.signature);
  TrustFunctionCallResponse resp;
  if (!tfcall(req, &resp)) {
    return false;
  }

  assert(resp.has_kvs());

  // kind == "commitment"
  // 返回{{"commitment": cm}}
  if (auth.kind == "commitment") {
    if (resp.kvs().kv(0).key() == "commitment") {
      *result = resp.kvs().kv(0).value();
      return true;
    }
    return false;
  }

  // kind == "ownership"
  // 返回{{"cipher": c}}
  if (resp.kvs().kv(0).key() == "cipher") {
    *result = resp.kvs().kv(0).value();
    return true;
  }
  return false;
}
