// Tell compiler we have those functions, do not print error when linking
mergeInto(LibraryManager.library, {
  call_method_v2: function(){},
  call_method: function(){},
  fetch_response: function(){},
});
