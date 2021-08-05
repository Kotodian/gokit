package byteconv

import (
	"testing"
)

func Test_conv(t *testing.T) {
	// t.Log("--->", BcdToDec([]byte{0x11, 0x22}))
	// t.Logf("--->[%x]\r\n", DecToBcdBig(1343206639796224000))
	t.Logf("--->[%x]\r\n", ByteToBcd([]byte("1696919958002335744")))
	// 3A5865674B010049CAAFCDC68F5E6B1B
	// 0000000000000000000000000000000A
	// checksum, err := strconv.ParseUint("3A5865674B010049CAAFCDC68F5E6B1B", 16, 16)
	// t.Log("--->", checksum, err)
}
