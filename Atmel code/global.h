/* Global declarations that should be included by all C files before any other includes.

The type sizes assume that the compiler option "-mint8" is used.
*/

#ifndef GLOBAL_H
#define GLOBAL_H

#include <stdint.h>

typedef uint8_t bool;

#undef NULL
#define NULL 0

#define true 1
#define false 0

// Needed for delays.
#define F_CPU 8000000

#endif
