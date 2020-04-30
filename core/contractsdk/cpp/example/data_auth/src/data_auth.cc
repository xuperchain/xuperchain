#include "data_auth.pb.h"
#include "xchain/table/table.tpl.h"
#include "xchain/table/types.h"
#include "xchain/trust_operators/trust_operators.h"
#include "xchain/xchain.h"
#include <inttypes.h>

struct DataAuth : public xchain::Contract {
public:
  DataAuth() : _data(this->context(), "data_auth") {}

  // get record by dataid and user address
  struct data : public data_auth::Data {
    DEFINE_ROWKEY(dataid, user);
    DEFINE_INDEX_BEGIN(0)
    DEFINE_INDEX_END();
  };

private:
  xchain::cdt::Table<data> _data;

public:
  decltype(_data) &get_data() { return _data; }
};

DEFINE_METHOD(DataAuth, initialize) {
  xchain::Context *ctx = self.context();
  ctx->ok("initialize succeed");
}

// store encrypted data with content and expire_time
DEFINE_METHOD(DataAuth, store) {
  xchain::Context *ctx = self.context();
  const std::string &dataid = ctx->arg("dataid");
  const std::string &content = ctx->arg("content");
  const std::string &expire = ctx->arg("expire");

  DataAuth::data dat;
  dat.set_dataid(std::stoll(dataid));
  dat.set_owner(ctx->initiator());
  dat.set_content(content);
  dat.set_expire(std::stoll(expire));
  dat.set_user(ctx->initiator());
  auto debug = self.get_data().put(dat);
  if (debug == false) {
    ctx->error("failed to store " + dataid);
    return;
  }
  ctx->ok("done");
}

// get a record from table
DEFINE_METHOD(DataAuth, get) {
  xchain::Context *ctx = self.context();
  const std::string &dataid = ctx->arg("dataid");
  const std::string &user = ctx->initiator();
  DataAuth::data dat;
  if (!self.get_data().find({{"dataid", dataid}, {"user", user}}, &dat)) {
    ctx->error("cannot find " + dataid);
    return;
  }
  // TODO: check if data has expired
  ctx->ok(dat.content());
}

// modify user's one record
DEFINE_METHOD(DataAuth, modify) {
  xchain::Context *ctx = self.context();
  const std::string &dataid = ctx->arg("dataid");
  const std::string &content = ctx->arg("content");
  const std::string &expire = ctx->arg("expire");
  DataAuth::data dat;
  if (!self.get_data().find({{"dataid", dataid}, {"user", ctx->initiator()}},
                            &dat)) {
    ctx->error("cannot find " + dataid);
    return;
  }
  dat.set_content(content);
  dat.set_expire(std::stoll(expire));
  auto debug = self.get_data().update(dat);
  if (debug == false) {
    ctx->error("failed to modify " + dataid);
    return;
  }
  ctx->ok("done");
}

// delete user's record
// TODO: owner can delete every owned record
DEFINE_METHOD(DataAuth, del) {
  xchain::Context *ctx = self.context();
  const std::string &dataid = ctx->arg("dataid");
  DataAuth::data dat;
  if (!self.get_data().find({{"dataid", dataid}, {"user", ctx->initiator()}},
                            &dat)) {
    ctx->error("cannot find " + dataid);
    return;
  }
  auto debug = self.get_data().del(dat);
  if (debug == false) {
    ctx->error("failed to delete " + dataid);
    return;
  }
  ctx->ok("done");
}

///////////////////// authorization related methods ////////////////////////////

// authorize user to use data
DEFINE_METHOD(DataAuth, authorize) {
  xchain::Context *ctx = self.context();
  const std::string &dataid = ctx->arg("dataid");
  const std::string &user = ctx->arg("user");
  const std::string &pubkey = ctx->arg("pubkey");
  const std::string &signature = ctx->arg("signature");
  DataAuth::data dat;
  if (!self.get_data().find({{"dataid", dataid}, {"user", ctx->initiator()}},
                            &dat)) {
    ctx->error("can not find " + dataid);
    return;
  }
  // only owner can authorize data
  if (dat.owner() != ctx->initiator()) {
    ctx->error("permission denied to authorize " + dataid);
    return;
  }

  // get authorization information
  TrustOperators::AuthInfo auth;
  auth.data = dat.content();
  auth.to = user;
  auth.pubkey = pubkey;
  auth.signature = signature;
  auth.kind = "commitment";
  // call trust operator to compute the commitment for user
  // kind = "commitment"，返回{{"commitment": commitment}}
  TrustOperators to(ctx, 0);
  std::map<std::string, std::string> result;
  auto debug = to.authorize(auth, &result);
  if (!debug) {
    ctx->error("failed to authorize data " + dataid);
    return;
  }
  const std::string commitment = result["commitment"];
  if (commitment == "") {
    ctx->error("failed to authorize data " + dataid);
    return;
  }

  // put a new record, dataid, owner, content, expire, user, commitment
  DataAuth::data newdat;
  newdat.set_dataid(std::stoll(dataid));
  newdat.set_owner(ctx->initiator());
  newdat.set_content(dat.content());
  newdat.set_expire(dat.expire());
  newdat.set_user(user);
  newdat.set_commitment(commitment);
  debug = self.get_data().put(newdat);
  if (debug == false) {
    ctx->error("failed to authorize " + dataid + " for " + user);
    return;
  }
  ctx->ok("done");
}

