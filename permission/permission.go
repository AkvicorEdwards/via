package permission

import (
	"github.com/AkvicorEdwards/util"
)

const (
	//	If this bit is 0, no one can read
	//	If this bit is 1, ReadPublic valid
	Read byte = iota

	//	If this bit is 0, no one can write
	//	if this bit is 1, WritePublic valid
	Write

	//	If this bit is 0, ReadProtectedPwd, ReadProtectedKey valid
	//	If this bit is 1, everyone can read
	ReadPublic

	//	If this bit is 0, WriteProtectedPwd, WriteProtectedKey valid
	//	If this bit is 1, everyone can write
	WritePublic

	//	If this bit is 0, Cannot be read by password
	//	If this bit is 1, Can be read by password
	ReadProtectedPwd

	//	If this bit is 0, Cannot be written by password
	//	If this bit is 1, Can be written by password
	WriteProtectedPwd

	//	If this bit is 0, Cannot be read by key
	//	If this bit is 1, Can be read by key
	ReadProtectedKey

	//	If this bit is 0, Cannot be written by key
	//	If this bit is 1, Can be written by key
	WriteProtectedKey

	//	If this bit is 0, Cannot be read after login
	//	If this bit is 1, Can be read after login
	ReadPrivate

	//	If this bit is 0, Cannot be written after login
	//	If this bit is 1, Can be written after login
	WritePrivate
)

func Empty() int64 {
	return 0
}

func ParseString(p string) int64 {
	if len(p) != 10 {
		return -1
	}
	var res uint64 = 0
	if p[0] == 'r' {
		util.BitSet(&res, Read, true)
	}
	if p[1] == 'w' {
		util.BitSet(&res, Write, true)
	}
	if p[2] == 'r' {
		util.BitSet(&res, ReadPublic, true)
	}
	if p[3] == 'w' {
		util.BitSet(&res, WritePublic, true)
	}
	if p[4] == 'r' {
		util.BitSet(&res, ReadProtectedPwd, true)
	}
	if p[5] == 'w' {
		util.BitSet(&res, WriteProtectedPwd, true)
	}
	if p[6] == 'r' {
		util.BitSet(&res, ReadProtectedKey, true)
	}
	if p[7] == 'w' {
		util.BitSet(&res, WriteProtectedKey, true)
	}
	if p[8] == 'r' {
		util.BitSet(&res, ReadPrivate, true)
	}
	if p[9] == 'w' {
		util.BitSet(&res, WritePrivate, true)
	}
	return int64(res)
}

func ToString(p int64) string {
	res := make([]byte, 10)
	if (p>>Read)&1 == 1 {
		res[0] = 'r'
	} else {
		res[0] = '-'
	}
	if (p>>Write)&1 == 1 {
		res[1] = 'w'
	} else {
		res[1] = '-'
	}
	if (p>>ReadPublic)&1 == 1 {
		res[2] = 'r'
	} else {
		res[2] = '-'
	}
	if (p>>WritePublic)&1 == 1 {
		res[3] = 'w'
	} else {
		res[3] = '-'
	}
	if (p>>ReadProtectedPwd)&1 == 1 {
		res[4] = 'r'
	} else {
		res[4] = '-'
	}
	if (p>>WriteProtectedPwd)&1 == 1 {
		res[5] = 'w'
	} else {
		res[5] = '-'
	}
	if (p>>ReadProtectedKey)&1 == 1 {
		res[6] = 'r'
	} else {
		res[6] = '-'
	}
	if (p>>WriteProtectedKey)&1 == 1 {
		res[7] = 'w'
	} else {
		res[7] = '-'
	}
	if (p>>ReadPrivate)&1 == 1 {
		res[8] = 'r'
	} else {
		res[8] = '-'
	}
	if (p>>WritePrivate)&1 == 1 {
		res[9] = 'w'
	} else {
		res[9] = '-'
	}
	return string(res)
}

func New(permissions ...byte) int64 {
	var res int64 = 0
	for _, v := range permissions {
		util.BitSet(&res, v, true)
	}
	return res
}

func Add(original *int64, permissions ...byte) {
	for _, v := range permissions {
		util.BitSet(original, v, true)
	}
}

func Del(original *int64, permissions ...byte) {
	for _, v := range permissions {
		util.BitSet(original, v, false)
	}
}
