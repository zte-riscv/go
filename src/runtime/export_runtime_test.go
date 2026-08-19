// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

package runtime

import "unsafe"

func TryClearSpanFunc() (needzero uint8, cleared bool) {
	buf := make([]byte, 128)
	for i := range buf {
		buf[i] = 0xff
	}

	ms := AllocMSpan()
	defer FreeMSpan(ms)

	s := (*mspan)(unsafe.Pointer(ms))
	s.startAddr = uintptr(unsafe.Pointer(&buf[0]))
	s.elemsize = 16
	s.nelems = uint16(len(buf) / int(s.elemsize))
	s.allocCount = 0
	s.needzero = 1

	clearSpanFunc(s)

	needzero = s.needzero
	cleared = true
	for _, b := range buf {
		if b != 0 {
			cleared = false
			break
		}
	}
	KeepAlive(buf)
	return
}