// share plain data to others, create a new record
DEFINE_METHOD(DataAuth, share) {
  xchain::Context *ctx = self.context();
  const std::string &dataid = ctx->arg("dataid");
  const std::string &addr = ctx->arg("toaddr");
  const std::string &newid = ctx->arg("newid");
  const std::string &pubkey = ctx->arg("pubkey");
  const std::string &signature = ctx->arg("signature");
  DataAuth::data dat;
  if (!self.get_data().find({{"dataid", dataid}, {"user", ctx->initiator()}},
                            &dat)) {
    ctx->error("cannot find " + dataid);
    return;
  }
  // only owner can share data
  if (dat.owner() != ctx->initiator()) {
    ctx->error("permission denied to share " + dataid);
    return;
  }

  // get share information
  TrustOperators::AuthInfo auth;
  auth.data = dat.content();
  auth.to = addr;
  auth.pubkey = pubkey;
  auth.signature = signature;
  auth.kind = "ownership";
  // call trust operator to compute the commitment for user
  // 若kind = "ownership"，返回{{"cipher": ciphertext}}
  TrustOperators to(ctx, 0);
  std::map<std::string, std::string> result;
  auto debug = to.authorize(auth, &result);
  if (!debug) {
    ctx->error("failed to share data " + dataid);
    return;
  }
  const std::string new_data = result["cipher"];
  if (new_data == "") {
    ctx->error("failed to share data " + dataid);
    return;
  }

  // put new_data into table
  DataAuth::data newdat;
  newdat.set_dataid(std::stoll(newid));
  newdat.set_owner(addr);
  newdat.set_content(new_data);
  newdat.set_expire(dat.expire());
  newdat.set_user(addr);
  debug = self.get_data().put(newdat);
  if (debug == false) {
    ctx->error("failed to store " + newid);
    return;
  }
  ctx->ok("done");
}

// add two ciphertext and create a new record
DEFINE_METHOD(DataAuth, add) {
  xchain::Context *ctx = self.context();
  const std::string &data1 = ctx->arg("data1");
  const std::string &data2 = ctx->arg("data2");
  const std::string &newid = ctx->arg("newid");
  // get two ciphertexts for addition
  DataAuth::data dat1;
  if (!self.get_data().find({{"dataid", data1}, {"user", ctx->initiator()}},
                            &dat1)) {
    ctx->error("cannot find " + data1);
    return;
  }
  DataAuth::data dat2;
  if (!self.get_data().find({{"dataid", data2}, {"user", ctx->initiator()}},
                            &dat2)) {
    ctx->error("cannot find " + data2);
    return;
  }

  // get left and right operands
  TrustOperators::Operand left_op, right_op;
  left_op.cipher = dat1.content();
  left_op.commitment = dat1.commitment();
  right_op.cipher = dat2.content();
  right_op.commitment = dat2.commitment();
  // call trust operator to add, returns {{"key", enc_result}}
  TrustOperators to(ctx, 0);
  std::map<std::string, std::string> result;
  auto debug = to.add(left_op, right_op, &result);
  if (!debug) {
    ctx->error("failed to add " + data1 + " and " + data2);
    return;
  }
  const std::string new_data = result["key"];
  if (new_data == "") {
    ctx->error("failed to add " + data1 + " and " + data2);
    return;
  }

  // put result into table, owner is the user
  DataAuth::data newdat;
  newdat.set_dataid(std::stoll(newid));
  newdat.set_owner(ctx->initiator());
  newdat.set_content(new_data);
  newdat.set_user(ctx->initiator());
  // expire set to be the most recently expire date of two operands
  if (dat1.expire() < dat2.expire()) {
    newdat.set_expire(dat1.expire());
  } else {
    newdat.set_expire(dat2.expire());
  }
  debug = self.get_data().put(newdat);
  if (debug == false) {
    ctx->error("failed to store " + newid);
    return;
  }
  ctx->ok("done");
}

