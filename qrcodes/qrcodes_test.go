package qrcodes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBuilder(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		b := NewBuilder(nil, nil)
		assert.NotNil(t, b)
	})
}

func Test_builder_BuildQRCode(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		b := NewBuilder(nil, nil)

		const expected = "data:image/jpeg;base64,iVBORw0KGgoAAAANSUhEUgAAAQAAAAEAEAAAAAApiSv5AAAFfElEQVR4nOzdy46jShAG4fGo3/+Ve5ZTC5dc6fwTkCK+5RGNrTkhBKYuP7+/fwT29+4voHsZAJwBwBkAnAHAGQCcAcAZAJwBwBkAnAHAGQCcAcAZAJwBwBkA3M/nQ16v/Meuw1DW81eHp+y+W+r81c86+dvd8dP/zu95BYAzADgDgDMAOAOAO3gKWE3cRU//7e6JoHP8yVNG9b/vzl9V+7fyCgBnAHAGAGcAcAYAV3wKWFXvqKvnXP+2endd/f2/ekznaSL1vmP3WTVeAeAMAM4A4AwAzgDgGk8B0zp3zqmnhok79tQIpQyvAHAGAGcAcAYAZwBwD3gKqP7GvkqNFDo5/8nonbvG/3/PKwCcAcAZAJwBwBkAXOMpIPU7dvWuuHp33XnK2J3/yt/wZz/LKwCcAcAZAJwBwBkAXPEpYPp37JM79urdeGr2bvX7TM8FyPAKAGcAcAYAZwBwBgD3un9k+k51Zc4nPKFUj7mfVwA4A4AzADgDgDMAuMZ+AZ278dRs3873SY0Uqs4XOFF9H1H9bv95BYAzADgDgDMAOAOAa+wXkL8j7c3G7azDPzFDubPSaWrPgs+8AsAZAJwBwBkAnAHADYwISo2T7xy/Sr1HmN4pbHq9o/e8AsAZAJwBwBkAnAHANWYHT+yi9YT19q8cvZPam/h7XgHgDADOAOAMAM4A4A7eBaRGsFSlfmNfTYzG6eiMqsrMv/AKAGcAcAYAZwBwBgB38C5g4jf/1cR7hNQTRGek0MT3P+G8ABUYAJwBwBkAnAHAFdcIOpmZe3KenYkdvqrfYeL9QmoUU/7dhFcAOAOAMwA4A4AzALjQ3sHTv8NX76KnnyaqrpwpXOMVAM4A4AwAzgDgDACuuEZQat2enYmx+hP7BTxtj+Dvv49XADgDgDMAOAOAMwC44b2DJ+YFnPxt6i59Z3pF0ImZ0e95BYAzADgDgDMAOAOAa4wImlj/PzUTeXocfuc8E2sZfc8rAJwBwBkAnAHAGQBcY7+AqtQql9Nr+1TPWT0mtRZQZqawVwA4A4AzADgDgDMAuOJ+ASmdUTQTexOcfLfdMR3TK4s6L0AfGACcAcAZAJwBwIVmBz9h7P2VO3+lnlzcL0A3MwA4A4AzADgDgGvMDp4Y676a/s3/5LOe8N5hd3yGVwA4A4AzADgDgDMAuOF5ARPr/HfeQZzsJjbxjmN6LsDKdwEqMAA4A4AzADgDgAvtGrbq3PGmVtrcnefknHc9KexU/x2cF6ACA4AzADgDgDMAuOLs4Ok1/1Nr70//nn/yWdVjUnwXoAIDgDMAOAOAMwC40K5hEyuIXjkCZ+KY3fc5OefsXm4rrwBwBgBnAHAGAGcAcI0RQavO2v47sytk1k2M4e9wpVAFGACcAcAZAJwBwIX2C0iNeJn+3b4zoqlz/s5nVTkiSAUGAGcAcAYAZwBwoRFBra/QeMpI3XVfOWfh5HNPzpP5P+cVAM4A4AwAzgDgDADuYETQ9Dr5qRU+Tz6res7OPghV93yuVwA4A4AzADgDgDMAuOK8gIk1dqq/dafeC3TmGlT/HVJ7GZ+cx3kBKjAAOAOAMwA4A4BrzA6e2Os2tYdv6u795Jw7d80RcF6ACgwAzgDgDADOAOBCawR1VO/wd8d3nhqq7x1Sn5VacfT79wteAeAMAM4A4AwAzgDgHvAU0NFZn3/nrlFGJ+fMz+X2CgBnAHAGAGcAcAYA13gKSN2RTt917/52ldrta6c6hv+6tYO8AsAZAJwBwBkAnAHAFZ8CnrYT1sTd9U5qh7LO7mnV4x0RpA8MAM4A4AwAzgDgHrBfgO7kFQDOAOAMAM4A4AwAzgDgDADOAOAMAM4A4AwAzgDgDADOAOAMAM4A4P4FAAD//zpsii5GqhDHAAAAAElFTkSuQmCC"

		actual := b.BuildQRCode(ctx, "username", "two-factor-secret")
		assert.Equal(t, expected, actual)
	})
}
