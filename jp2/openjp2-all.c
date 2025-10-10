#define OPJ_SKIP_POISON
#include "opj_includes.h"
#undef OPJ_SKIP_POISON

#include "opj_malloc.c"
#include "thread.c"
#include "bio.c"
#include "cio.c"
#include "event.c"
#include "ht_dec.c"
#include "image.c"
#include "invert.c"
#include "j2k.c"
#include "jp2.c"
#include "mct.c"
#include "mqc.c"
#include "opj_clock.c"
#include "dwt.c"
#include "t2.c"
#include "tcd.c"
#include "tgt.c"
#include "function_list.c"
#include "sparse_array.c"
#include "openjpeg.c"

#define opj_t1_allocate_buffers t1_opj_t1_allocate_buffers
#include "t1.c"
#undef opj_t1_allocate_buffers

#undef OPJ_UINT32_SEMANTICALLY_BUT_INT32
#include "pi.c"

