package ac

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

func Test_Bytetime(t *testing.T) {
	bt := NewBytetime(time.Now().Local())
	fmt.Printf("----->[%v]\r\n", bt.String())
}
func Test_BCDByte16(t *testing.T) {
	var tt BCDByte16

	orderID := "55031412782305012018061910262392"

	buf, _ := hex.DecodeString(orderID)
	fmt.Printf("------->[%x]\r\n", buf)
	copy(tt[:], buf)
	fmt.Printf("------->[%x]\r\n", tt)
	// orderID := uint64(1327396919688822785)
	// copy(tt[:], byteconv.DecToBcdBig(orderID))
	fmt.Printf("------>orderid1[%s]\r\n", tt.String())

}
