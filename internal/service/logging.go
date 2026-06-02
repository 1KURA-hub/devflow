package service

import "log"

// logSideEffectErr 用于记录主链路中非关键的副作用（缓存/MQ 等）失败。
// 这些错误不影响主流程返回，但需要被记录以便排查。
// 使用方式: logSideEffectErr("hot_score", err, "post_id", postID, "user_id", userID)
func logSideEffectErr(op string, err error, kvs ...any) {
	if err == nil {
		return
	}
	if len(kvs)%2 != 0 {
		log.Printf("[devflow] side_effect_failed op=%s err=%v fields=%v", op, err, kvs)
		return
	}
	// 拼成 key=value 形式
	args := make([]any, 0, len(kvs)/2*2+2)
	format := "[devflow] side_effect_failed op=%s err=%v"
	args = append(args, op, err)
	for i := 0; i < len(kvs); i += 2 {
		format += " %v=%v"
		args = append(args, kvs[i], kvs[i+1])
	}
	log.Printf(format, args...)
}
