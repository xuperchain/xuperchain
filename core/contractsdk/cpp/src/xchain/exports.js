// Tell compiler we have those functions, do not print error when linking
mergeInto(LibraryManager.library, {
  call_method_v2: function () { },
  call_method: function () { },
  fetch_response: function () { },
  xvm_hash: function () { },
  xvm_encode: function () { },
  xvm_decode: function () { },
  xvm_ecverify: function () { },
  xvm_make_tx: function () { },
  xvm_addr_from_pubkey: function () { },
});
