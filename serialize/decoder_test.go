package serialize

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xelaj/errs"
)

func TestPoppingInts(t *testing.T) {
	for i, tcase := range []struct {
		input         []byte
		leftBytes     []byte
		expectedValue int32
		expectPanic   bool
	}{
		{
			[]byte{0x48, 0x0f, 0x00, 0x00},
			[]byte{},
			0xf48,
			false,
		},
		{
			[]byte{},
			[]byte{},
			0,
			true,
		},
		{
			[]byte{0xff, 0xff, 0xff, 0x7f, 0xab, 0xcd},
			[]byte{0xab, 0xcd},
			0x7fffffff,
			false,
		},
		{
			[]byte{0xff, 0xff, 0xff},
			[]byte{},
			0,
			true,
		},
	} {
		t.Run("case "+strconv.Itoa(i), func(t *testing.T) {
			defer func() {
				r := recover()
				if tcase.expectPanic && r == nil {
					assert.Fail(t, "case %v: expected panic. didn't failed", i)
				}

				if !tcase.expectPanic && r != nil {
					assert.Fail(t, "didn't expect panic. failed with:", r)
				}
			}()

			d := NewDecoder(tcase.input)
			result := d.PopInt()
			assert.Equal(t, tcase.expectedValue, result)

			assert.Equal(t, tcase.leftBytes, d.GetRestOfMessage())
		})
	}
}

func TestPoppingBools(t *testing.T) {
	for i, tcase := range []struct {
		input         []byte
		leftBytes     []byte
		expectedValue bool
		expectPanic   bool
	}{
		{
			[]byte{0x37, 0x97, 0x79, 0xbc},
			[]byte{},
			false,
			false,
		},
		{
			[]byte{0xb5, 0x75, 0x72, 0x99},
			[]byte{},
			true,
			false,
		},
		{
			[]byte{0x00, 0x00, 0x00, 0x00},
			[]byte{},
			false,
			true,
		},
		{
			[]byte{0x12, 0x34, 0x56, 0x78},
			[]byte{},
			false,
			true,
		},
		{
			[]byte{},
			[]byte{},
			false,
			true,
		},
		{
			[]byte{0xb5, 0x75, 0x72, 0x99, 0xab, 0xcd},
			[]byte{0xab, 0xcd},
			true,
			false,
		},
		{
			[]byte{0x99, 0xab, 0xcd},
			[]byte{},
			false,
			true,
		},
	} {
		t.Run("case "+strconv.Itoa(i), func(t *testing.T) {
			defer func() {
				r := recover()
				if tcase.expectPanic && r == nil {
					assert.Fail(t, "case %v: expected panic. didn't failed", i)
				}

				if !tcase.expectPanic && r != nil {
					assert.Fail(t, "didn't expect panic. failed with:", r)
				}
			}()

			d := NewDecoder(tcase.input)
			result := d.PopBool()
			assert.Equal(t, tcase.expectedValue, result)

			assert.Equal(t, tcase.leftBytes, d.GetRestOfMessage())
		})
	}
}

