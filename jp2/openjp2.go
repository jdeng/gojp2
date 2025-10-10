package jp2

/*
#cgo CFLAGS: -I${SRCDIR} -I${SRCDIR}/openjp2 -std=c11
#cgo LDFLAGS: -lm
#include <stdlib.h>
#include <string.h>
#include "openjp2/openjpeg.h"

typedef struct {
    const unsigned char *data;
    size_t length;
    size_t offset;
} gojp2_buffer;

typedef struct {
    opj_image_t *image;
    int status;
    char message[256];
} gojp2_result;

static void gojp2_result_set_error(gojp2_result *result, const char *msg) {
    if (!result) {
        return;
    }
    result->status = -1;
    if (!msg) {
        result->message[0] = '\0';
        return;
    }
    size_t len = strlen(msg);
    if (len >= sizeof(result->message)) {
        len = sizeof(result->message) - 1U;
    }
    memcpy(result->message, msg, len);
    result->message[len] = '\0';
}

static OPJ_SIZE_T gojp2_read(void *p_buffer, OPJ_SIZE_T nb_bytes, void *user_data) {
    gojp2_buffer *buffer = (gojp2_buffer *)user_data;
    if (!buffer || !p_buffer) {
        return (OPJ_SIZE_T)0;
    }
    size_t remaining = 0;
    if (buffer->offset < buffer->length) {
        remaining = buffer->length - buffer->offset;
    }
    if (nb_bytes > remaining) {
        nb_bytes = (OPJ_SIZE_T)remaining;
    }
    if (nb_bytes > 0U) {
        memcpy(p_buffer, buffer->data + buffer->offset, (size_t)nb_bytes);
        buffer->offset += (size_t)nb_bytes;
    }
    return nb_bytes;
}

static OPJ_OFF_T gojp2_skip(OPJ_OFF_T nb_bytes, void *user_data) {
    gojp2_buffer *buffer = (gojp2_buffer *)user_data;
    if (!buffer) {
        return (OPJ_OFF_T)0;
    }
    if (nb_bytes == 0) {
        return (OPJ_OFF_T)0;
    }
    if (nb_bytes > 0) {
        size_t remaining = 0;
        if (buffer->offset < buffer->length) {
            remaining = buffer->length - buffer->offset;
        }
        size_t advance = (size_t)nb_bytes;
        if (advance > remaining) {
            advance = remaining;
        }
        buffer->offset += advance;
        return (OPJ_OFF_T)advance;
    }
    // negative skip
    size_t back = (size_t)(-nb_bytes);
    if (back > buffer->offset) {
        back = buffer->offset;
    }
    buffer->offset -= back;
    return -(OPJ_OFF_T)back;
}

static OPJ_BOOL gojp2_seek(OPJ_OFF_T nb_bytes, void *user_data) {
    gojp2_buffer *buffer = (gojp2_buffer *)user_data;
    if (!buffer || nb_bytes < 0) {
        return OPJ_FALSE;
    }
    size_t target = (size_t)nb_bytes;
    if (target > buffer->length) {
        return OPJ_FALSE;
    }
    buffer->offset = target;
    return OPJ_TRUE;
}

static void OPJ_CALLCONV gojp2_error_callback(const char *msg, void *client_data) {
    gojp2_result *result = (gojp2_result *)client_data;
    if (!result || !msg) {
        return;
    }
    if (result->message[0] != '\0') {
        return;
    }
    gojp2_result_set_error(result, msg);
}

static void OPJ_CALLCONV gojp2_warning_callback(const char *msg, void *client_data) {
    (void)msg;
    (void)client_data;
}

static void OPJ_CALLCONV gojp2_info_callback(const char *msg, void *client_data) {
    (void)msg;
    (void)client_data;
}

static opj_stream_t *gojp2_create_stream(gojp2_buffer *buffer) {
    opj_stream_t *stream = opj_stream_create(OPJ_J2K_STREAM_CHUNK_SIZE, OPJ_TRUE);
    if (!stream) {
        return NULL;
    }
    opj_stream_set_user_data(stream, buffer, NULL);
    opj_stream_set_user_data_length(stream, buffer->length);
    opj_stream_set_read_function(stream, gojp2_read);
    opj_stream_set_skip_function(stream, gojp2_skip);
    opj_stream_set_seek_function(stream, gojp2_seek);
    return stream;
}

static int gojp2_try_decode(OPJ_CODEC_FORMAT format, gojp2_buffer *buffer, gojp2_result *result) {
    opj_dparameters_t parameters;
    opj_codec_t *codec = NULL;
    opj_stream_t *stream = NULL;
    opj_image_t *image = NULL;
    int ok = 0;

    opj_set_default_decoder_parameters(&parameters);

    codec = opj_create_decompress(format);
    if (!codec) {
        gojp2_result_set_error(result, "opj_create_decompress failed");
        goto cleanup;
    }

    opj_set_error_handler(codec, gojp2_error_callback, result);
    opj_set_warning_handler(codec, gojp2_warning_callback, result);
    opj_set_info_handler(codec, gojp2_info_callback, result);

    stream = gojp2_create_stream(buffer);
    if (!stream) {
        gojp2_result_set_error(result, "stream creation failed");
        goto cleanup;
    }

    if (!opj_setup_decoder(codec, &parameters)) {
        gojp2_result_set_error(result, "opj_setup_decoder failed");
        goto cleanup;
    }

    if (!opj_read_header(stream, codec, &image)) {
        if (result->message[0] == '\0') {
            gojp2_result_set_error(result, "opj_read_header failed");
        }
        goto cleanup;
    }

    if (!opj_decode(codec, stream, image)) {
        if (result->message[0] == '\0') {
            gojp2_result_set_error(result, "opj_decode failed");
        }
        goto cleanup;
    }

    if (!opj_end_decompress(codec, stream)) {
        if (result->message[0] == '\0') {
            gojp2_result_set_error(result, "opj_end_decompress failed");
        }
        goto cleanup;
    }

    result->image = image;
    result->status = 0;
    result->message[0] = '\0';
    image = NULL;
    ok = 1;

cleanup:
    if (image) {
        opj_image_destroy(image);
    }
    if (stream) {
        opj_stream_destroy(stream);
    }
    if (codec) {
        opj_destroy_codec(codec);
    }
    buffer->offset = 0;
    return ok;
}

int gojp2_decode_memory(const unsigned char *data, size_t length, gojp2_result *result) {
    if (!result) {
        return 0;
    }
    result->image = NULL;
    result->status = -1;
    result->message[0] = '\0';

    if (!data || length == 0U) {
        gojp2_result_set_error(result, "empty input");
        return 0;
    }

    gojp2_buffer buffer;
    buffer.data = data;
    buffer.length = length;
    buffer.offset = 0;

    if (gojp2_try_decode(OPJ_CODEC_JP2, &buffer, result)) {
        return 1;
    }

    // Preserve existing message but try raw codestream as fallback.
    char previous[256];
    memcpy(previous, result->message, sizeof(previous));
    result->message[0] = '\0';
    if (gojp2_try_decode(OPJ_CODEC_J2K, &buffer, result)) {
        return 1;
    }

    if (result->message[0] == '\0') {
        memcpy(result->message, previous, sizeof(previous));
        result->message[sizeof(result->message) - 1U] = '\0';
    }
    return 0;
}

void gojp2_image_destroy(opj_image_t *image) {
    if (image) {
        opj_image_destroy(image);
    }
}
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"unsafe"
)

// Image holds decoded component buffers using 8-bit samples.
type Image struct {
	Width      int
	Height     int
	Components [][]byte
	AlphaIndex int
	BitDepth   int
	ColorSpace int
}

// Decode converts a JPEG 2000 codestream contained in data into 8-bit component buffers.
func Decode(data []byte) (*Image, error) {
	if len(data) == 0 {
		return nil, errors.New("jp2: empty input")
	}

	var result C.gojp2_result
	ok := C.gojp2_decode_memory(
		(*C.uchar)(unsafe.Pointer(&data[0])),
		C.size_t(len(data)),
		&result,
	)
	defer C.gojp2_image_destroy(result.image)

	runtime.KeepAlive(data)

	if ok == 0 || result.image == nil {
		msg := C.GoString(&result.message[0])
		msg = strings.TrimSpace(msg)
		if msg == "" {
			msg = "jp2 decode failed"
		}
		return nil, errors.New("jp2: " + msg)
	}

	img := result.image

	width := int(img.x1 - img.x0)
	height := int(img.y1 - img.y0)
	if width <= 0 || height <= 0 {
		return nil, errors.New("jp2: invalid image dimensions")
	}

	numComps := int(img.numcomps)
	if numComps <= 0 {
		return nil, errors.New("jp2: missing components")
	}

	compSlice := unsafe.Slice(
		(*C.opj_image_comp_t)(unsafe.Pointer(img.comps)),
		numComps,
	)

	totalPixels := width * height
	if totalPixels <= 0 {
		return nil, errors.New("jp2: invalid pixel count")
	}

	comps := make([][]byte, numComps)
	alphaIndex := -1
	maxBitDepth := 0

	for i := 0; i < numComps; i++ {
		comp := compSlice[i]

		compWidth := int(comp.w)
		compHeight := int(comp.h)
		if compWidth != width || compHeight != height || comp.dx != 1 || comp.dy != 1 {
			return nil, fmt.Errorf("jp2: unsupported component layout for component %d", i)
		}

		precision := int(comp.prec)
		if precision <= 0 {
			return nil, fmt.Errorf("jp2: invalid precision (%d) for component %d", precision, i)
		}
		if precision > maxBitDepth {
			maxBitDepth = precision
		}

		samples := unsafe.Slice(
			(*C.OPJ_INT32)(unsafe.Pointer(comp.data)),
			totalPixels,
		)
		component := make([]byte, totalPixels)

		offset := 0
		if comp.sgnd != 0 {
			if precision >= 32 {
				return nil, fmt.Errorf("jp2: unsupported signed precision (%d) for component %d", precision, i)
			}
			offset = 1 << (precision - 1)
		}

		var shiftLeft, shiftRight uint
		if precision < 8 {
			shiftLeft = uint(8 - precision)
		} else if precision > 8 {
			shiftRight = uint(precision - 8)
		}

		maxValue := (1 << precision) - 1

		for idx, sample := range samples {
			val := int(sample)
			if offset != 0 {
				val += offset
			}
			if val < 0 {
				val = 0
			} else if val > maxValue {
				val = maxValue
			}
			if shiftRight != 0 {
				val >>= shiftRight
			} else if shiftLeft != 0 {
				val <<= shiftLeft
				if val > 255 {
					val = 255
				}
			}
			component[idx] = byte(val & 0xFF)
		}

		comps[i] = component
		if comp.alpha != 0 && alphaIndex == -1 {
			alphaIndex = i
		}
	}

	return &Image{
		Width:      width,
		Height:     height,
		Components: comps,
		AlphaIndex: alphaIndex,
		BitDepth:   maxBitDepth,
		ColorSpace: int(img.color_space),
	}, nil
}
