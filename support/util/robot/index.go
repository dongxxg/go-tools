/**
 * @author  zhaoliang.liang
 * @date  2024/12/16 13:39
 */

package robot

import (
	"fmt"
	mw "ginmw"
	"strings"
)

func GetRobotCloudPushJobName(env, businessId string) string {
	// UUID 36位
	// ObjectId 24位
	// jobName长度不能超过52
	// aliyun-live-d-xxxxx
	// 6 + "-" + 4 + "-" + "1" + "-" +"businessId"
	// 14 + businessId长度
	// businessId长度不能超过38
	if len(businessId) > 38 {
		businessId = businessId[:38]
	}
	jobName := fmt.Sprintf(
		"%s-live-%s-%s",
		strings.ToLower(env),
		mw.GetShortEnv(),
		strings.ReplaceAll(businessId, "_", "-"),
	)
	return jobName
}
