/**
 * @author  zhaoliang.liang
 * @date  2024/11/27 18:47
 */

package support

import "encoding/json"

type BoolStr bool

func (b *BoolStr) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		*b = false
		return err
	}
	switch value := v.(type) {
	case bool:
		*b = BoolStr(value)
	case string:
		*b = value == "true"
	default:
		break
	}
	return nil

}

func (b *BoolStr) GetValue() bool {
	return bool(*b)
}
