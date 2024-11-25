// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "../../../../../runtime/textflag.h"

TEXT asmtest(SB),DUPOK|NOSPLIT,$0
start:
	FLIS $(NaN), F1                // d3801ff0
	FCVTMODWD F1, X11              // d39580c2
	FLEQD F2, F1, X11              // d3c520a2
	FLEQS F2, F1, X11              // d3c520a0
	FLTQD F18, F9, X11             // d3d524a3
	FLTQS F18, F9, X11             // d3d524a1
	FMAXMD F21, F20, F19           // d3395a2b
	FMAXMS F21, F20, F19           // d3395a29
	FMINMD F12, F11, F10           // 53a5c52a
	FMINMS F12, F11, F10           // 53a5c528
	FROUNDD F18, F9                // d3044942
	FROUNDS F18, F9                // d3044940
	FROUNDNXD F18, F9              // d3045942
	FROUNDNXS F18, F9              // d3045940

	FLIS $(-1.0), F1               // d30010f0
	FLIS $(-Inf), F1               // d38010f0
	FLIS $(1.52587890625e-05), F1  // d30011f0
	FLIS $(3.0517578125e-05),  F1  // d38011f0
	FLIS $(0.00390625), F1         // d30012f0
	FLIS $(0.0078125), F1          // d38012f0
	FLIS $(0.0625), F1             // d30013f0
	FLIS $(0.125), F1              // d38013f0
	FLIS $(0.25), F1               // d30014f0
	FLIS $(0.3125), F1             // d38014f0
	FLIS $(0.375), F1              // d30015f0
	FLIS $(0.4375), F1             // d38015f0
	FLIS $(0.5), F1                // d30016f0
	FLIS $(0.625), F1              // d38016f0
	FLIS $(0.75), F1               // d30017f0
	FLIS $(0.875), F1              // d38017f0
	FLIS $(1.0), F1                // d30018f0
	FLIS $(1.25), F1               // d38018f0
	FLIS $(1.5), F1                // d30019f0
	FLIS $(1.75), F1               // d38019f0
	FLIS $(2.0), F1                // d3001af0
	FLIS $(2.5), F1                // d3801af0
	FLIS $(3.0), F1                // d3001bf0
	FLIS $(4.0), F1                // d3801bf0
	FLIS $(8.0), F1                // d3001cf0
	FLIS $(1.6000000000000001), F1 // d3801cf0
	FLIS $(1.28), F1               // d3001df0
	FLIS $(2.5600000000000001), F1 // d3801df0
	FLIS $(3.27), F1               // d3001ef0
	FLIS $(6.5499999999999998), F1 // d3801ef0
	FLIS $(+Inf), F1               // d3001ff0
	FLIS $(NaN), F1                // d3801ff0

	FLID $(-1.0), F1               // d30010f2
	FLID $(-Inf), F1               // d38010f2
	FLID $(1.52587890625e-05), F1  // d30011f2
	FLID $(3.0517578125e-05), F1   // d38011f2
	FLID $(0.00390625), F1         // d30012f2
	FLID $(0.0078125), F1          // d38012f2
	FLID $(0.0625), F1             // d30013f2
	FLID $(0.125), F1              // d38013f2
	FLID $(0.25), F1               // d30014f2
	FLID $(0.3125), F1             // d38014f2
	FLID $(0.375), F1              // d30015f2
	FLID $(0.4375), F1             // d38015f2
	FLID $(0.5), F1                // d30016f2
	FLID $(0.625), F1              // d38016f2
	FLID $(0.75), F1               // d30017f2
	FLID $(0.875), F1              // d38017f2
	FLID $(1.0), F1                // d30018f2
	FLID $(1.25), F1               // d38018f2
	FLID $(1.5), F1                // d30019f2
	FLID $(1.75), F1               // d38019f2
	FLID $(2.0), F1                // d3001af2
	FLID $(2.5), F1                // d3801af2
	FLID $(3.0), F1                // d3001bf2
	FLID $(4.0), F1                // d3801bf2
	FLID $(8.0), F1                // d3001cf2
	FLID $(1.6000000000000001), F1 // d3801cf2
	FLID $(1.28), F1               // d3001df2
	FLID $(2.5600000000000001), F1 // d3801df2
	FLID $(3.27), F1               // d3001ef2
	FLID $(6.5499999999999998), F1 // d3801ef2
	FLID $(+Inf), F1               // d3001ff2
	FLID $(NaN), F1                // d3801ff2