// sub two ciphertext and create a new record
DEFINE_METHOD(DataAuth, sub) {
  xchain::Context *ctx = self.context();
  const std::string &data1 = ctx->arg("data1");
  const std::string &data2 = ctx->arg("data2");
  const std::string &newid = ctx->arg("newid");
  // get two ciphertexts for substraction
  DataAuth::data dat1;
  if (!self.get_data().find({{"dataid", data1}, {"user", ctx->initiator()}},
                            &dat1)) {
    ctx->error("cannot find " + data1);
    return;
  }
  DataAuth::data dat2;
  if (!self.get_data().find({{"dataid", data2}, {"user", ctx->initiator()}},
                            &dat2)) {
    ctx->error("cannot find " + data2);
    return;
  }

  // get left and right operands
  TrustOperators::Operand left_op, right_op;
  left_op.cipher = dat1.content();
  left_op.commitment = dat1.commitment();
  right_op.cipher = dat2.content();
  right_op.commitment = dat2.commitment();
  // call trust operator to sub, returns {{"key", enc_result}}
  TrustOperators to(ctx, 0);
  std::map<std::string, std::string> result;
  auto debug = to.sub(left_op, right_op, &result);
  if (!debug) {
    ctx->error("failed to sub " + data1 + " and " + data2);
    return;
  }
  const std::string new_data = result["key"];
  if (new_data == "") {
    ctx->error("failed to sub " + data1 + " and " + data2);
    return;
  }

  // put result into table, owner is the user
  DataAuth::data newdat;
  newdat.set_dataid(std::stoll(newid));
  newdat.set_owner(ctx->initiator());
  newdat.set_content(new_data);
  newdat.set_user(ctx->initiator());
  // expire set to be the most recently expire date of two operands
  if (dat1.expire() < dat2.expire()) {
    newdat.set_expire(dat1.expire());
  } else {
    newdat.set_expire(dat2.expire());
  }
  debug = self.get_data().put(newdat);
  if (debug == false) {
    ctx->error("failed to store " + newid);
    return;
  }
  ctx->ok("done");
}

// multiply two ciphertext and create a new record
DEFINE_METHOD(DataAuth, mul) {
  xchain::Context *ctx = self.context();
  const std::string &data1 = ctx->arg("data1");
  const std::string &data2 = ctx->arg("data2");
  const std::string &newid = ctx->arg("newid");
  // get two ciphertexts for multiplication
  DataAuth::data dat1;
  if (!self.get_data().find({{"dataid", data1}, {"user", ctx->initiator()}},
                            &dat1)) {
    ctx->error("cannot find " + data1);
    return;
  }
  DataAuth::data dat2;
  if (!self.get_data().find({{"dataid", data2}, {"user", ctx->initiator()}},
                            &dat2)) {
    ctx->error("cannot find " + data2);
    return;
  }

  // get left and right operands
  TrustOperators::Operand left_op, right_op;
  left_op.cipher = dat1.content();
  left_op.commitment = dat1.commitment();
  right_op.cipher = dat2.content();
  right_op.commitment = dat2.commitment();
  // call trust operator to mul, return {{"key", enc_result}}
  TrustOperators to(ctx, 0);
  std::map<std::string, std::string> result;
  auto debug = to.mul(left_op, right_op, &result);
  if (!debug) {
    ctx->error("failed to mul " + data1 + " and " + data2);
    return;
  }
  const std::string new_data = result["key"];
  if (new_data == "") {
    ctx->error("failed to mul " + data1 + " and " + data2);
    return;
  }

  // put result into table, owner is the user
  DataAuth::data newdat;
  newdat.set_dataid(std::stoll(newid));
  newdat.set_owner(ctx->initiator());
  newdat.set_content(new_data);
  newdat.set_user(ctx->initiator());
  // expire set to be the most recently expire date of two operands
  if (dat1.expire() < dat2.expire()) {
    newdat.set_expire(dat1.expire());
  } else {
    newdat.set_expire(dat2.expire());
  }
  debug = self.get_data().put(newdat);
  if (debug == false) {
    ctx->error("failed to store " + newid);
    return;
  }
  ctx->ok("done");
}
