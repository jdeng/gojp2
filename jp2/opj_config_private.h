#ifndef OPJ_CONFIG_PRIVATE_H
#define OPJ_CONFIG_PRIVATE_H

/* Manually generated from opj_config_private.h.cmake.in for Windows and Linux builds. */

#define OPJ_PACKAGE_VERSION "2.5.0"

#if defined(_WIN32)

#define OPJ_HAVE_MALLOC_H 1
#define OPJ_HAVE__ALIGNED_MALLOC 1

#else /* POSIX-like platforms */

#define _LARGEFILE_SOURCE
#define _LARGE_FILES 1
#ifndef _FILE_OFFSET_BITS
#define _FILE_OFFSET_BITS 64
#endif
#define OPJ_HAVE_FSEEKO 1
#define OPJ_HAVE_MALLOC_H 1
#define OPJ_HAVE_POSIX_MEMALIGN 1
#define OPJ_HAVE_MEMALIGN 1

#endif /* _WIN32 */

#if !defined(_POSIX_C_SOURCE)
#if defined(OPJ_HAVE_FSEEKO) || defined(OPJ_HAVE_POSIX_MEMALIGN)
#define _POSIX_C_SOURCE 200112L
#endif
#endif

#if !defined(__APPLE__)
#if defined(__BYTE_ORDER__) && defined(__ORDER_BIG_ENDIAN__)
#if (__BYTE_ORDER__ == __ORDER_BIG_ENDIAN__)
#define OPJ_BIG_ENDIAN
#endif
#endif
#elif defined(__BIG_ENDIAN__)
#define OPJ_BIG_ENDIAN
#endif

#endif /* OPJ_CONFIG_PRIVATE_H */