func TestPoppingMessages(t *testing.T) {
	for i, tcase := range []struct {
		input         []byte
		leftBytes     []byte
		expectedValue string
		expectPanic   bool
	}{
		{
			[]byte{0x00, 0x00, 0x00, 0x00},
			[]byte{},
			"",
			false,
		},
		{
			[]byte{0x03, 0x6d, 0x73, 0x67},
			[]byte{},
			"msg",
			false,
		},
		{
			[]byte{0x04, 0x68, 0x65, 0x6c, 0x6f, 0x00, 0x00, 0x00},
			[]byte{},
			"helo",
			false,
		},
		{ // оставшиеся байты должны быть нулевыми https://core.telegram.org/mtproto/serialize#base-types
			[]byte{0x04, 0x68, 0x65, 0x6c, 0x6f, 0x01, 0x00, 0x00},
			[]byte{},
			"",
			true,
		},
		{
			[]byte{
				0x04, 0x68, 0x65, 0x6c, 0x6f, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
			},
			[]byte{0x00, 0x00, 0x00, 0x00},
			"helo",
			false,
		},
		{
			[]byte{
				0x21, 0xd0, 0xbf, 0xd1, 0x80, 0xd0, 0xb8, 0xd0,
				0xb2, 0xd0, 0xb5, 0xd1, 0x82, 0x20, 0xd0, 0xbc,
				0xd0, 0xb8, 0xd1, 0x80, 0x21, 0x20, 0xf0, 0x9f,
				0x98, 0x9b, 0xf0, 0x9f, 0xa6, 0x84, 0xf0, 0x9f,
				0xa4, 0x94, 0x00, 0x00,
			},
			[]byte{},
			"привет мир! 😛🦄🤔",
			false,
		},
		{ // сообщение должно быть размером кратным 4
			[]byte{0x04, 0x68, 0x65, 0x6c, 0x6f},
			[]byte{},
			"",
			true,
		},
		{ // 0x00 тоже символ
			[]byte{0x01, 0x00, 0x00, 0x00},
			[]byte{},
			"\x00",
			false,
		},
	} {
		t.Run("case "+strconv.Itoa(i), func(t *testing.T) {
			defer func() {
				r := recover()
				if tcase.expectPanic && r == nil {
					assert.Fail(t, "case %v: expected panic. didn't failed", i)
				}

				if !tcase.expectPanic && r != nil {
					assert.Fail(t, "didn't expect panic. failed with:", r)
				}
			}()

			d := NewDecoder(tcase.input)
			result := d.PopString()
			assert.Equal(t, tcase.expectedValue, result)

			assert.Equal(t, tcase.leftBytes, d.GetRestOfMessage())
		})
	}
}

type simpleConstructor struct {
	SomeString string
	SomeInt    int32
	SomeBool   bool
	Someinner  *dummyConstructor
}

func (*simpleConstructor) CRC() uint32 {
	return 0xaaaaaaaa
}

func (t *simpleConstructor) Encode() []byte {
	panic("makes no sense")
}

type dummyConstructor struct{}

func (*dummyConstructor) CRC() uint32 {
	return 0xbbbbbbbb
}

func (t *dummyConstructor) Encode() []byte {
	panic("makes no sense")
}

type vectorConstructor struct {
	Veeveevector []bool
	Bytes        [][]byte
}

func (*vectorConstructor) CRC() uint32 {
	return 0xfedcba98
}

func (t *vectorConstructor) Encode() []byte {
	panic("makes no sense")
}

func generateDummyObjects(constructorID uint32) (obj TL, isEnum bool, err error) {
	switch constructorID {
	case 0xaaaaaaaa:
		return &simpleConstructor{}, false, nil
	case 0xbbbbbbbb:
		return &dummyConstructor{}, false, nil
	case 0xfedcba98:
		return &vectorConstructor{}, false, nil
	default:
		return nil, false, errs.NotFound("constructorID", fmt.Sprintf("%#v", constructorID))
	}
}

