#include "xchain/trust_operators/trust_operators.h"
#include "xchain/xchain.h"
#include <iostream>

struct Counter : public xchain::Contract {};

DEFINE_METHOD(Counter, initialize) {
  xchain::Context *ctx = self.context();
  const std::string &creator = ctx->arg("creator");
  if (creator.empty()) {
    ctx->error("missing creator");
    return;
  }
  ctx->put_object("creator", creator);
  ctx->ok("initialize succeed");
}

// get a number by key
DEFINE_METHOD(Counter, get) {
  xchain::Context *ctx = self.context();
  const std::string &key = ctx->arg("key");
  std::string value;
  if (ctx->get_object(key, &value)) {
    ctx->ok(value);
  } else {
    ctx->error("key not found");
  }
}

// store saves encrypted data directly
DEFINE_METHOD(Counter, store) {
  xchain::Context *ctx = self.context();
  std::string debug = "done";
  for (auto it = ctx->args().begin(); it != ctx->args().end(); ++it) {
    // args are already encrypted, just put
    auto ok = ctx->put_object(it->first, it->second);
    if (!ok) {
      debug = "error";
    }
  }
  ctx->ok(debug);
}

// add adds two encrypted numbers
// input is in format {"l":"key_l", "r":"key_r", "o":"key_output",
//                     "commitment": c1, "commitment2": c2}
DEFINE_METHOD(Counter, add) {
  xchain::Context *ctx = self.context();
  TrustOperators to(ctx, 0);

  std::map<std::string, std::string> args_map;
  for (auto it = ctx->args().begin(); it != ctx->args().end(); ++it) {
    std::string value;
    // get args map
    if (it->first == "l" || it->first == "r") {
      if (!ctx->get_object(it->second, &value)) {
        ctx->error("get object error");
        return;
      }
      args_map[it->first] = value;
    } else {
      args_map[it->first] = it->second;
    }
  }

  std::string result;
  auto ok = to.add(args_map["l"], args_map["r"], args_map["o"],
                   args_map["commitment"], args_map["commitment2"], &result);
  if (ok) {
    if(!ctx->put_object(args_map["o"], result)) {
        ctx->ok("error");
        return;
    }
    ctx->ok("done");
    return;
  }
  ctx->ok("error");
}

// sub substracts one number from another
// input is in format {"l":"key_l", "r":"key_r", "o":"key_output",
//                     "commitment": c1, "commitment2": c2}
DEFINE_METHOD(Counter, sub) {
  xchain::Context *ctx = self.context();
  TrustOperators to(ctx, 0);

  std::map<std::string, std::string> args_map;
  for (auto it = ctx->args().begin(); it != ctx->args().end(); ++it) {
    std::string value;
    // get args map
    if (it->first == "l" || it->first == "r") {
      if (!ctx->get_object(it->second, &value)) {
        ctx->error("get object error");
        return;
      }
      args_map[it->first] = value;
    } else {
      args_map[it->first] = it->second;
    }
  }

  std::string result;
  auto ok = to.sub(args_map["l"], args_map["r"], args_map["o"],
                   args_map["commitment"], args_map["commitment2"], &result);
  if (ok) {
    if(!ctx->put_object(args_map["o"], result)) {
        ctx->ok("error");
        return;
    }
    ctx->ok("done");
    return;
  }
  ctx->ok("error");
}

// mul multiplies two encrypted numbers
// input is in format {"l":"key_l", "r":"key_r", "o":"key_output"}
//                     "commitment": c1, "commitment2": c2}
DEFINE_METHOD(Counter, mul) {
  xchain::Context *ctx = self.context();
  TrustOperators to(ctx, 0);

  std::map<std::string, std::string> args_map;
  for (auto it = ctx->args().begin(); it != ctx->args().end(); ++it) {
    std::string value;
    // get args map
    if (it->first == "l" || it->first == "r") {
      if (!ctx->get_object(it->second, &value)) {
        ctx->error("get object error");
        return;
      }
      args_map[it->first] = value;
    } else {
      args_map[it->first] = it->second;
    }
  }

  std::string result;
  auto ok = to.mul(args_map["l"], args_map["r"], args_map["o"],
                   args_map["commitment"], args_map["commitment2"], &result);
  if (ok) {
    if(!ctx->put_object(args_map["o"], result)) {
        ctx->ok("error");
        return;
    }
    ctx->ok("done");
    return;
  }
  ctx->ok("error");
}
