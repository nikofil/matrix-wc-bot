package olm

/*
#cgo LDFLAGS: -lolm
#include <olm/olm.h>
#include <stdlib.h>
*/
import "C"
import (
	// "fmt"
	"unsafe"
)

func NewOlmGroupSession() OlmGroupSession {
	memSize := int(C.olm_inbound_group_session_size())
	buf := make([]byte, memSize)
	return OlmGroupSession{
		groupSess: C.olm_inbound_group_session(unsafe.Pointer(&buf[0])),
	}
}

type OlmGroupSession struct {
	groupSess *C.OlmInboundGroupSession
}

func (sess *OlmGroupSession) InitInbGroupSess(sessKey string) {
	keyBytes := []byte(sessKey)
	C.olm_init_inbound_group_session(sess.groupSess, (*C.uchar)(unsafe.Pointer(&keyBytes[0])), C.ulong(len(keyBytes)))
	// fmt.Println("Init returned:", rv)
	// fmt.Println(C.GoString(C.olm_inbound_group_session_last_error(sess.groupSess)))
}

func (sess *OlmGroupSession) DecryptGroupMsg(msg string) string {
	msgBytes := []byte(msg)
	ptxtLen := C.olm_group_decrypt_max_plaintext_length(sess.groupSess, (*C.uchar)(unsafe.Pointer(&msgBytes[0])), C.ulong(len(msgBytes)))
	msgBytes = []byte(msg)
	outBytes := make([]byte, ptxtLen)
	var msgIdx C.uint32_t
	plaintextLen := C.olm_group_decrypt(sess.groupSess, (*C.uchar)(unsafe.Pointer(&msgBytes[0])), C.ulong(len(msgBytes)), (*C.uchar)(unsafe.Pointer(&outBytes[0])), ptxtLen, &msgIdx)
	// fmt.Println("Decrypt returned:", retVal, msgIdx)
	// fmt.Println(C.GoString(C.olm_inbound_group_session_last_error(sess.groupSess)))
	return string(outBytes[:int(plaintextLen)])
}

/*
size_t olm_init_inbound_group_session(
    OlmInboundGroupSession *session,
    uint8_t const * session_key, size_t session_key_length
);

size_t olm_group_decrypt_max_plaintext_length(
    OlmInboundGroupSession *session,
    uint8_t * message, size_t message_length
);
size_t olm_group_decrypt(
    OlmInboundGroupSession *session,
    uint8_t * message, size_t message_length,
    uint8_t * plaintext, size_t max_plaintext_length,
    uint32_t * message_index
);
*/