func TestPoppingBasicObjects(t *testing.T) {
	for i, tcase := range []struct {
		input         []byte
		leftBytes     []byte
		expectedValue interface{}
		expectPanic   bool
	}{
		{
			[]byte{
				0xaa, 0xaa, 0xaa, 0xaa, 0x0c, 0x64, 0x75, 0x6d,
				0x6d, 0x79, 0x20, 0x73, 0x74, 0x72, 0x69, 0x6e,
				0x67, 0x00, 0x00, 0x00, 0xd2, 0x04, 0x00, 0x00,
				0xb5, 0x75, 0x72, 0x99, 0xbb, 0xbb, 0xbb, 0xbb,
			},
			[]byte{},
			&simpleConstructor{
				"dummy string",
				1234,
				true,
				&dummyConstructor{},
			},
			false,
		},
		{
			[]byte{
				0x98, 0xba, 0xdc, 0xfe, 0x15, 0xc4, 0xb5, 0x1c,
				0x03, 0x00, 0x00, 0x00, 0xb5, 0x75, 0x72, 0x99,
				0x37, 0x97, 0x79, 0xbc, 0xb5, 0x75, 0x72, 0x99,
				0x15, 0xc4, 0xb5, 0x1c, 0x01, 0x00, 0x00, 0x00,
				0x03, 0x31, 0x32, 0x33,
			},
			[]byte{},
			&vectorConstructor{
				[]bool{true, false, true},
				[][]byte{{0x31, 0x32, 0x33}},
			},
			false,
		},
	} {
		t.Run("case "+strconv.Itoa(i), func(t *testing.T) {
			defer func() {
				r := recover()
				if tcase.expectPanic && r == nil {
					assert.Fail(t, "case %v: expected panic. didn't failed", i)
				}

				if !tcase.expectPanic && r != nil {
					assert.Fail(t, "didn't expect panic. failed with:", r)
				}
			}()

			d := NewDecoder(tcase.input)
			result := d.PopObj()
			assert.Equal(t, tcase.expectedValue, result)

			assert.Equal(t, tcase.leftBytes, d.GetRestOfMessage())

		})
	}
}

