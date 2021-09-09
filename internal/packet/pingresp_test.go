package packet

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/yunqi/lighthouse/internal/xerror"
	"testing"
)

func TestNewPingresp(t *testing.T) {
	t.Run("correct test", func(t *testing.T) {
		fixedHeader := &FixedHeader{
			PacketType:   PINGRESP,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 0,
		}
		pingreq, err := NewPingresp(fixedHeader, &bytes.Buffer{})
		assert.NoError(t, err)
		assert.NotNil(t, pingreq)
	})

	t.Run("Flags error", func(t *testing.T) {
		fixedHeader := &FixedHeader{
			PacketType:   PINGRESP,
			Flags:        FixedHeaderFlagPubrel,
			RemainLength: 0,
		}
		pingreq, err := NewPingresp(fixedHeader, &bytes.Buffer{})
		assert.ErrorIs(t, err, xerror.ErrMalformed)
		assert.Nil(t, pingreq)
	})

	t.Run("RemainLength error", func(t *testing.T) {
		fixedHeader := &FixedHeader{
			PacketType:   PINGRESP,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 2,
		}
		pingreq, err := NewPingresp(fixedHeader, &bytes.Buffer{})
		assert.ErrorIs(t, err, xerror.ErrMalformed)
		assert.Nil(t, pingreq)
	})
}

func TestPingreps_Encode(t *testing.T) {

	fixedHeader := &FixedHeader{
		PacketType:   PINGRESP,
		Flags:        FixedHeaderFlagReserved,
		RemainLength: 0,
	}
	pingreq, err := NewPingresp(fixedHeader, &bytes.Buffer{})
	assert.NoError(t, err)
	assert.NotNil(t, pingreq)
	buff := &bytes.Buffer{}
	err = pingreq.Encode(buff)
	assert.NoError(t, err)
	assert.NotNil(t, buff)

}
func TestPingresp_String(t *testing.T) {

	fixedHeader := &FixedHeader{
		PacketType:   PINGRESP,
		Flags:        FixedHeaderFlagReserved,
		RemainLength: 0,
	}
	pingreq, err := NewPingresp(fixedHeader, &bytes.Buffer{})
	assert.NoError(t, err)
	assert.NotNil(t, pingreq)
	assert.Equal(t, "PINGRESP", pingreq.String())

}
