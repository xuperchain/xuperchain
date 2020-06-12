#include "paillier.pb.h"
#include "xchain/table/table.tpl.h"
#include "xchain/table/types.h"
#include "xchain/trust_operators/trust_operators.h"
#include "xchain/xchain.h"
#include <inttypes.h>

struct PaillierData : public xchain::Contract {
public:
  PaillierData() : _data(this->context(), "paillier_data") {}

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

DEFINE_METHOD(PaillierData, initialize) {
  xchain::Context *ctx = self.context();
  ctx->ok("initialize succeed");
}

// store encrypted data with content and expire_time
DEFINE_METHOD(PaillierData, store) {
  xchain::Context *ctx = self.context();
  const std::string &dataid = ctx->arg("dataid");
  const std::string &content = ctx->arg("content");
  const std::string &pubkey = ctx->arg("pubkey");

  PaillierData::data dat;
  dat.set_dataid(std::stoll(dataid));
  dat.set_owner(ctx->initiator());
  dat.set_content(content);
  dat.set_pubkey(pubkey);
  dat.set_user(ctx->initiator());
  auto debug = self.get_data().put(dat);
  if (debug == false) {
    ctx->error("failed to store " + dataid);
    return;
  }
  ctx->ok("done");
}

// get a record from table
DEFINE_METHOD(PaillierData, get) {
  xchain::Context *ctx = self.context();
  const std::string &dataid = ctx->arg("dataid");
  const std::string &user = ctx->initiator();
  PaillierData::data dat;
  if (!self.get_data().find({{"dataid", dataid}, {"user", user}}, &dat)) {
    ctx->error("cannot find " + dataid);
    return;
  }
  ctx->ok(dat.content());
}

// modify user's one record
DEFINE_METHOD(PaillierData, modify) {
  xchain::Context *ctx = self.context();
  const std::string &dataid = ctx->arg("dataid");
  const std::string &content = ctx->arg("content");
  const std::string &pubkey = ctx->arg("pubkey");
  PaillierData::data dat;
  if (!self.get_data().find({{"dataid", dataid}, {"user", ctx->initiator()}},
                            &dat)) {
    ctx->error("cannot find " + dataid);
    return;
  }
  dat.set_content(content);
  dat.set_pubkey(pubkey);
  auto debug = self.get_data().update(dat);
  if (debug == false) {
    ctx->error("failed to modify " + dataid);
    return;
  }
  ctx->ok("done");
}

// delete user's record
// TODO: owner can delete every owned record
DEFINE_METHOD(PaillierData, del) {
  xchain::Context *ctx = self.context();
  const std::string &dataid = ctx->arg("dataid");
  PaillierData::data dat;
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
DEFINE_METHOD(PaillierData, authorize) {
  xchain::Context *ctx = self.context();
  const std::string &dataid = ctx->arg("dataid");
  const std::string &user = ctx->arg("user");
  const std::string &commitment = ctx->arg("commitment");
  PaillierData::data dat;
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

  // put a new record, dataid, owner, content, expire, user, commitment
  PaillierData::data newdat;
  newdat.set_dataid(std::stoll(dataid));
  newdat.set_owner(ctx->initiator());
  newdat.set_content(dat.content());
  newdat.set_pubkey(dat.pubkey());
  newdat.set_user(user);
  newdat.set_commitment(commitment);
  auto debug = self.get_data().put(newdat);
  if (debug == false) {
    ctx->error("failed to authorize " + dataid + " for " + user);
    return;
  }
  ctx->ok("done");
}

// share plain data to others, create a new record
DEFINE_METHOD(PaillierData, share) {
  xchain::Context *ctx = self.context();
  const std::string &dataid = ctx->arg("dataid");
  const std::string &addr = ctx->arg("toaddr");
  const std::string &newid = ctx->arg("newid");
  const std::string &pubkey = ctx->arg("pubkey");
  const std::string &content = ctx->arg("content");
  PaillierData::data dat;
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

  // put new_data into table
  PaillierData::data newdat;
  newdat.set_dataid(std::stoll(newid));
  newdat.set_owner(addr);
  newdat.set_pubkey(pubkey);
  newdat.set_content(content);
  newdat.set_user(addr);
  auto debug = self.get_data().put(newdat);
  if (debug == false) {
    ctx->error("failed to share " + dataid + " for " + addr);
    return;
  }
  ctx->ok("done");
}

// add two ciphertext and create a new record
DEFINE_METHOD(PaillierData, add) {
  xchain::Context *ctx = self.context();
  const std::string &data1 = ctx->arg("data1");
  const std::string &data2 = ctx->arg("data2");
  const std::string &newid = ctx->arg("newid");
  // get two ciphertexts for addition
  PaillierData::data dat1;
  if (!self.get_data().find({{"dataid", data1}, {"user", ctx->initiator()}},
                            &dat1)) {
    ctx->error("cannot find " + data1);
    return;
  }
  PaillierData::data dat2;
  if (!self.get_data().find({{"dataid", data2}, {"user", ctx->initiator()}},
                            &dat2)) {
    ctx->error("cannot find " + data2);
    return;
  }

  if (dat1.pubkey() != dat2.pubkey()) {
    ctx->error(
        "addition of data with different public keys are not supported yet");
    return;
  }

  // get left and right operands
  TrustOperators::Operand left_op, right_op;
  left_op.cipher = dat1.content();
  left_op.commitment = dat1.commitment();
  right_op.cipher = dat2.content();
  right_op.commitment = dat2.commitment();
  // call trust operator to add, returns {{"ciphertext", add_result}}
  TrustOperators to(ctx, 0);
  std::map<std::string, std::string> result;
  auto debug = to.paillier_add(left_op, right_op, dat1.pubkey(), &result);
  if (!debug) {
    ctx->error("failed to add " + data1 + " and " + data2);
    return;
  }

  const std::string new_data = result["ciphertext"];
  if (new_data == "") {
    ctx->error("failed to get cipherAdd of " + data1 + " and " + data2);
    return;
  }

  // put result into table, owner is the user
  PaillierData::data newdat;
  newdat.set_dataid(std::stoll(newid));
  newdat.set_owner(dat1.owner());
  newdat.set_content(new_data);
  newdat.set_pubkey(dat1.pubkey());
  newdat.set_user(ctx->initiator());

  debug = self.get_data().put(newdat);
  if (debug == false) {
    ctx->error("failed to store " + newid);
    return;
  }
  ctx->ok("done");
}

// mul a ciphertext with a number and create a new record
DEFINE_METHOD(PaillierData, mul) {
  xchain::Context *ctx = self.context();
  const std::string &dataid = ctx->arg("dataid");
  const std::string &scalar = ctx->arg("scalar");
  const std::string &newid = ctx->arg("newid");
  // get two ciphertexts for addition
  PaillierData::data dat;
  if (!self.get_data().find({{"dataid", dataid}, {"user", ctx->initiator()}},
                            &dat)) {
    ctx->error("cannot find " + dataid);
    return;
  }

  // get operand
  TrustOperators::Operand left_op;
  left_op.cipher = dat.content();
  left_op.commitment = dat.commitment();
  // call trust operator to mul, returns {{"ciphertext", mul_result}}
  TrustOperators to(ctx, 0);
  std::map<std::string, std::string> result;
  auto debug = to.paillier_partial_mul(left_op, scalar, dat.pubkey(), &result);
  if (!debug) {
    ctx->error("failed to mul " + dataid + " and " + scalar);
    return;
  }
  const std::string new_data = result["ciphertext"];
  if (new_data == "") {
    ctx->error("failed to get cipherMul of " + dataid + " and " + scalar);
    return;
  }

  // put result into table, owner is the user
  PaillierData::data newdat;
  newdat.set_dataid(std::stoll(newid));
  newdat.set_owner(dat.owner());
  newdat.set_content(new_data);
  newdat.set_pubkey(dat.pubkey());
  newdat.set_user(ctx->initiator());

  debug = self.get_data().put(newdat);
  if (debug == false) {
    ctx->error("failed to store " + newid);
    return;
  }
  ctx->ok("done");
}