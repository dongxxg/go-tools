package support

import (
	"fmt"
	"testing"
	"time"
)

func TestTimeFormat(t *testing.T) {
	fmt.Println(FormatDateTime(time.Now(), DATE_FORMAT_MONTH_DAY))
}
