package mime

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zostay/go-addr/pkg/addr"
)

func TestHeaderDecoding(t *testing.T) {
	const emailText = "To: =?US-ASCII?Q?MIME=3A=3B?=: =?US-ASCII?Q?Winston=3A_Smith?= <winston.smith@recdep.minitrue>, =?US-ASCII?Q?Julia=3A=3B_?= <julia@ficdep.minitrue>\n\n"

	m, err := Parse([]byte(emailText))
	assert.NoError(t, err)

	str := m.HeaderGet("To")
	assert.Equal(t, "\"MIME:;\": \"Winston: Smith\" <winston.smith@recdep.minitrue>, \"Julia:; \" <julia@ficdep.minitrue>;", str)

	as, err := m.HeaderGetAddressList("To")
	assert.NoError(t, err)

	winstonEmail := addr.NewAddrSpecParsed(
		"winston.smith",
		"recdep.minitrue",
		"winston.smith@recdep.minitrue",
	)

	winston, err := addr.NewMailboxParsed(
		"Winston: Smith",
		winstonEmail,
		"",
		"\"Winston: Smith\" <winston.smith@recdep.minitrue>",
	)
	assert.NoError(t, err)

	juliaEmail := addr.NewAddrSpecParsed(
		"julia",
		"ficdep.minitrue",
		"julia@ficdep.minitrue",
	)

	julia, err := addr.NewMailboxParsed(
		"Julia:; ",
		juliaEmail,
		"",
		"\"Julia:; \" <julia@ficdep.minitrue>",
	)
	assert.NoError(t, err)

	mimegrp := addr.NewGroupParsed("MIME:;",
		addr.MailboxList{
			winston,
			julia,
		},
		"\"MIME:;\": \"Winston: Smith\" <winston.smith@recdep.minitrue>, \"Julia:; \" <julia@ficdep.minitrue>;",
	)

	assert.Equal(t, addr.AddressList{mimegrp}, as)
}
