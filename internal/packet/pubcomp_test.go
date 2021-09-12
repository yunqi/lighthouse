package packet

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/yunqi/lighthouse/internal/xerror"
	"testing"
)

func TestNewPubcomp(t *testing.T) {
	t.Run("correct test", func(t *testing.T) {
		fixedHeader := &FixedHeader{
			PacketType:   PUBCOMP,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 2,
		}
		puback, err := NewPubcomp(fixedHeader, Version311, bytes.NewBuffer([]byte{0x0, 0x0}))
		assert.NoError(t, err)
		assert.NotNil(t, puback)
		buff := &bytes.Buffer{}
		err = puback.Encode(buff)
		assert.NoError(t, err)
		assert.NotNil(t, buff)
	})

	t.Run("RemainLength error", func(t *testing.T) {
		fixedHeader := &FixedHeader{
			PacketType:   PUBACK,
			Flags:        FixedHeaderFlagReserved,
			RemainLength: 1,
		}
		pingreq, err := NewPubcomp(fixedHeader, Version311, bytes.NewBuffer([]byte{0x0, 0x0}))
		assert.ErrorIs(t, err, xerror.ErrMalformed)
		assert.Nil(t, pingreq)
	})
}
func TestPubcomp_String(t *testing.T) {
	fixedHeader := &FixedHeader{
		PacketType:   PUBCOMP,
		Flags:        FixedHeaderFlagReserved,
		RemainLength: 2,
	}
	puback, err := NewPubcomp(fixedHeader, Version311, bytes.NewBuffer([]byte{0x0, 0x0}))
	assert.NoError(t, err)
	assert.NotNil(t, puback)
	assert.Equal(t, "Pubcomp - Version:MQTT3.1.1, PacketId:0", puback.String())
}
