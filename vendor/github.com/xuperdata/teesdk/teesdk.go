package teesdk

/*
#cgo CFLAGS: -I${SRCDIR} 
#cgo LDFLAGS: -L ${SRCDIR}/lib -lmesatee_sdk_c -Wl,-rpath=${SRCDIR}/lib

#include "mesatee/mesatee.h"

#include <arpa/inet.h>
#include <string.h>
#include <mesatee/mesatee.h>
#include <netinet/in.h>
#include <stdio.h>
#include <stdlib.h>

mesatee_enclave_info_t *g_enclave_info = NULL;
mesatee_auditor_set_t *g_auditors = NULL;
int32_t g_tms_port = 5554;
int32_t g_tdfs_port = 5554;

int submit_task(const char* method, const char* args, const char* uid, 
	const char* token, char** output, size_t* output_size) {
  struct sockaddr_in tms_addr;
  struct sockaddr_in tdfs_addr;
  char recvbuf[2048] = {0};
  int ret;

  tms_addr.sin_family = AF_INET;
  tms_addr.sin_addr.s_addr = inet_addr("127.0.0.1");
  tms_addr.sin_port = htons(g_tms_port);

  tdfs_addr.sin_family = AF_INET;
  tdfs_addr.sin_addr.s_addr = inet_addr("127.0.0.1");
  tdfs_addr.sin_port = htons(g_tdfs_port);

  printf("[+] This is a single-party task: %s\n", method);
  
  mesatee_t *context = mesatee_context_new(g_enclave_info, uid, token,
                                           (struct sockaddr *)&tms_addr,
                                           (struct sockaddr *)&tdfs_addr);
  if (context == NULL) {
    return EXIT_FAILURE;
  }

  mesatee_task_t *task = mesatee_create_task(context, method);
  if (task == NULL) {
    return EXIT_FAILURE;
  }
  printf("args: %s, size=%lu\n", args, strlen(args));
  // BUG result truncating
  ret = mesatee_task_invoke_with_payload(task, args, strlen(args), 
	recvbuf, sizeof(recvbuf));
  if (ret <= 0) {
    return EXIT_FAILURE;
  }

  printf("Response: %s\n", recvbuf);
  *output_size = strlen(recvbuf);
  *output = (char*)malloc(strlen(recvbuf) + 1);
  memset(*output, 0, *output_size);
  memcpy(*output, recvbuf, *output_size);

  mesatee_task_free(task);
  mesatee_context_free(context);
  return 0;
}

int init(const char* pub, const char* pri, const char* conf, int32_t port1, int32_t port2) {
  mesatee_init();

  g_auditors = mesatee_auditor_set_new();
  mesatee_auditor_set_add_auditor(
      g_auditors, pub, pri);

  if (g_auditors == NULL) {
    return EXIT_FAILURE;
  }

  g_enclave_info =
      mesatee_enclave_info_load(g_auditors, conf);

  if (g_enclave_info ==  NULL) {
    return EXIT_FAILURE;
  }
  g_tms_port = port1;
  g_tdfs_port = port2;

  return 0;
}

int release() {
  mesatee_enclave_info_free(g_enclave_info);
  mesatee_auditor_set_free(g_auditors);
  return 0;
}

*/
import "C"

import (
	"unsafe"
	"errors"
	"sync"
)

type TEEClient struct {
    Uid                *C.char
    Token              *C.char
    PublicDer          *C.char
    SignSha256	       *C.char
    EnclaveInfoConfig  *C.char
    TMSPort	       C.int32_t
    TDFSPort           C.int32_t

}

// env_logger can be init once, so single instance patten is adapted
var kInstance *TEEClient
var once sync.Once
func NewTEEClient(uid, token, pd, ss, eic string, tmsport, tdfsport int32) *TEEClient {
	if kInstance != nil {
		return kInstance;
	}
	once.Do(func() {
		kInstance = &TEEClient{
		    Uid:                C.CString(uid),
		    Token:              C.CString(token),
		    PublicDer:          C.CString(pd),
		    SignSha256:         C.CString(ss),
		    EnclaveInfoConfig:  C.CString(eic),
		    TMSPort:		C.int32_t(tmsport),
		    TDFSPort:		C.int32_t(tdfsport),
		}
		s := kInstance
		C.init(s.PublicDer, s.SignSha256, s.EnclaveInfoConfig, s.TMSPort, s.TDFSPort)
		})
	return kInstance
}

func (s *TEEClient) Close() {
    C.release();
    C.free(unsafe.Pointer(s.Uid));
    C.free(unsafe.Pointer(s.Token));
    C.free(unsafe.Pointer(s.PublicDer));
    C.free(unsafe.Pointer(s.SignSha256));
    C.free(unsafe.Pointer(s.EnclaveInfoConfig));
}

func (s *TEEClient) Submit(method string, cipher string) (string, error) {
     cMethod, cArgs := C.CString(method), C.CString(cipher)
     defer C.free(unsafe.Pointer(cMethod))
     defer C.free(unsafe.Pointer(cArgs))
     // error handler TODO
     var output *C.char
     var outputSize C.size_t
     ret := C.submit_task(cMethod, cArgs, s.Uid, s.Token, &output, &outputSize)
     if ret != 0 {
	return "", errors.New("submit_task error, return nil")
     }
     defer C.free(unsafe.Pointer(output))
     return C.GoStringN(output, C.int(outputSize)), nil
}

