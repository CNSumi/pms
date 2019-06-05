package models

type ERR_CODE uint32

const (
	ERR_CODE_OK ERR_CODE = 0

	// 服务侧问题
	ERR_CODE_READ_REQUEST_FAIL    ERR_CODE = 5001
	ERR_CODE_UNKNOWN_ERROR        ERR_CODE = 5002 //
	ERR_CODE_UPDATE_PASSWORD_FAIL ERR_CODE = 5003 // 通过校验后往db写入新密码时失败

	// 错误的请求内容、行为、格式
	ERR_CODE_BODY_DECODE_FAIL             ERR_CODE = 4001 // 请求`body`不是合法的JSON编码，在解析时发生错误
	ERR_CODE_ARGS_MISS_REQUIRED           ERR_CODE = 4002 // 缺少关键参数
	ERR_CODE_USER_PASSWORD_NOT_MATCH      ERR_CODE = 4003 // name/password`不匹配
	ERR_CODE_ARGS_SAME_PASSWORD_AT_CHANGE ERR_CODE = 4004 // 请求`changePwd`时参数中的新旧密码相同。注意此返回值不代表输入的旧密码正确，只是提前过滤了这一情况
	ERR_CODE_ARGS_CHECK_FAIL              ERR_CODE = 4005
	ERR_CODE_GET_RTSP_FAIL				  ERR_CODE = 4006
	ERR_CODE_CMD_EXEC_FAIL				  ERR_CODE = 4007
	ERR_CODE_DB_QUERY_FAIL				  ERR_CODE = 4008
)
