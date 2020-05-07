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
// input format: {"l":"key_l", "r":"key_r", "commitment": c1,
//                "commitment2": c2}
DEFINE_METHOD(Counter, add) {
  xchain::Context *ctx = self.context();
  if (ctx->arg("l").empty() || ctx->arg("commitment").empty()) {
      ctx->error("missing left operand parameter ");
      return;
  }
  if (ctx->arg("r").empty() || ctx->arg("commitment2").empty()) {
      ctx->error("missing right operand parameter ");
      return;
  }

  TrustOperators to(ctx, 0);
  std::string value;
  TrustOperators::Operand left_op, right_op;
  // get left operand
  if (!ctx->get_object(ctx->arg("l"), &value)) {
      ctx->error("get left operand error ");
      return;
  }
  left_op.cipher = value;
  left_op.commitment = ctx->arg("commitment");
  // get right operand
  if (!ctx->get_object(ctx->arg("r"), &value)) {
      ctx->error("get right operand error ");
      return;
  }
  right_op.cipher = ctx->arg("r");
  right_op.commitment = ctx->arg("commitment2");

  // call trust operator to add, return {{"key", enc_result}}
  std::map<std::string, std::string> result;
  auto ok = to.add(left_op, right_op, &result);
  if (ok) {
    if (result["key"] == "") {
        ctx->error("error");
        return;
    }
    if (!ctx->put_object(ctx->arg("o"), result["key"])) {
        ctx->error("error");
        return;
    }
    ctx->ok("done");
    return;
  }
  ctx->error("error");
}

// sub substracts one number from another
// input format: {"l":"key_l", "r":"key_r", "commitment": c1,
//                "commitment2": c2}
DEFINE_METHOD(Counter, sub) {
  xchain::Context *ctx = self.context();
  if (ctx->arg("l").empty() || ctx->arg("commitment").empty()) {
      ctx->error("missing left operand parameter ");
      return;
  }
  if (ctx->arg("r").empty() || ctx->arg("commitment2").empty()) {
      ctx->error("missing right operand parameter ");
      return;
  }

  TrustOperators to(ctx, 0);
  std::string value;
  TrustOperators::Operand left_op, right_op;
  // get left operand
  if (!ctx->get_object(ctx->arg("l"), &value)) {
      ctx->error("get left operand error ");
      return;
  }
  left_op.cipher = value;
  left_op.commitment = ctx->arg("commitment");
  // get right operand
  if (!ctx->get_object(ctx->arg("r"), &value)) {
      ctx->error("get right operand error ");
      return;
  }
  right_op.cipher = ctx->arg("r");
  right_op.commitment = ctx->arg("commitment2");

  // call trust operator to sub, return {{"key", enc_result}}
  std::map<std::string, std::string> result;
  auto ok = to.sub(left_op, right_op, &result);
  if (ok) {
    if (result["key"] == "") {
        ctx->error("error");
        return;
    }
    if (!ctx->put_object(ctx->arg("o"), result["key"])) {
        ctx->error("error");
        return;
    }
    ctx->ok("done");
    return;
  }
  ctx->error("error");
}

// mul multiplies two encrypted numbers
// input format: {"l":"key_l", "r":"key_r", "commitment": c1,
//                "commitment2": c2}
DEFINE_METHOD(Counter, mul) {
  xchain::Context *ctx = self.context();
  if (ctx->arg("l").empty() || ctx->arg("commitment").empty()) {
      ctx->error("missing left operand parameter ");
      return;
  }
  if (ctx->arg("r").empty() || ctx->arg("commitment2").empty()) {
      ctx->error("missing right operand parameter ");
      return;
  }

  TrustOperators to(ctx, 0);
  std::string value;
  TrustOperators::Operand left_op, right_op;
  // get left operand
  if (!ctx->get_object(ctx->arg("l"), &value)) {
      ctx->error("get left operand error ");
      return;
  }
  left_op.cipher = value;
  left_op.commitment = ctx->arg("commitment");
  // get right operand
  if (!ctx->get_object(ctx->arg("r"), &value)) {
      ctx->error("get right operand error ");
      return;
  }
  right_op.cipher = ctx->arg("r");
  right_op.commitment = ctx->arg("commitment2");

  // call trust operator to mul, return {{"key", enc_result}}
  std::map<std::string, std::string> result;
  auto ok = to.mul(left_op, right_op, &result);
  if (ok) {
    if (result["key"] == "") {
        ctx->error("error");
        return;
    }
    if (!ctx->put_object(ctx->arg("o"), result["key"])) {
        ctx->error("error");
        return;
    }
    ctx->ok("done");
    return;
  }
  ctx->error("error");
}