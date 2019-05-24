#include "textflag.h"

TEXT ·callMethod(SB), NOSPLIT, $0
  CallImport
  RET

TEXT ·fetchResponse(SB), NOSPLIT, $0
  CallImport
  RET