/*
var data = []uint8{
	0x48, 0x0f, 0x00, 0x00, 0x51, 0xb0, 0x73, 0x5f, 0x82, 0xc0, 0x73, 0x5f, 0x37, 0x97, 0x79, 0xbc,
	0x02, 0x00, 0x00, 0x00, 0x15, 0xc4, 0xb5, 0x1c, 0x13, 0x00, 0x00, 0x00, 0x0d, 0xa1, 0xb7, 0x18,
	0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x0e, 0x31, 0x34, 0x39, 0x2e, 0x31, 0x35, 0x34,
	0x2e, 0x31, 0x37, 0x35, 0x2e, 0x35, 0x39, 0x00, 0xbb, 0x01, 0x00, 0x00, 0x0d, 0xa1, 0xb7, 0x18,
	0x10, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x0e, 0x31, 0x34, 0x39, 0x2e, 0x31, 0x35, 0x34,
	0x2e, 0x31, 0x37, 0x35, 0x2e, 0x35, 0x35, 0x00, 0xbb, 0x01, 0x00, 0x00, 0x0d, 0xa1, 0xb7, 0x18,
	0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x27, 0x32, 0x30, 0x30, 0x31, 0x3a, 0x30, 0x62,
	0x32, 0x38, 0x3a, 0x66, 0x32, 0x33, 0x64, 0x3a, 0x66, 0x30, 0x30, 0x31, 0x3a, 0x30, 0x30, 0x30,
	0x30, 0x3a, 0x30, 0x30, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x30, 0x61,
	0xbb, 0x01, 0x00, 0x00, 0x0d, 0xa1, 0xb7, 0x18, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00,
	0x0e, 0x31, 0x34, 0x39, 0x2e, 0x31, 0x35, 0x34, 0x2e, 0x31, 0x36, 0x37, 0x2e, 0x35, 0x30, 0x00,
	0xbb, 0x01, 0x00, 0x00, 0x0d, 0xa1, 0xb7, 0x18, 0x10, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00,
	0x0e, 0x31, 0x34, 0x39, 0x2e, 0x31, 0x35, 0x34, 0x2e, 0x31, 0x36, 0x37, 0x2e, 0x35, 0x31, 0x00,
	0xbb, 0x01, 0x00, 0x00, 0x0d, 0xa1, 0xb7, 0x18, 0x02, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00,
	0x0f, 0x31, 0x34, 0x39, 0x2e, 0x31, 0x35, 0x34, 0x2e, 0x31, 0x36, 0x37, 0x2e, 0x31, 0x35, 0x31,
	0xbb, 0x01, 0x00, 0x00, 0x0d, 0xa1, 0xb7, 0x18, 0x01, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00,
	0x27, 0x32, 0x30, 0x30, 0x31, 0x3a, 0x30, 0x36, 0x37, 0x63, 0x3a, 0x30, 0x34, 0x65, 0x38, 0x3a,
	0x66, 0x30, 0x30, 0x32, 0x3a, 0x30, 0x30, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x30, 0x30, 0x3a, 0x30,
	0x30, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x30, 0x61, 0xbb, 0x01, 0x00, 0x00, 0x0d, 0xa1, 0xb7, 0x18,
	0x03, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x27, 0x32, 0x30, 0x30, 0x31, 0x3a, 0x30, 0x36,
	0x37, 0x63, 0x3a, 0x30, 0x34, 0x65, 0x38, 0x3a, 0x66, 0x30, 0x30, 0x32, 0x3a, 0x30, 0x30, 0x30,
	0x30, 0x3a, 0x30, 0x30, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x30, 0x62,
	0xbb, 0x01, 0x00, 0x00, 0x0d, 0xa1, 0xb7, 0x18, 0x00, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00,
	0x0f, 0x31, 0x34, 0x39, 0x2e, 0x31, 0x35, 0x34, 0x2e, 0x31, 0x37, 0x35, 0x2e, 0x31, 0x30, 0x30,
	0xbb, 0x01, 0x00, 0x00, 0x0d, 0xa1, 0xb7, 0x18, 0x10, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00,
	0x0f, 0x31, 0x34, 0x39, 0x2e, 0x31, 0x35, 0x34, 0x2e, 0x31, 0x37, 0x35, 0x2e, 0x31, 0x30, 0x30,
	0xbb, 0x01, 0x00, 0x00, 0x0d, 0xa1, 0xb7, 0x18, 0x01, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00,
	0x27, 0x32, 0x30, 0x30, 0x31, 0x3a, 0x30, 0x62, 0x32, 0x38, 0x3a, 0x66, 0x32, 0x33, 0x64, 0x3a,
	0x66, 0x30, 0x30, 0x33, 0x3a, 0x30, 0x30, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x30, 0x30, 0x3a, 0x30,
	0x30, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x30, 0x61, 0xbb, 0x01, 0x00, 0x00, 0x0d, 0xa1, 0xb7, 0x18,
	0x00, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x0e, 0x31, 0x34, 0x39, 0x2e, 0x31, 0x35, 0x34,
	0x2e, 0x31, 0x36, 0x37, 0x2e, 0x39, 0x31, 0x00, 0xbb, 0x01, 0x00, 0x00, 0x0d, 0xa1, 0xb7, 0x18,
	0x10, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x0e, 0x31, 0x34, 0x39, 0x2e, 0x31, 0x35, 0x34,
	0x2e, 0x31, 0x36, 0x37, 0x2e, 0x39, 0x31, 0x00, 0xbb, 0x01, 0x00, 0x00, 0x0d, 0xa1, 0xb7, 0x18,
	0x01, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x27, 0x32, 0x30, 0x30, 0x31, 0x3a, 0x30, 0x36,
	0x37, 0x63, 0x3a, 0x30, 0x34, 0x65, 0x38, 0x3a, 0x66, 0x30, 0x30, 0x34, 0x3a, 0x30, 0x30, 0x30,
	0x30, 0x3a, 0x30, 0x30, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x30, 0x61,
	0xbb, 0x01, 0x00, 0x00, 0x0d, 0xa1, 0xb7, 0x18, 0x02, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00,
	0x0f, 0x31, 0x34, 0x39, 0x2e, 0x31, 0x35, 0x34, 0x2e, 0x31, 0x36, 0x36, 0x2e, 0x31, 0x32, 0x30,
	0xbb, 0x01, 0x00, 0x00, 0x0d, 0xa1, 0xb7, 0x18, 0x03, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00,
	0x27, 0x32, 0x30, 0x30, 0x31, 0x3a, 0x30, 0x36, 0x37, 0x63, 0x3a, 0x30, 0x34, 0x65, 0x38, 0x3a,
	0x66, 0x30, 0x30, 0x34, 0x3a, 0x30, 0x30, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x30, 0x30, 0x3a, 0x30,
	0x30, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x30, 0x62, 0xbb, 0x01, 0x00, 0x00, 0x0d, 0xa1, 0xb7, 0x18,
	0x01, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x27, 0x32, 0x30, 0x30, 0x31, 0x3a, 0x30, 0x62,
	0x32, 0x38, 0x3a, 0x66, 0x32, 0x33, 0x66, 0x3a, 0x66, 0x30, 0x30, 0x35, 0x3a, 0x30, 0x30, 0x30,
	0x30, 0x3a, 0x30, 0x30, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x30, 0x30, 0x3a, 0x30, 0x30, 0x30, 0x61,
	0xbb, 0x01, 0x00, 0x00, 0x0d, 0xa1, 0xb7, 0x18, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00,
	0x0d, 0x39, 0x31, 0x2e, 0x31, 0x30, 0x38, 0x2e, 0x35, 0x36, 0x2e, 0x31, 0x31, 0x36, 0x00, 0x00,
	0xbb, 0x01, 0x00, 0x00, 0x0d, 0xa1, 0xb7, 0x18, 0x10, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00,
	0x0d, 0x39, 0x31, 0x2e, 0x31, 0x30, 0x38, 0x2e, 0x35, 0x36, 0x2e, 0x31, 0x31, 0x36, 0x00, 0x00,
	0xbb, 0x01, 0x00, 0x00, 0x0d, 0x61, 0x70, 0x76, 0x33, 0x2e, 0x73, 0x74, 0x65, 0x6c, 0x2e, 0x63,
	0x6f, 0x6d, 0x00, 0x00, 0xc8, 0x00, 0x00, 0x00, 0x40, 0x0d, 0x03, 0x00, 0x64, 0x00, 0x00, 0x00,
	0x50, 0x34, 0x03, 0x00, 0x88, 0x13, 0x00, 0x00, 0x30, 0x75, 0x00, 0x00, 0xe0, 0x93, 0x04, 0x00,
	0x30, 0x75, 0x00, 0x00, 0xdc, 0x05, 0x00, 0x00, 0x60, 0xea, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00,
	0xc8, 0x00, 0x00, 0x00, 0x00, 0xa3, 0x02, 0x00, 0xff, 0xff, 0xff, 0x7f, 0xff, 0xff, 0xff, 0x7f,
	0x00, 0xea, 0x24, 0x00, 0xc8, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x80, 0x3a, 0x09, 0x00,
	0x05, 0x00, 0x00, 0x00, 0x64, 0x00, 0x00, 0x00, 0x20, 0x4e, 0x00, 0x00, 0x90, 0x5f, 0x01, 0x00,
	0x30, 0x75, 0x00, 0x00, 0x10, 0x27, 0x00, 0x00, 0x0d, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f,
	0x2f, 0x74, 0x2e, 0x6d, 0x65, 0x2f, 0x00, 0x00, 0x03, 0x67, 0x69, 0x66, 0x0a, 0x66, 0x6f, 0x75,
	0x72, 0x73, 0x71, 0x75, 0x61, 0x72, 0x65, 0x00, 0x04, 0x62, 0x69, 0x6e, 0x67, 0x00, 0x00, 0x00,
	0x00, 0x04, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00,
}

var res = &telegram.Config{
	PhonecallsEnabled:       false,
	DefaultP2PContacts:      true,
	PreloadFeaturedStickers: false,
	IgnorePhoneEntities:     false,
	RevokePmInbox:           true,
	BlockedMode:             true,
	PfsEnabled:              false,
	Date:                    1601417297,
	Expires:                 1601421442,
	TestMode:                false,
	ThisDc:                  2,
	DcOptions: []*telegram.DcOption{
		&telegram.DcOption{
			Ipv6:      false,
			MediaOnly: false,
			TcpoOnly:  false,
			Cdn:       false,
			Static:    false,
			Id:        1,
			IpAddress: "149.154.175.59",
			Port:      443,
			Secret:    []uint8{},
		},
		&telegram.DcOption{
			Ipv6:      false,
			MediaOnly: false,
			TcpoOnly:  false,
			Cdn:       false,
			Static:    true,
			Id:        1,
			IpAddress: "149.154.175.55",
			Port:      443,
			Secret:    []uint8{},
		},
		&telegram.DcOption{
			Ipv6:      true,
			MediaOnly: false,
			TcpoOnly:  false,
			Cdn:       false,
			Static:    false,
			Id:        1,
			IpAddress: "2001:0b28:f23d:f001:0000:0000:0000:000a",
			Port:      443,
			Secret:    []uint8{},
		},
		&telegram.DcOption{
			Ipv6:      false,
			MediaOnly: false,
			TcpoOnly:  false,
			Cdn:       false,
			Static:    false,
			Id:        2,
			IpAddress: "149.154.167.50",
			Port:      443,
			Secret:    []uint8{},
		},
		&telegram.DcOption{
			Ipv6:      false,
			MediaOnly: false,
			TcpoOnly:  false,
			Cdn:       false,
			Static:    true,
			Id:        2,
			IpAddress: "149.154.167.51",
			Port:      443,
			Secret:    []uint8{},
		},
		&telegram.DcOption{
			Ipv6:      false,
			MediaOnly: true,
			TcpoOnly:  false,
			Cdn:       false,
			Static:    false,
			Id:        2,
			IpAddress: "149.154.167.151",
			Port:      443,
			Secret:    []uint8{},
		},
		&telegram.DcOption{
			Ipv6:      true,
			MediaOnly: false,
			TcpoOnly:  false,
			Cdn:       false,
			Static:    false,
			Id:        2,
			IpAddress: "2001:067c:04e8:f002:0000:0000:0000:000a",
			Port:      443,
			Secret:    []uint8{},
		},
		&telegram.DcOption{
			Ipv6:      true,
			MediaOnly: true,
			TcpoOnly:  false,
			Cdn:       false,
			Static:    false,
			Id:        2,
			IpAddress: "2001:067c:04e8:f002:0000:0000:0000:000b",
			Port:      443,
			Secret:    []uint8{},
		},
		&telegram.DcOption{
			Ipv6:      false,
			MediaOnly: false,
			TcpoOnly:  false,
			Cdn:       false,
			Static:    false,
			Id:        3,
			IpAddress: "149.154.175.100",
			Port:      443,
			Secret:    []uint8{},
		},
		&telegram.DcOption{
			Ipv6:      false,
			MediaOnly: false,
			TcpoOnly:  false,
			Cdn:       false,
			Static:    true,
			Id:        3,
			IpAddress: "149.154.175.100",
			Port:      443,
			Secret:    []uint8{},
		},
		&telegram.DcOption{
			Ipv6:      true,
			MediaOnly: false,
			TcpoOnly:  false,
			Cdn:       false,
			Static:    false,
			Id:        3,
			IpAddress: "2001:0b28:f23d:f003:0000:0000:0000:000a",
			Port:      443,
			Secret:    []uint8{},
		},
		&telegram.DcOption{
			Ipv6:      false,
			MediaOnly: false,
			TcpoOnly:  false,
			Cdn:       false,
			Static:    false,
			Id:        4,
			IpAddress: "149.154.167.91",
			Port:      443,
			Secret:    []uint8{},
		},
		&telegram.DcOption{
			Ipv6:      false,
			MediaOnly: false,
			TcpoOnly:  false,
			Cdn:       false,
			Static:    true,
			Id:        4,
			IpAddress: "149.154.167.91",
			Port:      443,
			Secret:    []uint8{},
		},
		&telegram.DcOption{
			Ipv6:      true,
			MediaOnly: false,
			TcpoOnly:  false,
			Cdn:       false,
			Static:    false,
			Id:        4,
			IpAddress: "2001:067c:04e8:f004:0000:0000:0000:000a",
			Port:      443,
			Secret:    []uint8{},
		},
		&telegram.DcOption{
			Ipv6:      false,
			MediaOnly: true,
			TcpoOnly:  false,
			Cdn:       false,
			Static:    false,
			Id:        4,
			IpAddress: "149.154.166.120",
			Port:      443,
			Secret:    []uint8{},
		},
		&telegram.DcOption{
			Ipv6:      true,
			MediaOnly: true,
			TcpoOnly:  false,
			Cdn:       false,
			Static:    false,
			Id:        4,
			IpAddress: "2001:067c:04e8:f004:0000:0000:0000:000b",
			Port:      443,
			Secret:    []uint8{},
		},
		&telegram.DcOption{
			Ipv6:      true,
			MediaOnly: false,
			TcpoOnly:  false,
			Cdn:       false,
			Static:    false,
			Id:        5,
			IpAddress: "2001:0b28:f23f:f005:0000:0000:0000:000a",
			Port:      443,
			Secret:    []uint8{},
		},
		&telegram.DcOption{
			Ipv6:      false,
			MediaOnly: false,
			TcpoOnly:  false,
			Cdn:       false,
			Static:    false,
			Id:        5,
			IpAddress: "91.108.56.116",
			Port:      443,
			Secret:    []uint8{},
		},
		&telegram.DcOption{
			Ipv6:      false,
			MediaOnly: false,
			TcpoOnly:  false,
			Cdn:       false,
			Static:    true,
			Id:        5,
			IpAddress: "91.108.56.116",
			Port:      443,
			Secret:    []uint8{},
		},
	},
	DcTxtDomainName:         "apv3.stel.com",
	ChatSizeMax:             200,
	MegagroupSizeMax:        200000,
	ForwardedCountMax:       100,
	OnlineUpdatePeriodMs:    210000,
	OfflineBlurTimeoutMs:    5000,
	OfflineIdleTimeoutMs:    30000,
	OnlineCloudTimeoutMs:    300000,
	NotifyCloudDelayMs:      30000,
	NotifyDefaultDelayMs:    1500,
	PushChatPeriodMs:        60000,
	PushChatLimit:           2,
	SavedGifsLimit:          200,
	EditTimeLimit:           172800,
	RevokeTimeLimit:         2147483647,
	RevokePmTimeLimit:       2147483647,
	RatingEDecay:            2419200,
	StickersRecentLimit:     200,
	StickersFavedLimit:      5,
	ChannelsReadMediaPeriod: 604800,
	TmpSessions:             0,
	PinnedDialogsCountMax:   5,
	PinnedInfolderCountMax:  100,
	CallReceiveTimeoutMs:    20000,
	CallRingTimeoutMs:       90000,
	CallConnectTimeoutMs:    30000,
	CallPacketTimeoutMs:     10000,
	MeUrlPrefix:             "https://t.me/",
	AutoupdateUrlPrefix:     "",
	GifSearchUsername:       "gif",
	VenueSearchUsername:     "foursquare",
	ImgSearchUsername:       "bing",
	StaticMapsProvider:      "",
	CaptionLengthMax:        1024,
	MessageLengthMax:        4096,
	WebfileDcId:             4,
	SuggestedLangCode:       "",
	LangPackVersion:         0,
	BaseLangPackVersion:     0,
}
*/
