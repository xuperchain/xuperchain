// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

#ifndef MESATEE_H
#define MESATEE_H

#ifdef __cplusplus
extern "C" {
#endif

#include <mesatee/visibility.h>
#include <sys/socket.h>

typedef struct mesatee mesatee_t;
typedef struct mesatee_enclave_info mesatee_enclave_info_t;
typedef struct mesatee_task mesatee_task_t;
typedef struct mesatee_taskinfo mesatee_taskinfo_t;
typedef struct mesatee_auditor_set mesatee_auditor_set_t;

typedef struct sockaddr sockaddr_t;

typedef enum mesatee_task_status_t {
  TASK_CREATED,
  TASK_READY,
  TASK_RUNNING,
  TASK_FINISHED,
  TASK_FAILED,
} mesatee_task_status_t;

// Initialize logger
MESATEE_API int mesatee_init();

// MesaTEE Context APIs
MESATEE_API mesatee_t *mesatee_context_new(const mesatee_enclave_info_t *enclave_info_ptr,
                                           const char* user_id, const char* user_token,
                                           sockaddr_t * tms_addr, sockaddr_t * tdfs_addr);

MESATEE_API mesatee_t* mesatee_context_new2(const mesatee_enclave_info_t* enclave_info_ptr,
                                           const char* user_id, const char* user_token,
                                           const char* tms_addr_ptr/*ip:port*/,
                                           const char* tdfs_addr_ptr/*ip:port*/);

MESATEE_API int mesatee_context_free(mesatee_t *ctx_ptr);

// MesaTEE EnclaveInfo APIs
MESATEE_API mesatee_enclave_info_t *
mesatee_enclave_info_load(mesatee_auditor_set_t *auditors_ptr, const char *enclave_info_file_path_ptr);
MESATEE_API int mesatee_enclave_info_free(mesatee_enclave_info_t *enclave_info_ptr);

// Auditor APIs
MESATEE_API mesatee_auditor_set_t *mesatee_auditor_set_new();
MESATEE_API int mesatee_auditor_set_add_auditor(mesatee_auditor_set_t *ptr,
                                                const char *pub_key_path, const char *sig_path);
MESATEE_API int mesatee_auditor_set_free(mesatee_auditor_set_t *ptr);

// MesaTEE Task APIs
MESATEE_API mesatee_task_t *mesatee_create_task(mesatee_t *ctx_ptr, const char *func_name_ptr);
MESATEE_API int mesatee_task_free(mesatee_task_t *mesatee_task_ptr);
MESATEE_API int mesatee_task_invoke_with_payload(mesatee_task_t *mesatee_task_ptr, const char *payload_buf_ptr,
                                                 int payload_buf_len, char *result_buf_ptr, int result_buf_len);

#ifdef __cplusplus
} /* extern C */
#endif

#endif // MESATEE_H
